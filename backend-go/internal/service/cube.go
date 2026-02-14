package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type CubeService struct {
	db *sql.DB
}

func NewCubeService(db *sql.DB) *CubeService {
	return &CubeService{db: db}
}

type CubeResponse struct {
	CubeID        int64   `json:"cubeId"`
	CharacterName string  `json:"characterName"`
	Ban1          int     `json:"ban1"`
	Ban2          int     `json:"ban2"`
	Ban3          int     `json:"ban3"`
	Jewelry       int     `json:"jewelry"`
	LeapStone     int     `json:"leapStone"`
	ShillingGold  float64 `json:"shillingGold"`
}

type CreateCubeRequest struct {
	CharacterName string  `json:"characterName"`
	Ban1          int     `json:"ban1"`
	Ban2          int     `json:"ban2"`
	Ban3          int     `json:"ban3"`
	Jewelry       int     `json:"jewelry"`
	LeapStone     int     `json:"leapStone"`
	ShillingGold  float64 `json:"shillingGold"`
}

type UpdateCubeRequest struct {
	CubeID        int64   `json:"cubeId"`
	CharacterName string  `json:"characterName"`
	Ban1          int     `json:"ban1"`
	Ban2          int     `json:"ban2"`
	Ban3          int     `json:"ban3"`
	Jewelry       int     `json:"jewelry"`
	LeapStone     int     `json:"leapStone"`
	ShillingGold  float64 `json:"shillingGold"`
}

type CubeContent struct {
	Name      string `json:"name"`
	MinLevel  int    `json:"minLevel"`
	MaxLevel  int    `json:"maxLevel"`
}

type CubeCalculation struct {
	CharacterName string  `json:"characterName"`
	TotalProfit   float64 `json:"totalProfit"`
}

func (s *CubeService) GetCubeData(ctx context.Context, username string) ([]CubeResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT cu.cubes_id, c.character_name, cu.ban1, cu.ban2, cu.ban3
		 FROM cubes cu
		 JOIN characters c ON cu.character_id = c.characters_id
		 WHERE c.member_id = ? AND (c.is_deleted IS NULL OR c.is_deleted = false)
		 ORDER BY c.sort_number ASC`,
		memberID)
	if err != nil {
		return nil, fmt.Errorf("querying cubes: %w", err)
	}
	defer rows.Close()

	var cubes []CubeResponse
	for rows.Next() {
		var cube CubeResponse
		if err := rows.Scan(&cube.CubeID, &cube.CharacterName, &cube.Ban1, &cube.Ban2, &cube.Ban3); err != nil {
			return nil, fmt.Errorf("scanning cube: %w", err)
		}
		cubes = append(cubes, cube)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating cubes: %w", err)
	}

	if cubes == nil {
		cubes = []CubeResponse{}
	}
	return cubes, nil
}

func (s *CubeService) CreateOrUpdateCube(ctx context.Context, username string, req CreateCubeRequest) (*CubeResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	// Find the character by name belonging to the user
	var characterID int64
	err = s.db.QueryRowContext(ctx,
		`SELECT characters_id FROM characters
		 WHERE character_name = ? AND member_id = ? AND (is_deleted IS NULL OR is_deleted = false)`,
		req.CharacterName, memberID).Scan(&characterID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}

	// Check if cube already exists for this character
	var existingID int64
	err = s.db.QueryRowContext(ctx,
		`SELECT cube_id FROM cubes WHERE character_id = ?`, characterID).Scan(&existingID)

	now := time.Now()
	_ = now

	if err == sql.ErrNoRows {
		// Insert new cube
		result, err := s.db.ExecContext(ctx,
			`INSERT INTO cubes (character_id, ban1, ban2, ban3, ban4, ban5, unlock1, unlock2, unlock3, unlock4)
			 VALUES (?, ?, ?, ?, 0, 0, 0, 0, 0, 0)`,
			characterID, req.Ban1, req.Ban2, req.Ban3)
		if err != nil {
			return nil, fmt.Errorf("inserting cube: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("getting last insert id: %w", err)
		}

		return &CubeResponse{
			CubeID:        id,
			CharacterName: req.CharacterName,
			Ban1:          req.Ban1,
			Ban2:          req.Ban2,
			Ban3:          req.Ban3,
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("checking existing cube: %w", err)
	}

	// Update existing cube
	_, err = s.db.ExecContext(ctx,
		`UPDATE cubes SET ban1 = ?, ban2 = ?, ban3 = ?
		 WHERE cubes_id = ?`,
		req.Ban1, req.Ban2, req.Ban3, existingID)
	if err != nil {
		return nil, fmt.Errorf("updating cube: %w", err)
	}

	return &CubeResponse{
		CubeID:        existingID,
		CharacterName: req.CharacterName,
		Ban1:          req.Ban1,
		Ban2:          req.Ban2,
		Ban3:          req.Ban3,
	}, nil
}

func (s *CubeService) DeleteCube(ctx context.Context, username string, cubeID int64) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	// Verify cube belongs to user's character
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM cubes WHERE cube_id = ? AND character_id IN (
			SELECT characters_id FROM characters WHERE member_id = ?
		)`, cubeID, memberID)
	if err != nil {
		return fmt.Errorf("deleting cube: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("cube not found or not owned by user")
	}
	return nil
}

func (s *CubeService) GetCubeContent(ctx context.Context) ([]CubeContent, error) {
	return []CubeContent{
		{Name: "1금제", MinLevel: 1490, MaxLevel: 1539},
		{Name: "2금제", MinLevel: 1540, MaxLevel: 1579},
		{Name: "3금제", MinLevel: 1580, MaxLevel: 1619},
	}, nil
}

func (s *CubeService) UpdateCube(ctx context.Context, username string, req UpdateCubeRequest) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	// Verify cube belongs to user's character
	var ownerID int64
	err = s.db.QueryRowContext(ctx,
		`SELECT c.member_id FROM cubes cu
		 JOIN characters c ON cu.character_id = c.characters_id
		 WHERE cu.cubes_id = ?`, req.CubeID).Scan(&ownerID)
	if err != nil {
		return fmt.Errorf("cube not found: %w", err)
	}
	if ownerID != memberID {
		return fmt.Errorf("cube not owned by user")
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE cubes SET ban1 = ?, ban2 = ?, ban3 = ?
		 WHERE cubes_id = ?`,
		req.Ban1, req.Ban2, req.Ban3, req.CubeID)
	if err != nil {
		return fmt.Errorf("updating cube: %w", err)
	}
	return nil
}

func (s *CubeService) CalculateProfits(ctx context.Context, username string) ([]CubeCalculation, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT c.character_name, cu.ban1, cu.ban2, cu.ban3
		 FROM cubes cu
		 JOIN characters c ON cu.character_id = c.characters_id
		 WHERE c.member_id = ? AND (c.is_deleted IS NULL OR c.is_deleted = false)
		 ORDER BY c.sort_number ASC`,
		memberID)
	if err != nil {
		return nil, fmt.Errorf("querying cubes for calculation: %w", err)
	}
	defer rows.Close()

	// Market price per cube tier (gold per ticket)
	tierPrices := []float64{80.0, 160.0, 280.0}

	var calculations []CubeCalculation
	for rows.Next() {
		var characterName string
		var ban1, ban2, ban3 int
		if err := rows.Scan(&characterName, &ban1, &ban2, &ban3); err != nil {
			return nil, fmt.Errorf("scanning cube data: %w", err)
		}

		totalProfit := float64(ban1)*tierPrices[0] + float64(ban2)*tierPrices[1] + float64(ban3)*tierPrices[2]
		calculations = append(calculations, CubeCalculation{
			CharacterName: characterName,
			TotalProfit:   totalProfit,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating cube data: %w", err)
	}

	if calculations == nil {
		calculations = []CubeCalculation{}
	}
	return calculations, nil
}

func (s *CubeService) SpendCubeTicket(ctx context.Context, username string, characterID int64) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("querying member: %w", err)
	}
	// Decrement cube ticket count
	_, err = s.db.ExecContext(ctx,
		`UPDATE characters SET cube_ticket = GREATEST(0, cube_ticket - 1), last_modified_date = ?
		 WHERE characters_id = ? AND member_id = ?`,
		time.Now(), characterID, memberID,
	)
	return err
}
