package db

import (
	"context"
	"database/sql"
	"time"
)

// --- ServerTodoRepository ---

type ServerTodoRepository struct {
	db *sql.DB
}

func NewServerTodoRepository(db *sql.DB) *ServerTodoRepository {
	return &ServerTodoRepository{db: db}
}

const serverTodoColumns = `
	server_todo_id, content_name, default_enabled, member_id, frequency,
	created_date, last_modified_date
`

func scanServerTodo(scanner interface{ Scan(dest ...interface{}) error }) (*ServerTodo, error) {
	s := &ServerTodo{}
	err := scanner.Scan(
		&s.ID, &s.ContentName, &s.DefaultEnabled, &s.MemberID, &s.Frequency,
		&s.CreatedDate, &s.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *ServerTodoRepository) FindByID(ctx context.Context, id int64) (*ServerTodo, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+serverTodoColumns+` FROM server_todo WHERE server_todo_id = ?`, id,
	)
	return scanServerTodo(row)
}

func (r *ServerTodoRepository) FindAll(ctx context.Context) ([]ServerTodo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+serverTodoColumns+` FROM server_todo WHERE member_id IS NULL ORDER BY server_todo_id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []ServerTodo
	for rows.Next() {
		s, err := scanServerTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, *s)
	}
	return todos, rows.Err()
}

func (r *ServerTodoRepository) FindByMemberID(ctx context.Context, memberID int64) ([]ServerTodo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+serverTodoColumns+` FROM server_todo
		 WHERE member_id IS NULL OR member_id = ?
		 ORDER BY server_todo_id ASC`, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []ServerTodo
	for rows.Next() {
		s, err := scanServerTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, *s)
	}
	return todos, rows.Err()
}

func (r *ServerTodoRepository) Create(ctx context.Context, s *ServerTodo) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO server_todo (content_name, default_enabled, member_id, frequency,
		                          created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		s.ContentName, s.DefaultEnabled, s.MemberID, s.Frequency,
		now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ServerTodoRepository) Update(ctx context.Context, id int64, contentName string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE server_todo SET content_name = ?, last_modified_date = ? WHERE server_todo_id = ?`,
		contentName, time.Now(), id,
	)
	return err
}

func (r *ServerTodoRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM server_todo WHERE server_todo_id = ?`, id)
	return err
}

// --- ServerTodoVisibleWeekdayRepository ---

type ServerTodoVisibleWeekdayRepository struct {
	db *sql.DB
}

func NewServerTodoVisibleWeekdayRepository(db *sql.DB) *ServerTodoVisibleWeekdayRepository {
	return &ServerTodoVisibleWeekdayRepository{db: db}
}

func (r *ServerTodoVisibleWeekdayRepository) FindByServerTodoID(ctx context.Context, serverTodoID int64) ([]ServerTodoVisibleWeekday, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT server_todo_id, visible_weekday FROM server_todo_visible_weekday WHERE server_todo_id = ?`,
		serverTodoID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weekdays []ServerTodoVisibleWeekday
	for rows.Next() {
		var w ServerTodoVisibleWeekday
		if err := rows.Scan(&w.ServerTodoID, &w.VisibleWeekday); err != nil {
			return nil, err
		}
		weekdays = append(weekdays, w)
	}
	return weekdays, rows.Err()
}

func (r *ServerTodoVisibleWeekdayRepository) Create(ctx context.Context, serverTodoID int64, weekday string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO server_todo_visible_weekday (server_todo_id, visible_weekday) VALUES (?, ?)`,
		serverTodoID, weekday,
	)
	return err
}

func (r *ServerTodoVisibleWeekdayRepository) DeleteByServerTodoID(ctx context.Context, serverTodoID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM server_todo_visible_weekday WHERE server_todo_id = ?`, serverTodoID,
	)
	return err
}

// --- ServerTodoStateRepository ---

type ServerTodoStateRepository struct {
	db *sql.DB
}

func NewServerTodoStateRepository(db *sql.DB) *ServerTodoStateRepository {
	return &ServerTodoStateRepository{db: db}
}

const serverTodoStateColumns = `
	server_todo_state_id, server_todo_id, member_id, server_name, enabled, checked,
	created_date, last_modified_date
`

func scanServerTodoState(scanner interface{ Scan(dest ...interface{}) error }) (*ServerTodoState, error) {
	s := &ServerTodoState{}
	err := scanner.Scan(
		&s.ID, &s.ServerTodoID, &s.MemberID, &s.ServerName, &s.Enabled, &s.Checked,
		&s.CreatedDate, &s.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *ServerTodoStateRepository) FindByID(ctx context.Context, id int64) (*ServerTodoState, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+serverTodoStateColumns+` FROM server_todo_state WHERE server_todo_state_id = ?`, id,
	)
	return scanServerTodoState(row)
}

func (r *ServerTodoStateRepository) FindByMemberIDAndServerName(ctx context.Context, memberID int64, serverName string) ([]ServerTodoState, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+serverTodoStateColumns+` FROM server_todo_state
		 WHERE member_id = ? AND server_name = ?
		 ORDER BY server_todo_state_id ASC`, memberID, serverName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []ServerTodoState
	for rows.Next() {
		s, err := scanServerTodoState(rows)
		if err != nil {
			return nil, err
		}
		states = append(states, *s)
	}
	return states, rows.Err()
}

func (r *ServerTodoStateRepository) FindByMemberID(ctx context.Context, memberID int64) ([]ServerTodoState, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+serverTodoStateColumns+` FROM server_todo_state
		 WHERE member_id = ?
		 ORDER BY server_todo_state_id ASC`, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []ServerTodoState
	for rows.Next() {
		s, err := scanServerTodoState(rows)
		if err != nil {
			return nil, err
		}
		states = append(states, *s)
	}
	return states, rows.Err()
}

func (r *ServerTodoStateRepository) FindByServerTodoIDAndMemberID(ctx context.Context, serverTodoID, memberID int64, serverName string) (*ServerTodoState, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+serverTodoStateColumns+` FROM server_todo_state
		 WHERE server_todo_id = ? AND member_id = ? AND server_name = ?`,
		serverTodoID, memberID, serverName,
	)
	return scanServerTodoState(row)
}

func (r *ServerTodoStateRepository) Create(ctx context.Context, s *ServerTodoState) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO server_todo_state (server_todo_id, member_id, server_name, enabled, checked,
		                                created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		s.ServerTodoID, s.MemberID, s.ServerName, s.Enabled, s.Checked,
		now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ServerTodoStateRepository) UpdateEnabled(ctx context.Context, id int64, enabled bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE server_todo_state SET enabled = ?, last_modified_date = ? WHERE server_todo_state_id = ?`,
		enabled, time.Now(), id,
	)
	return err
}

func (r *ServerTodoStateRepository) UpdateChecked(ctx context.Context, id int64, checked bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE server_todo_state SET checked = ?, last_modified_date = ? WHERE server_todo_state_id = ?`,
		checked, time.Now(), id,
	)
	return err
}

func (r *ServerTodoStateRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM server_todo_state WHERE server_todo_state_id = ?`, id)
	return err
}

func (r *ServerTodoStateRepository) DeleteByServerTodoID(ctx context.Context, serverTodoID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM server_todo_state WHERE server_todo_id = ?`, serverTodoID)
	return err
}

func (r *ServerTodoStateRepository) ResetAllChecked(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE server_todo_state SET checked = false, last_modified_date = ?`, time.Now(),
	)
	return err
}

func (r *ServerTodoStateRepository) ResetCheckedByServerTodoIDs(ctx context.Context, serverTodoIDs []int64) error {
	if len(serverTodoIDs) == 0 {
		return nil
	}

	query := `UPDATE server_todo_state SET checked = false, last_modified_date = ? WHERE server_todo_id IN (`
	args := []interface{}{time.Now()}
	for i, id := range serverTodoIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += `)`

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}
