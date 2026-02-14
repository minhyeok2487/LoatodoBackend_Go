package db

import (
	"context"
	"database/sql"
	"time"
)

type RaidBusGoldRepository struct {
	db *sql.DB
}

func NewRaidBusGoldRepository(db *sql.DB) *RaidBusGoldRepository {
	return &RaidBusGoldRepository{db: db}
}

const raidBusGoldColumns = `
	raid_bus_gold_id, character_id, week_category, bus_gold, fixed,
	created_date, last_modified_date
`

func scanRaidBusGold(scanner interface{ Scan(dest ...interface{}) error }) (*RaidBusGold, error) {
	r := &RaidBusGold{}
	err := scanner.Scan(
		&r.ID, &r.CharacterID, &r.WeekCategory, &r.BusGold, &r.Fixed,
		&r.CreatedDate, &r.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RaidBusGoldRepository) FindByID(ctx context.Context, id int64) (*RaidBusGold, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+raidBusGoldColumns+` FROM raid_bus_gold WHERE raid_bus_gold_id = ?`, id,
	)
	return scanRaidBusGold(row)
}

func (r *RaidBusGoldRepository) FindByCharacterID(ctx context.Context, characterID int64) ([]RaidBusGold, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+raidBusGoldColumns+` FROM raid_bus_gold WHERE character_id = ? ORDER BY raid_bus_gold_id ASC`,
		characterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var golds []RaidBusGold
	for rows.Next() {
		g, err := scanRaidBusGold(rows)
		if err != nil {
			return nil, err
		}
		golds = append(golds, *g)
	}
	return golds, rows.Err()
}

func (r *RaidBusGoldRepository) FindByCharacterIDAndWeekCategory(ctx context.Context, characterID int64, weekCategory string) (*RaidBusGold, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+raidBusGoldColumns+` FROM raid_bus_gold WHERE character_id = ? AND week_category = ?`,
		characterID, weekCategory,
	)
	return scanRaidBusGold(row)
}

func (r *RaidBusGoldRepository) Create(ctx context.Context, rbg *RaidBusGold) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO raid_bus_gold (character_id, week_category, bus_gold, fixed,
		                            created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		rbg.CharacterID, rbg.WeekCategory, rbg.BusGold, rbg.Fixed,
		now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *RaidBusGoldRepository) Update(ctx context.Context, id int64, busGold int, fixed bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE raid_bus_gold SET bus_gold = ?, fixed = ?, last_modified_date = ?
		 WHERE raid_bus_gold_id = ?`,
		busGold, fixed, time.Now(), id,
	)
	return err
}

func (r *RaidBusGoldRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM raid_bus_gold WHERE raid_bus_gold_id = ?`, id)
	return err
}

func (r *RaidBusGoldRepository) DeleteByCharacterID(ctx context.Context, characterID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM raid_bus_gold WHERE character_id = ?`, characterID)
	return err
}

func (r *RaidBusGoldRepository) DeleteByCharacterIDAndWeekCategory(ctx context.Context, characterID int64, weekCategory string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM raid_bus_gold WHERE character_id = ? AND week_category = ?`,
		characterID, weekCategory,
	)
	return err
}
