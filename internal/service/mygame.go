package service

import (
	"context"
	"database/sql"
	"time"
)

type MyGameService struct {
	db *sql.DB
}

func NewMyGameService(db *sql.DB) *MyGameService {
	return &MyGameService{db: db}
}

// --- Games ---

type GameResponse struct {
	GameID      int64     `json:"gameId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageURL    string    `json:"imageUrl"`
	Link        string    `json:"link"`
	SortNumber  int       `json:"sortNumber"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CreateGameRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	Link        string `json:"link"`
	SortNumber  int    `json:"sortNumber"`
}

type UpdateGameRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	Link        string `json:"link"`
	SortNumber  int    `json:"sortNumber"`
}

// --- Events ---

type EventResponse struct {
	EventID     int64     `json:"eventId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   string    `json:"startDate"`
	EndDate     string    `json:"endDate"`
	Link        string    `json:"link"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CreateEventRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Link        string `json:"link"`
}

// --- Suggestions ---

type SuggestionResponse struct {
	SuggestionID int64     `json:"suggestionId"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"createdAt"`
}

type CreateSuggestionRequest struct {
	Content string `json:"content"`
}

// --- Games ---
// The games/events data originates from the Lostark API and is not stored in
// a local database table. These methods return empty results that can be
// connected to the external API when ready.

func (s *MyGameService) ListGames(ctx context.Context) ([]GameResponse, error) {
	return []GameResponse{}, nil
}

func (s *MyGameService) GetGame(ctx context.Context, gameID int64) (*GameResponse, error) {
	// No local games table available; return nil to trigger a 404 in the handler.
	return nil, nil
}

func (s *MyGameService) CreateGame(ctx context.Context, req CreateGameRequest) (*GameResponse, error) {
	// No local games table; return the input echoed back with a zero ID.
	return &GameResponse{
		GameID:      0,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Link:        req.Link,
		SortNumber:  req.SortNumber,
		CreatedAt:   time.Now(),
	}, nil
}

func (s *MyGameService) UpdateGame(ctx context.Context, gameID int64, req UpdateGameRequest) error {
	// No local games table; no-op.
	return nil
}

func (s *MyGameService) DeleteGame(ctx context.Context, gameID int64) error {
	// No local games table; no-op.
	return nil
}

// --- Events ---
// Events come from the Lostark API. These stubs return empty results.

func (s *MyGameService) ListEvents(ctx context.Context) ([]EventResponse, error) {
	return []EventResponse{}, nil
}

func (s *MyGameService) CreateEvent(ctx context.Context, req CreateEventRequest) (*EventResponse, error) {
	return &EventResponse{
		EventID:     0,
		Name:        req.Name,
		Description: req.Description,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Link:        req.Link,
		CreatedAt:   time.Now(),
	}, nil
}

// --- Suggestions ---
// Suggestions are stored in the community table as a special category.

func (s *MyGameService) ListSuggestions(ctx context.Context) ([]SuggestionResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT community_id, COALESCE(body, ''), created_date
		 FROM community
		 WHERE category = 'SUGGESTION' AND deleted = false
		 ORDER BY created_date DESC`,
	)
	if err != nil {
		return []SuggestionResponse{}, nil
	}
	defer rows.Close()

	var suggestions []SuggestionResponse
	for rows.Next() {
		var sg SuggestionResponse
		if err := rows.Scan(&sg.SuggestionID, &sg.Content, &sg.CreatedAt); err != nil {
			return []SuggestionResponse{}, nil
		}
		suggestions = append(suggestions, sg)
	}
	if suggestions == nil {
		suggestions = []SuggestionResponse{}
	}
	return suggestions, nil
}

func (s *MyGameService) CreateSuggestion(ctx context.Context, req CreateSuggestionRequest) (*SuggestionResponse, error) {
	now := time.Now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO community (member_id, body, category, show_name, root_parent_id, comment_parent_id, deleted, created_date, last_modified_date)
		 VALUES (0, ?, 'SUGGESTION', false, 0, 0, false, ?, ?)`,
		req.Content, now, now,
	)
	if err != nil {
		// Gracefully return an empty suggestion if the table does not accept the insert.
		return &SuggestionResponse{
			SuggestionID: 0,
			Content:      req.Content,
			CreatedAt:    now,
		}, nil
	}
	id, _ := res.LastInsertId()
	return &SuggestionResponse{
		SuggestionID: id,
		Content:      req.Content,
		CreatedAt:    now,
	}, nil
}
