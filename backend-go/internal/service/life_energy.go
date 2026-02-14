package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type LifeEnergyService struct {
	db *sql.DB
}

func NewLifeEnergyService(db *sql.DB) *LifeEnergyService {
	return &LifeEnergyService{db: db}
}

type LifeEnergyResponse struct {
	ID            int64  `json:"id"`
	CharacterName string `json:"characterName"`
	CurrentEnergy int    `json:"currentEnergy"`
	MaxEnergy     int    `json:"maxEnergy"`
}

type CreateLifeEnergyRequest struct {
	CharacterName string `json:"characterName"`
	CurrentEnergy int    `json:"currentEnergy"`
	MaxEnergy     int    `json:"maxEnergy"`
}

type UpdateLifeEnergyRequest struct {
	CharacterName string `json:"characterName"`
	CurrentEnergy int    `json:"currentEnergy"`
	MaxEnergy     int    `json:"maxEnergy"`
}

func (s *LifeEnergyService) GetLifeEnergies(ctx context.Context, username string) ([]LifeEnergyResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT life_energy_id, character_name, energy, max_energy
		 FROM life_energy
		 WHERE member_id = ?`,
		memberID)
	if err != nil {
		return nil, fmt.Errorf("querying life energies: %w", err)
	}
	defer rows.Close()

	var energies []LifeEnergyResponse
	for rows.Next() {
		var e LifeEnergyResponse
		if err := rows.Scan(&e.ID, &e.CharacterName, &e.CurrentEnergy, &e.MaxEnergy); err != nil {
			return nil, fmt.Errorf("scanning life energy: %w", err)
		}
		energies = append(energies, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating life energies: %w", err)
	}

	if energies == nil {
		energies = []LifeEnergyResponse{}
	}
	return energies, nil
}

func (s *LifeEnergyService) CreateLifeEnergy(ctx context.Context, username string, req CreateLifeEnergyRequest) (*LifeEnergyResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO life_energy (member_id, character_name, energy, max_energy, beatrice, created_date, last_modified_date)
		 VALUES (?, ?, ?, ?, false, ?, ?)`,
		memberID, req.CharacterName, req.CurrentEnergy, req.MaxEnergy, now, now)
	if err != nil {
		return nil, fmt.Errorf("inserting life energy: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}

	return &LifeEnergyResponse{
		ID:            id,
		CharacterName: req.CharacterName,
		CurrentEnergy: req.CurrentEnergy,
		MaxEnergy:     req.MaxEnergy,
	}, nil
}

func (s *LifeEnergyService) UpdateLifeEnergy(ctx context.Context, username string, id int64, req UpdateLifeEnergyRequest) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE life_energy SET character_name = ?, energy = ?, max_energy = ?, last_modified_date = ?
		 WHERE life_energy_id = ? AND member_id = ?`,
		req.CharacterName, req.CurrentEnergy, req.MaxEnergy, time.Now(), id, memberID)
	if err != nil {
		return fmt.Errorf("updating life energy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("life energy not found or not owned by user")
	}
	return nil
}

func (s *LifeEnergyService) DeleteLifeEnergy(ctx context.Context, username string, id int64) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	result, err := s.db.ExecContext(ctx,
		`DELETE FROM life_energy WHERE life_energy_id = ? AND member_id = ?`,
		id, memberID)
	if err != nil {
		return fmt.Errorf("deleting life energy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("life energy not found or not owned by user")
	}
	return nil
}

type SpendLifeEnergyRequest struct {
	CharacterName string `json:"characterName"`
	SpendAmount   int    `json:"spendAmount"`
}

func (s *LifeEnergyService) SpendLifeEnergy(ctx context.Context, username string, req SpendLifeEnergyRequest) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("querying member: %w", err)
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE life_energy SET current_energy = GREATEST(0, current_energy - ?), last_modified_date = ?
		 WHERE member_id = ? AND character_name = ?`,
		req.SpendAmount, time.Now(), memberID, req.CharacterName,
	)
	return err
}
