package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type FollowService struct {
	db *sql.DB
}

func NewFollowService(db *sql.DB) *FollowService {
	return &FollowService{db: db}
}

type FollowResponse struct {
	FollowID    int64     `json:"followId"`
	MemberID    int64     `json:"memberId"`
	FollowingID int64     `json:"followingId"`
	FollowName  string    `json:"followName"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ToggleFollowRequest struct {
	FollowingUsername string `json:"followingUsername"`
}

func (s *FollowService) GetFollowers(ctx context.Context, username string) ([]FollowResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("querying member: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT f.follow_id, f.member_id, f.following_id, m.username, f.created_date
		 FROM follow f JOIN member m ON f.following_id = m.member_id
		 WHERE f.member_id = ? ORDER BY f.created_date DESC`, memberID,
	)
	if err != nil {
		return []FollowResponse{}, nil
	}
	defer rows.Close()

	var followers []FollowResponse
	for rows.Next() {
		var f FollowResponse
		if err := rows.Scan(&f.FollowID, &f.MemberID, &f.FollowingID, &f.FollowName, &f.CreatedAt); err != nil {
			return []FollowResponse{}, nil
		}
		followers = append(followers, f)
	}
	if followers == nil {
		followers = []FollowResponse{}
	}
	return followers, rows.Err()
}

func (s *FollowService) ToggleFollow(ctx context.Context, username string, req ToggleFollowRequest) (map[string]interface{}, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("querying member: %w", err)
	}

	var followingID int64
	err = s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", req.FollowingUsername).Scan(&followingID)
	if err != nil {
		return nil, fmt.Errorf("대상 회원을 찾을 수 없습니다.")
	}

	// Check if already following
	var existingID int64
	err = s.db.QueryRowContext(ctx,
		"SELECT follow_id FROM follow WHERE member_id = ? AND following_id = ?", memberID, followingID,
	).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Add follow
		_, err = s.db.ExecContext(ctx,
			"INSERT INTO follow (member_id, following_id, created_date) VALUES (?, ?, ?)",
			memberID, followingID, time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("inserting follow: %w", err)
		}
		return map[string]interface{}{"following": true}, nil
	}

	// Remove follow
	_, err = s.db.ExecContext(ctx, "DELETE FROM follow WHERE follow_id = ?", existingID)
	if err != nil {
		return nil, fmt.Errorf("deleting follow: %w", err)
	}
	return map[string]interface{}{"following": false}, nil
}
