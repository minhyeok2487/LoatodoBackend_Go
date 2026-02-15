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
	CommunityID        int64    `json:"communityId"`
	CreatedDate        string   `json:"createdDate"`
	CharacterClassName string   `json:"characterClassName"`
	CharacterImage     string   `json:"characterImage"`
	Name               string   `json:"name"`
	MemberID           int64    `json:"memberId"`
	Body               string   `json:"body"`
	Category           string   `json:"category"`
	MyPost             bool     `json:"myPost"`
	LikeCount          int      `json:"likeCount"`
	MyLike             bool     `json:"myLike"`
	CommentCount       int      `json:"commentCount"`
	ImageList          []string `json:"imageList"`
}

type CommunityListResponse struct {
	Content []CommunityPostResponse `json:"content"`
	HasNext bool                    `json:"hasNext"`
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

// CommunityCommentResponse is similar to CommunityPostResponse but for comments in detail view
type CommunityCommentResponse struct {
	CommunityID        int64    `json:"communityId"`
	CreatedDate        string   `json:"createdDate"`
	CharacterClassName string   `json:"characterClassName"`
	CharacterImage     string   `json:"characterImage"`
	Name               string   `json:"name"`
	MemberID           int64    `json:"memberId"`
	Body               string   `json:"body"`
	Category           string   `json:"category"`
	MyPost             bool     `json:"myPost"`
	LikeCount          int      `json:"likeCount"`
	MyLike             bool     `json:"myLike"`
	CommentCount       int      `json:"commentCount"`
	ImageList          []string `json:"imageList"`
}

func (s *CommunityService) ListPosts(ctx context.Context, username string, category string, page, size int) (*CommunityListResponse, error) {
	// Get member ID for like check
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
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

	// Count total for hasNext calculation
	var totalCount int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM community c %s", baseWhere)
	err = s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("게시글 수 조회 실패: %w", err)
	}

	// Paginated query - include character info like Spring
	offset := page * size
	selectQuery := fmt.Sprintf(`
		SELECT c.community_id, c.created_date, c.member_id, c.name, c.body, c.category,
		       COALESCE(ch.character_class_name, '') as character_class_name,
		       COALESCE(ch.character_image, '') as character_image,
		       (SELECT COUNT(*) FROM community_like WHERE community_id = c.community_id) as like_count,
		       (SELECT COUNT(*) FROM community WHERE root_parent_id = c.community_id AND deleted = false) as comment_count,
		       EXISTS(SELECT 1 FROM community_like WHERE community_id = c.community_id AND member_id = ?) as my_like,
		       (c.member_id = ?) as my_post
		FROM community c
		LEFT JOIN characters ch ON ch.characters_id = (
			SELECT characters_id FROM characters WHERE member_id = c.member_id AND (is_deleted IS NULL OR is_deleted = false) ORDER BY sort_number LIMIT 1
		)
		%s
		ORDER BY c.created_date DESC
		LIMIT ? OFFSET ?
	`, baseWhere)

	queryArgs := []interface{}{memberID, memberID}
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, size+1, offset) // Fetch one extra to check hasNext

	rows, err := s.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("게시글 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	posts := []CommunityPostResponse{}
	for rows.Next() {
		var p CommunityPostResponse
		var createdDate time.Time
		if err := rows.Scan(
			&p.CommunityID, &createdDate, &p.MemberID, &p.Name, &p.Body, &p.Category,
			&p.CharacterClassName, &p.CharacterImage,
			&p.LikeCount, &p.CommentCount, &p.MyLike, &p.MyPost,
		); err != nil {
			return nil, fmt.Errorf("게시글 스캔 실패: %w", err)
		}
		p.CreatedDate = createdDate.Format("2006-01-02T15:04:05.000000")
		p.ImageList = []string{} // Initialize empty, will populate below
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("게시글 목록 반복 오류: %w", err)
	}

	// Calculate hasNext and trim extra item
	hasNext := len(posts) > size
	if hasNext {
		posts = posts[:size]
	}

	// Fetch images for each post
	for i := range posts {
		images, err := s.getPostImages(ctx, posts[i].CommunityID)
		if err == nil {
			posts[i].ImageList = images
		}
	}

	return &CommunityListResponse{
		Content: posts,
		HasNext: hasNext,
	}, nil
}

func (s *CommunityService) getPostImages(ctx context.Context, communityID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT url FROM community_images WHERE community_id = ? AND deleted = false ORDER BY ordering ASC",
		communityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		images = append(images, url)
	}
	if images == nil {
		images = []string{}
	}
	return images, rows.Err()
}

func (s *CommunityService) GetPost(ctx context.Context, username string, communityID int64) (*CommunityPostResponse, error) {
	// Get member ID for like check
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	query := `
		SELECT c.community_id, c.created_date, c.member_id, c.name, c.body, c.category,
		       COALESCE(ch.character_class_name, '') as character_class_name,
		       COALESCE(ch.character_image, '') as character_image,
		       (SELECT COUNT(*) FROM community_like WHERE community_id = c.community_id) as like_count,
		       (SELECT COUNT(*) FROM community WHERE root_parent_id = c.community_id AND deleted = false) as comment_count,
		       EXISTS(SELECT 1 FROM community_like WHERE community_id = c.community_id AND member_id = ?) as my_like,
		       (c.member_id = ?) as my_post
		FROM community c
		LEFT JOIN characters ch ON ch.characters_id = (
			SELECT characters_id FROM characters WHERE member_id = c.member_id AND (is_deleted IS NULL OR is_deleted = false) ORDER BY sort_number LIMIT 1
		)
		WHERE c.community_id = ? AND c.deleted = false
	`
	var p CommunityPostResponse
	var createdDate time.Time
	err = s.db.QueryRowContext(ctx, query, memberID, memberID, communityID).Scan(
		&p.CommunityID, &createdDate, &p.MemberID, &p.Name, &p.Body, &p.Category,
		&p.CharacterClassName, &p.CharacterImage,
		&p.LikeCount, &p.CommentCount, &p.MyLike, &p.MyPost,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("게시글을 찾을 수 없습니다")
		}
		return nil, fmt.Errorf("게시글 조회 실패: %w", err)
	}
	p.CreatedDate = createdDate.Format("2006-01-02T15:04:05.000000")
	p.ImageList, _ = s.getPostImages(ctx, communityID)

	return &p, nil
}

func (s *CommunityService) CreatePost(ctx context.Context, username string, req CreateCommunityRequest) (*CommunityPostResponse, error) {
	// Get member ID and character info
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// Get character info
	var characterClassName, characterImage string
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(character_class_name, ''), COALESCE(character_image, '')
		 FROM characters WHERE member_id = ? AND (is_deleted IS NULL OR is_deleted = false)
		 ORDER BY sort_number LIMIT 1`,
		memberID).Scan(&characterClassName, &characterImage)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("캐릭터 조회 실패: %w", err)
	}

	// Generate anonymous name like Spring: "익명의 {characterClassName} {memberId}"
	anonymousName := fmt.Sprintf("익명의 %s %d", characterClassName, memberID)

	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO community (member_id, name, body, category, root_parent_id, comment_parent_id, deleted, show_name, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, 0, 0, false, true, ?, ?)`,
		memberID, anonymousName, req.Body, req.Category, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("게시글 생성 실패: %w", err)
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("게시글 ID 조회 실패: %w", err)
	}

	return &CommunityPostResponse{
		CommunityID:        postID,
		CreatedDate:        now.Format("2006-01-02T15:04:05.000000"),
		CharacterClassName: characterClassName,
		CharacterImage:     characterImage,
		Name:               anonymousName,
		MemberID:           memberID,
		Body:               req.Body,
		Category:           req.Category,
		MyPost:             true,
		LikeCount:          0,
		MyLike:             false,
		CommentCount:       0,
		ImageList:          []string{},
	}, nil
}

func (s *CommunityService) UpdatePost(ctx context.Context, username string, communityID int64, req UpdateCommunityRequest) error {
	// Get member ID for ownership check
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	result, err := s.db.ExecContext(ctx,
		"UPDATE community SET name = ?, body = ?, category = ?, last_modified_date = ? WHERE community_id = ? AND member_id = ?",
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
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// Soft delete
	result, err := s.db.ExecContext(ctx,
		"UPDATE community SET deleted = true, last_modified_date = ? WHERE community_id = ? AND member_id = ?",
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
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return false, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// Check if like already exists
	var likeID int64
	err = s.db.QueryRowContext(ctx,
		"SELECT community_like_id FROM community_like WHERE member_id = ? AND community_id = ?",
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
		"DELETE FROM community_like WHERE community_like_id = ?",
		likeID,
	)
	if err != nil {
		return false, fmt.Errorf("좋아요 삭제 실패: %w", err)
	}
	return false, nil
}

func (s *CommunityService) GetImages(ctx context.Context, communityID int64) ([]CommunityImageResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT community_images_id, url FROM community_images WHERE community_id = ? AND deleted = false ORDER BY ordering ASC",
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
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	var postMemberID int64
	err = s.db.QueryRowContext(ctx, "SELECT member_id FROM community WHERE community_id = ?", communityID).Scan(&postMemberID)
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

func (s *CommunityService) GetComments(ctx context.Context, username string, communityID int64) ([]CommunityCommentResponse, error) {
	// Get member ID for myPost check
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// Get comments (root_parent_id = communityID means it's a comment on that post)
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.community_id, c.created_date, c.member_id, c.name, c.body, c.category,
		       COALESCE(ch.character_class_name, '') as character_class_name,
		       COALESCE(ch.character_image, '') as character_image,
		       (SELECT COUNT(*) FROM community_like WHERE community_id = c.community_id) as like_count,
		       (SELECT COUNT(*) FROM community WHERE root_parent_id = c.community_id AND deleted = false) as comment_count,
		       EXISTS(SELECT 1 FROM community_like WHERE community_id = c.community_id AND member_id = ?) as my_like,
		       (c.member_id = ?) as my_post
		FROM community c
		LEFT JOIN characters ch ON ch.characters_id = (
			SELECT characters_id FROM characters WHERE member_id = c.member_id AND (is_deleted IS NULL OR is_deleted = false) ORDER BY sort_number LIMIT 1
		)
		WHERE c.root_parent_id = ? AND c.deleted = false
		ORDER BY c.created_date ASC
	`, memberID, memberID, communityID)
	if err != nil {
		return nil, fmt.Errorf("댓글 조회 실패: %w", err)
	}
	defer rows.Close()

	comments := []CommunityCommentResponse{}
	for rows.Next() {
		var c CommunityCommentResponse
		var createdDate time.Time
		if err := rows.Scan(
			&c.CommunityID, &createdDate, &c.MemberID, &c.Name, &c.Body, &c.Category,
			&c.CharacterClassName, &c.CharacterImage,
			&c.LikeCount, &c.CommentCount, &c.MyLike, &c.MyPost,
		); err != nil {
			return nil, fmt.Errorf("댓글 스캔 실패: %w", err)
		}
		c.CreatedDate = createdDate.Format("2006-01-02T15:04:05.000000")
		c.ImageList = []string{}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("댓글 반복 오류: %w", err)
	}

	return comments, nil
}
