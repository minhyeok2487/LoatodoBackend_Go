package db

import (
	"context"
	"database/sql"
)

type CubeRepository struct {
	db *sql.DB
}

func NewCubeRepository(db *sql.DB) *CubeRepository {
	return &CubeRepository{db: db}
}

const cubesColumns = `
	cubes_id, character_id, ban1, ban2, ban3, ban4, ban5,
	unlock1, unlock2, unlock3, unlock4
`

func scanCubes(scanner interface{ Scan(dest ...interface{}) error }) (*Cubes, error) {
	c := &Cubes{}
	err := scanner.Scan(
		&c.ID, &c.CharacterID, &c.Ban1, &c.Ban2, &c.Ban3, &c.Ban4, &c.Ban5,
		&c.Unlock1, &c.Unlock2, &c.Unlock3, &c.Unlock4,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CubeRepository) FindByID(ctx context.Context, id int64) (*Cubes, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+cubesColumns+` FROM cubes WHERE cubes_id = ?`, id,
	)
	return scanCubes(row)
}

func (r *CubeRepository) FindByCharacterID(ctx context.Context, characterID int64) (*Cubes, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+cubesColumns+` FROM cubes WHERE character_id = ?`, characterID,
	)
	return scanCubes(row)
}

func (r *CubeRepository) FindByCharacterIDs(ctx context.Context, characterIDs []int64) ([]Cubes, error) {
	if len(characterIDs) == 0 {
		return nil, nil
	}

	query := `SELECT ` + cubesColumns + ` FROM cubes WHERE character_id IN (`
	args := make([]interface{}, len(characterIDs))
	for i, id := range characterIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += `)`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cubes []Cubes
	for rows.Next() {
		c, err := scanCubes(rows)
		if err != nil {
			return nil, err
		}
		cubes = append(cubes, *c)
	}
	return cubes, rows.Err()
}

func (r *CubeRepository) Create(ctx context.Context, c *Cubes) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO cubes (character_id, ban1, ban2, ban3, ban4, ban5,
		                    unlock1, unlock2, unlock3, unlock4)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.CharacterID, c.Ban1, c.Ban2, c.Ban3, c.Ban4, c.Ban5,
		c.Unlock1, c.Unlock2, c.Unlock3, c.Unlock4,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CubeRepository) Update(ctx context.Context, c *Cubes) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE cubes SET ban1 = ?, ban2 = ?, ban3 = ?, ban4 = ?, ban5 = ?,
		                  unlock1 = ?, unlock2 = ?, unlock3 = ?, unlock4 = ?
		 WHERE cubes_id = ?`,
		c.Ban1, c.Ban2, c.Ban3, c.Ban4, c.Ban5,
		c.Unlock1, c.Unlock2, c.Unlock3, c.Unlock4,
		c.ID,
	)
	return err
}

func (r *CubeRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cubes WHERE cubes_id = ?`, id)
	return err
}

func (r *CubeRepository) DeleteByCharacterID(ctx context.Context, characterID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cubes WHERE character_id = ?`, characterID)
	return err
}

func (r *CubeRepository) UpdateField(ctx context.Context, id int64, field string, value int) error {
	allowedFields := map[string]bool{
		"ban1": true, "ban2": true, "ban3": true, "ban4": true, "ban5": true,
		"unlock1": true, "unlock2": true, "unlock3": true, "unlock4": true,
	}
	if !allowedFields[field] {
		return sql.ErrNoRows
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE cubes SET `+field+` = ? WHERE cubes_id = ?`, value, id,
	)
	return err
}
