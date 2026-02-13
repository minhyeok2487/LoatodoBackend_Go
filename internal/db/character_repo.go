package db

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type CharacterRepository struct {
	db *sql.DB
}

func NewCharacterRepository(db *sql.DB) *CharacterRepository {
	return &CharacterRepository{db: db}
}

// characterColumns is the full column list for scanning a Character row.
const characterColumns = `
	characters_id, server_name, character_name, character_level, character_class_name,
	character_image, item_level, combat_power, sort_number, memo, member_id,
	gold_character, challenge_guardian, challenge_abyss, is_deleted,
	created_date, last_modified_date,
	chaos_name, chaos_id, guardian_id,
	chaos_check, chaos_gauge, chaos_gold,
	guardian_name, guardian_check, guardian_gauge, guardian_gold,
	epona_check2, epona_gauge, week_total_gold,
	before_epona_gauge, before_chaos_gauge, before_guardian_gauge,
	week_epona, silmael_change, cube_ticket, elysian_count,
	show_character, show_epona, threshold_epona,
	show_chaos, threshold_chaos, show_guardian, threshold_guardian,
	show_week_todo, show_week_epona, show_silmael_change, show_cube_ticket,
	gold_check_version, gold_check_policy_enum, link_cube_cal, show_more_button, show_elysian
`

func scanCharacter(scanner interface{ Scan(dest ...interface{}) error }) (*Character, error) {
	c := &Character{}
	err := scanner.Scan(
		&c.ID, &c.ServerName, &c.CharacterName, &c.CharacterLevel, &c.CharacterClassName,
		&c.CharacterImage, &c.ItemLevel, &c.CombatPower, &c.SortNumber, &c.Memo, &c.MemberID,
		&c.GoldCharacter, &c.ChallengeGuardian, &c.ChallengeAbyss, &c.IsDeleted,
		&c.CreatedDate, &c.LastModifiedDate,
		&c.ChaosName, &c.ChaosID, &c.GuardianID,
		&c.ChaosCheck, &c.ChaosGauge, &c.ChaosGold,
		&c.GuardianName, &c.GuardianCheck, &c.GuardianGauge, &c.GuardianGold,
		&c.EponaCheck2, &c.EponaGauge, &c.WeekTotalGold,
		&c.BeforeEponaGauge, &c.BeforeChaosGauge, &c.BeforeGuardianGauge,
		&c.WeekEpona, &c.SilmaelChange, &c.CubeTicket, &c.ElysianCount,
		&c.ShowCharacter, &c.ShowEpona, &c.ThresholdEpona,
		&c.ShowChaos, &c.ThresholdChaos, &c.ShowGuardian, &c.ThresholdGuardian,
		&c.ShowWeekTodo, &c.ShowWeekEpona, &c.ShowSilmaelChange, &c.ShowCubeTicket,
		&c.GoldCheckVersion, &c.GoldCheckPolicyEnum, &c.LinkCubeCal, &c.ShowMoreButton, &c.ShowElysian,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CharacterRepository) FindByID(ctx context.Context, id int64) (*Character, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+characterColumns+` FROM characters WHERE characters_id = ?`, id,
	)
	return scanCharacter(row)
}

func (r *CharacterRepository) FindByMemberID(ctx context.Context, memberID int64) ([]Character, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+characterColumns+` FROM characters WHERE member_id = ? ORDER BY sort_number ASC, item_level DESC`, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []Character
	for rows.Next() {
		c, err := scanCharacter(rows)
		if err != nil {
			return nil, err
		}
		chars = append(chars, *c)
	}
	return chars, rows.Err()
}

func (r *CharacterRepository) FindByMemberUsername(ctx context.Context, username string) ([]Character, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT c.`+strings.ReplaceAll(characterColumns, "\n", "")+`
		 FROM characters c
		 JOIN member m ON c.member_id = m.member_id
		 WHERE m.username = ?
		 ORDER BY c.sort_number ASC, c.item_level DESC`, username,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []Character
	for rows.Next() {
		c, err := scanCharacter(rows)
		if err != nil {
			return nil, err
		}
		chars = append(chars, *c)
	}
	return chars, rows.Err()
}

func (r *CharacterRepository) Create(ctx context.Context, c *Character) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO characters (
			server_name, character_name, character_level, character_class_name,
			character_image, item_level, combat_power, sort_number, memo, member_id,
			gold_character, challenge_guardian, challenge_abyss, is_deleted,
			created_date, last_modified_date,
			chaos_name, chaos_id, guardian_id,
			chaos_check, chaos_gauge, chaos_gold,
			guardian_name, guardian_check, guardian_gauge, guardian_gold,
			epona_check2, epona_gauge, week_total_gold,
			before_epona_gauge, before_chaos_gauge, before_guardian_gauge,
			week_epona, silmael_change, cube_ticket, elysian_count,
			show_character, show_epona, threshold_epona,
			show_chaos, threshold_chaos, show_guardian, threshold_guardian,
			show_week_todo, show_week_epona, show_silmael_change, show_cube_ticket,
			gold_check_version, gold_check_policy_enum, link_cube_cal, show_more_button, show_elysian
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ServerName, c.CharacterName, c.CharacterLevel, c.CharacterClassName,
		c.CharacterImage, c.ItemLevel, c.CombatPower, c.SortNumber, c.Memo, c.MemberID,
		c.GoldCharacter, c.ChallengeGuardian, c.ChallengeAbyss, c.IsDeleted,
		now, now,
		c.ChaosName, c.ChaosID, c.GuardianID,
		c.ChaosCheck, c.ChaosGauge, c.ChaosGold,
		c.GuardianName, c.GuardianCheck, c.GuardianGauge, c.GuardianGold,
		c.EponaCheck2, c.EponaGauge, c.WeekTotalGold,
		c.BeforeEponaGauge, c.BeforeChaosGauge, c.BeforeGuardianGauge,
		c.WeekEpona, c.SilmaelChange, c.CubeTicket, c.ElysianCount,
		c.ShowCharacter, c.ShowEpona, c.ThresholdEpona,
		c.ShowChaos, c.ThresholdChaos, c.ShowGuardian, c.ThresholdGuardian,
		c.ShowWeekTodo, c.ShowWeekEpona, c.ShowSilmaelChange, c.ShowCubeTicket,
		c.GoldCheckVersion, c.GoldCheckPolicyEnum, c.LinkCubeCal, c.ShowMoreButton, c.ShowElysian,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CharacterRepository) UpdateSettings(ctx context.Context, id int64, field string, value interface{}) error {
	allowedFields := map[string]bool{
		"show_character": true, "show_epona": true, "threshold_epona": true,
		"show_chaos": true, "threshold_chaos": true, "show_guardian": true,
		"threshold_guardian": true, "show_week_todo": true, "show_week_epona": true,
		"show_silmael_change": true, "show_cube_ticket": true,
		"gold_check_version": true, "gold_check_policy_enum": true,
		"link_cube_cal": true, "show_more_button": true, "show_elysian": true,
	}
	if !allowedFields[field] {
		return sql.ErrNoRows
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET `+field+` = ?, last_modified_date = ? WHERE characters_id = ?`,
		value, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateGoldCharacter(ctx context.Context, id int64, goldCharacter bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET gold_character = ?, last_modified_date = ? WHERE characters_id = ?`,
		goldCharacter, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateMemo(ctx context.Context, id int64, memo *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET memo = ?, last_modified_date = ? WHERE characters_id = ?`,
		memo, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateStatus(ctx context.Context, id int64, isDeleted bool, goldCharacter bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET is_deleted = ?, gold_character = ?, last_modified_date = ? WHERE characters_id = ?`,
		isDeleted, goldCharacter, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateSort(ctx context.Context, id int64, sortNumber int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET sort_number = ?, last_modified_date = ? WHERE characters_id = ?`,
		sortNumber, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateDayCheck(ctx context.Context, id int64, field string, value int, gaugeField string, gaugeValue int, weekTotalGold float64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET `+field+` = ?, `+gaugeField+` = ?, week_total_gold = ?, last_modified_date = ? WHERE characters_id = ?`,
		value, gaugeValue, weekTotalGold, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateDayGauge(ctx context.Context, id int64, chaosGauge, guardianGauge, eponaGauge int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET
			chaos_gauge = ?, before_chaos_gauge = ?,
			guardian_gauge = ?, before_guardian_gauge = ?,
			epona_gauge = ?, before_epona_gauge = ?,
			last_modified_date = ?
		 WHERE characters_id = ?`,
		chaosGauge, chaosGauge, guardianGauge, guardianGauge, eponaGauge, eponaGauge,
		time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateCharacter(ctx context.Context, c *Character) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET
			server_name = ?, character_name = ?, character_level = ?,
			character_class_name = ?, character_image = ?,
			item_level = ?, combat_power = ?,
			chaos_name = ?, chaos_id = ?,
			guardian_name = ?, guardian_id = ?,
			chaos_gold = ?, guardian_gold = ?,
			last_modified_date = ?
		 WHERE characters_id = ?`,
		c.ServerName, c.CharacterName, c.CharacterLevel,
		c.CharacterClassName, c.CharacterImage,
		c.ItemLevel, c.CombatPower,
		c.ChaosName, c.ChaosID,
		c.GuardianName, c.GuardianID,
		c.ChaosGold, c.GuardianGold,
		time.Now(), c.ID,
	)
	return err
}

func (r *CharacterRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM characters WHERE characters_id = ?`, id)
	return err
}

func (r *CharacterRepository) DeleteByMemberID(ctx context.Context, memberID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM characters WHERE member_id = ?`, memberID)
	return err
}

func (r *CharacterRepository) BulkUpdateDayContentGauge(ctx context.Context, memberID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET
			chaos_gauge = LEAST(chaos_gauge + IF(chaos_check < 2, (2 - chaos_check) * 10, 0), 200),
			guardian_gauge = LEAST(guardian_gauge + IF(guardian_check < 1, 10, 0), 100),
			epona_gauge = LEAST(epona_gauge + IF(epona_check2 < 3, (3 - epona_check2) * 10, 0), 100),
			last_modified_date = ?
		 WHERE member_id = ? AND is_deleted = false`,
		time.Now(), memberID,
	)
	return err
}

func (r *CharacterRepository) BulkSaveBeforeGauge(ctx context.Context, memberID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET
			before_chaos_gauge = chaos_gauge,
			before_guardian_gauge = guardian_gauge,
			before_epona_gauge = epona_gauge,
			last_modified_date = ?
		 WHERE member_id = ? AND is_deleted = false`,
		time.Now(), memberID,
	)
	return err
}

func (r *CharacterRepository) BulkUpdateDayContentCheck(ctx context.Context, memberID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET
			chaos_check = 0, guardian_check = 0, epona_check2 = 0,
			week_total_gold = 0,
			last_modified_date = ?
		 WHERE member_id = ? AND is_deleted = false`,
		time.Now(), memberID,
	)
	return err
}

func (r *CharacterRepository) BulkUpdateWeekContent(ctx context.Context, memberID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET
			week_epona = 0, silmael_change = false, cube_ticket = 0, elysian_count = 0,
			last_modified_date = ?
		 WHERE member_id = ? AND is_deleted = false`,
		time.Now(), memberID,
	)
	return err
}

func (r *CharacterRepository) BulkUpdateWeekDayTodoTotalGold(ctx context.Context, memberID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET week_total_gold = 0, last_modified_date = ?
		 WHERE member_id = ? AND is_deleted = false`,
		time.Now(), memberID,
	)
	return err
}

func (r *CharacterRepository) UpdateDayContentPriceGuardian(ctx context.Context, id int64, guardianGold float64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET guardian_gold = ?, last_modified_date = ? WHERE characters_id = ?`,
		guardianGold, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateWeekEpona(ctx context.Context, id int64, weekEpona int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET week_epona = ?, last_modified_date = ? WHERE characters_id = ?`,
		weekEpona, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateSilmaelChange(ctx context.Context, id int64, silmaelChange bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET silmael_change = ?, last_modified_date = ? WHERE characters_id = ?`,
		silmaelChange, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateCubeTicket(ctx context.Context, id int64, cubeTicket int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET cube_ticket = ?, last_modified_date = ? WHERE characters_id = ?`,
		cubeTicket, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateElysianCount(ctx context.Context, id int64, elysianCount int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET elysian_count = ?, last_modified_date = ? WHERE characters_id = ?`,
		elysianCount, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateChallengeGuardian(ctx context.Context, id int64, challengeGuardian bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET challenge_guardian = ?, last_modified_date = ? WHERE characters_id = ?`,
		challengeGuardian, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateChallengeAbyss(ctx context.Context, id int64, challengeAbyss bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET challenge_abyss = ?, last_modified_date = ? WHERE characters_id = ?`,
		challengeAbyss, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) UpdateCharacterName(ctx context.Context, id int64, characterName string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET character_name = ?, last_modified_date = ? WHERE characters_id = ?`,
		characterName, time.Now(), id,
	)
	return err
}

func (r *CharacterRepository) FindByCharacterName(ctx context.Context, characterName string) (*Character, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+characterColumns+` FROM characters WHERE character_name = ? AND is_deleted = false LIMIT 1`,
		characterName,
	)
	return scanCharacter(row)
}

func (r *CharacterRepository) FindGoldCharactersByMemberID(ctx context.Context, memberID int64) ([]Character, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+characterColumns+` FROM characters
		 WHERE member_id = ? AND gold_character = true AND is_deleted = false
		 ORDER BY sort_number ASC, item_level DESC`, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []Character
	for rows.Next() {
		c, err := scanCharacter(rows)
		if err != nil {
			return nil, err
		}
		chars = append(chars, *c)
	}
	return chars, rows.Err()
}

func (r *CharacterRepository) CountByMemberID(ctx context.Context, memberID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM characters WHERE member_id = ? AND is_deleted = false`, memberID,
	).Scan(&count)
	return count, err
}

func (r *CharacterRepository) UpdateByAdmin(ctx context.Context, id int64, characterName *string, itemLevel *float64, sortNumber *int, memo *string, goldCharacter *bool, isDeleted *bool) error {
	query := "UPDATE characters SET last_modified_date = ?"
	args := []interface{}{time.Now()}

	if characterName != nil {
		query += ", character_name = ?"
		args = append(args, *characterName)
	}
	if itemLevel != nil {
		query += ", item_level = ?"
		args = append(args, *itemLevel)
	}
	if sortNumber != nil {
		query += ", sort_number = ?"
		args = append(args, *sortNumber)
	}
	if memo != nil {
		query += ", memo = ?"
		args = append(args, *memo)
	}
	if goldCharacter != nil {
		query += ", gold_character = ?"
		args = append(args, *goldCharacter)
	}
	if isDeleted != nil {
		query += ", is_deleted = ?"
		args = append(args, *isDeleted)
	}

	query += " WHERE characters_id = ?"
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *CharacterRepository) BulkResetDayContent(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET
			chaos_check = 0, guardian_check = 0, epona_check2 = 0,
			last_modified_date = ?
		 WHERE is_deleted = false`, time.Now(),
	)
	return err
}

func (r *CharacterRepository) BulkResetWeekContent(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET
			week_epona = 0, silmael_change = false, cube_ticket = 0, elysian_count = 0,
			week_total_gold = 0,
			last_modified_date = ?
		 WHERE is_deleted = false`, time.Now(),
	)
	return err
}

func (r *CharacterRepository) BulkAccumulateGauges(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE characters SET
			chaos_gauge = LEAST(chaos_gauge + IF(chaos_check < 2, (2 - chaos_check) * 10, 0), 200),
			guardian_gauge = LEAST(guardian_gauge + IF(guardian_check < 1, 10, 0), 100),
			epona_gauge = LEAST(epona_gauge + IF(epona_check2 < 3, (3 - epona_check2) * 10, 0), 100),
			before_chaos_gauge = LEAST(chaos_gauge + IF(chaos_check < 2, (2 - chaos_check) * 10, 0), 200),
			before_guardian_gauge = LEAST(guardian_gauge + IF(guardian_check < 1, 10, 0), 100),
			before_epona_gauge = LEAST(epona_gauge + IF(epona_check2 < 3, (3 - epona_check2) * 10, 0), 100),
			last_modified_date = ?
		 WHERE is_deleted = false`, time.Now(),
	)
	return err
}
