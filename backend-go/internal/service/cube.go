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
	CharacterID   int64   `json:"characterId"`
	CharacterName string  `json:"characterName"`
	ItemLevel     float64 `json:"itemLevel"`
	Ban1          int     `json:"ban1"`
	Ban2          int     `json:"ban2"`
	Ban3          int     `json:"ban3"`
	Ban4          int     `json:"ban4"`
	Ban5          int     `json:"ban5"`
	Unlock1       int     `json:"unlock1"`
	Unlock2       int     `json:"unlock2"`
	Unlock3       int     `json:"unlock3"`
	Unlock4       int     `json:"unlock4"`
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

// CubeReward represents the reward table for each cube ticket tier
type CubeReward struct {
	Name           string `json:"name"`
	Jewelry        int    `json:"jewelry"`
	LeapStone      int    `json:"leapStone"`
	Shilling       int    `json:"shilling"`
	SolarGrace     int    `json:"solarGrace"`
	SolarBlessing  int    `json:"solarBlessing"`
	SolarProtection int   `json:"solarProtection"`
	CardExp        int    `json:"cardExp"`
	JewelryPrice   int    `json:"jewelryPrice"`
	LavasBreath    int    `json:"lavasBreath"`
	GlaciersBreath int    `json:"glaciersBreath"`
}

func (s *CubeService) GetCubeData(ctx context.Context, username string) ([]CubeResponse, error) {
	var memberID int64
	err := s.db.QueryRowContext(ctx, "SELECT member_id FROM member WHERE username = ?", username).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT cu.cubes_id, c.characters_id, c.character_name, c.item_level,
		        cu.ban1, cu.ban2, cu.ban3, cu.ban4, cu.ban5,
		        cu.unlock1, cu.unlock2, cu.unlock3, cu.unlock4
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
		if err := rows.Scan(&cube.CubeID, &cube.CharacterID, &cube.CharacterName, &cube.ItemLevel,
			&cube.Ban1, &cube.Ban2, &cube.Ban3, &cube.Ban4, &cube.Ban5,
			&cube.Unlock1, &cube.Unlock2, &cube.Unlock3, &cube.Unlock4); err != nil {
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

	// Get character item level for the response
	var itemLevel float64
	err = s.db.QueryRowContext(ctx,
		`SELECT item_level FROM characters WHERE characters_id = ?`,
		characterID).Scan(&itemLevel)
	if err != nil {
		itemLevel = 0.0 // Default if not found
	}

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
			CharacterID:   characterID,
			CharacterName: req.CharacterName,
			ItemLevel:     itemLevel,
			Ban1:          req.Ban1,
			Ban2:          req.Ban2,
			Ban3:          req.Ban3,
			Ban4:          0,
			Ban5:          0,
			Unlock1:       0,
			Unlock2:       0,
			Unlock3:       0,
			Unlock4:       0,
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

	// Fetch the full cube data to return
	var cube CubeResponse
	err = s.db.QueryRowContext(ctx,
		`SELECT cu.cubes_id, c.characters_id, c.character_name, c.item_level,
		        cu.ban1, cu.ban2, cu.ban3, cu.ban4, cu.ban5,
		        cu.unlock1, cu.unlock2, cu.unlock3, cu.unlock4
		 FROM cubes cu
		 JOIN characters c ON cu.character_id = c.characters_id
		 WHERE cu.cubes_id = ?`,
		existingID).Scan(&cube.CubeID, &cube.CharacterID, &cube.CharacterName, &cube.ItemLevel,
		&cube.Ban1, &cube.Ban2, &cube.Ban3, &cube.Ban4, &cube.Ban5,
		&cube.Unlock1, &cube.Unlock2, &cube.Unlock3, &cube.Unlock4)
	if err != nil {
		return nil, fmt.Errorf("fetching updated cube: %w", err)
	}

	return &cube, nil
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

// GetCubeStatistics returns the static cube reward table matching Spring's response
func (s *CubeService) GetCubeStatistics(ctx context.Context) ([]CubeReward, error) {
	// Static reward table matching Spring's CubeService.getCubeStatistics()
	rewards := []CubeReward{
		{Name: "1금제", Jewelry: 21, LeapStone: 20, Shilling: 79859, SolarGrace: 6, SolarBlessing: 3, SolarProtection: 1, CardExp: 3000, JewelryPrice: 13, LavasBreath: 0, GlaciersBreath: 0},
		{Name: "2금제", Jewelry: 36, LeapStone: 14, Shilling: 100142, SolarGrace: 8, SolarBlessing: 4, SolarProtection: 2, CardExp: 9000, JewelryPrice: 13, LavasBreath: 0, GlaciersBreath: 0},
		{Name: "3금제", Jewelry: 54, LeapStone: 25, Shilling: 110370, SolarGrace: 11, SolarBlessing: 6, SolarProtection: 2, CardExp: 12000, JewelryPrice: 13, LavasBreath: 0, GlaciersBreath: 0},
		{Name: "4금제", Jewelry: 72, LeapStone: 14, Shilling: 120518, SolarGrace: 12, SolarBlessing: 7, SolarProtection: 3, CardExp: 13000, JewelryPrice: 13, LavasBreath: 0, GlaciersBreath: 0},
		{Name: "5금제", Jewelry: 81, LeapStone: 25, Shilling: 129802, SolarGrace: 13, SolarBlessing: 8, SolarProtection: 4, CardExp: 13500, JewelryPrice: 13, LavasBreath: 0, GlaciersBreath: 0},
		{Name: "1해금", Jewelry: 9, LeapStone: 14, Shilling: 140173, SolarGrace: 0, SolarBlessing: 0, SolarProtection: 0, CardExp: 14000, JewelryPrice: 126, LavasBreath: 0, GlaciersBreath: 0},
		{Name: "2해금", Jewelry: 12, LeapStone: 25, Shilling: 151741, SolarGrace: 0, SolarBlessing: 0, SolarProtection: 0, CardExp: 14500, JewelryPrice: 126, LavasBreath: 0, GlaciersBreath: 0},
		{Name: "3해금", Jewelry: 24, LeapStone: 32, Shilling: 161235, SolarGrace: 0, SolarBlessing: 0, SolarProtection: 0, CardExp: 15000, JewelryPrice: 126, LavasBreath: 6, GlaciersBreath: 6},
		{Name: "4해금", Jewelry: 33, LeapStone: 41, Shilling: 0, SolarGrace: 0, SolarBlessing: 0, SolarProtection: 0, CardExp: 15500, JewelryPrice: 126, LavasBreath: 8, GlaciersBreath: 8},
	}
	return rewards, nil
}
