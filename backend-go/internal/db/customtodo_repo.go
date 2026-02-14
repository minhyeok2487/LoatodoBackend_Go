package db

import (
	"context"
	"database/sql"
	"time"
)

type CustomTodoRepository struct {
	db *sql.DB
}

func NewCustomTodoRepository(db *sql.DB) *CustomTodoRepository {
	return &CustomTodoRepository{db: db}
}

const customTodoColumns = `
	custom_todo_id, content_name, is_checked, frequency, character_id,
	created_date, last_modified_date
`

func scanCustomTodo(scanner interface{ Scan(dest ...interface{}) error }) (*CustomTodo, error) {
	ct := &CustomTodo{}
	err := scanner.Scan(
		&ct.ID, &ct.ContentName, &ct.IsChecked, &ct.Frequency, &ct.CharacterID,
		&ct.CreatedDate, &ct.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return ct, nil
}

func (r *CustomTodoRepository) FindByID(ctx context.Context, id int64) (*CustomTodo, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+customTodoColumns+` FROM custom_todo WHERE custom_todo_id = ?`, id,
	)
	return scanCustomTodo(row)
}

func (r *CustomTodoRepository) FindByCharacterID(ctx context.Context, characterID int64) ([]CustomTodo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+customTodoColumns+` FROM custom_todo WHERE character_id = ? ORDER BY custom_todo_id ASC`,
		characterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []CustomTodo
	for rows.Next() {
		ct, err := scanCustomTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, *ct)
	}
	return todos, rows.Err()
}

func (r *CustomTodoRepository) FindByCharacterIDAndFrequency(ctx context.Context, characterID int64, frequency string) ([]CustomTodo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+customTodoColumns+` FROM custom_todo
		 WHERE character_id = ? AND frequency = ?
		 ORDER BY custom_todo_id ASC`, characterID, frequency,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []CustomTodo
	for rows.Next() {
		ct, err := scanCustomTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, *ct)
	}
	return todos, rows.Err()
}

func (r *CustomTodoRepository) Create(ctx context.Context, ct *CustomTodo) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO custom_todo (content_name, is_checked, frequency, character_id,
		                          created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		ct.ContentName, ct.IsChecked, ct.Frequency, ct.CharacterID,
		now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CustomTodoRepository) UpdateCheck(ctx context.Context, id int64, isChecked bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE custom_todo SET is_checked = ?, last_modified_date = ? WHERE custom_todo_id = ?`,
		isChecked, time.Now(), id,
	)
	return err
}

func (r *CustomTodoRepository) UpdateContentName(ctx context.Context, id int64, contentName string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE custom_todo SET content_name = ?, last_modified_date = ? WHERE custom_todo_id = ?`,
		contentName, time.Now(), id,
	)
	return err
}

func (r *CustomTodoRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM custom_todo WHERE custom_todo_id = ?`, id)
	return err
}

func (r *CustomTodoRepository) DeleteByCharacterID(ctx context.Context, characterID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM custom_todo WHERE character_id = ?`, characterID)
	return err
}

func (r *CustomTodoRepository) ResetDailyByCharacterIDs(ctx context.Context, characterIDs []int64) error {
	if len(characterIDs) == 0 {
		return nil
	}

	query := `UPDATE custom_todo SET is_checked = false, last_modified_date = ?
	          WHERE frequency = 'DAILY' AND character_id IN (`
	args := []interface{}{time.Now()}
	for i, id := range characterIDs {
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

func (r *CustomTodoRepository) ResetWeeklyByCharacterIDs(ctx context.Context, characterIDs []int64) error {
	if len(characterIDs) == 0 {
		return nil
	}

	query := `UPDATE custom_todo SET is_checked = false, last_modified_date = ?
	          WHERE frequency = 'WEEKLY' AND character_id IN (`
	args := []interface{}{time.Now()}
	for i, id := range characterIDs {
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

func (r *CustomTodoRepository) ResetAllDaily(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE custom_todo SET is_checked = false, last_modified_date = ? WHERE frequency = 'DAILY'`,
		time.Now(),
	)
	return err
}

func (r *CustomTodoRepository) ResetAllWeekly(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE custom_todo SET is_checked = false, last_modified_date = ? WHERE frequency = 'WEEKLY'`,
		time.Now(),
	)
	return err
}
