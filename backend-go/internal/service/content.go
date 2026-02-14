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

// WeekRaidCategoryResponse represents a unique raid category (matching Spring's format)
type WeekRaidCategoryResponse struct {
	CategoryID          int64   `json:"categoryId"`
	Name                string  `json:"name"`
	WeekContentCategory string  `json:"weekContentCategory"`
	Level               float64 `json:"level"`
}

func (s *ContentService) GetWeekContent(ctx context.Context) ([]WeekContentResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT content_id, COALESCE(name, ''), COALESCE(week_category, ''), COALESCE(week_content_category, ''), gate, gold
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

// GetWeekRaidCategory returns unique raid categories (matching Spring format)
// Spring returns gate=1 contents grouped by (id, name, week_content_category)
func (s *ContentService) GetWeekRaidCategory(ctx context.Context) ([]WeekRaidCategoryResponse, error) {
	// Spring query:
	// - Selects weekContent.id, weekContent.name, weekContentCategory, MAX(level)
	// - WHERE week_content_category IS NOT NULL AND gate = 1
	// - GROUP BY id, name, week_content_category
	// - ORDER BY MAX(level) DESC
	rows, err := s.db.QueryContext(ctx,
		`SELECT content_id as category_id, COALESCE(name, ''), COALESCE(week_content_category, ''), MAX(level) as level
		 FROM content
		 WHERE dtype = 'WeekContent' AND week_content_category IS NOT NULL AND week_content_category != '' AND gate = 1
		 GROUP BY content_id, name, week_content_category
		 ORDER BY MAX(level) DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying week raid categories: %w", err)
	}
	defer rows.Close()

	var categories []WeekRaidCategoryResponse
	for rows.Next() {
		var c WeekRaidCategoryResponse
		if err := rows.Scan(&c.CategoryID, &c.Name, &c.WeekContentCategory, &c.Level); err != nil {
			return nil, fmt.Errorf("scanning week raid category: %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating week raid categories: %w", err)
	}

	if categories == nil {
		categories = []WeekRaidCategoryResponse{}
	}
	return categories, nil
}
