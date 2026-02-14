package db

import (
	"context"
	"database/sql"
	"time"
)

type ShedLockRepository struct {
	db *sql.DB
}

func NewShedLockRepository(db *sql.DB) *ShedLockRepository {
	return &ShedLockRepository{db: db}
}

// TryLock attempts to acquire a distributed lock. Returns true if the lock was acquired.
func (r *ShedLockRepository) TryLock(ctx context.Context, name string, lockUntil time.Time, lockedBy string) (bool, error) {
	now := time.Now()

	// Try to insert the lock. If it already exists, try to update it only if expired.
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO shedlock (name, lock_until, locked_at, locked_by)
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
			lock_until = IF(lock_until <= ?, VALUES(lock_until), lock_until),
			locked_at = IF(lock_until <= ?, VALUES(locked_at), locked_at),
			locked_by = IF(lock_until <= ?, VALUES(locked_by), locked_by)`,
		name, lockUntil, now, lockedBy,
		now, now, now,
	)
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	// For INSERT: affected = 1; for UPDATE with change: affected = 2; for no change: affected = 0
	return affected > 0, nil
}

// Unlock releases a lock by setting lock_until to now.
func (r *ShedLockRepository) Unlock(ctx context.Context, name string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE shedlock SET lock_until = ? WHERE name = ?`,
		time.Now(), name,
	)
	return err
}

// ExtendLock extends an existing lock's duration.
func (r *ShedLockRepository) ExtendLock(ctx context.Context, name string, lockUntil time.Time, lockedBy string) (bool, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE shedlock SET lock_until = ?, locked_at = ?
		 WHERE name = ? AND locked_by = ? AND lock_until > ?`,
		lockUntil, now, name, lockedBy, now,
	)
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

// FindByName retrieves a lock entry by name.
func (r *ShedLockRepository) FindByName(ctx context.Context, name string) (*ShedLock, error) {
	sl := &ShedLock{}
	err := r.db.QueryRowContext(ctx,
		`SELECT name, lock_until, locked_at, locked_by FROM shedlock WHERE name = ?`, name,
	).Scan(&sl.Name, &sl.LockUntil, &sl.LockedAt, &sl.LockedBy)
	if err != nil {
		return nil, err
	}
	return sl, nil
}

// FindAll retrieves all lock entries.
func (r *ShedLockRepository) FindAll(ctx context.Context) ([]ShedLock, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT name, lock_until, locked_at, locked_by FROM shedlock ORDER BY name ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locks []ShedLock
	for rows.Next() {
		var sl ShedLock
		if err := rows.Scan(&sl.Name, &sl.LockUntil, &sl.LockedAt, &sl.LockedBy); err != nil {
			return nil, err
		}
		locks = append(locks, sl)
	}
	return locks, rows.Err()
}

// Delete removes a lock entry.
func (r *ShedLockRepository) Delete(ctx context.Context, name string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM shedlock WHERE name = ?`, name)
	return err
}
