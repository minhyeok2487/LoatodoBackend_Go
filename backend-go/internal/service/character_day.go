package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CharacterDayService handles daily content business logic.
type CharacterDayService struct {
	db *sql.DB
}

// NewCharacterDayService creates a new CharacterDayService.
func NewCharacterDayService(db *sql.DB) *CharacterDayService {
	return &CharacterDayService{db: db}
}

// DayTodoResult is the internal response for day todo operations.
type DayTodoResult struct {
	ChaosCheck    int     `json:"chaosCheck"`
	ChaosGauge    int     `json:"chaosGauge"`
	ChaosGold     float64 `json:"chaosGold"`
	GuardianCheck int     `json:"guardianCheck"`
	GuardianGauge int     `json:"guardianGauge"`
	GuardianGold  float64 `json:"guardianGold"`
	EponaCheck    int     `json:"eponaCheck"`
	EponaGauge    int     `json:"eponaGauge"`
	WeekTotalGold float64 `json:"weekDayTodoGold"`
}

// DayCheckRequest is the request for checking daily content.
type DayCheckRequest struct {
	CharacterID int64  `json:"characterId"`
	Category    string `json:"category"`
}

// DayGaugeRequest is the request for updating rest gauge.
type DayGaugeRequest struct {
	CharacterID  int64 `json:"characterId"`
	ChaosGauge   int   `json:"chaosGauge"`
	GuardianGauge int  `json:"guardianGauge"`
}

// DayCheckAllRequest is the request for checking all daily content for one character.
type DayCheckAllRequest struct {
	CharacterID int64 `json:"characterId"`
}

// DayCheckAllCharactersRequest is the request for checking all daily for all characters.
type DayCheckAllCharactersRequest struct {
	ServerName string `json:"serverName"`
}

// dayTodoEntity holds the day todo columns from the DB.
type dayTodoEntity struct {
	CharacterID        int64
	MemberID           int64
	ChaosCheck         int
	ChaosGauge         int
	BeforeChaosGauge   int
	ChaosGold          float64
	GuardianCheck      int
	GuardianGauge      int
	BeforeGuardianGauge int
	GuardianGold       float64
	EponaCheck         int
	EponaGauge         int
	BeforeEponaGauge   int
	WeekTotalGold      float64
}

func (s *CharacterDayService) loadDayTodo(ctx context.Context, characterID int64) (*dayTodoEntity, error) {
	var d dayTodoEntity
	err := s.db.QueryRowContext(ctx,
		`SELECT characters_id, member_id,
		        COALESCE(chaos_check, 0), COALESCE(chaos_gauge, 0),
		        COALESCE(before_chaos_gauge, 0), COALESCE(chaos_gold, 0),
		        COALESCE(guardian_check, 0), COALESCE(guardian_gauge, 0),
		        COALESCE(before_guardian_gauge, 0), COALESCE(guardian_gold, 0),
		        COALESCE(epona_check2, 0), COALESCE(epona_gauge, 0),
		        COALESCE(before_epona_gauge, 0),
		        COALESCE(week_total_gold, 0)
		 FROM characters WHERE characters_id = ?`, characterID,
	).Scan(
		&d.CharacterID, &d.MemberID,
		&d.ChaosCheck, &d.ChaosGauge,
		&d.BeforeChaosGauge, &d.ChaosGold,
		&d.GuardianCheck, &d.GuardianGauge,
		&d.BeforeGuardianGauge, &d.GuardianGold,
		&d.EponaCheck, &d.EponaGauge,
		&d.BeforeEponaGauge,
		&d.WeekTotalGold,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("캐릭터를 찾을 수 없습니다.")
		}
		return nil, fmt.Errorf("loading day todo: %w", err)
	}
	return &d, nil
}

func (s *CharacterDayService) verifyOwnership(ctx context.Context, username string, memberID int64) error {
	var ownerID int64
	err := s.db.QueryRowContext(ctx,
		"SELECT member_id FROM member WHERE username = ?", username,
	).Scan(&ownerID)
	if err != nil {
		return fmt.Errorf("querying member: %w", err)
	}
	if ownerID != memberID {
		return errors.New("권한이 없습니다.")
	}
	return nil
}

// CheckDayContent checks a specific daily content (chaos/guardian/epona).
func (s *CharacterDayService) CheckDayContent(ctx context.Context, username string, req DayCheckRequest) (*DayTodoResult, error) {
	d, err := s.loadDayTodo(ctx, req.CharacterID)
	if err != nil {
		return nil, err
	}

	if err := s.verifyOwnership(ctx, username, d.MemberID); err != nil {
		return nil, err
	}

	switch req.Category {
	case "chaos":
		d.ChaosCheck, d.ChaosGauge, d.WeekTotalGold = updateCheckChaos(
			d.ChaosCheck, d.ChaosGauge, d.BeforeChaosGauge,
			d.ChaosGold, d.WeekTotalGold,
		)
	case "guardian":
		d.GuardianCheck, d.GuardianGauge, d.WeekTotalGold = updateCheckGuardian(
			d.GuardianCheck, d.GuardianGauge, d.BeforeGuardianGauge,
			d.GuardianGold, d.WeekTotalGold,
		)
	case "epona":
		d.EponaCheck, d.EponaGauge = updateCheckEpona(
			d.EponaCheck, d.EponaGauge, d.BeforeEponaGauge,
		)
	default:
		return nil, fmt.Errorf("알 수 없는 카테고리: %s", req.Category)
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE characters
		 SET chaos_check = ?, chaos_gauge = ?,
		     guardian_check = ?, guardian_gauge = ?,
		     epona_check2 = ?, epona_gauge = ?,
		     week_total_gold = ?, last_modified_date = ?
		 WHERE characters_id = ?`,
		d.ChaosCheck, d.ChaosGauge,
		d.GuardianCheck, d.GuardianGauge,
		d.EponaCheck, d.EponaGauge,
		d.WeekTotalGold, time.Now(), req.CharacterID,
	)
	if err != nil {
		return nil, fmt.Errorf("updating day content: %w", err)
	}

	return &DayTodoResult{
		ChaosCheck:    d.ChaosCheck,
		ChaosGauge:    d.ChaosGauge,
		ChaosGold:     d.ChaosGold,
		GuardianCheck: d.GuardianCheck,
		GuardianGauge: d.GuardianGauge,
		GuardianGold:  d.GuardianGold,
		EponaCheck:    d.EponaCheck,
		EponaGauge:    d.EponaGauge,
		WeekTotalGold: d.WeekTotalGold,
	}, nil
}

// UpdateGauge updates the rest gauge for a character.
func (s *CharacterDayService) UpdateGauge(ctx context.Context, username string, req DayGaugeRequest) error {
	d, err := s.loadDayTodo(ctx, req.CharacterID)
	if err != nil {
		return err
	}

	if err := s.verifyOwnership(ctx, username, d.MemberID); err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE characters
		 SET chaos_gauge = ?, before_chaos_gauge = ?,
		     guardian_gauge = ?, before_guardian_gauge = ?,
		     last_modified_date = ?
		 WHERE characters_id = ?`,
		req.ChaosGauge, req.ChaosGauge,
		req.GuardianGauge, req.GuardianGauge,
		time.Now(), req.CharacterID,
	)
	return err
}

// CheckAllDayContent checks all daily content for a single character.
func (s *CharacterDayService) CheckAllDayContent(ctx context.Context, username string, req DayCheckAllRequest) (*DayTodoResult, error) {
	d, err := s.loadDayTodo(ctx, req.CharacterID)
	if err != nil {
		return nil, err
	}

	if err := s.verifyOwnership(ctx, username, d.MemberID); err != nil {
		return nil, err
	}

	// Check chaos if not already checked
	if d.ChaosCheck != 2 {
		d.ChaosCheck, d.ChaosGauge, d.WeekTotalGold = updateCheckChaos(
			d.ChaosCheck, d.ChaosGauge, d.BeforeChaosGauge,
			d.ChaosGold, d.WeekTotalGold,
		)
	}

	// Check guardian if not already checked
	if d.GuardianCheck < 1 {
		d.GuardianCheck, d.GuardianGauge, d.WeekTotalGold = updateCheckGuardian(
			d.GuardianCheck, d.GuardianGauge, d.BeforeGuardianGauge,
			d.GuardianGold, d.WeekTotalGold,
		)
	}

	// Check epona to max (3)
	for d.EponaCheck < 3 {
		d.EponaCheck, d.EponaGauge = updateCheckEpona(
			d.EponaCheck, d.EponaGauge, d.BeforeEponaGauge,
		)
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE characters
		 SET chaos_check = ?, chaos_gauge = ?,
		     guardian_check = ?, guardian_gauge = ?,
		     epona_check2 = ?, epona_gauge = ?,
		     week_total_gold = ?, last_modified_date = ?
		 WHERE characters_id = ?`,
		d.ChaosCheck, d.ChaosGauge,
		d.GuardianCheck, d.GuardianGauge,
		d.EponaCheck, d.EponaGauge,
		d.WeekTotalGold, time.Now(), req.CharacterID,
	)
	if err != nil {
		return nil, fmt.Errorf("updating all day content: %w", err)
	}

	return &DayTodoResult{
		ChaosCheck:    d.ChaosCheck,
		ChaosGauge:    d.ChaosGauge,
		ChaosGold:     d.ChaosGold,
		GuardianCheck: d.GuardianCheck,
		GuardianGauge: d.GuardianGauge,
		GuardianGold:  d.GuardianGold,
		EponaCheck:    d.EponaCheck,
		EponaGauge:    d.EponaGauge,
		WeekTotalGold: d.WeekTotalGold,
	}, nil
}

// CheckAllDayCharacters checks all daily content for all characters on a server.
func (s *CharacterDayService) CheckAllDayCharacters(ctx context.Context, username string, req DayCheckAllCharactersRequest) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx,
		"SELECT member_id FROM member WHERE username = ?", username,
	).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("querying member: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT characters_id FROM characters
		 WHERE member_id = ? AND server_name = ?
		   AND is_deleted = false`,
		memberID, req.ServerName,
	)
	if err != nil {
		return fmt.Errorf("querying characters: %w", err)
	}
	defer rows.Close()

	var charIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("scanning character id: %w", err)
		}
		charIDs = append(charIDs, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, charID := range charIDs {
		_, err := s.CheckAllDayContent(ctx, username, DayCheckAllRequest{CharacterID: charID})
		if err != nil {
			return fmt.Errorf("checking character %d: %w", charID, err)
		}
	}

	return nil
}

// --- Business Logic Functions (ported from Java) ---

// updateCheckEpona handles the epona check toggle logic.
func updateCheckEpona(eponaCheck, eponaGauge, beforeEponaGauge int) (newCheck, newGauge int) {
	if eponaCheck < 3 {
		eponaCheck += 1
		if eponaGauge >= 20 {
			eponaGauge -= 20
		}
	} else {
		eponaCheck = 0
		eponaGauge = beforeEponaGauge
	}
	return eponaCheck, eponaGauge
}

// updateCheckChaos handles the chaos check toggle logic (0 or 2, no intermediate).
func updateCheckChaos(chaosCheck int, chaosGauge, beforeChaosGauge int, chaosGold, weekTotalGold float64) (newCheck, newGauge int, newWeekGold float64) {
	if chaosCheck != 2 {
		// Calculate gold based on gauge state
		var gold float64
		if beforeChaosGauge >= 20 && beforeChaosGauge < 40 {
			if chaosGauge >= 20 && chaosGauge < 40 {
				gold = chaosGold * 2 / 3
			} else {
				gold = chaosGold / 3
			}
		} else {
			gold = chaosGold / 2
		}
		chaosCheck = 2
		if chaosGauge >= 40 && beforeChaosGauge == chaosGauge {
			chaosGauge -= 40
		}
		weekTotalGold += gold
	} else {
		chaosCheck = 0
		weekTotalGold -= chaosGold
		if weekTotalGold < 0 {
			weekTotalGold = 0
		}
		chaosGauge = beforeChaosGauge
	}
	return chaosCheck, chaosGauge, weekTotalGold
}

// updateCheckGuardian handles the guardian check toggle logic.
func updateCheckGuardian(guardianCheck int, guardianGauge, beforeGuardianGauge int, guardianGold, weekTotalGold float64) (newCheck, newGauge int, newWeekGold float64) {
	if guardianCheck < 1 {
		guardianCheck += 1
		if guardianGauge >= 20 && beforeGuardianGauge == guardianGauge {
			guardianGauge -= 20
		}
		weekTotalGold += guardianGold
	} else {
		guardianCheck = 0
		if beforeGuardianGauge >= 20 {
			if guardianGauge >= 20 {
				weekTotalGold -= guardianGold
			} else {
				weekTotalGold -= guardianGold * 2
			}
		} else {
			weekTotalGold -= guardianGold
		}
		if weekTotalGold < 0 {
			weekTotalGold = 0
		}
		guardianGauge = beforeGuardianGauge
	}
	return guardianCheck, guardianGauge, weekTotalGold
}
