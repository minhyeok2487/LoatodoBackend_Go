package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type ServerTodoService struct {
	db *sql.DB
}

func NewServerTodoService(db *sql.DB) *ServerTodoService {
	return &ServerTodoService{db: db}
}

// ServerTodoItem matches Spring's ServerTodo entity response
type ServerTodoItem struct {
	TodoID          int64    `json:"todoId"`
	ContentName     string   `json:"contentName"`
	DefaultEnabled  bool     `json:"defaultEnabled"`
	VisibleWeekdays []string `json:"visibleWeekdays"`
	Frequency       *string  `json:"frequency"`
	Custom          bool     `json:"custom"`
}

// ServerTodoState matches Spring's ServerTodoStateResponse
type ServerTodoState struct {
	TodoID     int64  `json:"todoId"`
	ServerName string `json:"serverName"`
	Enabled    bool   `json:"enabled"`
	Checked    bool   `json:"checked"`
}

// ServerTodoListResponse matches Spring's response format
type ServerTodoListResponse struct {
	Todos  []ServerTodoItem  `json:"todos"`
	States []ServerTodoState `json:"states"`
}

// Legacy response for backward compatibility
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

func (s *ServerTodoService) GetServerTodos(ctx context.Context, username string) (*ServerTodoListResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	// Get distinct server names from member's characters (matching Spring's extractServerNames)
	serverRows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT server_name FROM characters
		 WHERE member_id = ? AND is_deleted = 0`, memberID)
	if err != nil {
		return nil, fmt.Errorf("querying server names: %w", err)
	}
	defer serverRows.Close()

	var serverNames []string
	for serverRows.Next() {
		var serverName string
		if err := serverRows.Scan(&serverName); err != nil {
			return nil, fmt.Errorf("scanning server name: %w", err)
		}
		serverNames = append(serverNames, serverName)
	}

	// Get all server todos (without member_id filter for global todos, or for current user's custom todos)
	rows, err := s.db.QueryContext(ctx,
		`SELECT st.server_todo_id, st.content_name, st.default_enabled,
		        COALESCE(st.frequency, ''), (st.member_id IS NOT NULL) as custom
		 FROM server_todo st
		 WHERE st.member_id IS NULL OR st.member_id = ?
		 ORDER BY st.server_todo_id ASC`, memberID)
	if err != nil {
		return nil, fmt.Errorf("querying server todos: %w", err)
	}
	defer rows.Close()

	var todos []ServerTodoItem
	var todoIDs []int64
	for rows.Next() {
		var t ServerTodoItem
		var frequencyStr string
		var defaultEnabled []byte // MySQL BIT type
		var custom []byte         // MySQL BIT type
		if err := rows.Scan(&t.TodoID, &t.ContentName, &defaultEnabled, &frequencyStr, &custom); err != nil {
			return nil, fmt.Errorf("scanning server todo: %w", err)
		}

		// Convert BIT to bool
		t.DefaultEnabled = len(defaultEnabled) > 0 && defaultEnabled[0] == 1
		t.Custom = len(custom) > 0 && custom[0] == 1

		// Handle frequency
		if frequencyStr != "" {
			t.Frequency = &frequencyStr
		}

		t.VisibleWeekdays = []string{} // Initialize empty, will be populated below
		todos = append(todos, t)
		todoIDs = append(todoIDs, t.TodoID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating server todos: %w", err)
	}

	// Get weekdays for each todo from server_todo_visible_weekday table
	if len(todoIDs) > 0 {
		weekdayRows, err := s.db.QueryContext(ctx,
			`SELECT server_todo_id, visible_weekday
			 FROM server_todo_visible_weekday
			 WHERE server_todo_id IN (`+placeholders(len(todoIDs))+`)`,
			int64SliceToInterface(todoIDs)...)
		if err != nil {
			return nil, fmt.Errorf("querying server todo weekdays: %w", err)
		}
		defer weekdayRows.Close()

		weekdayMap := make(map[int64][]string)
		for weekdayRows.Next() {
			var todoID int64
			var weekday string
			if err := weekdayRows.Scan(&todoID, &weekday); err != nil {
				return nil, fmt.Errorf("scanning weekday: %w", err)
			}
			weekdayMap[todoID] = append(weekdayMap[todoID], weekday)
		}

		// Assign weekdays to todos
		for i := range todos {
			if days, ok := weekdayMap[todos[i].TodoID]; ok {
				todos[i].VisibleWeekdays = days
			}
		}
	}

	// Get user's states - only for servers where user has characters (matching Spring's behavior)
	var states []ServerTodoState
	if len(serverNames) > 0 {
		stateRows, err := s.db.QueryContext(ctx,
			`SELECT server_todo_id, server_name, enabled+0, checked+0
			 FROM server_todo_state
			 WHERE member_id = ? AND server_name IN (`+placeholders(len(serverNames))+`)`,
			append([]interface{}{memberID}, stringSliceToInterface(serverNames)...)...)
		if err != nil {
			return nil, fmt.Errorf("querying server todo states: %w", err)
		}
		defer stateRows.Close()

		for stateRows.Next() {
			var st ServerTodoState
			var enabledInt, checkedInt int // BIT+0 returns int
			if err := stateRows.Scan(&st.TodoID, &st.ServerName, &enabledInt, &checkedInt); err != nil {
				return nil, fmt.Errorf("scanning server todo state: %w", err)
			}
			// Convert int to bool
			st.Enabled = enabledInt != 0
			st.Checked = checkedInt != 0
			states = append(states, st)
		}
	}

	if todos == nil {
		todos = []ServerTodoItem{}
	}
	if states == nil {
		states = []ServerTodoState{}
	}

	return &ServerTodoListResponse{
		Todos:  todos,
		States: states,
	}, nil
}

// placeholders generates "?,?,?" for SQL IN clause
func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	result := strings.Repeat("?,", n)
	return result[:len(result)-1] // Remove trailing comma
}

// int64SliceToInterface converts []int64 to []interface{} for SQL query
func int64SliceToInterface(ids []int64) []interface{} {
	result := make([]interface{}, len(ids))
	for i, id := range ids {
		result[i] = id
	}
	return result
}

// stringSliceToInterface converts []string to []interface{} for SQL query
func stringSliceToInterface(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}

func (s *ServerTodoService) ToggleCheck(ctx context.Context, username string, req ToggleServerTodoRequest) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
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
		 WHERE server_todo_id = ?`,
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
		`DELETE FROM server_todo WHERE server_todo_id = ?`, id)
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
