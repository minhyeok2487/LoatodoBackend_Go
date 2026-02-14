package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// CharacterJsonDto represents a character from the Lostark API siblings endpoint.
type CharacterJsonDto struct {
	ServerName         string `json:"ServerName"`
	CharacterName      string `json:"CharacterName"`
	CharacterLevel     int    `json:"CharacterLevel"`
	CharacterClassName string `json:"CharacterClassName"`
	CharacterImage     string `json:"CharacterImage"`
	ItemAvgLevel       string `json:"ItemAvgLevel"` // "1,620.00" format
	CombatPower        string `json:"CombatPower"`  // needs comma removal
}

// ParseItemLevel parses the ItemAvgLevel string (e.g., "1,620.00") into a float64.
func (c *CharacterJsonDto) ParseItemLevel() (float64, error) {
	cleaned := strings.ReplaceAll(c.ItemAvgLevel, ",", "")
	return strconv.ParseFloat(cleaned, 64)
}

// ParseCombatPower parses the CombatPower string into a float64.
func (c *CharacterJsonDto) ParseCombatPower() (float64, error) {
	if c.CombatPower == "" {
		return 0, nil
	}
	cleaned := strings.ReplaceAll(c.CombatPower, ",", "")
	return strconv.ParseFloat(cleaned, 64)
}

// CharacterProfile represents profile data from the armories/profiles endpoint.
type CharacterProfile struct {
	CharacterImage     *string `json:"CharacterImage"`
	CharacterClassName string  `json:"CharacterClassName"`
	CharacterLevel     int     `json:"CharacterLevel"`
	ServerName         string  `json:"ServerName"`
	CharacterName      string  `json:"CharacterName"`
	ItemAvgLevel       string  `json:"ItemAvgLevel"`
	CombatPower        string  `json:"CombatPower"`
}

// LostarkCharacterClient handles character-related Lostark API calls.
type LostarkCharacterClient struct {
	client *LostarkClient
}

// NewLostarkCharacterClient creates a new LostarkCharacterClient.
func NewLostarkCharacterClient(client *LostarkClient) *LostarkCharacterClient {
	return &LostarkCharacterClient{client: client}
}

// FindSiblings gets all characters for an account via representative character name.
// GET https://developer-lostark.game.onstove.com/characters/{name}/siblings
// Only returns characters with ItemAvgLevel >= 1415, sorted by item level descending.
func (c *LostarkCharacterClient) FindSiblings(characterName, apiKey string) ([]CharacterJsonDto, error) {
	encodedName := url.PathEscape(characterName)
	apiURL := fmt.Sprintf("https://developer-lostark.game.onstove.com/characters/%s/siblings", encodedName)

	body, err := c.client.Get(apiURL, apiKey)
	if err != nil {
		return nil, fmt.Errorf("fetching siblings: %w", err)
	}

	var characters []CharacterJsonDto
	if err := json.Unmarshal(body, &characters); err != nil {
		return nil, fmt.Errorf("parsing siblings response: %w", err)
	}

	if characters == nil {
		return nil, fmt.Errorf("존재하지 않는 캐릭터명 입니다")
	}

	// Filter characters with ItemAvgLevel >= 1415
	filtered := make([]CharacterJsonDto, 0, len(characters))
	for _, ch := range characters {
		level, err := ch.ParseItemLevel()
		if err != nil {
			continue
		}
		if level >= 1415.0 {
			filtered = append(filtered, ch)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("아이템 레벨 1415 이상 캐릭터가 없습니다")
	}

	// Sort by item level descending
	sort.Slice(filtered, func(i, j int) bool {
		li, _ := filtered[i].ParseItemLevel()
		lj, _ := filtered[j].ParseItemLevel()
		return li > lj
	})

	return filtered, nil
}

// GetCharacterProfile gets a character's profile details.
// GET https://developer-lostark.game.onstove.com/armories/characters/{name}/profiles
func (c *LostarkCharacterClient) GetCharacterProfile(characterName, apiKey string) (*CharacterProfile, error) {
	encodedName := url.PathEscape(characterName)
	apiURL := fmt.Sprintf("https://developer-lostark.game.onstove.com/armories/characters/%s/profiles", encodedName)

	body, err := c.client.Get(apiURL, apiKey)
	if err != nil {
		return nil, fmt.Errorf("fetching profile: %w", err)
	}

	var profile CharacterProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("parsing profile response: %w", err)
	}

	return &profile, nil
}

// GetCharacterImageURL fetches only the character image URL from the profile endpoint.
func (c *LostarkCharacterClient) GetCharacterImageURL(characterName, apiKey string) (string, error) {
	profile, err := c.GetCharacterProfile(characterName, apiKey)
	if err != nil {
		return "", err
	}
	if profile.CharacterImage == nil {
		return "", nil
	}
	return *profile.CharacterImage, nil
}
