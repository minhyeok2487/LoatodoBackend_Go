package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type LogsService struct {
	db *sql.DB
}

func NewLogsService(db *sql.DB) *LogsService {
	return &LogsService{db: db}
}

type LogResponse struct {
	LogsID             int64   `json:"logsId"`
	CreatedDate        string  `json:"createdDate"`
	LocalDate          string  `json:"localDate"`
	LogType            string  `json:"logType"`
	LogContent         string  `json:"logContent"`
	Name               string  `json:"name"`
	Message            string  `json:"message"`
	Profit             float64 `json:"profit"`
	CharacterClassName string  `json:"characterClassName"`
	CharacterName      string  `json:"characterName"`
}

type LogListResponse struct {
	Content []LogResponse `json:"content"`
	HasNext bool          `json:"hasNext"`
}

type CreateLogRequest struct {
	LogContent string  `json:"logContent"`
	Profit     float64 `json:"profit"`
}

type GoldSumResponse struct {
	TotalGold float64 `json:"totalGold"`
}

func (s *LogsService) GetLogs(ctx context.Context, username string, page, size int) (*LogListResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	var totalCount int64
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM logs WHERE member_id = ? AND deleted = false`,
		memberID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("counting logs: %w", err)
	}

	offset := page * size
	rows, err := s.db.QueryContext(ctx,
		`SELECT l.logs_id, l.last_modified_date, l.local_date,
		        COALESCE(l.log_type, ''), COALESCE(l.log_content, ''),
		        COALESCE(l.name, ''), COALESCE(l.message, ''), l.profit,
		        COALESCE(c.character_class_name, ''), COALESCE(c.character_name, '')
		 FROM logs l
		 LEFT JOIN characters c ON c.characters_id = l.character_id
		 WHERE l.member_id = ? AND l.deleted = false
		 ORDER BY l.last_modified_date DESC
		 LIMIT ? OFFSET ?`,
		memberID, size, offset)
	if err != nil {
		return nil, fmt.Errorf("querying logs: %w", err)
	}
	defer rows.Close()

	var logs []LogResponse
	for rows.Next() {
		var l LogResponse
		var lastModifiedDate time.Time
		var localDate sql.NullTime
		if err := rows.Scan(&l.LogsID, &lastModifiedDate, &localDate,
			&l.LogType, &l.LogContent, &l.Name, &l.Message, &l.Profit,
			&l.CharacterClassName, &l.CharacterName); err != nil {
			return nil, fmt.Errorf("scanning log: %w", err)
		}
		// Format dates to match Spring's ISO format
		l.CreatedDate = lastModifiedDate.Format("2006-01-02T15:04:05")
		if localDate.Valid {
			l.LocalDate = localDate.Time.Format("2006-01-02")
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating logs: %w", err)
	}

	if logs == nil {
		logs = []LogResponse{}
	}

	// HasNext: true if there are more logs beyond current page
	hasNext := int64(offset+len(logs)) < totalCount

	return &LogListResponse{
		Content: logs,
		HasNext: hasNext,
	}, nil
}

func (s *LogsService) CreateLog(ctx context.Context, username string, req CreateLogRequest) (*LogResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	now := time.Now()
	localDate := now.Format("2006-01-02")
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO logs (member_id, log_content, profit, deleted, created_date, last_modified_date, local_date)
		 VALUES (?, ?, ?, false, ?, ?, ?)`,
		memberID, req.LogContent, req.Profit, now, now, localDate)
	if err != nil {
		return nil, fmt.Errorf("inserting log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}

	return &LogResponse{
		LogsID:      id,
		LogContent:  req.LogContent,
		Profit:      req.Profit,
		CreatedDate: now.Format("2006-01-02T15:04:05"),
		LocalDate:   localDate,
	}, nil
}

func (s *LogsService) DeleteLog(ctx context.Context, username string, logID int64) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE logs SET deleted = true, last_modified_date = ?
		 WHERE logs_id = ? AND member_id = ?`,
		time.Now(), logID, memberID)
	if err != nil {
		return fmt.Errorf("soft deleting log: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("log not found or not owned by user")
	}
	return nil
}

func (s *LogsService) GetGoldSum(ctx context.Context, username string) (*GoldSumResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	var totalGold float64
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(profit), 0) FROM logs WHERE member_id = ? AND deleted = false`,
		memberID).Scan(&totalGold)
	if err != nil {
		return nil, fmt.Errorf("summing gold: %w", err)
	}

	return &GoldSumResponse{TotalGold: totalGold}, nil
}

func (s *LogsService) GetLogsProfit(ctx context.Context, username string) ([]map[string]interface{}, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("querying member: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT DATE(created_date) as log_date, SUM(profit) as total_gold
		 FROM logs WHERE member_id = ? AND created_date >= DATE_SUB(NOW(), INTERVAL 14 DAY)
		 GROUP BY DATE(created_date) ORDER BY log_date DESC`, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying logs profit: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var logDate string
		var totalGold float64
		if err := rows.Scan(&logDate, &totalGold); err != nil {
			return nil, fmt.Errorf("scanning: %w", err)
		}
		results = append(results, map[string]interface{}{"date": logDate, "gold": totalGold})
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	return results, rows.Err()
}
