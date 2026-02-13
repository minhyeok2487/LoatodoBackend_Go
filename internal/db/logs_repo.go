package db

import (
	"context"
	"database/sql"
	"time"
)

type LogsRepository struct {
	db *sql.DB
}

func NewLogsRepository(db *sql.DB) *LogsRepository {
	return &LogsRepository{db: db}
}

const logsColumns = `
	logs_id, local_date, member_id, character_id, log_type, log_content,
	name, message, profit, deleted, created_date, last_modified_date
`

func scanLogs(scanner interface{ Scan(dest ...interface{}) error }) (*Logs, error) {
	l := &Logs{}
	err := scanner.Scan(
		&l.ID, &l.LocalDate, &l.MemberID, &l.CharacterID, &l.LogType, &l.LogContent,
		&l.Name, &l.Message, &l.Profit, &l.Deleted, &l.CreatedDate, &l.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *LogsRepository) FindByID(ctx context.Context, id int64) (*Logs, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+logsColumns+` FROM logs WHERE logs_id = ?`, id,
	)
	return scanLogs(row)
}

func (r *LogsRepository) FindByMemberID(ctx context.Context, memberID int64, offset, limit int) ([]Logs, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+logsColumns+` FROM logs
		 WHERE member_id = ? AND deleted = false
		 ORDER BY last_modified_date DESC
		 LIMIT ? OFFSET ?`, memberID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Logs
	for rows.Next() {
		l, err := scanLogs(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *l)
	}
	return logs, rows.Err()
}

func (r *LogsRepository) FindByMemberIDAndCharacterID(ctx context.Context, memberID, characterID int64, offset, limit int) ([]Logs, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+logsColumns+` FROM logs
		 WHERE member_id = ? AND character_id = ? AND deleted = false
		 ORDER BY last_modified_date DESC
		 LIMIT ? OFFSET ?`, memberID, characterID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Logs
	for rows.Next() {
		l, err := scanLogs(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *l)
	}
	return logs, rows.Err()
}

func (r *LogsRepository) FindByMemberIDAndLogContent(ctx context.Context, memberID int64, logContent string, offset, limit int) ([]Logs, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+logsColumns+` FROM logs
		 WHERE member_id = ? AND log_content = ? AND deleted = false
		 ORDER BY last_modified_date DESC
		 LIMIT ? OFFSET ?`, memberID, logContent, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Logs
	for rows.Next() {
		l, err := scanLogs(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *l)
	}
	return logs, rows.Err()
}

func (r *LogsRepository) Create(ctx context.Context, l *Logs) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO logs (local_date, member_id, character_id, log_type, log_content,
		                   name, message, profit, deleted, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.LocalDate, l.MemberID, l.CharacterID, l.LogType, l.LogContent,
		l.Name, l.Message, l.Profit, l.Deleted, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *LogsRepository) Update(ctx context.Context, l *Logs) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE logs SET local_date = ?, message = ?, profit = ?, deleted = ?, last_modified_date = ?
		 WHERE logs_id = ?`,
		l.LocalDate, l.Message, l.Profit, l.Deleted, time.Now(), l.ID,
	)
	return err
}

func (r *LogsRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE logs SET deleted = true, last_modified_date = ? WHERE logs_id = ?`,
		time.Now(), id,
	)
	return err
}

func (r *LogsRepository) HardDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM logs WHERE logs_id = ?`, id)
	return err
}

func (r *LogsRepository) FindByMemberIDAndDateAndName(ctx context.Context, memberID, characterID int64, logContent, localDate, name string) (*Logs, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+logsColumns+` FROM logs
		 WHERE member_id = ? AND character_id = ? AND log_content = ? AND local_date = ? AND name = ?`,
		memberID, characterID, logContent, localDate, name,
	)
	return scanLogs(row)
}

func (r *LogsRepository) CountByMemberID(ctx context.Context, memberID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM logs WHERE member_id = ? AND deleted = false`, memberID,
	).Scan(&count)
	return count, err
}

func (r *LogsRepository) SumProfitByMemberIDAndCharacterID(ctx context.Context, memberID, characterID int64, logContent string) (float64, error) {
	var total sql.NullFloat64
	err := r.db.QueryRowContext(ctx,
		`SELECT SUM(profit) FROM logs
		 WHERE member_id = ? AND character_id = ? AND log_content = ? AND deleted = false`,
		memberID, characterID, logContent,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}
