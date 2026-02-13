package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"
)

type CommunityService struct {
	db *sql.DB
}

func NewCommunityService(db *sql.DB) *CommunityService {
	return &CommunityService{db: db}
}

type CommunityPostResponse struct {
	CommunityID int64     `json:"communityId"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Username    string    `json:"username"`
	LikeCount   int       `json:"likeCount"`
	CommentCount int      `json:"commentCount"`
	Liked       bool      `json:"liked"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CommunityListResponse struct {
	Posts      []CommunityPostResponse `json:"posts"`
	TotalCount int64                   `json:"totalCount"`
	Page       int                     `json:"page"`
	Size       int                     `json:"size"`
}

type CreateCommunityRequest struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

type UpdateCommunityRequest struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

type CommunityImageResponse struct {
	ImageID  int64  `json:"imageId"`
	ImageURL string `json:"imageUrl"`
}

func (s *CommunityService) ListPosts(ctx context.Context, username string, category string, page, size int) (*CommunityListResponse, error) {
	// Get member ID for like check
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// Build query with optional category filter
	baseWhere := "WHERE c.root_parent_id = 0 AND c.deleted = false"
	args := []interface{}{}
	if category != "" {
		baseWhere += " AND c.category = ?"
		args = append(args, category)
	}

	// Count total
	var totalCount int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM community c %s", baseWhere)
	err = s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("게시글 수 조회 실패: %w", err)
	}

	// Paginated query
	offset := page * size
	selectQuery := fmt.Sprintf(`
		SELECT c.id, c.category, c.name, c.body, m.username,
		       (SELECT COUNT(*) FROM community_like WHERE community_id = c.id) as like_count,
		       (SELECT COUNT(*) FROM comments WHERE community_id = c.id) as comment_count,
		       EXISTS(SELECT 1 FROM community_like WHERE community_id = c.id AND member_id = ?) as liked,
		       c.created_date, c.last_modified_date
		FROM community c
		JOIN member m ON c.member_id = m.id
		%s
		ORDER BY c.created_date DESC
		LIMIT ? OFFSET ?
	`, baseWhere)

	queryArgs := []interface{}{memberID}
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, size, offset)

	rows, err := s.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("게시글 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	posts := []CommunityPostResponse{}
	for rows.Next() {
		var p CommunityPostResponse
		if err := rows.Scan(
			&p.CommunityID, &p.Category, &p.Title, &p.Body, &p.Username,
			&p.LikeCount, &p.CommentCount, &p.Liked,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("게시글 스캔 실패: %w", err)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("게시글 목록 반복 오류: %w", err)
	}

	return &CommunityListResponse{
		Posts:      posts,
		TotalCount: totalCount,
		Page:       page,
		Size:       size,
	}, nil
}

func (s *CommunityService) GetPost(ctx context.Context, username string, communityID int64) (*CommunityPostResponse, error) {
	// Get member ID for like check
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	query := `
		SELECT c.id, c.category, c.name, c.body, m.username,
		       (SELECT COUNT(*) FROM community_like WHERE community_id = c.id) as like_count,
		       (SELECT COUNT(*) FROM comments WHERE community_id = c.id) as comment_count,
		       EXISTS(SELECT 1 FROM community_like WHERE community_id = c.id AND member_id = ?) as liked,
		       c.created_date, c.last_modified_date
		FROM community c
		JOIN member m ON c.member_id = m.id
		WHERE c.id = ? AND c.deleted = false
	`
	var p CommunityPostResponse
	err = s.db.QueryRowContext(ctx, query, memberID, communityID).Scan(
		&p.CommunityID, &p.Category, &p.Title, &p.Body, &p.Username,
		&p.LikeCount, &p.CommentCount, &p.Liked,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("게시글을 찾을 수 없습니다")
		}
		return nil, fmt.Errorf("게시글 조회 실패: %w", err)
	}

	return &p, nil
}

func (s *CommunityService) CreatePost(ctx context.Context, username string, req CreateCommunityRequest) (*CommunityPostResponse, error) {
	// Get member ID
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO community (member_id, name, body, category, root_parent_id, comment_parent_id, deleted, show_name, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, 0, 0, false, true, ?, ?)`,
		memberID, req.Title, req.Body, req.Category, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("게시글 생성 실패: %w", err)
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("게시글 ID 조회 실패: %w", err)
	}

	return &CommunityPostResponse{
		CommunityID:  postID,
		Category:     req.Category,
		Title:        req.Title,
		Body:         req.Body,
		Username:     username,
		LikeCount:    0,
		CommentCount: 0,
		Liked:        false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (s *CommunityService) UpdatePost(ctx context.Context, username string, communityID int64, req UpdateCommunityRequest) error {
	// Get member ID for ownership check
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	result, err := s.db.ExecContext(ctx,
		"UPDATE community SET name = ?, body = ?, category = ?, last_modified_date = ? WHERE id = ? AND member_id = ?",
		req.Title, req.Body, req.Category, time.Now(), communityID, memberID,
	)
	if err != nil {
		return fmt.Errorf("게시글 수정 실패: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("게시글을 찾을 수 없거나 권한이 없습니다")
	}

	return nil
}

func (s *CommunityService) DeletePost(ctx context.Context, username string, communityID int64) error {
	// Get member ID for ownership check
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// Soft delete
	result, err := s.db.ExecContext(ctx,
		"UPDATE community SET deleted = true, last_modified_date = ? WHERE id = ? AND member_id = ?",
		time.Now(), communityID, memberID,
	)
	if err != nil {
		return fmt.Errorf("게시글 삭제 실패: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("게시글을 찾을 수 없거나 권한이 없습니다")
	}

	return nil
}

func (s *CommunityService) ToggleLike(ctx context.Context, username string, communityID int64) (bool, error) {
	// Get member ID
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return false, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// Check if like already exists
	var likeID int64
	err = s.db.QueryRowContext(ctx,
		"SELECT id FROM community_like WHERE member_id = ? AND community_id = ?",
		memberID, communityID,
	).Scan(&likeID)

	if err == sql.ErrNoRows {
		// Like does not exist, create it
		now := time.Now()
		_, err = s.db.ExecContext(ctx,
			"INSERT INTO community_like (member_id, community_id, created_date, last_modified_date) VALUES (?, ?, ?, ?)",
			memberID, communityID, now, now,
		)
		if err != nil {
			return false, fmt.Errorf("좋아요 추가 실패: %w", err)
		}
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("좋아요 조회 실패: %w", err)
	}

	// Like exists, remove it
	_, err = s.db.ExecContext(ctx,
		"DELETE FROM community_like WHERE id = ?",
		likeID,
	)
	if err != nil {
		return false, fmt.Errorf("좋아요 삭제 실패: %w", err)
	}
	return false, nil
}

func (s *CommunityService) GetImages(ctx context.Context, communityID int64) ([]CommunityImageResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, url FROM community_images WHERE community_id = ? AND deleted = false ORDER BY ordering ASC",
		communityID,
	)
	if err != nil {
		return nil, fmt.Errorf("이미지 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	images := []CommunityImageResponse{}
	for rows.Next() {
		var img CommunityImageResponse
		if err := rows.Scan(&img.ImageID, &img.ImageURL); err != nil {
			return nil, fmt.Errorf("이미지 스캔 실패: %w", err)
		}
		images = append(images, img)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("이미지 목록 반복 오류: %w", err)
	}

	return images, nil
}

func (s *CommunityService) UploadImage(ctx context.Context, username string, communityID int64, filename string, file io.Reader) (*CommunityImageResponse, error) {
	// Verify ownership
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	var postMemberID int64
	err = s.db.QueryRowContext(ctx, "SELECT member_id FROM community WHERE id = ?", communityID).Scan(&postMemberID)
	if err != nil {
		return nil, fmt.Errorf("게시글을 찾을 수 없습니다: %w", err)
	}
	if postMemberID != memberID {
		return nil, fmt.Errorf("권한이 없습니다")
	}

	// Get next ordering
	var maxOrdering int
	err = s.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(ordering), 0) FROM community_images WHERE community_id = ?",
		communityID,
	).Scan(&maxOrdering)
	if err != nil {
		return nil, fmt.Errorf("이미지 순서 조회 실패: %w", err)
	}

	// Insert with placeholder URL (actual S3 upload would replace this)
	now := time.Now()
	placeholderURL := fmt.Sprintf("/images/community/%d/%s", communityID, filename)
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO community_images (community_id, url, file_name, ordering, deleted, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, false, ?, ?)`,
		communityID, placeholderURL, filename, maxOrdering+1, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("이미지 저장 실패: %w", err)
	}

	imageID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("이미지 ID 조회 실패: %w", err)
	}

	return &CommunityImageResponse{
		ImageID:  imageID,
		ImageURL: placeholderURL,
	}, nil
}
