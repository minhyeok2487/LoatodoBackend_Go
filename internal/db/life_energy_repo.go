package db

import (
	"context"
	"database/sql"
	"time"
)

type LifeEnergyRepository struct {
	db *sql.DB
}

func NewLifeEnergyRepository(db *sql.DB) *LifeEnergyRepository {
	return &LifeEnergyRepository{db: db}
}

const lifeEnergyColumns = `
	life_energy_id, member_id, character_name, energy, max_energy, beatrice,
	created_date, last_modified_date
`

func scanLifeEnergy(scanner interface{ Scan(dest ...interface{}) error }) (*LifeEnergy, error) {
	le := &LifeEnergy{}
	err := scanner.Scan(
		&le.ID, &le.MemberID, &le.CharacterName, &le.Energy, &le.MaxEnergy, &le.Beatrice,
		&le.CreatedDate, &le.LastModifiedDate,
	)
	if err != nil {
		return nil, err
	}
	return le, nil
}

func (r *LifeEnergyRepository) FindByID(ctx context.Context, id int64) (*LifeEnergy, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+lifeEnergyColumns+` FROM life_energy WHERE life_energy_id = ?`, id,
	)
	return scanLifeEnergy(row)
}

func (r *LifeEnergyRepository) FindByMemberID(ctx context.Context, memberID int64) ([]LifeEnergy, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+lifeEnergyColumns+` FROM life_energy WHERE member_id = ? ORDER BY life_energy_id ASC`, memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var energies []LifeEnergy
	for rows.Next() {
		le, err := scanLifeEnergy(rows)
		if err != nil {
			return nil, err
		}
		energies = append(energies, *le)
	}
	return energies, rows.Err()
}

func (r *LifeEnergyRepository) Create(ctx context.Context, le *LifeEnergy) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO life_energy (member_id, character_name, energy, max_energy, beatrice,
		                          created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		le.MemberID, le.CharacterName, le.Energy, le.MaxEnergy, le.Beatrice,
		now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *LifeEnergyRepository) Update(ctx context.Context, le *LifeEnergy) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE life_energy SET character_name = ?, energy = ?, max_energy = ?, beatrice = ?,
		        last_modified_date = ?
		 WHERE life_energy_id = ?`,
		le.CharacterName, le.Energy, le.MaxEnergy, le.Beatrice,
		time.Now(), le.ID,
	)
	return err
}

func (r *LifeEnergyRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM life_energy WHERE life_energy_id = ?`, id)
	return err
}

func (r *LifeEnergyRepository) DeleteByMemberID(ctx context.Context, memberID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM life_energy WHERE member_id = ?`, memberID)
	return err
}

func (r *LifeEnergyRepository) SpendEnergy(ctx context.Context, id int64, amount int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE life_energy SET energy = energy - ?, last_modified_date = ? WHERE life_energy_id = ?`,
		amount, time.Now(), id,
	)
	return err
}

func (r *LifeEnergyRepository) BulkAddEnergy(ctx context.Context, memberID int64, amount int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE life_energy SET energy = LEAST(energy + ?, max_energy), last_modified_date = ?
		 WHERE member_id = ?`,
		amount, time.Now(), memberID,
	)
	return err
}

func (r *LifeEnergyRepository) BulkAddEnergyAll(ctx context.Context, amount int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE life_energy SET energy = LEAST(energy + ?, max_energy), last_modified_date = ?`,
		amount, time.Now(),
	)
	return err
}
