package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ShedLock provides distributed locking using the existing shedlock MySQL table.
// Compatible with the Java ShedLock library's table schema.
type ShedLock struct {
	db       *sql.DB
	lockedBy string
}

// NewShedLock creates a new ShedLock instance.
func NewShedLock(db *sql.DB) *ShedLock {
	return &ShedLock{
		db:       db,
		lockedBy: "go-server",
	}
}

// TryLock attempts to acquire a named lock. Returns true if the lock was successfully acquired.
// lockAtMost is the maximum duration the lock should be held.
// lockAtLeast is the minimum duration the lock should be held (prevents immediate re-acquisition).
func (s *ShedLock) TryLock(ctx context.Context, name string, lockAtMost, lockAtLeast time.Duration) (bool, error) {
	now := time.Now()
	lockUntil := now.Add(lockAtMost)

	// Use INSERT ... ON DUPLICATE KEY UPDATE to atomically acquire the lock.
	// The lock is only acquired if the current lock_until has passed (expired lock).
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO shedlock (name, lock_until, locked_at, locked_by)
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		 lock_until = IF(lock_until <= NOW(), VALUES(lock_until), lock_until),
		 locked_at = IF(lock_until <= NOW(), VALUES(locked_at), locked_at),
		 locked_by = IF(lock_until <= NOW(), VALUES(locked_by), locked_by)`,
		name, lockUntil, now, s.lockedBy,
	)
	if err != nil {
		return false, fmt.Errorf("acquiring lock %q: %w", name, err)
	}

	// For INSERT: RowsAffected=1, for UPDATE that changed: RowsAffected=2,
	// for UPDATE that didn't change (lock not expired): RowsAffected=0
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("checking lock result %q: %w", name, err)
	}

	// If no rows were affected, the lock was not acquired (still held by someone else)
	if rows == 0 {
		return false, nil
	}

	// Verify we actually got the lock by checking locked_by and locked_at
	var lockedBy string
	var lockedAt time.Time
	err = s.db.QueryRowContext(ctx,
		`SELECT locked_by, locked_at FROM shedlock WHERE name = ?`, name,
	).Scan(&lockedBy, &lockedAt)
	if err != nil {
		return false, fmt.Errorf("verifying lock %q: %w", name, err)
	}

	// Check that we are the lock holder and the timestamp is approximately ours
	acquired := lockedBy == s.lockedBy && now.Sub(lockedAt).Abs() < 2*time.Second
	return acquired, nil
}

// Unlock releases the lock by setting lock_until to now + lockAtLeast.
// This ensures the lock is held for at least the lockAtLeast duration.
func (s *ShedLock) Unlock(ctx context.Context, name string, lockAtLeast time.Duration) error {
	lockUntil := time.Now().Add(lockAtLeast)

	_, err := s.db.ExecContext(ctx,
		`UPDATE shedlock SET lock_until = ? WHERE name = ? AND locked_by = ?`,
		lockUntil, name, s.lockedBy,
	)
	if err != nil {
		return fmt.Errorf("releasing lock %q: %w", name, err)
	}

	return nil
}

// RunWithLock attempts to acquire the named lock and runs the given function if successful.
// After the function completes, the lock is released with the lockAtLeast duration.
func (s *ShedLock) RunWithLock(ctx context.Context, name string, lockAtMost, lockAtLeast time.Duration, fn func(ctx context.Context) error) error {
	acquired, err := s.TryLock(ctx, name, lockAtMost, lockAtLeast)
	if err != nil {
		return fmt.Errorf("trying lock %q: %w", name, err)
	}
	if !acquired {
		return nil // Another instance holds the lock
	}

	defer func() {
		_ = s.Unlock(ctx, name, lockAtLeast)
	}()

	return fn(ctx)
}
