package db

import (
	"context"
	"database/sql"
)

type KeyValueRepository struct {
	db *sql.DB
}

func NewKeyValueRepository(db *sql.DB) *KeyValueRepository {
	return &KeyValueRepository{db: db}
}

func (r *KeyValueRepository) FindByID(ctx context.Context, id int64) (*KeyValue, error) {
	kv := &KeyValue{}
	err := r.db.QueryRowContext(ctx,
		`SELECT key_value_id, key_name, value_name FROM key_value WHERE key_value_id = ?`, id,
	).Scan(&kv.ID, &kv.KeyName, &kv.ValueName)
	if err != nil {
		return nil, err
	}
	return kv, nil
}

func (r *KeyValueRepository) FindByKeyName(ctx context.Context, keyName string) (*KeyValue, error) {
	kv := &KeyValue{}
	err := r.db.QueryRowContext(ctx,
		`SELECT key_value_id, key_name, value_name FROM key_value WHERE key_name = ?`, keyName,
	).Scan(&kv.ID, &kv.KeyName, &kv.ValueName)
	if err != nil {
		return nil, err
	}
	return kv, nil
}

func (r *KeyValueRepository) FindAll(ctx context.Context) ([]KeyValue, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT key_value_id, key_name, value_name FROM key_value ORDER BY key_value_id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var kvs []KeyValue
	for rows.Next() {
		var kv KeyValue
		if err := rows.Scan(&kv.ID, &kv.KeyName, &kv.ValueName); err != nil {
			return nil, err
		}
		kvs = append(kvs, kv)
	}
	return kvs, rows.Err()
}

func (r *KeyValueRepository) Create(ctx context.Context, kv *KeyValue) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO key_value (key_name, value_name) VALUES (?, ?)`,
		kv.KeyName, kv.ValueName,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *KeyValueRepository) UpdateValue(ctx context.Context, id int64, valueName string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE key_value SET value_name = ? WHERE key_value_id = ?`, valueName, id,
	)
	return err
}

func (r *KeyValueRepository) UpdateValueByKeyName(ctx context.Context, keyName, valueName string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE key_value SET value_name = ? WHERE key_name = ?`, valueName, keyName,
	)
	return err
}

func (r *KeyValueRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM key_value WHERE key_value_id = ?`, id)
	return err
}

func (r *KeyValueRepository) Upsert(ctx context.Context, keyName, valueName string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO key_value (key_name, value_name) VALUES (?, ?)
		 ON DUPLICATE KEY UPDATE value_name = VALUES(value_name)`,
		keyName, valueName,
	)
	return err
}
