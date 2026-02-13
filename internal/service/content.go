package service

import (
	"context"
	"database/sql"
	"fmt"
)

type ContentService struct {
	db *sql.DB
}

func NewContentService(db *sql.DB) *ContentService {
	return &ContentService{db: db}
}

type WeekContentResponse struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	WeekCategory        string `json:"weekCategory"`
	WeekContentCategory string `json:"weekContentCategory"`
	Gate                int    `json:"gate"`
	Gold                int    `json:"gold"`
}

func (s *ContentService) GetWeekContent(ctx context.Context) ([]WeekContentResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, COALESCE(name, ''), COALESCE(week_category, ''), COALESCE(week_content_category, ''), gate, gold
		 FROM content
		 WHERE dtype = 'WeekContent'
		 ORDER BY level ASC`)
	if err != nil {
		return nil, fmt.Errorf("querying week content: %w", err)
	}
	defer rows.Close()

	var contents []WeekContentResponse
	for rows.Next() {
		var c WeekContentResponse
		if err := rows.Scan(&c.ID, &c.Name, &c.WeekCategory, &c.WeekContentCategory, &c.Gate, &c.Gold); err != nil {
			return nil, fmt.Errorf("scanning week content: %w", err)
		}
		contents = append(contents, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating week content: %w", err)
	}

	if contents == nil {
		contents = []WeekContentResponse{}
	}
	return contents, nil
}
