package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// MemberService handles member-related business logic.
type MemberService struct {
	db *sql.DB
}

// NewMemberService creates a new MemberService.
func NewMemberService(db *sql.DB) *MemberService {
	return &MemberService{db: db}
}

// MemberResponse is the response DTO for member info.
type MemberResponse struct {
	ID            int64               `json:"id"`
	Username      string              `json:"username"`
	MainCharacter *string             `json:"mainCharacter"`
	APIKey        *string             `json:"apiKey"`
	Characters    []CharacterResponse `json:"characters"`
	AuthProvider  string              `json:"authProvider"`
	AdsDate       *time.Time          `json:"adsDate"`
}

// UpdatePasswordRequest is the request body for password update.
type UpdatePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// EditMainCharacterRequest is the request for updating main character name.
type EditMainCharacterRequest struct {
	MainCharacter string `json:"mainCharacter"`
}

// ChangeProviderRequest is the request for changing auth provider.
type ChangeProviderRequest struct {
	Password string `json:"password"`
}

// UpdateAPIKeyRequest is the request for updating the API key.
type UpdateAPIKeyRequest struct {
	APIKey string `json:"apiKey"`
}

// GetMember returns member info with characters.
func (s *MemberService) GetMember(ctx context.Context, username string) (*MemberResponse, error) {
	var m MemberResponse
	var mainChar, apiKey sql.NullString
	var adsDate sql.NullTime

	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, main_character, api_key, auth_provider, ads_date
		 FROM member WHERE username = ?`, username,
	).Scan(&m.ID, &m.Username, &mainChar, &apiKey, &m.AuthProvider, &adsDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("등록되지 않은 회원입니다.")
		}
		return nil, fmt.Errorf("querying member: %w", err)
	}

	if mainChar.Valid {
		m.MainCharacter = &mainChar.String
	}
	if apiKey.Valid {
		m.APIKey = &apiKey.String
	}
	if adsDate.Valid {
		m.AdsDate = &adsDate.Time
	}

	characters, err := s.getCharactersByMemberID(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	m.Characters = characters

	return &m, nil
}

// SaveCharacter fetches characters from the Lostark API using the member's API key and saves them.
func (s *MemberService) SaveCharacter(ctx context.Context, username string) error {
	var memberID int64
	var apiKey sql.NullString
	err := s.db.QueryRowContext(ctx,
		"SELECT id, api_key FROM member WHERE username = ?", username,
	).Scan(&memberID, &apiKey)
	if err != nil {
		return fmt.Errorf("querying member: %w", err)
	}

	if !apiKey.Valid || apiKey.String == "" {
		return errors.New("API Key가 등록되지 않았습니다.")
	}

	// TODO: Call Lostark API to fetch character list using apiKey
	// This requires the LostarkClient which will be implemented separately.
	// For now, return a placeholder error indicating the client is needed.
	return errors.New("Lostark API client not yet implemented")
}

// UpdatePassword changes the member's password.
func (s *MemberService) UpdatePassword(ctx context.Context, username string, req UpdatePasswordRequest) error {
	var hashedPassword string
	err := s.db.QueryRowContext(ctx,
		"SELECT password FROM member WHERE username = ?", username,
	).Scan(&hashedPassword)
	if err != nil {
		return fmt.Errorf("querying member: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.OldPassword)); err != nil {
		return errors.New("기존 비밀번호가 일치하지 않습니다.")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE member SET password = ?, last_modified_date = ? WHERE username = ?",
		string(newHash), time.Now(), username,
	)
	return err
}

// EditMainCharacter updates the member's main character name.
func (s *MemberService) EditMainCharacter(ctx context.Context, username string, req EditMainCharacterRequest) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE member SET main_character = ?, last_modified_date = ? WHERE username = ?",
		req.MainCharacter, time.Now(), username,
	)
	return err
}

// ChangeProvider changes auth from Google to normal (sets a password).
func (s *MemberService) ChangeProvider(ctx context.Context, username string, req ChangeProviderRequest) error {
	if req.Password == "" {
		return errors.New("비밀번호를 입력해주세요.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE member SET password = ?, auth_provider = 'none', last_modified_date = ? WHERE username = ?",
		string(hashedPassword), time.Now(), username,
	)
	return err
}

// SaveAds records that the user viewed an ad.
func (s *MemberService) SaveAds(ctx context.Context, username string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE member SET ads_date = ?, last_modified_date = ? WHERE username = ?",
		time.Now(), time.Now(), username,
	)
	return err
}

// DeleteAllCharacters removes all characters for the given member.
func (s *MemberService) DeleteAllCharacters(ctx context.Context, username string) error {
	var memberID int64
	err := s.db.QueryRowContext(ctx,
		"SELECT id FROM member WHERE username = ?", username,
	).Scan(&memberID)
	if err != nil {
		return fmt.Errorf("querying member: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		"DELETE FROM character_entity WHERE member_id = ?", memberID,
	)
	return err
}

// UpdateAPIKey validates and updates the member's Lostark API key.
func (s *MemberService) UpdateAPIKey(ctx context.Context, username string, req UpdateAPIKeyRequest) error {
	if req.APIKey == "" {
		return errors.New("API Key를 입력해주세요.")
	}

	// Validate API key by calling Lostark events API
	valid, err := validateLostarkAPIKey(req.APIKey)
	if err != nil {
		return fmt.Errorf("validating API key: %w", err)
	}
	if !valid {
		return errors.New("유효하지 않은 API Key 입니다.")
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE member SET api_key = ?, last_modified_date = ? WHERE username = ?",
		req.APIKey, time.Now(), username,
	)
	return err
}

// validateLostarkAPIKey checks if the API key is valid by calling the Lostark events endpoint.
func validateLostarkAPIKey(apiKey string) (bool, error) {
	req, err := http.NewRequest("GET", "https://developer-lostark.game.onstove.com/news/events", nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return resp.StatusCode == http.StatusOK, nil
}

// getCharactersByMemberID returns all non-deleted characters for a member.
func (s *MemberService) getCharactersByMemberID(ctx context.Context, memberID int64) ([]CharacterResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.server_name, c.character_name, c.character_level,
		        c.character_class_name, c.character_image, c.item_level,
		        c.combat_power, c.sort_number, c.memo, c.gold_character,
		        c.challenge_guardian, c.challenge_abyss
		 FROM character_entity c
		 WHERE c.member_id = ? AND (c.deleted IS NULL OR c.deleted = false)
		 ORDER BY c.sort_number ASC`, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying characters: %w", err)
	}
	defer rows.Close()

	var characters []CharacterResponse
	for rows.Next() {
		var c CharacterResponse
		var image, memo sql.NullString
		var combatPower sql.NullFloat64
		err := rows.Scan(
			&c.ID, &c.ServerName, &c.CharacterName, &c.CharacterLevel,
			&c.CharacterClassName, &image, &c.ItemLevel,
			&combatPower, &c.SortNumber, &memo, &c.GoldCharacter,
			&c.ChallengeGuardian, &c.ChallengeAbyss,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning character: %w", err)
		}
		if image.Valid {
			c.CharacterImage = &image.String
		}
		if memo.Valid {
			c.Memo = &memo.String
		}
		if combatPower.Valid {
			c.CombatPower = combatPower.Float64
		}
		characters = append(characters, c)
	}

	if characters == nil {
		characters = []CharacterResponse{}
	}
	return characters, rows.Err()
}
