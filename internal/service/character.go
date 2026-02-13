package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CharacterService handles character-related business logic.
type CharacterService struct {
	db *sql.DB
}

// NewCharacterService creates a new CharacterService.
func NewCharacterService(db *sql.DB) *CharacterService {
	return &CharacterService{db: db}
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

	// TODO: Call Lostark API to fetch updated character info
	// For now, return existing character data
	resp := &CharacterResponse{
		ID:                 char.ID,
		ServerName:         char.ServerName,
		CharacterName:      char.CharacterName,
		CharacterLevel:     char.CharacterLevel,
		CharacterClassName: char.CharacterClassName,
		ItemLevel:          char.ItemLevel,
		SortNumber:         char.SortNumber,
		GoldCharacter:      char.GoldCharacter,
		ChallengeGuardian:  char.ChallengeGuardian,
		ChallengeAbyss:     char.ChallengeAbyss,
	}
	if char.CharacterImage.Valid {
		resp.CharacterImage = &char.CharacterImage.String
	}
	if char.Memo.Valid {
		resp.Memo = &char.Memo.String
	}
	if char.CombatPower.Valid {
		resp.CombatPower = char.CombatPower.Float64
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
		ID:            id,
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
		`SELECT c.id, c.server_name, c.character_name, c.character_level,
		        c.character_class_name, c.character_image, c.item_level,
		        c.combat_power, c.sort_number, c.memo, c.gold_character,
		        c.challenge_guardian, c.challenge_abyss,
		        COALESCE(c.chaos_check, 0), COALESCE(c.chaos_gauge, 0), COALESCE(c.chaos_gold, 0),
		        COALESCE(c.guardian_check, 0), COALESCE(c.guardian_gauge, 0), COALESCE(c.guardian_gold, 0),
		        COALESCE(c.epona_check, 0), COALESCE(c.epona_gauge, 0),
		        COALESCE(c.week_total_gold, 0),
		        COALESCE(c.week_epona, 0), COALESCE(c.week_epona_check, false),
		        COALESCE(c.silmael_exchange, false), COALESCE(c.cube_ticket, 0),
		        COALESCE(c.gold_check_version, 0),
		        COALESCE(c.elysian_count, 0), COALESCE(c.elysian_all_check, false),
		        COALESCE(c.show_character, true), COALESCE(c.show_chaos, true),
		        COALESCE(c.show_guardian, true), COALESCE(c.show_epona, true),
		        COALESCE(c.show_week_epona, true), COALESCE(c.show_silmael, true),
		        COALESCE(c.show_cube, true)
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
			&c.ID, &c.ServerName, &c.CharacterName, &c.CharacterLevel,
			&c.CharacterClassName, &image, &c.ItemLevel,
			&combatPower, &c.SortNumber, &memo, &c.GoldCharacter,
			&c.ChallengeGuardian, &c.ChallengeAbyss,
			&c.DayTodo.ChaosCheck, &c.DayTodo.ChaosGauge, &c.DayTodo.ChaosGold,
			&c.DayTodo.GuardianCheck, &c.DayTodo.GuardianGauge, &c.DayTodo.GuardianGold,
			&c.DayTodo.EponaCheck, &c.DayTodo.EponaGauge,
			&c.DayTodo.WeekTotalGold,
			&c.WeekTodo.WeekEpona, &c.WeekTodo.WeekEponaCheck,
			&c.WeekTodo.SilmaelExchange, &c.WeekTodo.CubeTicket,
			&c.WeekTodo.GoldCheckVersion,
			&c.WeekTodo.ElysianCount, &c.WeekTodo.ElysianAllCheck,
			&c.Settings.ShowCharacter, &c.Settings.ShowChaos,
			&c.Settings.ShowGuardian, &c.Settings.ShowEpona,
			&c.Settings.ShowWeekEpona, &c.Settings.ShowSilmael,
			&c.Settings.ShowCube,
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
		todoList, err := s.getTodoV2List(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		c.TodoV2List = todoList

		// Load bus gold list
		busGoldList, err := s.getRaidBusGoldList(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		c.RaidBusGoldList = busGoldList

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
		`SELECT c.id, c.server_name, c.character_name, c.character_level,
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
			&c.ID, &c.ServerName, &c.CharacterName, &c.CharacterLevel,
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

// getTodoV2List returns the raid/weekly todo list for a character.
func (s *CharacterService) getTodoV2List(ctx context.Context, characterID int64) ([]TodoV2Response, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT t.id, t.week_category, t.week_content_name, COALESCE(t.week_content_gate, 0),
		        COALESCE(t.gold, 0), COALESCE(t.gold_check, false), COALESCE(t.is_checked, false),
		        COALESCE(t.sort_number, 0), t.message,
		        COALESCE(t.current_gate, 0), COALESCE(t.total_gate, 0),
		        COALESCE(t.more_reward, false), COALESCE(t.more_reward_gold, 0)
		 FROM todo_v2 t
		 WHERE t.character_id = ?
		 ORDER BY t.sort_number ASC`, characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying todo_v2: %w", err)
	}
	defer rows.Close()

	var todos []TodoV2Response
	for rows.Next() {
		var t TodoV2Response
		var message sql.NullString
		err := rows.Scan(
			&t.ID, &t.WeekCategory, &t.WeekContentName, &t.WeekContentGate,
			&t.Gold, &t.GoldCheck, &t.Check,
			&t.SortNumber, &message,
			&t.CurrentGate, &t.TotalGate,
			&t.MoreReward, &t.MoreRewardGold,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning todo_v2: %w", err)
		}
		if message.Valid {
			t.Message = &message.String
		}
		todos = append(todos, t)
	}

	if todos == nil {
		todos = []TodoV2Response{}
	}
	return todos, rows.Err()
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
