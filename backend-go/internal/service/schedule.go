package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ScheduleService struct {
	db *sql.DB
}

func NewScheduleService(db *sql.DB) *ScheduleService {
	return &ScheduleService{db: db}
}

type ScheduleResponse struct {
	ScheduleID           int64    `json:"scheduleId"`
	ScheduleCategory     string   `json:"scheduleCategory"`
	ScheduleRaidCategory string   `json:"scheduleRaidCategory"`
	RaidName             string   `json:"raidName"`
	DayOfWeek            string   `json:"dayOfWeek"`
	RepeatWeek           bool     `json:"repeatWeek"`
	Date                 *string  `json:"date"`
	Time                 string   `json:"time"`
	Memo                 string   `json:"memo"`
	IsLeader             bool     `json:"isLeader"`
	LeaderScheduleId     int64    `json:"leaderScheduleId"`
	CharacterName        string   `json:"characterName"`
	LeaderCharacterName  string   `json:"leaderCharacterName"`
	FriendCharacterNames []string `json:"friendCharacterNames"`
	AutoCheck            bool     `json:"autoCheck"`
}

type CreateScheduleRequest struct {
	ScheduleCategory     string `json:"scheduleCategory"`
	ScheduleRaidCategory string `json:"scheduleRaidCategory"`
	RaidName             string `json:"raidName"`
	DayOfWeek            int    `json:"dayOfWeek"`
	TimeSlot             string `json:"timeSlot"`
	Memo                 string `json:"memo"`
}

type UpdateScheduleRequest struct {
	ScheduleCategory     string `json:"scheduleCategory"`
	ScheduleRaidCategory string `json:"scheduleRaidCategory"`
	RaidName             string `json:"raidName"`
	DayOfWeek            int    `json:"dayOfWeek"`
	TimeSlot             string `json:"timeSlot"`
	Memo                 string `json:"memo"`
}

type RaidCategoryResponse struct {
	CategoryID          int64   `json:"categoryId"`
	Name                string  `json:"name"`
	WeekContentCategory string  `json:"weekContentCategory"`
	Level               float64 `json:"level"`
}

type RaidCategoryGroupResponse struct {
	Name      string   `json:"name"`
	RaidNames []string `json:"raidNames"`
}

type ScheduleCharacter struct {
	CharacterID   int64  `json:"characterId"`
	CharacterName string `json:"characterName"`
	ServerName    string `json:"serverName"`
	ItemLevel     float64 `json:"itemLevel"`
}

type WeekScheduleResponse struct {
	DayOfWeek int               `json:"dayOfWeek"`
	Schedules []ScheduleResponse `json:"schedules"`
}

// dayOfWeekIntToString converts day of week integer to Spring's DayOfWeek enum string
func dayOfWeekIntToString(day int) string {
	days := []string{"SUNDAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY"}
	if day >= 0 && day < 7 {
		return days[day]
	}
	return ""
}

// getMemberID resolves the member_id for a given username.
func (s *ScheduleService) getMemberID(ctx context.Context, username string) (int64, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx,
		`SELECT member_id FROM member WHERE username = ?`, username,
	).Scan(&memberID)
	if err != nil {
		return 0, fmt.Errorf("member not found: %w", err)
	}
	return memberID, nil
}

// getCharacterIDs returns all character IDs belonging to a member.
func (s *ScheduleService) getCharacterIDs(ctx context.Context, memberID int64) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT characters_id FROM characters WHERE member_id = ? AND is_deleted = false ORDER BY sort_number ASC`,
		memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *ScheduleService) GetSchedules(ctx context.Context, username, startDate, endDate string) ([]ScheduleResponse, error) {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return nil, err
	}

	charIDs, err := s.getCharacterIDs(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("get character ids: %w", err)
	}
	if len(charIDs) == 0 {
		return []ScheduleResponse{}, nil
	}

	// Build the IN clause
	query := `SELECT s.schedule_id, s.schedule_category, s.schedule_raid_category,
	                 COALESCE(s.raid_name, ''), COALESCE(s.day_of_week, 0),
	                 IFNULL(s.repeat_week, 0), s.date, COALESCE(s.time, ''), COALESCE(s.memo, ''),
	                 IFNULL(s.leader, 0), COALESCE(s.leader_schedule_id, 0),
	                 COALESCE(c.character_name, ''), IFNULL(s.auto_check, 0)
	          FROM schedule s
	          LEFT JOIN characters c ON s.character_id = c.characters_id
	          WHERE s.character_id IN (`
	args := make([]interface{}, 0, len(charIDs)+2)
	for i, id := range charIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ")"

	// Optional date range filter (for month view)
	if startDate != "" && endDate != "" {
		query += " AND (s.repeat_week = 1 OR (s.date >= ? AND s.date <= ?))"
		args = append(args, startDate, endDate)
	}

	query += " ORDER BY s.day_of_week ASC, s.time ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query schedules: %w", err)
	}
	defer rows.Close()

	var schedules []ScheduleResponse
	for rows.Next() {
		var sr ScheduleResponse
		var dayOfWeekInt int
		var repeatWeekInt int
		var leaderInt int
		var autoCheckInt int
		var dateStr sql.NullString

		if err := rows.Scan(&sr.ScheduleID, &sr.ScheduleCategory, &sr.ScheduleRaidCategory,
			&sr.RaidName, &dayOfWeekInt, &repeatWeekInt, &dateStr, &sr.Time, &sr.Memo,
			&leaderInt, &sr.LeaderScheduleId, &sr.CharacterName, &autoCheckInt); err != nil {
			return nil, err
		}

		// Convert day of week int to string enum
		sr.DayOfWeek = dayOfWeekIntToString(dayOfWeekInt)

		// Convert BIT fields (now as int due to IFNULL) to bool
		sr.RepeatWeek = repeatWeekInt == 1
		sr.IsLeader = leaderInt == 1
		sr.AutoCheck = autoCheckInt == 1

		// Handle date
		if dateStr.Valid {
			sr.Date = &dateStr.String
		}

		// Set leader character name (same as character name if leader)
		if sr.IsLeader {
			sr.LeaderCharacterName = sr.CharacterName
		}

		// Initialize friendCharacterNames as empty array (required by frontend)
		sr.FriendCharacterNames = []string{}

		schedules = append(schedules, sr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if schedules == nil {
		schedules = []ScheduleResponse{}
	}
	return schedules, nil
}

func (s *ScheduleService) CreateSchedule(ctx context.Context, username string, req CreateScheduleRequest) (*ScheduleResponse, error) {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return nil, err
	}

	// Get the first character of the member for the schedule
	var characterID int64
	err = s.db.QueryRowContext(ctx,
		`SELECT characters_id FROM characters WHERE member_id = ? AND is_deleted = false ORDER BY sort_number ASC LIMIT 1`,
		memberID,
	).Scan(&characterID)
	if err != nil {
		return nil, fmt.Errorf("no characters found for member: %w", err)
	}

	now := time.Now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO schedule (character_id, schedule_raid_category, schedule_category,
		                       raid_name, day_of_week, time, memo,
		                       repeat_week, leader, leader_schedule_id, checked,
		                       auto_check, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, false, false, 0, false, false, ?, ?)`,
		characterID, req.ScheduleRaidCategory, req.ScheduleCategory,
		req.RaidName, req.DayOfWeek, req.TimeSlot, req.Memo,
		now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert schedule: %w", err)
	}
	scheduleID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &ScheduleResponse{
		ScheduleID:            scheduleID,
		ScheduleCategory:      req.ScheduleCategory,
		ScheduleRaidCategory:  req.ScheduleRaidCategory,
		RaidName:              req.RaidName,
		DayOfWeek:             dayOfWeekIntToString(req.DayOfWeek),
		Time:                  req.TimeSlot,
		Memo:                  req.Memo,
		FriendCharacterNames:  []string{},
	}, nil
}

func (s *ScheduleService) UpdateSchedule(ctx context.Context, username string, scheduleID int64, req UpdateScheduleRequest) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	charIDs, err := s.getCharacterIDs(ctx, memberID)
	if err != nil {
		return fmt.Errorf("get character ids: %w", err)
	}
	if len(charIDs) == 0 {
		return fmt.Errorf("no characters found for member")
	}

	// Build IN clause for ownership check
	query := `UPDATE schedule SET schedule_raid_category = ?, schedule_category = ?,
	                              raid_name = ?, day_of_week = ?, time = ?, memo = ?,
	                              last_modified_date = ?
	          WHERE schedule_id = ? AND character_id IN (`
	args := []interface{}{
		req.ScheduleRaidCategory, req.ScheduleCategory,
		req.RaidName, req.DayOfWeek, req.TimeSlot, req.Memo,
		time.Now(), scheduleID,
	}
	for i, id := range charIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ")"

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update schedule: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("schedule not found or not owned by user")
	}
	return nil
}

func (s *ScheduleService) DeleteSchedule(ctx context.Context, username string, scheduleID int64) error {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return err
	}

	charIDs, err := s.getCharacterIDs(ctx, memberID)
	if err != nil {
		return fmt.Errorf("get character ids: %w", err)
	}
	if len(charIDs) == 0 {
		return fmt.Errorf("no characters found for member")
	}

	// Build IN clause for ownership check
	query := `DELETE FROM schedule WHERE schedule_id = ? AND character_id IN (`
	args := []interface{}{scheduleID}
	for i, id := range charIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ")"

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("schedule not found or not owned by user")
	}
	return nil
}

func (s *ScheduleService) GetRaidCategories(ctx context.Context) ([]RaidCategoryResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT content_id, COALESCE(name, '') AS name,
		        COALESCE(week_content_category, '') AS week_content_category,
		        COALESCE(level, 0) AS level
		 FROM content
		 WHERE dtype = 'WeekContent'
		   AND week_content_category IS NOT NULL
		   AND week_content_category != ''
		   AND gate = 1
		 ORDER BY level DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query raid categories: %w", err)
	}
	defer rows.Close()

	var result []RaidCategoryResponse
	for rows.Next() {
		var rc RaidCategoryResponse
		if err := rows.Scan(&rc.CategoryID, &rc.Name, &rc.WeekContentCategory, &rc.Level); err != nil {
			return nil, err
		}
		result = append(result, rc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if result == nil {
		result = []RaidCategoryResponse{}
	}
	return result, nil
}

func (s *ScheduleService) GetCharacters(ctx context.Context, username string) ([]ScheduleCharacter, error) {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT characters_id, character_name, server_name, item_level
		 FROM characters
		 WHERE member_id = ? AND is_deleted = false
		 ORDER BY sort_number ASC, item_level DESC`,
		memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("query characters: %w", err)
	}
	defer rows.Close()

	var characters []ScheduleCharacter
	for rows.Next() {
		var c ScheduleCharacter
		if err := rows.Scan(&c.CharacterID, &c.CharacterName, &c.ServerName, &c.ItemLevel); err != nil {
			return nil, err
		}
		characters = append(characters, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if characters == nil {
		characters = []ScheduleCharacter{}
	}
	return characters, nil
}

func (s *ScheduleService) GetWeekSchedule(ctx context.Context, username string) ([]WeekScheduleResponse, error) {
	memberID, err := s.getMemberID(ctx, username)
	if err != nil {
		return nil, err
	}

	charIDs, err := s.getCharacterIDs(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("get character ids: %w", err)
	}
	if len(charIDs) == 0 {
		// Return empty week with all 7 days
		result := make([]WeekScheduleResponse, 7)
		for i := 0; i < 7; i++ {
			result[i] = WeekScheduleResponse{DayOfWeek: i, Schedules: []ScheduleResponse{}}
		}
		return result, nil
	}

	// Build the IN clause
	query := `SELECT s.schedule_id, s.schedule_category, s.schedule_raid_category,
	                 COALESCE(s.raid_name, ''), COALESCE(s.day_of_week, 0),
	                 IFNULL(s.repeat_week, 0), s.date, COALESCE(s.time, ''), COALESCE(s.memo, ''),
	                 IFNULL(s.leader, 0), COALESCE(s.leader_schedule_id, 0),
	                 COALESCE(c.character_name, ''), IFNULL(s.auto_check, 0)
	          FROM schedule s
	          LEFT JOIN characters c ON s.character_id = c.characters_id
	          WHERE s.character_id IN (`
	args := make([]interface{}, 0, len(charIDs))
	for i, id := range charIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ") ORDER BY s.day_of_week ASC, s.time ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query week schedules: %w", err)
	}
	defer rows.Close()

	// Group schedules by day of week
	dayMap := make(map[int][]ScheduleResponse)
	for rows.Next() {
		var sr ScheduleResponse
		var dayOfWeekInt int
		var repeatWeekInt int
		var leaderInt int
		var autoCheckInt int
		var dateStr sql.NullString

		if err := rows.Scan(&sr.ScheduleID, &sr.ScheduleCategory, &sr.ScheduleRaidCategory,
			&sr.RaidName, &dayOfWeekInt, &repeatWeekInt, &dateStr, &sr.Time, &sr.Memo,
			&leaderInt, &sr.LeaderScheduleId, &sr.CharacterName, &autoCheckInt); err != nil {
			return nil, err
		}

		// Convert day of week int to string enum
		sr.DayOfWeek = dayOfWeekIntToString(dayOfWeekInt)

		// Convert BIT fields (now as int due to IFNULL) to bool
		sr.RepeatWeek = repeatWeekInt == 1
		sr.IsLeader = leaderInt == 1
		sr.AutoCheck = autoCheckInt == 1

		// Handle date
		if dateStr.Valid {
			sr.Date = &dateStr.String
		}

		// Set leader character name
		if sr.IsLeader {
			sr.LeaderCharacterName = sr.CharacterName
		}

		// Initialize friendCharacterNames as empty array
		sr.FriendCharacterNames = []string{}

		dayMap[dayOfWeekInt] = append(dayMap[dayOfWeekInt], sr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build response for all 7 days (0 = Sunday .. 6 = Saturday)
	result := make([]WeekScheduleResponse, 7)
	for i := 0; i < 7; i++ {
		schedules := dayMap[i]
		if schedules == nil {
			schedules = []ScheduleResponse{}
		}
		result[i] = WeekScheduleResponse{
			DayOfWeek: i,
			Schedules: schedules,
		}
	}

	return result, nil
}

func (s *ScheduleService) EditFriend(ctx context.Context, username string, scheduleID int64, friendUsername string) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("querying member: %w", err)
	}
	// Verify schedule ownership
	var ownerID int64
	err = s.db.QueryRowContext(ctx, "SELECT member_id FROM schedule_raid WHERE schedule_id = ?", scheduleID).Scan(&ownerID)
	if err != nil {
		return fmt.Errorf("schedule not found")
	}
	if ownerID != memberID {
		return fmt.Errorf("권한이 없습니다.")
	}

	var friendID int64
	err = s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", friendUsername).Scan(&friendID)
	if err != nil {
		return fmt.Errorf("친구를 찾을 수 없습니다.")
	}

	// Toggle friend in schedule (check if exists, add/remove)
	var existingID int64
	err = s.db.QueryRowContext(ctx,
		"SELECT id FROM schedule_raid_friend WHERE schedule_raid_id = ? AND member_id = ?",
		scheduleID, friendID,
	).Scan(&existingID)

	if err == sql.ErrNoRows {
		_, err = s.db.ExecContext(ctx,
			"INSERT INTO schedule_raid_friend (schedule_raid_id, member_id) VALUES (?, ?)",
			scheduleID, friendID,
		)
		return err
	}
	_, err = s.db.ExecContext(ctx, "DELETE FROM schedule_raid_friend WHERE id = ?", existingID)
	return err
}
