package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type AdminService struct {
	db *sql.DB
}

func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{db: db}
}

// --- Dashboard ---

type DashboardResponse struct {
	TotalMembers    int64 `json:"totalMembers"`
	TotalCharacters int64 `json:"totalCharacters"`
	TotalFriends    int64 `json:"totalFriends"`
	TotalComments   int64 `json:"totalComments"`
	TodaySignups    int64 `json:"todaySignups"`
	TodayLogins     int64 `json:"todayLogins"`
}

// --- Member ---

type AdminMemberResponse struct {
	MemberID  int64     `json:"memberId"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type AdminMemberListResponse struct {
	Members    []AdminMemberResponse `json:"members"`
	TotalCount int64                 `json:"totalCount"`
	Page       int                   `json:"page"`
	Size       int                   `json:"size"`
}

type AdminMemberDetailResponse struct {
	MemberID   int64     `json:"memberId"`
	Username   string    `json:"username"`
	Role       string    `json:"role"`
	Characters []interface{} `json:"characters"`
	CreatedAt  time.Time `json:"createdAt"`
}

type UpdateMemberRequest struct {
	Role string `json:"role"`
}

// --- Character ---

type AdminCharacterResponse struct {
	CharacterID   int64   `json:"characterId"`
	CharacterName string  `json:"characterName"`
	ServerName    string  `json:"serverName"`
	ItemLevel     float64 `json:"itemLevel"`
	ClassName     string  `json:"className"`
	Username      string  `json:"username"`
}

type AdminCharacterListResponse struct {
	Characters []AdminCharacterResponse `json:"characters"`
	TotalCount int64                    `json:"totalCount"`
	Page       int                      `json:"page"`
	Size       int                      `json:"size"`
}

type AdminUpdateCharacterRequest struct {
	CharacterName string  `json:"characterName"`
	ServerName    string  `json:"serverName"`
	ItemLevel     float64 `json:"itemLevel"`
	ClassName     string  `json:"className"`
}

// --- Friends ---

type AdminFriendResponse struct {
	FriendID  int64  `json:"friendId"`
	Username1 string `json:"username1"`
	Username2 string `json:"username2"`
	AreWeFriends bool `json:"areWeFriends"`
}

// --- Comments ---

type AdminCommentResponse struct {
	CommentID   int64     `json:"commentId"`
	CommunityID int64     `json:"communityId"`
	Username    string    `json:"username"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"createdAt"`
}

// --- Notification ---

type AdminNotificationResponse struct {
	NotificationID int64     `json:"notificationId"`
	Username       string    `json:"username"`
	Content        string    `json:"content"`
	IsRead         bool      `json:"isRead"`
	CreatedAt      time.Time `json:"createdAt"`
}

type CreateNotificationRequest struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

// --- Content ---

type AdminContentResponse struct {
	ContentID  int64  `json:"contentId"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	Enabled    bool   `json:"enabled"`
}

type UpdateContentRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Enabled  bool   `json:"enabled"`
}

// --- Ads ---

type ManageAdsRequest struct {
	Action string `json:"action"`
	AdData interface{} `json:"adData"`
}

func (s *AdminService) GetDashboard(ctx context.Context) (*DashboardResponse, error) {
	var resp DashboardResponse
	err := s.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM member) as total_members,
			(SELECT COUNT(*) FROM character_entity WHERE deleted = false) as total_characters,
			(SELECT COUNT(*) FROM friends) as total_friends,
			(SELECT COUNT(*) FROM comments) as total_comments,
			(SELECT COUNT(*) FROM member WHERE DATE(created_date) = CURDATE()) as today_signups,
			0 as today_logins
	`).Scan(
		&resp.TotalMembers,
		&resp.TotalCharacters,
		&resp.TotalFriends,
		&resp.TotalComments,
		&resp.TodaySignups,
		&resp.TodayLogins,
	)
	if err != nil {
		return nil, fmt.Errorf("대시보드 조회 실패: %w", err)
	}
	return &resp, nil
}

func (s *AdminService) ListMembers(ctx context.Context, search string, page, size int) (*AdminMemberListResponse, error) {
	// Build WHERE clause
	where := ""
	args := []interface{}{}
	if search != "" {
		where = "WHERE username LIKE ?"
		args = append(args, "%"+search+"%")
	}

	// Count total
	var totalCount int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM member %s", where)
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("회원 수 조회 실패: %w", err)
	}

	// Paginated query
	offset := page * size
	selectQuery := fmt.Sprintf(
		"SELECT id, username, role, created_date FROM member %s ORDER BY created_date DESC LIMIT ? OFFSET ?",
		where,
	)
	queryArgs := append(args, size, offset)

	rows, err := s.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("회원 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	members := []AdminMemberResponse{}
	for rows.Next() {
		var m AdminMemberResponse
		if err := rows.Scan(&m.MemberID, &m.Username, &m.Role, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("회원 스캔 실패: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("회원 목록 반복 오류: %w", err)
	}

	return &AdminMemberListResponse{
		Members:    members,
		TotalCount: totalCount,
		Page:       page,
		Size:       size,
	}, nil
}

func (s *AdminService) GetMemberDetail(ctx context.Context, memberID int64) (*AdminMemberDetailResponse, error) {
	// Get member info
	var resp AdminMemberDetailResponse
	err := s.db.QueryRowContext(ctx,
		"SELECT id, username, role, created_date FROM member WHERE id = ?",
		memberID,
	).Scan(&resp.MemberID, &resp.Username, &resp.Role, &resp.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("회원을 찾을 수 없습니다")
		}
		return nil, fmt.Errorf("회원 상세 조회 실패: %w", err)
	}

	// Get characters
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, character_name, server_name, item_level, character_class_name
		 FROM character_entity WHERE member_id = ? AND deleted = false ORDER BY sort_number ASC`,
		memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("캐릭터 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	characters := []interface{}{}
	for rows.Next() {
		var c AdminCharacterResponse
		if err := rows.Scan(&c.CharacterID, &c.CharacterName, &c.ServerName, &c.ItemLevel, &c.ClassName); err != nil {
			return nil, fmt.Errorf("캐릭터 스캔 실패: %w", err)
		}
		c.Username = resp.Username
		characters = append(characters, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("캐릭터 목록 반복 오류: %w", err)
	}

	resp.Characters = characters
	return &resp, nil
}

func (s *AdminService) UpdateMember(ctx context.Context, memberID int64, req UpdateMemberRequest) error {
	// Validate role
	if req.Role != "USER" && req.Role != "ADMIN" {
		return fmt.Errorf("유효하지 않은 역할입니다: %s", req.Role)
	}

	result, err := s.db.ExecContext(ctx,
		"UPDATE member SET role = ?, last_modified_date = ? WHERE id = ?",
		req.Role, time.Now(), memberID,
	)
	if err != nil {
		return fmt.Errorf("회원 업데이트 실패: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("회원을 찾을 수 없습니다")
	}

	return nil
}

func (s *AdminService) DeleteMember(ctx context.Context, memberID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback()

	// Delete cascade: notifications
	_, err = tx.ExecContext(ctx, "DELETE FROM notification WHERE receiver_id = ?", memberID)
	if err != nil {
		return fmt.Errorf("알림 삭제 실패: %w", err)
	}

	// Delete cascade: community likes
	_, err = tx.ExecContext(ctx, "DELETE FROM community_like WHERE member_id = ?", memberID)
	if err != nil {
		return fmt.Errorf("좋아요 삭제 실패: %w", err)
	}

	// Delete cascade: comments
	_, err = tx.ExecContext(ctx, "DELETE FROM comments WHERE member_id = ?", memberID)
	if err != nil {
		return fmt.Errorf("댓글 삭제 실패: %w", err)
	}

	// Delete cascade: community images for member's posts
	_, err = tx.ExecContext(ctx,
		"DELETE FROM community_images WHERE community_id IN (SELECT id FROM community WHERE member_id = ?)",
		memberID,
	)
	if err != nil {
		return fmt.Errorf("커뮤니티 이미지 삭제 실패: %w", err)
	}

	// Delete cascade: community posts
	_, err = tx.ExecContext(ctx, "DELETE FROM community WHERE member_id = ?", memberID)
	if err != nil {
		return fmt.Errorf("커뮤니티 게시글 삭제 실패: %w", err)
	}

	// Delete cascade: friends (both directions)
	_, err = tx.ExecContext(ctx, "DELETE FROM friends WHERE member_id = ? OR from_member = ?", memberID, memberID)
	if err != nil {
		return fmt.Errorf("친구 삭제 실패: %w", err)
	}

	// Delete cascade: characters
	_, err = tx.ExecContext(ctx, "DELETE FROM character_entity WHERE member_id = ?", memberID)
	if err != nil {
		return fmt.Errorf("캐릭터 삭제 실패: %w", err)
	}

	// Delete the member
	result, err := tx.ExecContext(ctx, "DELETE FROM member WHERE id = ?", memberID)
	if err != nil {
		return fmt.Errorf("회원 삭제 실패: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("회원을 찾을 수 없습니다")
	}

	return tx.Commit()
}

func (s *AdminService) ListCharacters(ctx context.Context, page, size int) (*AdminCharacterListResponse, error) {
	// Count total
	var totalCount int64
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM character_entity WHERE deleted = false").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("캐릭터 수 조회 실패: %w", err)
	}

	// Paginated query
	offset := page * size
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.character_name, c.server_name, c.item_level, c.character_class_name, m.username
		 FROM character_entity c
		 JOIN member m ON c.member_id = m.id
		 WHERE c.deleted = false
		 ORDER BY c.item_level DESC
		 LIMIT ? OFFSET ?`,
		size, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("캐릭터 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	characters := []AdminCharacterResponse{}
	for rows.Next() {
		var c AdminCharacterResponse
		if err := rows.Scan(&c.CharacterID, &c.CharacterName, &c.ServerName, &c.ItemLevel, &c.ClassName, &c.Username); err != nil {
			return nil, fmt.Errorf("캐릭터 스캔 실패: %w", err)
		}
		characters = append(characters, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("캐릭터 목록 반복 오류: %w", err)
	}

	return &AdminCharacterListResponse{
		Characters: characters,
		TotalCount: totalCount,
		Page:       page,
		Size:       size,
	}, nil
}

func (s *AdminService) GetCharacterDetail(ctx context.Context, characterID int64) (*AdminCharacterResponse, error) {
	var c AdminCharacterResponse
	err := s.db.QueryRowContext(ctx,
		`SELECT c.id, c.character_name, c.server_name, c.item_level, c.character_class_name, m.username
		 FROM character_entity c
		 JOIN member m ON c.member_id = m.id
		 WHERE c.id = ?`,
		characterID,
	).Scan(&c.CharacterID, &c.CharacterName, &c.ServerName, &c.ItemLevel, &c.ClassName, &c.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("캐릭터를 찾을 수 없습니다")
		}
		return nil, fmt.Errorf("캐릭터 상세 조회 실패: %w", err)
	}
	return &c, nil
}

func (s *AdminService) UpdateCharacter(ctx context.Context, characterID int64, req AdminUpdateCharacterRequest) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE character_entity
		 SET character_name = ?, server_name = ?, item_level = ?, character_class_name = ?, last_modified_date = ?
		 WHERE id = ?`,
		req.CharacterName, req.ServerName, req.ItemLevel, req.ClassName, time.Now(), characterID,
	)
	if err != nil {
		return fmt.Errorf("캐릭터 업데이트 실패: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("캐릭터를 찾을 수 없습니다")
	}

	return nil
}

func (s *AdminService) ListFriends(ctx context.Context) ([]AdminFriendResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT f.id, m1.username, m2.username, f.are_we_friend
		 FROM friends f
		 JOIN member m1 ON f.member_id = m1.id
		 JOIN member m2 ON f.from_member = m2.id
		 ORDER BY f.id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("친구 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	friends := []AdminFriendResponse{}
	for rows.Next() {
		var f AdminFriendResponse
		if err := rows.Scan(&f.FriendID, &f.Username1, &f.Username2, &f.AreWeFriends); err != nil {
			return nil, fmt.Errorf("친구 스캔 실패: %w", err)
		}
		friends = append(friends, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("친구 목록 반복 오류: %w", err)
	}

	return friends, nil
}

func (s *AdminService) ListAllComments(ctx context.Context) ([]AdminCommentResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.community_id, m.username, c.body, c.created_date
		 FROM comments c
		 JOIN member m ON c.member_id = m.id
		 ORDER BY c.created_date DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("댓글 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	comments := []AdminCommentResponse{}
	for rows.Next() {
		var c AdminCommentResponse
		if err := rows.Scan(&c.CommentID, &c.CommunityID, &c.Username, &c.Body, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("댓글 스캔 실패: %w", err)
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("댓글 목록 반복 오류: %w", err)
	}

	return comments, nil
}

func (s *AdminService) DeleteComment(ctx context.Context, commentID int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		return fmt.Errorf("댓글 삭제 실패: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("댓글을 찾을 수 없습니다")
	}

	return nil
}

func (s *AdminService) ListAllNotifications(ctx context.Context) ([]AdminNotificationResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT n.id, m.username, n.content, n.is_read, n.created_date
		 FROM notification n
		 JOIN member m ON n.receiver_id = m.id
		 ORDER BY n.created_date DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("알림 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	notifications := []AdminNotificationResponse{}
	for rows.Next() {
		var n AdminNotificationResponse
		if err := rows.Scan(&n.NotificationID, &n.Username, &n.Content, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("알림 스캔 실패: %w", err)
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("알림 목록 반복 오류: %w", err)
	}

	return notifications, nil
}

func (s *AdminService) CreateNotification(ctx context.Context, req CreateNotificationRequest) error {
	// Get receiver member ID
	var receiverID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM member WHERE username = ?", req.Username).Scan(&receiverID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("회원을 찾을 수 없습니다: %s", req.Username)
		}
		return fmt.Errorf("회원 조회 실패: %w", err)
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO notification (content, is_read, notification_type, receiver_id, created_date, last_modified_date)
		 VALUES (?, false, 'ADMIN', ?, ?, ?)`,
		req.Content, receiverID, now, now,
	)
	if err != nil {
		return fmt.Errorf("알림 생성 실패: %w", err)
	}

	return nil
}

func (s *AdminService) ListContent(ctx context.Context) ([]AdminContentResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, category, enabled FROM content ORDER BY id ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("컨텐츠 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	contents := []AdminContentResponse{}
	for rows.Next() {
		var c AdminContentResponse
		if err := rows.Scan(&c.ContentID, &c.Name, &c.Category, &c.Enabled); err != nil {
			return nil, fmt.Errorf("컨텐츠 스캔 실패: %w", err)
		}
		contents = append(contents, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("컨텐츠 목록 반복 오류: %w", err)
	}

	return contents, nil
}

func (s *AdminService) UpdateContent(ctx context.Context, contentID int64, req UpdateContentRequest) error {
	result, err := s.db.ExecContext(ctx,
		"UPDATE content SET name = ?, category = ?, enabled = ? WHERE id = ?",
		req.Name, req.Category, req.Enabled, contentID,
	)
	if err != nil {
		return fmt.Errorf("컨텐츠 업데이트 실패: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("컨텐츠를 찾을 수 없습니다")
	}

	return nil
}

func (s *AdminService) ManageAds(ctx context.Context, req ManageAdsRequest) error {
	switch req.Action {
	case "UPDATE_ALL":
		// Update ads_date for all members
		now := time.Now()
		_, err := s.db.ExecContext(ctx,
			"UPDATE member SET ads_date = ?, last_modified_date = ?",
			now, now,
		)
		if err != nil {
			return fmt.Errorf("광고 일괄 업데이트 실패: %w", err)
		}
		return nil
	case "RESET_ALL":
		// Reset ads_date to NULL for all members
		now := time.Now()
		_, err := s.db.ExecContext(ctx,
			"UPDATE member SET ads_date = NULL, last_modified_date = ?",
			now,
		)
		if err != nil {
			return fmt.Errorf("광고 일괄 초기화 실패: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("지원하지 않는 광고 작업입니다: %s", req.Action)
	}
}
