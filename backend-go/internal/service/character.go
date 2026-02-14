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
			"SELECT member_id FROM member WHERE username = ?", username,
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
		`SELECT f.friends_id FROM friends f
		 JOIN member m1 ON f.member_id = m1.member_id
		 JOIN member m2 ON f.from_member = m2.member_id
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
		`SELECT characters_id, member_id, server_name, character_name, character_level,
		        character_class_name, character_image, item_level, combat_power,
		        sort_number, memo, gold_character+0 AS gold_character, challenge_guardian+0 AS challenge_guardian, challenge_abyss+0 AS challenge_abyss,
		        COALESCE(is_deleted+0, 0) AS is_deleted
		 FROM characters WHERE characters_id = ?`, id,
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

	query := fmt.Sprintf("UPDATE characters SET %s = ?, last_modified_date = ? WHERE characters_id = ?", column)
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
		"UPDATE characters SET gold_character = ?, last_modified_date = ? WHERE characters_id = ?",
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
		"UPDATE characters SET memo = ?, last_modified_date = ? WHERE characters_id = ?",
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
		"UPDATE characters SET is_deleted = ?, last_modified_date = ? WHERE characters_id = ?",
		!char.Deleted, time.Now(), req.CharacterID,
	)
	return err
}

// ToggleDeletedWithFriend toggles the character's deleted status with friend support.
func (s *CharacterService) ToggleDeletedWithFriend(ctx context.Context, username, friendUsername string, req DeleteCharacterRequest) error {
	char, err := s.getUpdateCharacter(ctx, username, friendUsername, req.CharacterID)
	if err != nil {
		return err
	}

	// If gold character, unset it first
	if char.GoldCharacter {
		_, err = s.db.ExecContext(ctx,
			"UPDATE characters SET gold_character = false WHERE characters_id = ?",
			req.CharacterID,
		)
		if err != nil {
			return err
		}
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE characters SET is_deleted = ?, last_modified_date = ? WHERE characters_id = ?",
		!char.Deleted, time.Now(), req.CharacterID,
	)
	return err
}

// UpdateSettingsAndReturn updates settings and returns the updated character.
func (s *CharacterService) UpdateSettingsAndReturn(ctx context.Context, username, friendUsername string, req UpdateSettingsRequest) (*CharacterResponse, error) {
	if err := s.UpdateSettings(ctx, username, friendUsername, req); err != nil {
		return nil, err
	}
	return s.GetCharacterByID(ctx, req.CharacterID)
}

// ToggleGoldCharacterAndReturn toggles gold character and returns the updated character.
func (s *CharacterService) ToggleGoldCharacterAndReturn(ctx context.Context, username, friendUsername string, req GoldCharacterRequest) (*CharacterResponse, error) {
	char, err := s.getUpdateCharacter(ctx, username, friendUsername, req.CharacterID)
	if err != nil {
		return nil, err
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE characters SET gold_character = ?, last_modified_date = ? WHERE characters_id = ?",
		!char.GoldCharacter, time.Now(), req.CharacterID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetCharacterByID(ctx, req.CharacterID)
}

// UpdateMemoAndReturn updates memo and returns the updated character.
func (s *CharacterService) UpdateMemoAndReturn(ctx context.Context, username, friendUsername string, req UpdateMemoRequest) (*CharacterResponse, error) {
	_, err := s.getUpdateCharacter(ctx, username, friendUsername, req.CharacterID)
	if err != nil {
		return nil, err
	}

	var memo interface{}
	if req.Memo != "" {
		memo = req.Memo
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE characters SET memo = ?, last_modified_date = ? WHERE characters_id = ?",
		memo, time.Now(), req.CharacterID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetCharacterByID(ctx, req.CharacterID)
}

// ChangeCharacterNameAndReturn changes character name and returns the updated character.
func (s *CharacterService) ChangeCharacterNameAndReturn(ctx context.Context, username, friendUsername string, req ChangeCharacterNameRequest) (*CharacterResponse, error) {
	if err := s.ChangeCharacterName(ctx, username, req); err != nil {
		return nil, err
	}
	return s.GetCharacterByID(ctx, req.CharacterID)
}

// GetCharacterByID returns a single character by ID.
func (s *CharacterService) GetCharacterByID(ctx context.Context, characterID int64) (*CharacterResponse, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT c.characters_id, c.member_id, c.server_name, c.character_name,
		        c.character_class_name, c.character_image, c.item_level,
		        c.combat_power, c.sort_number, c.memo, c.gold_character+0 AS gold_character,
		        c.challenge_guardian+0 AS challenge_guardian, c.challenge_abyss+0 AS challenge_abyss,
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
		 FROM characters c
		 WHERE c.characters_id = ?`, characterID,
	)

	var c CharacterResponse
	var image, serverName, memo sql.NullString
	var combatPower sql.NullFloat64

	err := row.Scan(
		&c.CharacterID, &c.MemberID, &serverName, &c.CharacterName,
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
		&c.BeforeEponaGauge, &c.BeforeChaosGauge, &c.BeforeGuardianGauge,
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
		return nil, fmt.Errorf("querying character: %w", err)
	}

	if image.Valid {
		c.CharacterImage = &image.String
	}
	if serverName.Valid {
		c.ServerName = serverName.String
	}
	if memo.Valid {
		c.Memo = &memo.String
	}
	if combatPower.Valid {
		c.CombatPower = combatPower.Float64
	}

	// Build todo list
	todoList, err := s.buildTodoList(ctx, c.CharacterID, c.CharacterClassName, c.GoldCharacter, c.Settings.GoldCheckPolicyEnum)
	if err == nil {
		c.TodoList = todoList
	} else {
		c.TodoList = []TodoResponseDto{}
	}

	// Load chaos/guardian content based on itemLevel
	c.Chaos = s.getDayContent(ctx, "카오스던전", c.ItemLevel)
	c.Guardian = s.getDayContent(ctx, "가디언토벌", c.ItemLevel)

	// Calculate gold
	s.calculateWeekRaidGold(&c)

	return &c, nil
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
		"SELECT api_key FROM member WHERE member_id = ?", char.MemberID,
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
		cp, parseErr := (&client.CharacterJsonDto{CombatPower: profile.CombatPower}).ParseCombatPower()
		if parseErr != nil {
			return nil, fmt.Errorf("parsing combat power '%s': %w", profile.CombatPower, parseErr)
		}
		if cp > 0 {
			combatPower = sql.NullFloat64{Float64: cp, Valid: true}
		}
	}
	// If combatPower is still not valid, use existing value from character
	if !combatPower.Valid && char.CombatPower.Valid {
		combatPower = char.CombatPower
	}

	var charImage sql.NullString
	if profile.CharacterImage != nil {
		charImage = sql.NullString{String: *profile.CharacterImage, Valid: true}
	}

	// Find appropriate content IDs based on new item level
	var chaosID, guardianID sql.NullInt64
	err = s.db.QueryRowContext(ctx,
		`SELECT content_id FROM content WHERE dtype = 'DayContent' AND category = '카오스던전' AND level <= ? ORDER BY level DESC LIMIT 1`,
		itemLevel,
	).Scan(&chaosID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("querying chaos content: %w", err)
	}
	err = s.db.QueryRowContext(ctx,
		`SELECT content_id FROM content WHERE dtype = 'DayContent' AND category = '가디언토벌' AND level <= ? ORDER BY level DESC LIMIT 1`,
		itemLevel,
	).Scan(&guardianID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("querying guardian content: %w", err)
	}

	// Update character in DB
	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		`UPDATE characters SET
			character_name = ?, character_level = ?, character_class_name = ?,
			character_image = ?, item_level = ?, combat_power = ?,
			server_name = ?, chaos_id = ?, guardian_id = ?,
			last_modified_date = ?
		 WHERE characters_id = ?`,
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
		"UPDATE characters SET character_name = ?, last_modified_date = ? WHERE characters_id = ?",
		req.CharacterName, time.Now(), req.CharacterID,
	)
	return err
}

// AddCharacter adds a new character to the member.
func (s *CharacterService) AddCharacter(ctx context.Context, username string, req AddCharacterRequest) (*CharacterResponse, error) {
	// 1. Get member info including API key
	var memberID int64
	var apiKey sql.NullString
	err := s.db.QueryRowContext(ctx,
		"SELECT member_id, api_key FROM member WHERE username = ?", username,
	).Scan(&memberID, &apiKey)
	if err != nil {
		return nil, fmt.Errorf("querying member: %w", err)
	}
	if !apiKey.Valid || apiKey.String == "" {
		return nil, errors.New("API Key가 등록되지 않았습니다.")
	}

	// 2. Check for duplicate character name
	var existingCount int
	err = s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM characters WHERE member_id = ? AND character_name = ?",
		memberID, req.CharacterName,
	).Scan(&existingCount)
	if err != nil {
		return nil, fmt.Errorf("checking duplicate character: %w", err)
	}
	if existingCount > 0 {
		return nil, errors.New("이미 등록된 캐릭터 이름입니다. 삭제된 캐릭터도 확인해주세요.")
	}

	// 3. Get character info from Lostark API
	profile, err := s.charClient.GetCharacterProfile(req.CharacterName, apiKey.String)
	if err != nil {
		return nil, fmt.Errorf("fetching character from Lostark API: %w", err)
	}

	// Parse item level
	itemLevel := 0.0
	if profile.ItemAvgLevel != "" {
		parsed, _ := (&client.CharacterJsonDto{ItemAvgLevel: profile.ItemAvgLevel}).ParseItemLevel()
		if parsed >= 1415.0 {
			itemLevel = parsed
		}
	}
	if itemLevel < 1415.0 {
		return nil, errors.New("아이템 레벨 1415 이상 캐릭터만 등록 가능합니다.")
	}

	// Parse combat power
	var combatPower sql.NullFloat64
	if profile.CombatPower != "" {
		cp, _ := (&client.CharacterJsonDto{CombatPower: profile.CombatPower}).ParseCombatPower()
		if cp > 0 {
			combatPower = sql.NullFloat64{Float64: cp, Valid: true}
		}
	}

	// Get character image
	var charImage sql.NullString
	if profile.CharacterImage != nil {
		charImage = sql.NullString{String: *profile.CharacterImage, Valid: true}
	}

	// 4. Find content IDs based on item level
	var chaosID, guardianID sql.NullInt64
	err = s.db.QueryRowContext(ctx,
		`SELECT content_id FROM content WHERE dtype = 'DayContent' AND category = '카오스던전' AND level <= ? ORDER BY level DESC LIMIT 1`,
		itemLevel,
	).Scan(&chaosID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("querying chaos content: %w", err)
	}
	err = s.db.QueryRowContext(ctx,
		`SELECT content_id FROM content WHERE dtype = 'DayContent' AND category = '가디언토벌' AND level <= ? ORDER BY level DESC LIMIT 1`,
		itemLevel,
	).Scan(&guardianID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("querying guardian content: %w", err)
	}

	// 5. Get max sort number
	var maxSort sql.NullInt64
	err = s.db.QueryRowContext(ctx,
		"SELECT MAX(sort_number) FROM characters WHERE member_id = ?", memberID,
	).Scan(&maxSort)
	if err != nil {
		return nil, fmt.Errorf("querying max sort: %w", err)
	}

	nextSort := 0
	if maxSort.Valid {
		nextSort = int(maxSort.Int64) + 1
	}

	// 6. Insert character with full data (all NOT NULL fields must be included)
	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO characters (
			member_id, server_name, character_name, character_level,
			character_class_name, character_image, item_level, combat_power,
			sort_number, gold_character, challenge_guardian, challenge_abyss,
			chaos_id, guardian_id, is_deleted, created_date, last_modified_date,
			before_chaos_gauge, before_epona_gauge, before_guardian_gauge,
			chaos_check, chaos_gauge, chaos_gold,
			epona_check2, epona_gauge,
			guardian_check, guardian_gauge, guardian_gold,
			week_total_gold, cube_ticket, silmael_change, week_epona, elysian_count
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, false, false, false,
			?, ?, false, ?, ?,
			0, 0, 0,
			0, 0, 0,
			0, 0,
			0, 0, 0,
			0, 0, false, 0, 0
		)`,
		memberID, profile.ServerName, profile.CharacterName, profile.CharacterLevel,
		profile.CharacterClassName, charImage, itemLevel, combatPower,
		nextSort,
		chaosID, guardianID, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting character: %w", err)
	}

	id, _ := result.LastInsertId()
	resp := &CharacterResponse{
		CharacterID:        id,
		ServerName:         profile.ServerName,
		CharacterName:      profile.CharacterName,
		CharacterClassName: profile.CharacterClassName,
		ItemLevel:          itemLevel,
		SortNumber:         nextSort,
	}
	if charImage.Valid {
		resp.CharacterImage = &charImage.String
	}
	if combatPower.Valid {
		resp.CombatPower = combatPower.Float64
	}
	return resp, nil
}

// GetCharacterList returns the full character list with todo info for a member.
// Optimized to use batch queries instead of N+1 queries.
func (s *CharacterService) GetCharacterList(ctx context.Context, username string) ([]CharacterResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx,
		"SELECT member_id FROM member WHERE username = ?", username,
	).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("querying member: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT c.characters_id, c.member_id, c.server_name, c.character_name,
		        c.character_class_name, c.character_image, c.item_level,
		        c.combat_power, c.sort_number, c.memo, c.gold_character+0 AS gold_character,
		        c.challenge_guardian+0 AS challenge_guardian, c.challenge_abyss+0 AS challenge_abyss,
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
		 FROM characters c
		 WHERE c.member_id = ? AND COALESCE(c.is_deleted+0, 0) = 0
		 ORDER BY c.sort_number ASC`, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying characters: %w", err)
	}
	defer rows.Close()

	var characters []CharacterResponse
	var characterIDs []int64
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
		c.TodoList = []TodoResponseDto{}
		c.RaidBusGoldList = []RaidBusGoldResponse{}

		characters = append(characters, c)
		characterIDs = append(characterIDs, c.CharacterID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(characters) == 0 {
		return []CharacterResponse{}, nil
	}

	// Batch load all TodoV2 data for all characters at once
	todoMap, err := s.buildTodoListBatch(ctx, characterIDs)
	if err != nil {
		return nil, err
	}

	// Batch load all bus gold data for all characters at once
	busGoldMap, err := s.getRaidBusGoldListBatch(ctx, characterIDs)
	if err != nil {
		return nil, err
	}

	// Cache day content by item level to avoid duplicate queries
	dayContentCache := make(map[string]*DayContentResponse)

	// Assign data to each character
	for i := range characters {
		c := &characters[i]

		// Assign todo list from batch result
		if todos, ok := todoMap[c.CharacterID]; ok {
			c.TodoList = s.processTodoList(todos, c.CharacterClassName, c.GoldCharacter, c.Settings.GoldCheckPolicyEnum)
		}

		// Load chaos/guardian content with caching
		chaosKey := fmt.Sprintf("카오스던전_%.0f", c.ItemLevel)
		if cached, ok := dayContentCache[chaosKey]; ok {
			c.Chaos = cached
		} else {
			c.Chaos = s.getDayContent(ctx, "카오스던전", c.ItemLevel)
			dayContentCache[chaosKey] = c.Chaos
		}

		guardianKey := fmt.Sprintf("가디언토벌_%.0f", c.ItemLevel)
		if cached, ok := dayContentCache[guardianKey]; ok {
			c.Guardian = cached
		} else {
			c.Guardian = s.getDayContent(ctx, "가디언토벌", c.ItemLevel)
			dayContentCache[guardianKey] = c.Guardian
		}

		// Assign bus gold list from batch result
		if busGold, ok := busGoldMap[c.CharacterID]; ok {
			c.RaidBusGoldList = busGold
		}

		// Calculate week raid gold from todo list
		s.calculateWeekRaidGold(c)
	}

	return characters, nil
}

// GetDeletedCharacters returns deleted characters for a member.
func (s *CharacterService) GetDeletedCharacters(ctx context.Context, username string) ([]CharacterResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx,
		"SELECT member_id FROM member WHERE username = ?", username,
	).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("querying member: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT c.characters_id, c.server_name, c.character_name,
		        c.character_class_name, c.character_image, c.item_level,
		        c.combat_power, c.sort_number, c.memo, c.gold_character+0 AS gold_character,
		        c.challenge_guardian+0 AS challenge_guardian, c.challenge_abyss+0 AS challenge_abyss
		 FROM characters c
		 WHERE c.member_id = ? AND c.is_deleted+0 = 1
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
		"SELECT member_id FROM member WHERE username = ?", username,
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
		"UPDATE characters SET sort_number = ?, last_modified_date = ? WHERE characters_id = ? AND member_id = ?",
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

// todoV2Row is an internal representation of a single todov2 DB row.
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
		`SELECT t.todo_id, wc.week_category,
		        COALESCE(wc.week_content_category, ''),
		        COALESCE(wc.gate, 0),
		        COALESCE(wc.gold, 0), COALESCE(wc.character_gold, 0),
		        COALESCE(t.gold_check, false), COALESCE(t.is_checked, false),
		        COALESCE(t.sort_number, 0), t.message,
		        COALESCE(t.cool_time, 1),
		        COALESCE(t.more_reward_check, false), COALESCE(wc.more_reward_gold, 0)
		 FROM todov2 t
		 LEFT JOIN content wc ON t.content_id = wc.content_id
		 WHERE t.character_id = ? AND COALESCE(t.cool_time, 1) >= 1
		 ORDER BY COALESCE(wc.gate, 0) ASC`, characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying todov2: %w", err)
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
			return nil, fmt.Errorf("scanning todov2: %w", err)
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

// getDayContent returns the appropriate day content (chaos/guardian) for a given item level.
func (s *CharacterService) getDayContent(ctx context.Context, category string, itemLevel float64) *DayContentResponse {
	row := s.db.QueryRowContext(ctx,
		`SELECT content_id, category, name, level,
		        COALESCE(shilling, 0), COALESCE(honor_shard, 0), COALESCE(leap_stone, 0),
		        COALESCE(destruction_stone, 0), COALESCE(guardian_stone, 0), COALESCE(jewelry, 0)
		 FROM content
		 WHERE category = ? AND level <= ?
		 ORDER BY level DESC, content_id DESC
		 LIMIT 1`, category, itemLevel,
	)

	var c DayContentResponse
	err := row.Scan(
		&c.ID, &c.Category, &c.Name, &c.Level,
		&c.Shilling, &c.HonorShard, &c.LeapStone,
		&c.DestructionStone, &c.GuardianStone, &c.Jewelry,
	)
	if err != nil {
		return nil
	}
	return &c
}

// buildTodoListBatch loads all TodoV2 rows for multiple characters in a single query.
func (s *CharacterService) buildTodoListBatch(ctx context.Context, characterIDs []int64) (map[int64][]todoV2Row, error) {
	if len(characterIDs) == 0 {
		return make(map[int64][]todoV2Row), nil
	}

	// Build IN clause
	placeholders := make([]string, len(characterIDs))
	args := make([]interface{}, len(characterIDs))
	for i, id := range characterIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	inClause := "(" + joinStrings(placeholders, ",") + ")"

	query := `SELECT t.character_id, t.todo_id, wc.week_category,
	                 COALESCE(wc.week_content_category, ''),
	                 COALESCE(wc.gate, 0),
	                 COALESCE(wc.gold, 0), COALESCE(wc.character_gold, 0),
	                 COALESCE(t.gold_check, false), COALESCE(t.is_checked, false),
	                 COALESCE(t.sort_number, 0), t.message,
	                 COALESCE(t.cool_time, 1),
	                 COALESCE(t.more_reward_check, false), COALESCE(wc.more_reward_gold, 0)
	          FROM todov2 t
	          LEFT JOIN content wc ON t.content_id = wc.content_id
	          WHERE t.character_id IN ` + inClause + ` AND COALESCE(t.cool_time, 1) >= 1
	          ORDER BY t.character_id, COALESCE(wc.gate, 0) ASC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("batch querying todov2: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]todoV2Row)
	for rows.Next() {
		var charID int64
		var t todoV2Row
		var message sql.NullString
		err := rows.Scan(
			&charID, &t.ID, &t.WeekCategory,
			&t.WeekContentCategory, &t.Gate,
			&t.Gold, &t.CharacterGold,
			&t.GoldCheck, &t.IsChecked,
			&t.SortNumber, &message,
			&t.CoolTime,
			&t.MoreRewardCheck, &t.MoreRewardGold,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning todov2 batch: %w", err)
		}
		if message.Valid {
			t.Message = &message.String
		}
		result[charID] = append(result[charID], t)
	}

	return result, rows.Err()
}

// processTodoList converts raw todoV2 rows to TodoResponseDto list (matching Spring Boot grouping logic).
func (s *CharacterService) processTodoList(todoRows []todoV2Row, charClassName string, goldCharacter bool, goldPolicy string) []TodoResponseDto {
	if len(todoRows) == 0 {
		return []TodoResponseDto{}
	}

	// Group by weekCategory into TodoResponseDto
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

	return result
}

// getRaidBusGoldListBatch loads all bus gold data for multiple characters in a single query.
func (s *CharacterService) getRaidBusGoldListBatch(ctx context.Context, characterIDs []int64) (map[int64][]RaidBusGoldResponse, error) {
	if len(characterIDs) == 0 {
		return make(map[int64][]RaidBusGoldResponse), nil
	}

	// Build IN clause
	placeholders := make([]string, len(characterIDs))
	args := make([]interface{}, len(characterIDs))
	for i, id := range characterIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	inClause := "(" + joinStrings(placeholders, ",") + ")"

	query := `SELECT character_id, week_category, COALESCE(bus_gold, 0)
	          FROM raid_bus_gold
	          WHERE character_id IN ` + inClause

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("batch querying raid_bus_gold: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]RaidBusGoldResponse)
	for rows.Next() {
		var charID int64
		var r RaidBusGoldResponse
		if err := rows.Scan(&charID, &r.WeekCategory, &r.BusGold); err != nil {
			return nil, fmt.Errorf("scanning raid_bus_gold batch: %w", err)
		}
		result[charID] = append(result[charID], r)
	}

	// Initialize empty slices for characters with no bus gold
	for _, id := range characterIDs {
		if _, ok := result[id]; !ok {
			result[id] = []RaidBusGoldResponse{}
		}
	}

	return result, rows.Err()
}

// joinStrings joins strings with a separator (simple helper to avoid importing strings package).
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
