package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CharacterWeekService handles weekly content business logic.
type CharacterWeekService struct {
	db *sql.DB
}

// NewCharacterWeekService creates a new CharacterWeekService.
func NewCharacterWeekService(db *sql.DB) *CharacterWeekService {
	return &CharacterWeekService{db: db}
}

// WeekRaidRequest is the request for adding/removing week raids.
type WeekRaidRequest struct {
	CharacterID       int64   `json:"characterId"`
	WeekContentIDList []int64 `json:"weekContentIdList"`
}

// WeekRaidBusRequest is the request for updating bus gold.
type WeekRaidBusRequest struct {
	CharacterID  int64   `json:"characterId"`
	WeekCategory string  `json:"weekCategory"`
	BusGold      float64 `json:"busGold"`
}

// WeekRaidCheckRequest is the request for checking a raid.
type WeekRaidCheckRequest struct {
	CharacterID  int64  `json:"characterId"`
	WeekCategory string `json:"weekCategory"`
}

// WeekRaidMessageRequest is the request for updating a raid message.
type WeekRaidMessageRequest struct {
	CharacterID  int64  `json:"characterId"`
	WeekCategory string `json:"weekCategory"`
	Message      string `json:"message"`
}

// WeekRaidSortRequest is the request for sorting raids.
type WeekRaidSortRequest struct {
	CharacterID     int64              `json:"characterId"`
	SortRequestList []RaidSortEntry    `json:"sortRequestList"`
}

// RaidSortEntry is a single sort entry for a raid.
type RaidSortEntry struct {
	TodoID     int64 `json:"todoId"`
	SortNumber int   `json:"sortNumber"`
}

// WeekEponaRequest is the request for checking weekly epona.
type WeekEponaRequest struct {
	CharacterID int64 `json:"characterId"`
	AllCheck    bool  `json:"allCheck"`
}

// WeekSilmaelRequest is the request for toggling silmael.
type WeekSilmaelRequest struct {
	CharacterID int64 `json:"characterId"`
}

// WeekCubeRequest is the request for updating cube ticket.
type WeekCubeRequest struct {
	CharacterID int64 `json:"characterId"`
	Num         int   `json:"num"`
}

// WeekGoldCheckRequest is the request for toggling gold check.
type WeekGoldCheckRequest struct {
	CharacterID  int64  `json:"characterId"`
	WeekCategory string `json:"weekCategory"`
	UpdateValue  bool   `json:"updateValue"`
}

// WeekGoldCheckVersionRequest is the request for toggling gold check version.
type WeekGoldCheckVersionRequest struct {
	CharacterID int64 `json:"characterId"`
}

// WeekMoreRewardRequest is the request for toggling more reward.
type WeekMoreRewardRequest struct {
	CharacterID  int64  `json:"characterId"`
	WeekCategory string `json:"weekCategory"`
}

// WeekElysianRequest is the request for updating elysian count.
type WeekElysianRequest struct {
	CharacterID int64  `json:"characterId"`
	Action      string `json:"action"`
}

// WeekElysianAllRequest is the request for toggling all elysian.
type WeekElysianAllRequest struct {
	CharacterID int64 `json:"characterId"`
}

func (s *CharacterWeekService) verifyCharacterOwnership(ctx context.Context, username string, characterID int64) (int64, error) {
	var memberID, charMemberID int64
	err := s.db.QueryRowContext(ctx,
		"SELECT id FROM member WHERE username = ?", username,
	).Scan(&memberID)
	if err != nil {
		return 0, fmt.Errorf("querying member: %w", err)
	}

	err = s.db.QueryRowContext(ctx,
		"SELECT member_id FROM character_entity WHERE id = ?", characterID,
	).Scan(&charMemberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("캐릭터를 찾을 수 없습니다.")
		}
		return 0, fmt.Errorf("querying character: %w", err)
	}

	if memberID != charMemberID {
		return 0, errors.New("권한이 없습니다.")
	}
	return memberID, nil
}

// AddRemoveWeekRaid adds or removes week raids for a character.
func (s *CharacterWeekService) AddRemoveWeekRaid(ctx context.Context, username string, req WeekRaidRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove existing raids for this character
	_, err = tx.ExecContext(ctx,
		"DELETE FROM todo_v2 WHERE character_id = ?", req.CharacterID,
	)
	if err != nil {
		return fmt.Errorf("deleting existing raids: %w", err)
	}

	// Add new raids from week content list
	now := time.Now()
	for i, contentID := range req.WeekContentIDList {
		var weekCategory, weekContentName string
		var weekContentGate int
		var gold, itemLevel float64

		err := s.db.QueryRowContext(ctx,
			`SELECT week_category, week_content_name, COALESCE(week_content_gate, 0),
			        COALESCE(gold, 0), COALESCE(item_level, 0)
			 FROM week_content WHERE id = ?`, contentID,
		).Scan(&weekCategory, &weekContentName, &weekContentGate, &gold, &itemLevel)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return fmt.Errorf("querying week content %d: %w", contentID, err)
		}

		// Count total gates for this category
		var totalGate int
		err = s.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM week_content WHERE week_category = ?", weekCategory,
		).Scan(&totalGate)
		if err != nil {
			totalGate = 1
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO todo_v2 (character_id, week_category, week_content_name,
			 week_content_gate, gold, gold_check, is_checked, sort_number,
			 current_gate, total_gate, more_reward, more_reward_gold,
			 created_date, last_modified_date)
			 VALUES (?, ?, ?, ?, ?, false, false, ?, ?, ?, false, 0, ?, ?)`,
			req.CharacterID, weekCategory, weekContentName,
			weekContentGate, gold, i,
			weekContentGate, totalGate,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("inserting todo_v2: %w", err)
		}
	}

	return tx.Commit()
}

// GetWeekRaidForm returns the available raid form for a character.
func (s *CharacterWeekService) GetWeekRaidForm(ctx context.Context, username string, characterID int64) ([]WeekRaidFormResponse, error) {
	_, err := s.verifyCharacterOwnership(ctx, username, characterID)
	if err != nil {
		return nil, err
	}

	// Get character item level
	var itemLevel float64
	err = s.db.QueryRowContext(ctx,
		"SELECT COALESCE(item_level, 0) FROM character_entity WHERE id = ?", characterID,
	).Scan(&itemLevel)
	if err != nil {
		return nil, fmt.Errorf("querying character item level: %w", err)
	}

	// Get all week content that matches the character's item level
	rows, err := s.db.QueryContext(ctx,
		`SELECT wc.id, wc.week_category, wc.week_content_name,
		        COALESCE(wc.week_content_gate, 0), COALESCE(wc.gold, 0),
		        COALESCE(wc.item_level, 0)
		 FROM week_content wc
		 WHERE wc.item_level <= ?
		 ORDER BY wc.item_level DESC, wc.week_category, wc.week_content_gate`, itemLevel,
	)
	if err != nil {
		return nil, fmt.Errorf("querying week content: %w", err)
	}
	defer rows.Close()

	// Get existing selected raids
	selectedMap := make(map[string]bool)
	selectedRows, err := s.db.QueryContext(ctx,
		"SELECT week_category FROM todo_v2 WHERE character_id = ?", characterID,
	)
	if err == nil {
		defer selectedRows.Close()
		for selectedRows.Next() {
			var cat string
			if err := selectedRows.Scan(&cat); err == nil {
				selectedMap[cat] = true
			}
		}
	}

	var forms []WeekRaidFormResponse
	for rows.Next() {
		var f WeekRaidFormResponse
		err := rows.Scan(&f.ID, &f.WeekCategory, &f.WeekContentName,
			&f.WeekContentGate, &f.Gold, &f.ItemLevel)
		if err != nil {
			return nil, fmt.Errorf("scanning week content: %w", err)
		}
		f.Selected = selectedMap[f.WeekCategory]
		forms = append(forms, f)
	}

	if forms == nil {
		forms = []WeekRaidFormResponse{}
	}
	return forms, rows.Err()
}

// UpdateBusGold updates the bus gold for a raid.
func (s *CharacterWeekService) UpdateBusGold(ctx context.Context, username string, req WeekRaidBusRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	// Upsert raid_bus_gold
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO raid_bus_gold (character_id, week_category, bus_gold)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE bus_gold = ?`,
		req.CharacterID, req.WeekCategory, req.BusGold, req.BusGold,
	)
	return err
}

// CheckRaid toggles the check status of a raid.
func (s *CharacterWeekService) CheckRaid(ctx context.Context, username string, req WeekRaidCheckRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE todo_v2 SET is_checked = NOT is_checked, last_modified_date = ?
		 WHERE character_id = ? AND week_category = ?`,
		time.Now(), req.CharacterID, req.WeekCategory,
	)
	return err
}

// UpdateRaidMessage updates the message for a raid.
func (s *CharacterWeekService) UpdateRaidMessage(ctx context.Context, username string, req WeekRaidMessageRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE todo_v2 SET message = ?, last_modified_date = ?
		 WHERE character_id = ? AND week_category = ?`,
		req.Message, time.Now(), req.CharacterID, req.WeekCategory,
	)
	return err
}

// SortRaids updates the sort order of raids.
func (s *CharacterWeekService) SortRaids(ctx context.Context, username string, req WeekRaidSortRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"UPDATE todo_v2 SET sort_number = ?, last_modified_date = ? WHERE id = ? AND character_id = ?",
	)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, entry := range req.SortRequestList {
		_, err := stmt.ExecContext(ctx, entry.SortNumber, now, entry.TodoID, req.CharacterID)
		if err != nil {
			return fmt.Errorf("updating sort for todo %d: %w", entry.TodoID, err)
		}
	}

	return tx.Commit()
}

// CheckWeekEpona handles weekly epona check/uncheck.
func (s *CharacterWeekService) CheckWeekEpona(ctx context.Context, username string, req WeekEponaRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	if req.AllCheck {
		// Toggle: if all checked (3), reset to 0; otherwise set to 3
		var current int
		err := s.db.QueryRowContext(ctx,
			"SELECT COALESCE(week_epona, 0) FROM character_entity WHERE id = ?", req.CharacterID,
		).Scan(&current)
		if err != nil {
			return fmt.Errorf("querying week_epona: %w", err)
		}

		newVal := 3
		if current >= 3 {
			newVal = 0
		}

		_, err = s.db.ExecContext(ctx,
			"UPDATE character_entity SET week_epona = ?, week_epona_check = ?, last_modified_date = ? WHERE id = ?",
			newVal, newVal >= 3, time.Now(), req.CharacterID,
		)
		return err
	}

	// Increment by 1, wrap around at 3
	_, err = s.db.ExecContext(ctx,
		`UPDATE character_entity
		 SET week_epona = CASE WHEN COALESCE(week_epona, 0) >= 3 THEN 0 ELSE COALESCE(week_epona, 0) + 1 END,
		     week_epona_check = CASE WHEN COALESCE(week_epona, 0) >= 2 THEN true ELSE false END,
		     last_modified_date = ?
		 WHERE id = ?`,
		time.Now(), req.CharacterID,
	)
	return err
}

// ToggleSilmael toggles the silmael exchange flag.
func (s *CharacterWeekService) ToggleSilmael(ctx context.Context, username string, req WeekSilmaelRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE character_entity
		 SET silmael_change = NOT COALESCE(silmael_change, false),
		     last_modified_date = ?
		 WHERE id = ?`,
		time.Now(), req.CharacterID,
	)
	return err
}

// UpdateCubeTicket updates the cube ticket count.
func (s *CharacterWeekService) UpdateCubeTicket(ctx context.Context, username string, req WeekCubeRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE character_entity SET cube_ticket = ?, last_modified_date = ? WHERE id = ?",
		req.Num, time.Now(), req.CharacterID,
	)
	return err
}

// ToggleGoldCheck toggles the gold check for a specific raid.
func (s *CharacterWeekService) ToggleGoldCheck(ctx context.Context, username string, req WeekGoldCheckRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE todo_v2 SET gold_check = ?, last_modified_date = ?
		 WHERE character_id = ? AND week_category = ?`,
		req.UpdateValue, time.Now(), req.CharacterID, req.WeekCategory,
	)
	return err
}

// ToggleGoldCheckVersion toggles the gold check version for a character.
func (s *CharacterWeekService) ToggleGoldCheckVersion(ctx context.Context, username string, req WeekGoldCheckVersionRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE character_entity
		 SET gold_check_version = CASE WHEN COALESCE(gold_check_version, 0) = 0 THEN 1 ELSE 0 END,
		     last_modified_date = ?
		 WHERE id = ?`,
		time.Now(), req.CharacterID,
	)
	return err
}

// ToggleMoreReward toggles the more reward flag for a raid.
func (s *CharacterWeekService) ToggleMoreReward(ctx context.Context, username string, req WeekMoreRewardRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE todo_v2 SET more_reward = NOT COALESCE(more_reward, false), last_modified_date = ?
		 WHERE character_id = ? AND week_category = ?`,
		time.Now(), req.CharacterID, req.WeekCategory,
	)
	return err
}

// UpdateElysian updates the elysian count (increment or decrement).
func (s *CharacterWeekService) UpdateElysian(ctx context.Context, username string, req WeekElysianRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	switch req.Action {
	case "increment":
		_, err = s.db.ExecContext(ctx,
			`UPDATE character_entity
			 SET elysian_count = COALESCE(elysian_count, 0) + 1,
			     last_modified_date = ?
			 WHERE id = ?`,
			time.Now(), req.CharacterID,
		)
	case "decrement":
		_, err = s.db.ExecContext(ctx,
			`UPDATE character_entity
			 SET elysian_count = CASE WHEN COALESCE(elysian_count, 0) > 0 THEN elysian_count - 1 ELSE 0 END,
			     last_modified_date = ?
			 WHERE id = ?`,
			time.Now(), req.CharacterID,
		)
	default:
		return fmt.Errorf("알 수 없는 액션: %s", req.Action)
	}
	return err
}

// ToggleElysianAll toggles all elysian for a character.
func (s *CharacterWeekService) ToggleElysianAll(ctx context.Context, username string, req WeekElysianAllRequest) error {
	_, err := s.verifyCharacterOwnership(ctx, username, req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE character_entity
		 SET elysian_all_check = NOT COALESCE(elysian_all_check, false),
		     last_modified_date = ?
		 WHERE id = ?`,
		time.Now(), req.CharacterID,
	)
	return err
}
