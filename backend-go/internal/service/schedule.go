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
	ScheduleID            int64  `json:"scheduleId"`
	ScheduleCategory      string `json:"scheduleCategory"`
	ScheduleRaidCategory  string `json:"scheduleRaidCategory"`
	RaidName              string `json:"raidName"`
	DayOfWeek             int    `json:"dayOfWeek"`
	TimeSlot              string `json:"timeSlot"`
	Memo                  string `json:"memo"`
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
	Name       string   `json:"name"`
	RaidNames  []string `json:"raidNames"`
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
	query := `SELECT schedule_id, schedule_category, schedule_raid_category,
	                 COALESCE(raid_name, ''), COALESCE(day_of_week, 0), COALESCE(time, ''), COALESCE(memo, '')
	          FROM schedule WHERE character_id IN (`
	args := make([]interface{}, 0, len(charIDs)+2)
	for i, id := range charIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ")"

	// Optional date range filter
	if startDate != "" && endDate != "" {
		query += " AND date >= ? AND date <= ?"
		args = append(args, startDate, endDate)
	}

	query += " ORDER BY day_of_week ASC, time ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query schedules: %w", err)
	}
	defer rows.Close()

	var schedules []ScheduleResponse
	for rows.Next() {
		var sr ScheduleResponse
		if err := rows.Scan(&sr.ScheduleID, &sr.ScheduleCategory, &sr.ScheduleRaidCategory,
			&sr.RaidName, &sr.DayOfWeek, &sr.TimeSlot, &sr.Memo); err != nil {
			return nil, err
		}
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
		ScheduleID:           scheduleID,
		ScheduleCategory:     req.ScheduleCategory,
		ScheduleRaidCategory: req.ScheduleRaidCategory,
		RaidName:             req.RaidName,
		DayOfWeek:            req.DayOfWeek,
		TimeSlot:             req.TimeSlot,
		Memo:                 req.Memo,
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
		`SELECT COALESCE(week_category, '') AS wc, COALESCE(name, '') AS n
		 FROM content
		 WHERE dtype = 'WeekContent'
		 ORDER BY week_category, level DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query raid categories: %w", err)
	}
	defer rows.Close()

	// Group by week_category, preserving order
	categoryMap := make(map[string]*RaidCategoryResponse)
	var categoryOrder []string

	for rows.Next() {
		var weekCategory, raidName string
		if err := rows.Scan(&weekCategory, &raidName); err != nil {
			return nil, err
		}
		if weekCategory == "" {
			continue
		}

		existing, ok := categoryMap[weekCategory]
		if !ok {
			resp := &RaidCategoryResponse{
				Name:      weekCategory,
				RaidNames: []string{},
			}
			categoryMap[weekCategory] = resp
			categoryOrder = append(categoryOrder, weekCategory)
			existing = resp
		}

		// Avoid duplicate raid names within the same category
		found := false
		for _, n := range existing.RaidNames {
			if n == raidName {
				found = true
				break
			}
		}
		if !found && raidName != "" {
			existing.RaidNames = append(existing.RaidNames, raidName)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]RaidCategoryResponse, 0, len(categoryOrder))
	for _, cat := range categoryOrder {
		result = append(result, *categoryMap[cat])
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
	query := `SELECT schedule_id, schedule_category, schedule_raid_category,
	                 COALESCE(raid_name, ''), COALESCE(day_of_week, 0), COALESCE(time, ''), COALESCE(memo, '')
	          FROM schedule WHERE character_id IN (`
	args := make([]interface{}, 0, len(charIDs))
	for i, id := range charIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ") ORDER BY day_of_week ASC, time ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query week schedules: %w", err)
	}
	defer rows.Close()

	// Group schedules by day of week
	dayMap := make(map[int][]ScheduleResponse)
	for rows.Next() {
		var sr ScheduleResponse
		if err := rows.Scan(&sr.ScheduleID, &sr.ScheduleCategory, &sr.ScheduleRaidCategory,
			&sr.RaidName, &sr.DayOfWeek, &sr.TimeSlot, &sr.Memo); err != nil {
			return nil, err
		}
		dayMap[sr.DayOfWeek] = append(dayMap[sr.DayOfWeek], sr)
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
