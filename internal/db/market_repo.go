package db

import (
	"context"
	"database/sql"
	"time"
)

type MarketRepository struct {
	db *sql.DB
}

func NewMarketRepository(db *sql.DB) *MarketRepository {
	return &MarketRepository{db: db}
}

const marketColumns = `
	market_id, lostark_market_id, name, grade, icon,
	bundle_count, y_day_avg_price, recent_price, current_min_price, category_code,
	created_date, last_modified_date
`

func scanMarket(scanner interface{ Scan(dest ...interface{}) error }) (*Market, error) {
	m := &Market{}
	err := scanner.Scan(
		&m.ID, &m.LostarkMarketID, &m.Name, &m.Grade, &m.Icon,
		&m.BundleCount, &m.YDayAvgPrice, &m.RecentPrice, &m.CurrentMinPrice, &m.CategoryCode,
		&m.CreatedDate, &m.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MarketRepository) FindByID(ctx context.Context, id int64) (*Market, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+marketColumns+` FROM market WHERE market_id = ?`, id,
	)
	return scanMarket(row)
}

func (r *MarketRepository) FindAll(ctx context.Context) ([]Market, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+marketColumns+` FROM market ORDER BY name ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []Market
	for rows.Next() {
		m, err := scanMarket(rows)
		if err != nil {
			return nil, err
		}
		markets = append(markets, *m)
	}
	return markets, rows.Err()
}

func (r *MarketRepository) FindByName(ctx context.Context, name string) (*Market, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+marketColumns+` FROM market WHERE name = ?`, name,
	)
	return scanMarket(row)
}

func (r *MarketRepository) Upsert(ctx context.Context, m *Market) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO market (lostark_market_id, name, grade, icon,
		                     bundle_count, y_day_avg_price, recent_price, current_min_price,
		                     category_code, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
			y_day_avg_price = VALUES(y_day_avg_price),
			recent_price = VALUES(recent_price),
			current_min_price = VALUES(current_min_price),
			bundle_count = VALUES(bundle_count),
			last_modified_date = VALUES(last_modified_date)`,
		m.LostarkMarketID, m.Name, m.Grade, m.Icon,
		m.BundleCount, m.YDayAvgPrice, m.RecentPrice, m.CurrentMinPrice,
		m.CategoryCode, now, now,
	)
	return err
}

func (r *MarketRepository) UpdatePrice(ctx context.Context, id int64, yDayAvgPrice float64, recentPrice int, currentMinPrice int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE market SET y_day_avg_price = ?, recent_price = ?, current_min_price = ?, last_modified_date = ?
		 WHERE market_id = ?`,
		yDayAvgPrice, recentPrice, currentMinPrice, time.Now(), id,
	)
	return err
}

func (r *MarketRepository) FindByCategoryCode(ctx context.Context, categoryCode int) ([]Market, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+marketColumns+` FROM market WHERE category_code = ? ORDER BY name ASC`, categoryCode,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []Market
	for rows.Next() {
		m, err := scanMarket(rows)
		if err != nil {
			return nil, err
		}
		markets = append(markets, *m)
	}
	return markets, rows.Err()
}

func (r *MarketRepository) FindByNames(ctx context.Context, names []string) ([]Market, error) {
	if len(names) == 0 {
		return nil, nil
	}

	query := `SELECT ` + marketColumns + ` FROM market WHERE name IN (`
	args := make([]interface{}, len(names))
	for i, name := range names {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = name
	}
	query += `)`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []Market
	for rows.Next() {
		m, err := scanMarket(rows)
		if err != nil {
			return nil, err
		}
		markets = append(markets, *m)
	}
	return markets, rows.Err()
}
