package db

import (
	"context"
	"database/sql"
	"time"
)

type AdsRepository struct {
	db *sql.DB
}

func NewAdsRepository(db *sql.DB) *AdsRepository {
	return &AdsRepository{db: db}
}

const adsColumns = `
	ads_id, name, proposer_email, member_id, checked,
	created_date, last_modified_date
`

func scanAds(scanner interface{ Scan(dest ...interface{}) error }) (*Ads, error) {
	a := &Ads{}
	err := scanner.Scan(
		&a.ID, &a.Name, &a.ProposerEmail, &a.MemberID, &a.Checked,
		&a.CreatedDate, &a.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AdsRepository) FindByID(ctx context.Context, id int64) (*Ads, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+adsColumns+` FROM ads WHERE ads_id = ?`, id,
	)
	return scanAds(row)
}

func (r *AdsRepository) FindByMemberID(ctx context.Context, memberID int64) ([]Ads, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+adsColumns+` FROM ads WHERE member_id = ? ORDER BY created_date DESC`, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adsList []Ads
	for rows.Next() {
		a, err := scanAds(rows)
		if err != nil {
			return nil, err
		}
		adsList = append(adsList, *a)
	}
	return adsList, rows.Err()
}

func (r *AdsRepository) FindAll(ctx context.Context) ([]Ads, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+adsColumns+` FROM ads ORDER BY created_date DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adsList []Ads
	for rows.Next() {
		a, err := scanAds(rows)
		if err != nil {
			return nil, err
		}
		adsList = append(adsList, *a)
	}
	return adsList, rows.Err()
}

func (r *AdsRepository) FindUnchecked(ctx context.Context) ([]Ads, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+adsColumns+` FROM ads WHERE checked = false ORDER BY created_date DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adsList []Ads
	for rows.Next() {
		a, err := scanAds(rows)
		if err != nil {
			return nil, err
		}
		adsList = append(adsList, *a)
	}
	return adsList, rows.Err()
}

func (r *AdsRepository) Create(ctx context.Context, a *Ads) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO ads (name, proposer_email, member_id, checked, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		a.Name, a.ProposerEmail, a.MemberID, false, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *AdsRepository) UpdateCheck(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE ads SET checked = true, last_modified_date = ? WHERE ads_id = ?`,
		time.Now(), id,
	)
	return err
}

func (r *AdsRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM ads WHERE ads_id = ?`, id)
	return err
}
