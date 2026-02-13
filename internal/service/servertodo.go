package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ServerTodoService struct {
	db *sql.DB
}

func NewServerTodoService(db *sql.DB) *ServerTodoService {
	return &ServerTodoService{db: db}
}

type ServerTodoResponse struct {
	ServerTodoID int64  `json:"serverTodoId"`
	Content      string `json:"content"`
	SortNumber   int    `json:"sortNumber"`
	Checked      bool   `json:"checked"`
}

type ToggleServerTodoRequest struct {
	ServerTodoID int64 `json:"serverTodoId"`
}

type CreateServerTodoRequest struct {
	Content    string `json:"content"`
	SortNumber int    `json:"sortNumber"`
}

type UpdateServerTodoRequest struct {
	Content    string `json:"content"`
	SortNumber int    `json:"sortNumber"`
}

func (s *ServerTodoService) GetServerTodos(ctx context.Context, username string) ([]ServerTodoResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT st.id, st.content_name, COALESCE(sts.checked, false)
		 FROM server_todo st
		 LEFT JOIN server_todo_state sts ON st.id = sts.server_todo_id AND sts.member_id = ?
		 ORDER BY st.id ASC`,
		memberID)
	if err != nil {
		return nil, fmt.Errorf("querying server todos: %w", err)
	}
	defer rows.Close()

	var todos []ServerTodoResponse
	sortNum := 0
	for rows.Next() {
		var t ServerTodoResponse
		if err := rows.Scan(&t.ServerTodoID, &t.Content, &t.Checked); err != nil {
			return nil, fmt.Errorf("scanning server todo: %w", err)
		}
		t.SortNumber = sortNum
		sortNum++
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating server todos: %w", err)
	}

	if todos == nil {
		todos = []ServerTodoResponse{}
	}
	return todos, nil
}

func (s *ServerTodoService) ToggleCheck(ctx context.Context, username string, req ToggleServerTodoRequest) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO server_todo_state (server_todo_id, member_id, server_name, enabled, checked, created_date, last_modified_date)
		 VALUES (?, ?, '', true, true, ?, ?)
		 ON DUPLICATE KEY UPDATE checked = NOT checked, last_modified_date = ?`,
		req.ServerTodoID, memberID, now, now, now)
	if err != nil {
		return fmt.Errorf("toggling server todo check: %w", err)
	}
	return nil
}

func (s *ServerTodoService) CreateServerTodo(ctx context.Context, req CreateServerTodoRequest) (*ServerTodoResponse, error) {
	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO server_todo (content_name, default_enabled, created_date, last_modified_date)
		 VALUES (?, true, ?, ?)`,
		req.Content, now, now)
	if err != nil {
		return nil, fmt.Errorf("inserting server todo: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}

	return &ServerTodoResponse{
		ServerTodoID: id,
		Content:      req.Content,
		SortNumber:   req.SortNumber,
		Checked:      false,
	}, nil
}

func (s *ServerTodoService) UpdateServerTodo(ctx context.Context, id int64, req UpdateServerTodoRequest) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE server_todo SET content_name = ?, last_modified_date = ?
		 WHERE id = ?`,
		req.Content, time.Now(), id)
	if err != nil {
		return fmt.Errorf("updating server todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("server todo not found")
	}
	return nil
}

func (s *ServerTodoService) DeleteServerTodo(ctx context.Context, id int64) error {
	// Delete associated states first
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM server_todo_state WHERE server_todo_id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting server todo states: %w", err)
	}

	result, err := s.db.ExecContext(ctx,
		`DELETE FROM server_todo WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting server todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("server todo not found")
	}
	return nil
}
