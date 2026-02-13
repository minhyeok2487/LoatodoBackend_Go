package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"lostark-todo-backend/internal/client"
)

// CharacterService handles character-related business logic.
type CharacterService struct {
	db         *sql.DB
	charClient *client.LostarkCharacterClient
}

// NewCharacterService creates a new CharacterService.
func NewCharacterService(db *sql.DB, charClient *client.LostarkCharacterClient) *CharacterService {
	return &CharacterService{db: db, charClient: charClient}
}

// UpdateSettingsRequest is the request to update character display settings.
type UpdateSettingsRequest struct {
	CharacterID int64  `json:"characterId"`
	Name        string `json:"name"`
	Value       bool   `json:"value"`
}

// GoldCharacterRequest is the request to toggle gold character.
type GoldCharacterRequest struct {
	CharacterID int64 `json:"characterId"`
}

// UpdateMemoRequest is the request to update character memo.
type UpdateMemoRequest struct {
	CharacterID int64  `json:"characterId"`
	Memo        string `json:"memo"`
}

// DeleteCharacterRequest is the request to toggle character deleted status.
type DeleteCharacterRequest struct {
	CharacterID int64 `json:"characterId"`
}

// UpdateCharacterRequest is the request to update a character from Lostark API.
type UpdateCharacterRequest struct {
	CharacterID int64 `json:"characterId"`
}

// ChangeCharacterNameRequest is the request to change character name.
type ChangeCharacterNameRequest struct {
	CharacterID   int64  `json:"characterId"`
	CharacterName string `json:"characterName"`
}

// AddCharacterRequest is the request to add a new character.
type AddCharacterRequest struct {
	CharacterName string `json:"characterName"`
}

// getUpdateCharacter verifies ownership or friend permission for a character.
func (s *CharacterService) getUpdateCharacter(ctx context.Context, username, friendUsername string, characterID int64) (*characterEntity, error) {
	if friendUsername == "" {
		// Own character - verify ownership
		char, err := s.findCharacterByID(ctx, characterID)
		if err != nil {
			return nil, err
		}
		var memberID int64
		err = s.db.QueryRowContext(ctx,
			"SELECT id FROM member WHERE username = ?", username,
		).Scan(&memberID)
		if err != nil {
			return nil, fmt.Errorf("querying member: %w", err)
		}
		if char.MemberID != memberID {
			return nil, errors.New("권한이 없습니다.")
		}
		return char, nil
	}

	// Friend's character - check friend permission
	var friendID int64
	err := s.db.QueryRowContext(ctx,
		`SELECT f.id FROM friends f
		 JOIN member m1 ON f.member_id = m1.id
		 JOIN member m2 ON f.friend_id = m2.id
		 WHERE m1.username = ? AND m2.username = ?`,
		friendUsername, username,
	).Scan(&friendID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("친구 관계가 아닙니다.")
		}
		return nil, fmt.Errorf("querying friend: %w", err)
	}

	char, err := s.findCharacterByID(ctx, characterID)
	if err != nil {
		return nil, err
	}
	return char, nil
}

// characterEntity is an internal DB model for characters.
type characterEntity struct {
	ID                 int64
	MemberID           int64
	ServerName         string
	CharacterName      string
	CharacterLevel     int
	CharacterClassName string
	CharacterImage     sql.NullString
	ItemLevel          float64
	CombatPower        sql.NullFloat64
	SortNumber         int
	Memo               sql.NullString
	GoldCharacter      bool
	ChallengeGuardian  bool
	ChallengeAbyss     bool
	Deleted            bool
}

func (s *CharacterService) findCharacterByID(ctx context.Context, id int64) (*characterEntity, error) {
	var c characterEntity
	err := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, server_name, character_name, character_level,
		        character_class_name, character_image, item_level, combat_power,
		        sort_number, memo, gold_character, challenge_guardian, challenge_abyss,
		        COALESCE(deleted, false)
		 FROM character_entity WHERE id = ?`, id,
	).Scan(
		&c.ID, &c.MemberID, &c.ServerName, &c.CharacterName, &c.CharacterLevel,
		&c.CharacterClassName, &c.CharacterImage, &c.ItemLevel, &c.CombatPower,
		&c.SortNumber, &c.Memo, &c.GoldCharacter, &c.ChallengeGuardian, &c.ChallengeAbyss,
		&c.Deleted,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("캐릭터를 찾을 수 없습니다.")
		}
		return nil, fmt.Errorf("querying character: %w", err)
	}
	return &c, nil
}

// UpdateSettings updates a character's display setting.
func (s *CharacterService) UpdateSettings(ctx context.Context, username, friendUsername string, req UpdateSettingsRequest) error {
	_, err := s.getUpdateCharacter(ctx, username, friendUsername, req.CharacterID)
	if err != nil {
		return err
	}

	// Map setting name to column
	columnMap := map[string]string{
		"showCharacter": "show_character",
		"showChaos":     "show_chaos",
		"showGuardian":  "show_guardian",
		"showEpona":     "show_epona",
		"showWeekEpona": "show_week_epona",
		"showSilmael":   "show_silmael",
		"showCube":      "show_cube",
	}

	column, ok := columnMap[req.Name]
	if !ok {
		return fmt.Errorf("알 수 없는 설정: %s", req.Name)
	}

	query := fmt.Sprintf("UPDATE character_entity SET %s = ?, last_modified_date = ? WHERE id = ?", column)
	_, err = s.db.ExecContext(ctx, query, req.Value, time.Now(), req.CharacterID)
	return err
}

// ToggleGoldCharacter toggles the gold character flag.
func (s *CharacterService) ToggleGoldCharacter(ctx context.Context, username string, req GoldCharacterRequest) error {
	char, err := s.getUpdateCharacter(ctx, username, "", req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE character_entity SET gold_character = ?, last_modified_date = ? WHERE id = ?",
		!char.GoldCharacter, time.Now(), req.CharacterID,
	)
	return err
}

// UpdateMemo updates a character's memo.
func (s *CharacterService) UpdateMemo(ctx context.Context, username string, req UpdateMemoRequest) error {
	_, err := s.getUpdateCharacter(ctx, username, "", req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE character_entity SET memo = ?, last_modified_date = ? WHERE id = ?",
		req.Memo, time.Now(), req.CharacterID,
	)
	return err
}

// ToggleDeleted toggles the character's deleted status.
func (s *CharacterService) ToggleDeleted(ctx context.Context, username string, req DeleteCharacterRequest) error {
	char, err := s.getUpdateCharacter(ctx, username, "", req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE character_entity SET deleted = ?, last_modified_date = ? WHERE id = ?",
		!char.Deleted, time.Now(), req.CharacterID,
	)
	return err
}

// UpdateSingleCharacter updates a character's info from the Lostark API.
func (s *CharacterService) UpdateSingleCharacter(ctx context.Context, username string, req UpdateCharacterRequest) (*CharacterResponse, error) {
	char, err := s.getUpdateCharacter(ctx, username, "", req.CharacterID)
	if err != nil {
		return nil, err
	}

	// Get member's API key
	var apiKey sql.NullString
	err = s.db.QueryRowContext(ctx,
		"SELECT api_key FROM member WHERE id = ?", char.MemberID,
	).Scan(&apiKey)
	if err != nil {
		return nil, fmt.Errorf("querying member api key: %w", err)
	}
	if !apiKey.Valid || apiKey.String == "" {
		return nil, errors.New("API Key가 등록되지 않았습니다.")
	}

	// Fetch profile from Lostark API
	profile, err := s.charClient.GetCharacterProfile(char.CharacterName, apiKey.String)
	if err != nil {
		return nil, fmt.Errorf("fetching character profile: %w", err)
	}

	// Parse updated values
	itemLevel := char.ItemLevel
	if profile.ItemAvgLevel != "" {
		parsed, _ := (&client.CharacterJsonDto{ItemAvgLevel: profile.ItemAvgLevel}).ParseItemLevel()
		if parsed >= 1415.0 {
			itemLevel = parsed
		}
	}

	var combatPower sql.NullFloat64
	if profile.CombatPower != "" {
		cp, _ := (&client.CharacterJsonDto{CombatPower: profile.CombatPower}).ParseCombatPower()
		if cp > 0 {
			combatPower = sql.NullFloat64{Float64: cp, Valid: true}
		}
	}

	var charImage sql.NullString
	if profile.CharacterImage != nil {
		charImage = sql.NullString{String: *profile.CharacterImage, Valid: true}
	}

	// Find appropriate content IDs based on new item level
	var chaosID, guardianID sql.NullInt64
	err = s.db.QueryRowContext(ctx,
		`SELECT id FROM content WHERE dtype = 'DayContent' AND category = '카오스던전' AND level <= ? ORDER BY level DESC LIMIT 1`,
		itemLevel,
	).Scan(&chaosID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("querying chaos content: %w", err)
	}
	err = s.db.QueryRowContext(ctx,
		`SELECT id FROM content WHERE dtype = 'DayContent' AND category = '가디언토벌' AND level <= ? ORDER BY level DESC LIMIT 1`,
		itemLevel,
	).Scan(&guardianID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("querying guardian content: %w", err)
	}

	// Update character in DB
	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		`UPDATE character_entity SET
			character_name = ?, character_level = ?, character_class_name = ?,
			character_image = ?, item_level = ?, combat_power = ?,
			server_name = ?, chaos_id = ?, guardian_id = ?,
			last_modified_date = ?
		 WHERE id = ?`,
		profile.CharacterName, profile.CharacterLevel, profile.CharacterClassName,
		charImage, itemLevel, combatPower,
		profile.ServerName, chaosID, guardianID,
		now, char.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("updating character: %w", err)
	}

	resp := &CharacterResponse{
		CharacterID:        char.ID,
		ServerName:         profile.ServerName,
		CharacterName:      profile.CharacterName,
		CharacterClassName: profile.CharacterClassName,
		ItemLevel:          itemLevel,
		SortNumber:         char.SortNumber,
		GoldCharacter:      char.GoldCharacter,
		ChallengeGuardian:  char.ChallengeGuardian,
		ChallengeAbyss:     char.ChallengeAbyss,
	}
	if charImage.Valid {
		resp.CharacterImage = &charImage.String
	}
	if char.Memo.Valid {
		resp.Memo = &char.Memo.String
	}
	if combatPower.Valid {
		resp.CombatPower = combatPower.Float64
	}
	return resp, nil
}

// ChangeCharacterName changes a character's name.
func (s *CharacterService) ChangeCharacterName(ctx context.Context, username string, req ChangeCharacterNameRequest) error {
	_, err := s.getUpdateCharacter(ctx, username, "", req.CharacterID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE character_entity SET character_name = ?, last_modified_date = ? WHERE id = ?",
		req.CharacterName, time.Now(), req.CharacterID,
	)
	return err
}

// AddCharacter adds a new character to the member.
func (s *CharacterService) AddCharacter(ctx context.Context, username string, req AddCharacterRequest) (*CharacterResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx,
		"SELECT id FROM member WHERE username = ?", username,
	).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("querying member: %w", err)
	}

	// Get max sort number
	var maxSort sql.NullInt64
	err = s.db.QueryRowContext(ctx,
		"SELECT MAX(sort_number) FROM character_entity WHERE member_id = ?", memberID,
	).Scan(&maxSort)
	if err != nil {
		return nil, fmt.Errorf("querying max sort: %w", err)
	}

	nextSort := 0
	if maxSort.Valid {
		nextSort = int(maxSort.Int64) + 1
	}

	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO character_entity (member_id, character_name, sort_number, gold_character,
		 challenge_guardian, challenge_abyss, deleted, created_date, last_modified_date)
		 VALUES (?, ?, ?, false, false, false, false, ?, ?)`,
		memberID, req.CharacterName, nextSort, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting character: %w", err)
	}

	id, _ := result.LastInsertId()
	return &CharacterResponse{
		CharacterID:   id,
		CharacterName: req.CharacterName,
		SortNumber:    nextSort,
	}, nil
}

// GetCharacterList returns the full character list with todo info for a member.
func (s *CharacterService) GetCharacterList(ctx context.Context, username string) ([]CharacterResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx,
		"SELECT id FROM member WHERE username = ?", username,
	).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("querying member: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.member_id, c.server_name, c.character_name,
		        c.character_class_name, c.character_image, c.item_level,
		        c.combat_power, c.sort_number, c.memo, c.gold_character,
		        c.challenge_guardian, c.challenge_abyss,
		        COALESCE(c.chaos_check, 0), COALESCE(c.chaos_gauge, 0), COALESCE(c.chaos_gold, 0),
		        COALESCE(c.guardian_check, 0), COALESCE(c.guardian_gauge, 0), COALESCE(c.guardian_gold, 0),
		        COALESCE(c.epona_check2, 0), COALESCE(c.epona_gauge, 0),
		        COALESCE(c.week_total_gold, 0),
		        COALESCE(c.week_epona, 0),
		        COALESCE(c.silmael_change, false), COALESCE(c.cube_ticket, 0),
		        COALESCE(c.elysian_count, 0),
		        COALESCE(c.before_epona_gauge, 0), COALESCE(c.before_chaos_gauge, 0),
		        COALESCE(c.before_guardian_gauge, 0),
		        COALESCE(c.show_character, true), COALESCE(c.show_epona, true),
		        COALESCE(c.threshold_epona, 0),
		        COALESCE(c.show_chaos, true), COALESCE(c.threshold_chaos, 0),
		        COALESCE(c.show_guardian, true), COALESCE(c.threshold_guardian, 0),
		        COALESCE(c.show_week_todo, true), COALESCE(c.show_week_epona, true),
		        COALESCE(c.show_silmael_change, true), COALESCE(c.show_cube_ticket, true),
		        COALESCE(c.gold_check_version, false),
		        COALESCE(c.gold_check_policy_enum, 'TOP_THREE_POLICY'),
		        COALESCE(c.link_cube_cal, false), COALESCE(c.show_more_button, true),
		        COALESCE(c.show_elysian, true)
		 FROM character_entity c
		 WHERE c.member_id = ? AND (c.deleted IS NULL OR c.deleted = false)
		 ORDER BY c.sort_number ASC`, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying characters: %w", err)
	}
	defer rows.Close()

	var characters []CharacterResponse
	for rows.Next() {
		var c CharacterResponse
		var image, memo sql.NullString
		var combatPower sql.NullFloat64
		err := rows.Scan(
			&c.CharacterID, &c.MemberID, &c.ServerName, &c.CharacterName,
			&c.CharacterClassName, &image, &c.ItemLevel,
			&combatPower, &c.SortNumber, &memo, &c.GoldCharacter,
			&c.ChallengeGuardian, &c.ChallengeAbyss,
			&c.ChaosCheck, &c.ChaosGauge, &c.ChaosGold,
			&c.GuardianCheck, &c.GuardianGauge, &c.GuardianGold,
			&c.EponaCheck, &c.EponaGauge,
			&c.WeekDayTodoGold,
			&c.WeekEpona,
			&c.SilmaelChange, &c.CubeTicket,
			&c.ElysianCount,
			&c.BeforeEponaGauge, &c.BeforeChaosGauge,
			&c.BeforeGuardianGauge,
			&c.Settings.ShowCharacter, &c.Settings.ShowEpona,
			&c.Settings.ThresholdEpona,
			&c.Settings.ShowChaos, &c.Settings.ThresholdChaos,
			&c.Settings.ShowGuardian, &c.Settings.ThresholdGuardian,
			&c.Settings.ShowWeekTodo, &c.Settings.ShowWeekEpona,
			&c.Settings.ShowSilmaelChange, &c.Settings.ShowCubeTicket,
			&c.Settings.GoldCheckVersion,
			&c.Settings.GoldCheckPolicyEnum,
			&c.Settings.LinkCubeCal, &c.Settings.ShowMoreButton,
			&c.Settings.ShowElysian,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning character: %w", err)
		}
		if image.Valid {
			c.CharacterImage = &image.String
		}
		if memo.Valid {
			c.Memo = &memo.String
		}
		if combatPower.Valid {
			c.CombatPower = combatPower.Float64
		}

		// Load TodoV2 (raid) list for this character
		todoList, err := s.buildTodoList(ctx, c.CharacterID, c.CharacterClassName, c.GoldCharacter, c.Settings.GoldCheckPolicyEnum)
		if err != nil {
			return nil, err
		}
		c.TodoList = todoList

		// Load bus gold list (internal use for gold calc)
		busGoldList, err := s.getRaidBusGoldList(ctx, c.CharacterID)
		if err != nil {
			return nil, err
		}
		c.RaidBusGoldList = busGoldList

		// Calculate week raid gold from todo list
		s.calculateWeekRaidGold(&c)

		characters = append(characters, c)
	}

	if characters == nil {
		characters = []CharacterResponse{}
	}
	return characters, rows.Err()
}

// GetDeletedCharacters returns deleted characters for a member.
func (s *CharacterService) GetDeletedCharacters(ctx context.Context, username string) ([]CharacterResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx,
		"SELECT id FROM member WHERE username = ?", username,
	).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("querying member: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.server_name, c.character_name,
		        c.character_class_name, c.character_image, c.item_level,
		        c.combat_power, c.sort_number, c.memo, c.gold_character,
		        c.challenge_guardian, c.challenge_abyss
		 FROM character_entity c
		 WHERE c.member_id = ? AND c.deleted = true
		 ORDER BY c.sort_number ASC`, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying deleted characters: %w", err)
	}
	defer rows.Close()

	var characters []CharacterResponse
	for rows.Next() {
		var c CharacterResponse
		var image, memo sql.NullString
		var combatPower sql.NullFloat64
		err := rows.Scan(
			&c.CharacterID, &c.ServerName, &c.CharacterName,
			&c.CharacterClassName, &image, &c.ItemLevel,
			&combatPower, &c.SortNumber, &memo, &c.GoldCharacter,
			&c.ChallengeGuardian, &c.ChallengeAbyss,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning character: %w", err)
		}
		if image.Valid {
			c.CharacterImage = &image.String
		}
		if memo.Valid {
			c.Memo = &memo.String
		}
		if combatPower.Valid {
			c.CombatPower = combatPower.Float64
		}
		characters = append(characters, c)
	}

	if characters == nil {
		characters = []CharacterResponse{}
	}
	return characters, rows.Err()
}

// SortCharacterRequest is a single entry in the sort request.
type SortCharacterRequest struct {
	CharacterID int64 `json:"characterId"`
	SortNumber  int   `json:"sortNumber"`
}

// UpdateCharacterSorting updates the sort order for characters.
func (s *CharacterService) UpdateCharacterSorting(ctx context.Context, username string, sortList []SortCharacterRequest) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx,
		"SELECT id FROM member WHERE username = ?", username,
	).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("querying member: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"UPDATE character_entity SET sort_number = ?, last_modified_date = ? WHERE id = ? AND member_id = ?",
	)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, s := range sortList {
		_, err := stmt.ExecContext(ctx, s.SortNumber, now, s.CharacterID, memberID)
		if err != nil {
			return fmt.Errorf("updating sort for character %d: %w", s.CharacterID, err)
		}
	}

	return tx.Commit()
}

// todoV2Row is an internal representation of a single todo_v2 DB row.
type todoV2Row struct {
	ID                  int64
	WeekCategory        string
	WeekContentCategory string
	Gate                int
	Gold                int
	CharacterGold       int
	GoldCheck           bool
	IsChecked           bool
	SortNumber          int
	Message             *string
	CoolTime            int
	MoreRewardCheck     bool
	MoreRewardGold      int
}

// buildTodoList loads TodoV2 rows and groups them into TodoResponseDto list (matching Spring Boot).
func (s *CharacterService) buildTodoList(ctx context.Context, characterID int64, charClassName string, goldCharacter bool, goldPolicy string) ([]TodoResponseDto, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT t.id, t.week_category,
		        COALESCE(wc.week_content_category, ''),
		        COALESCE(wc.gate, 0),
		        COALESCE(wc.gold, 0), COALESCE(wc.character_gold, 0),
		        COALESCE(t.gold_check, false), COALESCE(t.is_checked, false),
		        COALESCE(t.sort_number, 0), t.message,
		        COALESCE(t.cool_time, 1),
		        COALESCE(t.more_reward_check, false), COALESCE(wc.more_reward_gold, 0)
		 FROM todo_v2 t
		 LEFT JOIN content wc ON t.week_content_id = wc.content_id
		 WHERE t.character_id = ? AND COALESCE(t.cool_time, 1) >= 1
		 ORDER BY COALESCE(wc.gate, 0) ASC`, characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying todo_v2: %w", err)
	}
	defer rows.Close()

	var todoRows []todoV2Row
	for rows.Next() {
		var t todoV2Row
		var message sql.NullString
		err := rows.Scan(
			&t.ID, &t.WeekCategory,
			&t.WeekContentCategory, &t.Gate,
			&t.Gold, &t.CharacterGold,
			&t.GoldCheck, &t.IsChecked,
			&t.SortNumber, &message,
			&t.CoolTime,
			&t.MoreRewardCheck, &t.MoreRewardGold,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning todo_v2: %w", err)
		}
		if message.Valid {
			t.Message = &message.String
		}
		todoRows = append(todoRows, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Group by weekCategory into TodoResponseDto (matching Spring Boot grouping logic)
	dtoMap := make(map[string]*TodoResponseDto)
	var dtoOrder []string

	for _, t := range todoRows {
		existing, ok := dtoMap[t.WeekCategory]
		if !ok {
			realGold := 0
			charGold := 0
			if goldCharacter {
				realGold = t.Gold
				if t.MoreRewardCheck && t.CharacterGold == 0 {
					realGold -= t.MoreRewardGold
				}
				if t.CharacterGold != 0 {
					charGold = t.CharacterGold
					if t.MoreRewardCheck {
						charGold -= t.MoreRewardGold
					}
				}
			}
			dto := &TodoResponseDto{
				ID:                  t.ID,
				WeekCategory:        t.WeekCategory,
				WeekContentCategory: t.WeekContentCategory,
				Gold:                t.Gold,
				RealGold:            realGold,
				GoldCheck:           t.GoldCheck,
				Message:             t.Message,
				CurrentGate:         0,
				TotalGate:           t.Gate,
				SortNumber:          t.SortNumber,
				CharacterClassName:  charClassName,
				MoreRewardCheckList: []bool{t.MoreRewardCheck},
				MoreRewardGoldList:  []int{t.MoreRewardGold},
				CharacterGold:       charGold,
			}
			if t.IsChecked {
				dto.CurrentGate = t.Gate
			}
			dtoMap[t.WeekCategory] = dto
			dtoOrder = append(dtoOrder, t.WeekCategory)
		} else {
			// Merge into existing entry (multi-gate raid)
			existing.Gold += t.Gold
			existing.TotalGate = t.Gate
			existing.MoreRewardCheckList = append(existing.MoreRewardCheckList, t.MoreRewardCheck)
			existing.MoreRewardGoldList = append(existing.MoreRewardGoldList, t.MoreRewardGold)

			if goldCharacter {
				addRealGold := t.Gold
				if t.MoreRewardCheck && t.CharacterGold == 0 {
					addRealGold -= t.MoreRewardGold
				}
				existing.RealGold += addRealGold
				if t.CharacterGold != 0 {
					addCharGold := t.CharacterGold
					if t.MoreRewardCheck {
						addCharGold -= t.MoreRewardGold
					}
					existing.CharacterGold += addCharGold
				}
			}

			if t.IsChecked {
				existing.CurrentGate = t.Gate
			}
		}
	}

	// Apply RAID_CHECK_POLICY if applicable
	if goldPolicy == "RAID_CHECK_POLICY" {
		for _, dto := range dtoMap {
			dto.RealGold -= dto.Gold
		}
	}

	// Build result in order, mark completed
	result := make([]TodoResponseDto, 0, len(dtoOrder))
	for _, cat := range dtoOrder {
		dto := dtoMap[cat]
		if dto.CurrentGate == dto.TotalGate {
			dto.Check = true
		}
		result = append(result, *dto)
	}

	if result == nil {
		result = []TodoResponseDto{}
	}
	return result, nil
}

// calculateWeekRaidGold calculates the weekly raid gold for a character response.
func (s *CharacterService) calculateWeekRaidGold(c *CharacterResponse) {
	if !c.GoldCharacter || len(c.TodoList) == 0 {
		return
	}

	for i := range c.TodoList {
		todo := &c.TodoList[i]
		if todo.Check && todo.GoldCheck {
			c.WeekRaidGold += todo.RealGold
			c.WeekCharacterRaidGold += todo.CharacterGold
		}
	}

	// Apply bus gold
	busMap := make(map[string]float64)
	for _, bg := range c.RaidBusGoldList {
		busMap[bg.WeekCategory] = bg.BusGold
	}
	for i := range c.TodoList {
		todo := &c.TodoList[i]
		if busGold, ok := busMap[todo.WeekCategory]; ok {
			todo.RealGold += int(busGold)
			if todo.Check {
				c.WeekRaidGold += int(busGold)
			}
		}
	}
}

// getRaidBusGoldList returns the bus gold list for a character.
func (s *CharacterService) getRaidBusGoldList(ctx context.Context, characterID int64) ([]RaidBusGoldResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT week_category, COALESCE(bus_gold, 0)
		 FROM raid_bus_gold
		 WHERE character_id = ?`, characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying raid_bus_gold: %w", err)
	}
	defer rows.Close()

	var list []RaidBusGoldResponse
	for rows.Next() {
		var r RaidBusGoldResponse
		if err := rows.Scan(&r.WeekCategory, &r.BusGold); err != nil {
			return nil, fmt.Errorf("scanning raid_bus_gold: %w", err)
		}
		list = append(list, r)
	}

	if list == nil {
		list = []RaidBusGoldResponse{}
	}
	return list, rows.Err()
}
