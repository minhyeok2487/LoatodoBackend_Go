package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// bitToBool converts MySQL BIT(1)/TINYINT(1) to bool
func bitToBool(b []byte) bool {
	return len(b) > 0 && b[0] == 1
}

type FriendService struct {
	db *sql.DB
}

func NewFriendService(db *sql.DB) *FriendService {
	return &FriendService{db: db}
}

// FriendSettings mirrors the embedded FriendSettings entity.
// Each field is a permission toggle for what a friend can see/do.
type FriendSettings struct {
	ShowDayTodo  bool `json:"showDayTodo"`
	ShowRaid     bool `json:"showRaid"`
	ShowWeekTodo bool `json:"showWeekTodo"`
	CheckDayTodo bool `json:"checkDayTodo"`
	CheckRaid    bool `json:"checkRaid"`
	CheckWeekTodo bool `json:"checkWeekTodo"`
	UpdateGauge  bool `json:"updateGauge"`
	UpdateRaid   bool `json:"updateRaid"`
	Setting      bool `json:"setting"`
}

// NewFriendSettings returns default settings matching Java constructor.
func NewFriendSettings() FriendSettings {
	return FriendSettings{
		ShowDayTodo:   true,
		ShowRaid:      true,
		ShowWeekTodo:  true,
		CheckDayTodo:  false,
		CheckRaid:     false,
		CheckWeekTodo: false,
		UpdateGauge:   false,
		UpdateRaid:    false,
		Setting:       false,
	}
}

// Update sets a field by name, mirroring the Java reflection-based update.
func (fs *FriendSettings) Update(name string, value bool) error {
	switch name {
	case "showDayTodo":
		fs.ShowDayTodo = value
	case "showRaid":
		fs.ShowRaid = value
	case "showWeekTodo":
		fs.ShowWeekTodo = value
	case "checkDayTodo":
		fs.CheckDayTodo = value
	case "checkRaid":
		fs.CheckRaid = value
	case "checkWeekTodo":
		fs.CheckWeekTodo = value
	case "updateGauge":
		fs.UpdateGauge = value
	case "updateRaid":
		fs.UpdateRaid = value
	case "setting":
		fs.Setting = value
	default:
		return fmt.Errorf("없는 필드 값 입니다: %s", name)
	}
	return nil
}

// FriendsResponse is the main DTO returned by GET /api/v1/friend.
// Matches Java FriendsResponse exactly.
type FriendsResponse struct {
	FriendID           int64            `json:"friendId"`
	FriendUsername      string          `json:"friendUsername"`
	AreWeFriend        string           `json:"areWeFriend"`
	NickName           string           `json:"nickName"`
	Ordering           int              `json:"ordering"`
	CharacterList      []interface{}    `json:"characterList"`
	ToFriendSettings   *FriendSettings  `json:"toFriendSettings"`
	FromFriendSettings *FriendSettings  `json:"fromFriendSettings"`
}

// FriendFindCharacterResponse is the DTO for character search results.
// Matches Java FriendFindCharacterResponse.
type FriendFindCharacterResponse struct {
	ID                int64  `json:"id"`
	Username          string `json:"username"`
	CharacterName     string `json:"characterName"`
	CharacterListSize int    `json:"characterListSize"`
	AreWeFriend       string `json:"areWeFriend"`
}

// FriendRequest is the body for POST /api/v1/friend (send friend request).
type FriendRequest struct {
	FriendUsername string `json:"friendUsername"`
}

// FriendRequestAction is the body for POST /api/v1/friend/request (accept/reject/delete).
// Category must be one of: "OK", "REJECT", "DELETE".
type FriendRequestAction struct {
	FriendUsername string `json:"friendUsername"`
	Category       string `json:"category"`
}

// FriendSettingsUpdateRequest is the body for PATCH /api/v1/friend/settings.
// ID is the friend entity ID. Name is the settings field name. Value is the new boolean value.
type FriendSettingsUpdateRequest struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

// FriendSortRequest is the body for PUT /api/v1/friend/sort.
type FriendSortRequest struct {
	FriendIDList []int64 `json:"friendIdList"`
}

// friendRecord holds data from the friends table for batch processing
type friendRecord struct {
	FriendID       int64
	FriendUsername string
	AreWeFriend    bool
	Ordering       int
	Settings       FriendSettings
}

// reverseRecord holds reverse friendship data
type reverseRecord struct {
	AreWeFriend bool
	Settings    FriendSettings
}

// GetFriends returns the friend list for a member.
// Optimized: Uses batch queries to avoid N+1 problem.
func (s *FriendService) GetFriends(ctx context.Context, username string) ([]FriendsResponse, error) {
	// 1. Get member by username
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// 2. Query all friend records for this member (my -> friend direction)
	query := `
		SELECT f.friends_id, m2.username, f.are_we_friend, f.ordering,
		       f.show_day_todo, f.show_raid, f.show_week_todo,
		       f.check_day_todo, f.check_raid, f.check_week_todo,
		       f.update_gauge, f.update_raid, f.setting
		FROM friends f
		JOIN member m2 ON f.from_member = m2.member_id
		WHERE f.member_id = ?
		ORDER BY f.ordering ASC
	`
	rows, err := s.db.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("친구 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	var friends []friendRecord
	var friendUsernames []string
	for rows.Next() {
		var (
			friendID       int64
			friendUsername string
			areWeFriendRaw []byte
			ordering       int
			showDayTodo, showRaid, showWeekTodo       []byte
			checkDayTodo, checkRaid, checkWeekTodo    []byte
			updateGauge, updateRaid, setting          []byte
		)
		if err := rows.Scan(
			&friendID, &friendUsername, &areWeFriendRaw, &ordering,
			&showDayTodo, &showRaid, &showWeekTodo,
			&checkDayTodo, &checkRaid, &checkWeekTodo,
			&updateGauge, &updateRaid, &setting,
		); err != nil {
			return nil, fmt.Errorf("친구 목록 스캔 실패: %w", err)
		}

		friends = append(friends, friendRecord{
			FriendID:       friendID,
			FriendUsername: friendUsername,
			AreWeFriend:    bitToBool(areWeFriendRaw),
			Ordering:       ordering,
			Settings: FriendSettings{
				ShowDayTodo:   bitToBool(showDayTodo),
				ShowRaid:      bitToBool(showRaid),
				ShowWeekTodo:  bitToBool(showWeekTodo),
				CheckDayTodo:  bitToBool(checkDayTodo),
				CheckRaid:     bitToBool(checkRaid),
				CheckWeekTodo: bitToBool(checkWeekTodo),
				UpdateGauge:   bitToBool(updateGauge),
				UpdateRaid:    bitToBool(updateRaid),
				Setting:       bitToBool(setting),
			},
		})
		friendUsernames = append(friendUsernames, friendUsername)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("친구 목록 반복 오류: %w", err)
	}

	if len(friends) == 0 {
		return []FriendsResponse{}, nil
	}

	// 3. Batch query: Get all reverse records in one query (friend -> me direction)
	reverseMap := make(map[string]reverseRecord)
	reverseQuery := `
		SELECT m1.username, f.are_we_friend,
		       f.show_day_todo, f.show_raid, f.show_week_todo,
		       f.check_day_todo, f.check_raid, f.check_week_todo,
		       f.update_gauge, f.update_raid, f.setting
		FROM friends f
		JOIN member m1 ON f.member_id = m1.member_id
		WHERE f.from_member = ?
	`
	reverseRows, err := s.db.QueryContext(ctx, reverseQuery, memberID)
	if err != nil {
		return nil, fmt.Errorf("역방향 친구 조회 실패: %w", err)
	}
	defer reverseRows.Close()

	for reverseRows.Next() {
		var (
			friendUsername string
			areWeFriendRaw []byte
			showDayTodo, showRaid, showWeekTodo       []byte
			checkDayTodo, checkRaid, checkWeekTodo    []byte
			updateGauge, updateRaid, setting          []byte
		)
		if err := reverseRows.Scan(
			&friendUsername, &areWeFriendRaw,
			&showDayTodo, &showRaid, &showWeekTodo,
			&checkDayTodo, &checkRaid, &checkWeekTodo,
			&updateGauge, &updateRaid, &setting,
		); err != nil {
			return nil, fmt.Errorf("역방향 친구 스캔 실패: %w", err)
		}

		reverseMap[friendUsername] = reverseRecord{
			AreWeFriend: bitToBool(areWeFriendRaw),
			Settings: FriendSettings{
				ShowDayTodo:   bitToBool(showDayTodo),
				ShowRaid:      bitToBool(showRaid),
				ShowWeekTodo:  bitToBool(showWeekTodo),
				CheckDayTodo:  bitToBool(checkDayTodo),
				CheckRaid:     bitToBool(checkRaid),
				CheckWeekTodo: bitToBool(checkWeekTodo),
				UpdateGauge:   bitToBool(updateGauge),
				UpdateRaid:    bitToBool(updateRaid),
				Setting:       bitToBool(setting),
			},
		}
	}
	if err := reverseRows.Err(); err != nil {
		return nil, fmt.Errorf("역방향 친구 반복 오류: %w", err)
	}

	// 4. Build response by combining forward and reverse data
	var results []FriendsResponse
	for _, f := range friends {
		reverse, hasReverse := reverseMap[f.FriendUsername]

		// Determine status string
		var status string
		reverseAreWeFriend := hasReverse && reverse.AreWeFriend
		switch {
		case f.AreWeFriend && reverseAreWeFriend:
			status = "깐부"
		case f.AreWeFriend && !reverseAreWeFriend:
			status = "깐부 요청 진행중"
		case !f.AreWeFriend && reverseAreWeFriend:
			status = "깐부 요청 받음"
		default:
			status = "요청 거부"
		}

		fromSettings := FriendSettings{}
		if hasReverse {
			fromSettings = reverse.Settings
		}

		toSettings := f.Settings
		resp := FriendsResponse{
			FriendID:           f.FriendID,
			FriendUsername:     f.FriendUsername,
			AreWeFriend:        status,
			NickName:           f.FriendUsername,
			Ordering:           f.Ordering,
			CharacterList:      []interface{}{},
			ToFriendSettings:   &toSettings,
			FromFriendSettings: &fromSettings,
		}
		results = append(results, resp)
	}

	if results == nil {
		results = []FriendsResponse{}
	}
	return results, nil
}

// SearchCharacter searches for characters by name, excluding the requester.
// Optimized: Uses batch query to avoid N+1 problem.
func (s *FriendService) SearchCharacter(ctx context.Context, username, characterName string) ([]FriendFindCharacterResponse, error) {
	// 1. Get requester member ID
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// 2. Query characters by exact name match with friendship status in one query using LEFT JOINs
	query := `
		SELECT DISTINCT m.member_id, m.username, c.character_name,
		       (SELECT COUNT(*) FROM characters WHERE member_id = m.member_id AND is_deleted = false) as char_count,
		       my_f.are_we_friend as my_are_we_friend,
		       their_f.are_we_friend as their_are_we_friend
		FROM characters c
		JOIN member m ON c.member_id = m.member_id
		LEFT JOIN friends my_f ON my_f.member_id = ? AND my_f.from_member = m.member_id
		LEFT JOIN friends their_f ON their_f.member_id = m.member_id AND their_f.from_member = ?
		WHERE c.character_name = ? AND m.username != ?
	`
	rows, err := s.db.QueryContext(ctx, query, memberID, memberID, characterName, username)
	if err != nil {
		return nil, fmt.Errorf("캐릭터 검색 실패: %w", err)
	}
	defer rows.Close()

	var results []FriendFindCharacterResponse
	for rows.Next() {
		var resp FriendFindCharacterResponse
		var myAreWeFriendRaw, theirAreWeFriendRaw []byte

		if err := rows.Scan(&resp.ID, &resp.Username, &resp.CharacterName, &resp.CharacterListSize,
			&myAreWeFriendRaw, &theirAreWeFriendRaw); err != nil {
			return nil, fmt.Errorf("캐릭터 검색 스캔 실패: %w", err)
		}

		// Determine friend status
		myRecordExists := len(myAreWeFriendRaw) > 0
		theirRecordExists := len(theirAreWeFriendRaw) > 0

		if myRecordExists || theirRecordExists {
			toFriends := myRecordExists && bitToBool(myAreWeFriendRaw)
			fromFriends := theirRecordExists && bitToBool(theirAreWeFriendRaw)
			switch {
			case toFriends && fromFriends:
				resp.AreWeFriend = "깐부"
			case toFriends:
				resp.AreWeFriend = "깐부 요청 진행중"
			case fromFriends:
				resp.AreWeFriend = "깐부 요청 받음"
			default:
				resp.AreWeFriend = "요청 거부"
			}
		} else {
			resp.AreWeFriend = "깐부 요청"
		}

		results = append(results, resp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("캐릭터 검색 반복 오류: %w", err)
	}

	if results == nil {
		results = []FriendFindCharacterResponse{}
	}
	return results, nil
}

// SendFriendRequest creates a bidirectional friend request.
// Java: friendsService.addFriendsRequest(toMember, fromMember)
// Creates two Friends records: one for requester (areWeFriend=true) and one for target (areWeFriend=false).
func (s *FriendService) SendFriendRequest(ctx context.Context, username string, req FriendRequest) error {
	// 1. Get requester (toMember) by username
	var toMemberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&toMemberID)
	if err != nil {
		return fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// Verify requester has characters
	var charCount int
	err = s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM characters WHERE member_id = ? AND is_deleted = false",
		toMemberID,
	).Scan(&charCount)
	if err != nil {
		return fmt.Errorf("캐릭터 확인 실패: %w", err)
	}
	if charCount == 0 {
		return errors.New("등록된 캐릭터가 없습니다")
	}

	// 2. Get target (fromMember) by username
	var fromMemberID int64
	err = s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", req.FriendUsername).Scan(&fromMemberID)
	if err != nil {
		return fmt.Errorf("대상 회원을 찾을 수 없습니다: %w", err)
	}

	// Cannot friend yourself
	if toMemberID == fromMemberID {
		return errors.New("자기 자신에게 깐부 요청을 보낼 수 없습니다")
	}

	// 3. Check no existing friendship
	var existingCount int
	err = s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM friends WHERE member_id = ? AND from_member = ?",
		toMemberID, fromMemberID,
	).Scan(&existingCount)
	if err != nil {
		return fmt.Errorf("기존 친구 확인 실패: %w", err)
	}
	if existingCount > 0 {
		return errors.New("이미 깐부이거나 요청이 진행중입니다")
	}

	// 4. Insert two rows in a transaction
	now := time.Now()
	ds := NewFriendSettings()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO friends (member_id, from_member, are_we_friend, ordering,
		                     show_day_todo, show_raid, show_week_todo,
		                     check_day_todo, check_raid, check_week_todo,
		                     update_gauge, update_raid, setting,
		                     created_date, last_modified_date)
		VALUES (?, ?, ?, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Requester side: areWeFriend=true
	_, err = tx.ExecContext(ctx, insertQuery,
		toMemberID, fromMemberID, true,
		ds.ShowDayTodo, ds.ShowRaid, ds.ShowWeekTodo,
		ds.CheckDayTodo, ds.CheckRaid, ds.CheckWeekTodo,
		ds.UpdateGauge, ds.UpdateRaid, ds.Setting,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("친구 요청 삽입 실패 (요청자): %w", err)
	}

	// Target side: areWeFriend=false
	_, err = tx.ExecContext(ctx, insertQuery,
		fromMemberID, toMemberID, false,
		ds.ShowDayTodo, ds.ShowRaid, ds.ShowWeekTodo,
		ds.CheckDayTodo, ds.CheckRaid, ds.CheckWeekTodo,
		ds.UpdateGauge, ds.UpdateRaid, ds.Setting,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("친구 요청 삽입 실패 (대상): %w", err)
	}

	// 5. Create notification for the target
	_, err = tx.ExecContext(ctx,
		`INSERT INTO notification (content, is_read, notification_type, member_id, friend_username, created_date, last_modified_date)
		 VALUES (?, false, 'FRIEND_REQUEST', ?, ?, ?, ?)`,
		username+"님이 깐부 요청을 보냈습니다", fromMemberID, username, now, now,
	)
	if err != nil {
		return fmt.Errorf("알림 생성 실패: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}
	return nil
}

// HandleFriendRequest processes accept/reject/delete of a friend request.
// Java: friendsService.updateFriendsRequestV2(toMember, fromMember, category)
func (s *FriendService) HandleFriendRequest(ctx context.Context, username string, req FriendRequestAction) error {
	// Get both member IDs
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	var friendMemberID int64
	err = s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", req.FriendUsername).Scan(&friendMemberID)
	if err != nil {
		return fmt.Errorf("대상 회원을 찾을 수 없습니다: %w", err)
	}

	now := time.Now()

	switch req.Category {
	case "OK":
		// Accept: set areWeFriend=true on the receiver's Friends record
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("트랜잭션 시작 실패: %w", err)
		}
		defer tx.Rollback()

		result, err := tx.ExecContext(ctx,
			"UPDATE friends SET are_we_friend = true, last_modified_date = ? WHERE member_id = ? AND from_member = ?",
			now, memberID, friendMemberID,
		)
		if err != nil {
			return fmt.Errorf("친구 요청 수락 실패: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return errors.New("해당 깐부 요청을 찾을 수 없습니다")
		}

		// Create notification for the requester
		_, err = tx.ExecContext(ctx,
			`INSERT INTO notification (content, is_read, notification_type, member_id, friend_username, created_date, last_modified_date)
			 VALUES (?, false, 'FRIEND_ACCEPTED', ?, ?, ?, ?)`,
			username+"님이 깐부 요청을 수락했습니다", friendMemberID, username, now, now,
		)
		if err != nil {
			return fmt.Errorf("알림 생성 실패: %w", err)
		}

		return tx.Commit()

	case "REJECT":
		// Reject: delete both Friends records
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("트랜잭션 시작 실패: %w", err)
		}
		defer tx.Rollback()

		_, err = tx.ExecContext(ctx,
			"DELETE FROM friends WHERE member_id = ? AND from_member = ?",
			memberID, friendMemberID,
		)
		if err != nil {
			return fmt.Errorf("친구 요청 거절 실패 (내 레코드): %w", err)
		}
		_, err = tx.ExecContext(ctx,
			"DELETE FROM friends WHERE member_id = ? AND from_member = ?",
			friendMemberID, memberID,
		)
		if err != nil {
			return fmt.Errorf("친구 요청 거절 실패 (상대 레코드): %w", err)
		}

		// Create notification for the requester
		_, err = tx.ExecContext(ctx,
			`INSERT INTO notification (content, is_read, notification_type, member_id, friend_username, created_date, last_modified_date)
			 VALUES (?, false, 'FRIEND_REJECTED', ?, ?, ?, ?)`,
			username+"님이 깐부 요청을 거절했습니다", friendMemberID, username, now, now,
		)
		if err != nil {
			return fmt.Errorf("알림 생성 실패: %w", err)
		}

		return tx.Commit()

	case "DELETE":
		// Delete: remove both Friends records
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("트랜잭션 시작 실패: %w", err)
		}
		defer tx.Rollback()

		_, err = tx.ExecContext(ctx,
			"DELETE FROM friends WHERE member_id = ? AND from_member = ?",
			memberID, friendMemberID,
		)
		if err != nil {
			return fmt.Errorf("친구 삭제 실패 (내 레코드): %w", err)
		}
		_, err = tx.ExecContext(ctx,
			"DELETE FROM friends WHERE member_id = ? AND from_member = ?",
			friendMemberID, memberID,
		)
		if err != nil {
			return fmt.Errorf("친구 삭제 실패 (상대 레코드): %w", err)
		}

		return tx.Commit()

	default:
		return errors.New("지원하지 않는 친구 요청 카테고리입니다: " + req.Category)
	}
}

// UpdateSettings updates a single friend settings field by name.
// Java: friendsService.updateSetting(request) using reflection-based FriendSettings.update(name, value).
func (s *FriendService) UpdateSettings(ctx context.Context, username string, req FriendSettingsUpdateRequest) (*FriendSettings, error) {
	// Map field names to column names
	columnMap := map[string]string{
		"showDayTodo":  "show_day_todo",
		"showRaid":     "show_raid",
		"showWeekTodo": "show_week_todo",
		"checkDayTodo": "check_day_todo",
		"checkRaid":    "check_raid",
		"checkWeekTodo": "check_week_todo",
		"updateGauge":  "update_gauge",
		"updateRaid":   "update_raid",
		"setting":      "setting",
	}

	// Validate the field name using FriendSettings.Update
	settings := NewFriendSettings()
	if err := settings.Update(req.Name, req.Value); err != nil {
		return nil, err
	}

	column, ok := columnMap[req.Name]
	if !ok {
		return nil, fmt.Errorf("없는 필드 값 입니다: %s", req.Name)
	}

	// Verify ownership: the friend record must belong to the requesting user
	var ownerUsername string
	err := s.db.QueryRowContext(ctx,
		"SELECT m.username FROM friends f JOIN member m ON f.member_id = m.member_id WHERE f.friends_id = ?",
		req.ID,
	).Scan(&ownerUsername)
	if err != nil {
		return nil, fmt.Errorf("친구 설정을 찾을 수 없습니다: %w", err)
	}
	if ownerUsername != username {
		return nil, errors.New("권한이 없습니다")
	}

	// Update the specific column
	query := fmt.Sprintf("UPDATE friends SET %s = ?, last_modified_date = ? WHERE friends_id = ?", column)
	_, err = s.db.ExecContext(ctx, query, req.Value, time.Now(), req.ID)
	if err != nil {
		return nil, fmt.Errorf("친구 설정 업데이트 실패: %w", err)
	}

	// Read back the full settings (BIT fields)
	var showDayTodo, showRaid, showWeekTodo       []byte
	var checkDayTodo, checkRaid, checkWeekTodo    []byte
	var updateGauge, updateRaid, settingRaw       []byte
	err = s.db.QueryRowContext(ctx,
		`SELECT show_day_todo, show_raid, show_week_todo,
		        check_day_todo, check_raid, check_week_todo,
		        update_gauge, update_raid, setting
		 FROM friends WHERE friends_id = ?`, req.ID,
	).Scan(
		&showDayTodo, &showRaid, &showWeekTodo,
		&checkDayTodo, &checkRaid, &checkWeekTodo,
		&updateGauge, &updateRaid, &settingRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("설정 조회 실패: %w", err)
	}

	updated := FriendSettings{
		ShowDayTodo:   bitToBool(showDayTodo),
		ShowRaid:      bitToBool(showRaid),
		ShowWeekTodo:  bitToBool(showWeekTodo),
		CheckDayTodo:  bitToBool(checkDayTodo),
		CheckRaid:     bitToBool(checkRaid),
		CheckWeekTodo: bitToBool(checkWeekTodo),
		UpdateGauge:   bitToBool(updateGauge),
		UpdateRaid:    bitToBool(updateRaid),
		Setting:       bitToBool(settingRaw),
	}

	return &updated, nil
}

// DeleteFriend deletes both sides of a friendship.
// Java: friendsService.deleteById(member, friendId) deletes both directions.
func (s *FriendService) DeleteFriend(ctx context.Context, username string, friendID int64) error {
	// 1. Get member ID
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// 2. Find the friend record to get from_member
	var fromMember int64
	err = s.db.QueryRowContext(ctx,
		"SELECT from_member FROM friends WHERE friends_id = ? AND member_id = ?",
		friendID, memberID,
	).Scan(&fromMember)
	if err != nil {
		return fmt.Errorf("친구 관계를 찾을 수 없습니다: %w", err)
	}

	// 3. Delete both directions in a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"DELETE FROM friends WHERE member_id = ? AND from_member = ?",
		memberID, fromMember,
	)
	if err != nil {
		return fmt.Errorf("친구 삭제 실패 (내 레코드): %w", err)
	}

	_, err = tx.ExecContext(ctx,
		"DELETE FROM friends WHERE member_id = ? AND from_member = ?",
		fromMember, memberID,
	)
	if err != nil {
		return fmt.Errorf("친구 삭제 실패 (상대 레코드): %w", err)
	}

	return tx.Commit()
}

// UpdateSortOrder updates the ordering of friends.
// Java: friendsService.updateSort(member, friendIdList)
func (s *FriendService) UpdateSortOrder(ctx context.Context, username string, req FriendSortRequest) error {
	// 1. Get member ID
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("회원을 찾을 수 없습니다: %w", err)
	}

	// 2. Update ordering in a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	for i, friendID := range req.FriendIDList {
		result, err := tx.ExecContext(ctx,
			"UPDATE friends SET ordering = ?, last_modified_date = ? WHERE friends_id = ? AND member_id = ?",
			i+1, now, friendID, memberID,
		)
		if err != nil {
			return fmt.Errorf("정렬 순서 업데이트 실패: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return fmt.Errorf("친구 ID %d를 찾을 수 없거나 권한이 없습니다", friendID)
		}
	}

	return tx.Commit()
}
