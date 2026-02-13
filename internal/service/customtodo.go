package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type CustomTodoService struct {
	db *sql.DB
}

func NewCustomTodoService(db *sql.DB) *CustomTodoService {
	return &CustomTodoService{db: db}
}

type CustomTodoResponse struct {
	ID          int64  `json:"id"`
	CharacterID int64  `json:"characterId"`
	ContentName string `json:"contentName"`
	Frequency   string `json:"frequency"`
	Checked     bool   `json:"checked"`
}

type CreateCustomTodoRequest struct {
	CharacterID int64  `json:"characterId"`
	ContentName string `json:"contentName"`
	Frequency   string `json:"frequency"`
}

type UpdateCustomTodoRequest struct {
	ContentName string `json:"contentName"`
	Frequency   string `json:"frequency"`
}

func (s *CustomTodoService) GetCustomTodos(ctx context.Context, username string, characterID int64) ([]CustomTodoResponse, error) {
	// Verify the character belongs to the user
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM character_entity c
			JOIN member m ON c.member_id = m.id
			WHERE c.id = ? AND m.username = ?
		)`, characterID, username).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("verifying character ownership: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("character not found or not owned by user")
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, character_id, content_name, frequency, is_checked
		 FROM custom_todo
		 WHERE character_id = ?`,
		characterID)
	if err != nil {
		return nil, fmt.Errorf("querying custom todos: %w", err)
	}
	defer rows.Close()

	var todos []CustomTodoResponse
	for rows.Next() {
		var t CustomTodoResponse
		if err := rows.Scan(&t.ID, &t.CharacterID, &t.ContentName, &t.Frequency, &t.Checked); err != nil {
			return nil, fmt.Errorf("scanning custom todo: %w", err)
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating custom todos: %w", err)
	}

	if todos == nil {
		todos = []CustomTodoResponse{}
	}
	return todos, nil
}

func (s *CustomTodoService) CreateCustomTodo(ctx context.Context, username string, req CreateCustomTodoRequest) (*CustomTodoResponse, error) {
	// Verify the character belongs to the user
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM character_entity c
			JOIN member m ON c.member_id = m.id
			WHERE c.id = ? AND m.username = ?
		)`, req.CharacterID, username).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("verifying character ownership: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("character not found or not owned by user")
	}

	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO custom_todo (character_id, content_name, is_checked, frequency, created_date, last_modified_date)
		 VALUES (?, ?, false, ?, ?, ?)`,
		req.CharacterID, req.ContentName, req.Frequency, now, now)
	if err != nil {
		return nil, fmt.Errorf("inserting custom todo: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}

	return &CustomTodoResponse{
		ID:          id,
		CharacterID: req.CharacterID,
		ContentName: req.ContentName,
		Frequency:   req.Frequency,
		Checked:     false,
	}, nil
}

func (s *CustomTodoService) ToggleCheck(ctx context.Context, username string, id int64) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE custom_todo SET is_checked = NOT is_checked, last_modified_date = ?
		 WHERE id = ?`,
		time.Now(), id)
	if err != nil {
		return fmt.Errorf("toggling custom todo check: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("custom todo not found")
	}
	return nil
}

func (s *CustomTodoService) UpdateCustomTodo(ctx context.Context, username string, id int64, req UpdateCustomTodoRequest) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE custom_todo SET content_name = ?, frequency = ?, last_modified_date = ?
		 WHERE id = ?`,
		req.ContentName, req.Frequency, time.Now(), id)
	if err != nil {
		return fmt.Errorf("updating custom todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("custom todo not found")
	}
	return nil
}

func (s *CustomTodoService) DeleteCustomTodo(ctx context.Context, username string, id int64) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM custom_todo WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting custom todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("custom todo not found")
	}
	return nil
}
