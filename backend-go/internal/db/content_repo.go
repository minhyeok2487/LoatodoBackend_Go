package db

import (
	"context"
	"database/sql"
)

type ContentRepository struct {
	db *sql.DB
}

func NewContentRepository(db *sql.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

const contentColumns = `
	content_id, dtype, category, name, level,
	COALESCE(shilling, 0), COALESCE(honor_shard, 0), COALESCE(leap_stone, 0),
	COALESCE(destruction_stone, 0), COALESCE(guardian_stone, 0), COALESCE(jewelry, 0),
	week_category, week_content_category, COALESCE(gate, 0),
	COALESCE(gold, 0), COALESCE(character_gold, 0), COALESCE(cool_time, 0),
	COALESCE(more_reward_gold, 0)
`

func scanContent(scanner interface{ Scan(dest ...interface{}) error }) (*Content, error) {
	c := &Content{}
	err := scanner.Scan(
		&c.ID, &c.DType, &c.Category, &c.Name, &c.Level,
		&c.Shilling, &c.HonorShard, &c.LeapStone,
		&c.DestructionStone, &c.GuardianStone, &c.Jewelry,
		&c.WeekCategory, &c.WeekContentCategory, &c.Gate,
		&c.Gold, &c.CharacterGold, &c.CoolTime,
		&c.MoreRewardGold,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ContentRepository) FindByID(ctx context.Context, id int64) (*Content, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+contentColumns+` FROM content WHERE content_id = ?`, id,
	)
	return scanContent(row)
}

func (r *ContentRepository) FindAllDayContent(ctx context.Context) ([]Content, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+contentColumns+` FROM content WHERE dtype = 'DayContent' ORDER BY level DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []Content
	for rows.Next() {
		c, err := scanContent(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, *c)
	}
	return contents, rows.Err()
}

func (r *ContentRepository) FindAllWeekContent(ctx context.Context) ([]Content, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+contentColumns+` FROM content WHERE dtype = 'WeekContent' ORDER BY level DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []Content
	for rows.Next() {
		c, err := scanContent(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, *c)
	}
	return contents, rows.Err()
}

func (r *ContentRepository) FindDayContentByCategory(ctx context.Context, category string) ([]Content, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+contentColumns+` FROM content WHERE dtype = 'DayContent' AND category = ? ORDER BY level DESC`,
		category,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []Content
	for rows.Next() {
		c, err := scanContent(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, *c)
	}
	return contents, rows.Err()
}

func (r *ContentRepository) FindWeekContentByID(ctx context.Context, id int64) (*Content, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+contentColumns+` FROM content WHERE content_id = ? AND dtype = 'WeekContent'`, id,
	)
	return scanContent(row)
}

func (r *ContentRepository) FindWeekContentByWeekCategory(ctx context.Context, weekCategory string) ([]Content, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+contentColumns+` FROM content
		 WHERE dtype = 'WeekContent' AND week_category = ?
		 ORDER BY level DESC, gate ASC`,
		weekCategory,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []Content
	for rows.Next() {
		c, err := scanContent(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, *c)
	}
	return contents, rows.Err()
}

func (r *ContentRepository) FindAll(ctx context.Context) ([]Content, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+contentColumns+` FROM content ORDER BY dtype, level DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []Content
	for rows.Next() {
		c, err := scanContent(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, *c)
	}
	return contents, rows.Err()
}

func (r *ContentRepository) FindWeekContentByLevelAndCategory(ctx context.Context, level float64, weekContentCategory string) ([]Content, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+contentColumns+` FROM content
		 WHERE dtype = 'WeekContent' AND level <= ? AND week_content_category = ?
		 ORDER BY level DESC`,
		level, weekContentCategory,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []Content
	for rows.Next() {
		c, err := scanContent(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, *c)
	}
	return contents, rows.Err()
}
