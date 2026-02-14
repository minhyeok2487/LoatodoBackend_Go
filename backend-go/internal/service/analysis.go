package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type AnalysisService struct {
	db *sql.DB
}

func NewAnalysisService(db *sql.DB) *AnalysisService {
	return &AnalysisService{db: db}
}

type AnalysisResponse struct {
	ID        int64     `json:"id"`
	MemberID  int64     `json:"memberId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type AnalysisListResponse struct {
	Content []AnalysisResponse `json:"content"`
	HasNext bool               `json:"hasNext"`
}

type CreateAnalysisRequest struct {
	Content string `json:"content"`
}

func (s *AnalysisService) GetAnalyses(ctx context.Context, username string) (*AnalysisListResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return &AnalysisListResponse{Content: []AnalysisResponse{}, HasNext: false}, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT analysis_id, member_id, content_name, created_date FROM analysis
		 WHERE member_id = ? ORDER BY created_date DESC LIMIT 100`, memberID,
	)
	if err != nil {
		return &AnalysisListResponse{Content: []AnalysisResponse{}, HasNext: false}, nil
	}
	defer rows.Close()

	var analyses []AnalysisResponse
	for rows.Next() {
		var a AnalysisResponse
		if err := rows.Scan(&a.ID, &a.MemberID, &a.Content, &a.CreatedAt); err != nil {
			return &AnalysisListResponse{Content: []AnalysisResponse{}, HasNext: false}, nil
		}
		analyses = append(analyses, a)
	}
	if analyses == nil {
		analyses = []AnalysisResponse{}
	}
	return &AnalysisListResponse{Content: analyses, HasNext: false}, rows.Err()
}

func (s *AnalysisService) CreateAnalysis(ctx context.Context, username string, req CreateAnalysisRequest) (*AnalysisResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("querying member: %w", err)
	}

	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO analysis (member_id, content_name, created_date) VALUES (?, ?, ?)",
		memberID, req.Content, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting analysis: %w", err)
	}
	id, _ := result.LastInsertId()
	return &AnalysisResponse{ID: id, MemberID: memberID, Content: req.Content, CreatedAt: now}, nil
}
