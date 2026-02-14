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
	ID          int64  `json:"customTodoId"`
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
	var rows *sql.Rows
	var err error

	if characterID > 0 {
		// Verify the character belongs to the user
		var exists bool
		err = s.db.QueryRowContext(ctx,
			`SELECT EXISTS(
				SELECT 1 FROM characters c
				JOIN member m ON c.member_id = m.member_id
				WHERE c.characters_id = ? AND m.username = ?
			)`, characterID, username).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("verifying character ownership: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("character not found or not owned by user")
		}

		rows, err = s.db.QueryContext(ctx,
			`SELECT custom_todo_id, character_id, content_name, frequency, is_checked+0
			 FROM custom_todo
			 WHERE character_id = ?`,
			characterID)
	} else {
		// Return all custom todos for the user's characters
		rows, err = s.db.QueryContext(ctx,
			`SELECT ct.custom_todo_id, ct.character_id, ct.content_name, ct.frequency, ct.is_checked+0
			 FROM custom_todo ct
			 JOIN characters c ON ct.character_id = c.characters_id
			 JOIN member m ON c.member_id = m.member_id
			 WHERE m.username = ?`,
			username)
	}
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
			SELECT 1 FROM characters c
			JOIN member m ON c.member_id = m.member_id
			WHERE c.characters_id = ? AND m.username = ?
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
		 WHERE custom_todo_id = ?`,
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
		 WHERE custom_todo_id = ?`,
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
		`DELETE FROM custom_todo WHERE custom_todo_id = ?`, id)
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
