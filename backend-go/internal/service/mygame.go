package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"
)

type MyGameService struct {
	db *sql.DB
}

func NewMyGameService(db *sql.DB) *MyGameService {
	return &MyGameService{db: db}
}

// --- Wrapper Response ---

type ApiResponseWrapper struct {
	Success    bool               `json:"success"`
	Data       interface{}        `json:"data"`
	Pagination *PaginationWrapper `json:"pagination,omitempty"`
	Error      *ErrorWrapper      `json:"error,omitempty"`
}

type PaginationWrapper struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}

type ErrorWrapper struct {
	Message string `json:"message"`
}

// --- Game DTOs ---

type GameResponse struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Image       *string    `json:"image"`
	Color       *string    `json:"color"`
	Description *string    `json:"description"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

type CreateGameRequest struct {
	Name        string  `json:"name"`
	Image       *string `json:"image"`
	Color       *string `json:"color"`
	Description *string `json:"description"`
}

// --- Event DTOs ---

type EventResponse struct {
	ID          int64            `json:"id"`
	Game        *GameResponse    `json:"game"`
	Title       string           `json:"title"`
	Description *string          `json:"description"`
	Type        *string          `json:"type"`
	StartDate   time.Time        `json:"startDate"`
	EndDate     time.Time        `json:"endDate"`
	Images      []EventImageResp `json:"images"`
	Videos      []EventVideoResp `json:"videos"`
	CreatedAt   *time.Time       `json:"createdAt"`
	UpdatedAt   *time.Time       `json:"updatedAt"`
}

type EventImageResp struct {
	ID       int64  `json:"id"`
	URL      string `json:"url"`
	FileName string `json:"fileName"`
	Ordering int    `json:"ordering"`
}

type EventVideoResp struct {
	ID       int64  `json:"id"`
	URL      string `json:"url"`
	Ordering int    `json:"ordering"`
}

type CreateEventRequest struct {
	GameID      int64    `json:"gameId"`
	Title       string   `json:"title"`
	Description *string  `json:"description"`
	Type        *string  `json:"type"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	ImageIds    []int64  `json:"imageIds"`
	Videos      []string `json:"videos"`
}

// --- Suggestion DTOs ---

type SuggestionResponse struct {
	ID        int64      `json:"id"`
	Content   string     `json:"content"`
	CreatedAt *time.Time `json:"createdAt"`
}

type CreateSuggestionRequest struct {
	Content string `json:"content"`
}

// --- Game Methods ---

func (s *MyGameService) SearchGames(ctx context.Context, search string, page, limit int) (*ApiResponseWrapper, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	var countQuery, dataQuery string
	var countArgs, dataArgs []interface{}

	if search != "" {
		countQuery = "SELECT COUNT(*) FROM my_game WHERE name LIKE ?"
		countArgs = []interface{}{"%" + search + "%"}
		dataQuery = "SELECT my_game_id, name, image, color, description, created_date, last_modified_date FROM my_game WHERE name LIKE ? ORDER BY created_date DESC LIMIT ? OFFSET ?"
		dataArgs = []interface{}{"%" + search + "%", limit, offset}
	} else {
		countQuery = "SELECT COUNT(*) FROM my_game"
		dataQuery = "SELECT my_game_id, name, image, color, description, created_date, last_modified_date FROM my_game ORDER BY created_date DESC LIMIT ? OFFSET ?"
		dataArgs = []interface{}{limit, offset}
	}

	if err := s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting games: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("querying games: %w", err)
	}
	defer rows.Close()

	var games []GameResponse
	for rows.Next() {
		var g GameResponse
		if err := rows.Scan(&g.ID, &g.Name, &g.Image, &g.Color, &g.Description, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning game: %w", err)
		}
		games = append(games, g)
	}
	if games == nil {
		games = []GameResponse{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ApiResponseWrapper{
		Success: true,
		Data:    games,
		Pagination: &PaginationWrapper{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *MyGameService) GetAllGames(ctx context.Context) (*ApiResponseWrapper, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT my_game_id, name, image, color, description, created_date, last_modified_date FROM my_game ORDER BY created_date DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("querying all games: %w", err)
	}
	defer rows.Close()

	var games []GameResponse
	for rows.Next() {
		var g GameResponse
		if err := rows.Scan(&g.ID, &g.Name, &g.Image, &g.Color, &g.Description, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning game: %w", err)
		}
		games = append(games, g)
	}
	if games == nil {
		games = []GameResponse{}
	}

	return &ApiResponseWrapper{Success: true, Data: games}, nil
}

func (s *MyGameService) GetGameById(ctx context.Context, gameID int64) (*ApiResponseWrapper, error) {
	var g GameResponse
	err := s.db.QueryRowContext(ctx,
		"SELECT my_game_id, name, image, color, description, created_date, last_modified_date FROM my_game WHERE my_game_id = ?",
		gameID,
	).Scan(&g.ID, &g.Name, &g.Image, &g.Color, &g.Description, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("존재하지 않는 게임입니다.")
		}
		return nil, fmt.Errorf("querying game: %w", err)
	}
	return &ApiResponseWrapper{Success: true, Data: g}, nil
}

func (s *MyGameService) CreateGame(ctx context.Context, req CreateGameRequest) (*ApiResponseWrapper, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("게임 이름은 필수입니다.")
	}
	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO my_game (name, image, color, description, created_date, last_modified_date) VALUES (?, ?, ?, ?, ?, ?)",
		req.Name, req.Image, req.Color, req.Description, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting game: %w", err)
	}
	id, _ := result.LastInsertId()
	g := GameResponse{ID: id, Name: req.Name, Image: req.Image, Color: req.Color, Description: req.Description, CreatedAt: &now, UpdatedAt: &now}
	return &ApiResponseWrapper{Success: true, Data: g}, nil
}

// --- Event Methods ---

func (s *MyGameService) SearchEvents(ctx context.Context, gameIdsStr string, startDate, endDate, eventType string, page, limit int) (*ApiResponseWrapper, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 100
	}
	offset := (page - 1) * limit

	query := `SELECT e.my_event_id, e.my_game_id, e.title, e.description, e.type, e.start_date, e.end_date,
                     e.created_date, e.last_modified_date,
                     g.my_game_id, g.name, g.image, g.color, g.description
              FROM my_event e LEFT JOIN my_game g ON e.my_game_id = g.my_game_id
              ORDER BY e.start_date DESC LIMIT ? OFFSET ?`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("querying events: %w", err)
	}
	defer rows.Close()

	var events []EventResponse
	for rows.Next() {
		var e EventResponse
		var gameID sql.NullInt64
		var gameName, gameImage, gameColor, gameDesc sql.NullString
		var eDesc, eType sql.NullString
		if err := rows.Scan(
			&e.ID, &gameID, &e.Title, &eDesc, &eType, &e.StartDate, &e.EndDate,
			&e.CreatedAt, &e.UpdatedAt,
			&gameID, &gameName, &gameImage, &gameColor, &gameDesc,
		); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		if eDesc.Valid {
			e.Description = &eDesc.String
		}
		if eType.Valid {
			e.Type = &eType.String
		}
		if gameID.Valid {
			e.Game = &GameResponse{ID: gameID.Int64, Name: gameName.String, Image: nilIfInvalid(gameImage), Color: nilIfInvalid(gameColor), Description: nilIfInvalid(gameDesc)}
		}
		// Load images and videos
		e.Images = s.loadEventImages(ctx, e.ID)
		e.Videos = s.loadEventVideos(ctx, e.ID)
		events = append(events, e)
	}
	if events == nil {
		events = []EventResponse{}
	}

	var total int64
	s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM my_event").Scan(&total)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ApiResponseWrapper{
		Success: true, Data: events,
		Pagination: &PaginationWrapper{Total: total, Page: page, Limit: limit, TotalPages: totalPages},
	}, nil
}

func (s *MyGameService) GetEventById(ctx context.Context, eventID int64) (*ApiResponseWrapper, error) {
	var e EventResponse
	var gameID sql.NullInt64
	var gameName, gameImage, gameColor, gameDesc sql.NullString
	var eDesc, eType sql.NullString

	err := s.db.QueryRowContext(ctx,
		`SELECT e.my_event_id, e.my_game_id, e.title, e.description, e.type, e.start_date, e.end_date,
                e.created_date, e.last_modified_date,
                g.my_game_id, g.name, g.image, g.color, g.description
         FROM my_event e LEFT JOIN my_game g ON e.my_game_id = g.my_game_id
         WHERE e.my_event_id = ?`, eventID,
	).Scan(
		&e.ID, &gameID, &e.Title, &eDesc, &eType, &e.StartDate, &e.EndDate,
		&e.CreatedAt, &e.UpdatedAt,
		&gameID, &gameName, &gameImage, &gameColor, &gameDesc,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("존재하지 않는 이벤트입니다.")
		}
		return nil, fmt.Errorf("querying event: %w", err)
	}
	if eDesc.Valid {
		e.Description = &eDesc.String
	}
	if eType.Valid {
		e.Type = &eType.String
	}
	if gameID.Valid {
		e.Game = &GameResponse{ID: gameID.Int64, Name: gameName.String, Image: nilIfInvalid(gameImage), Color: nilIfInvalid(gameColor), Description: nilIfInvalid(gameDesc)}
	}
	e.Images = s.loadEventImages(ctx, e.ID)
	e.Videos = s.loadEventVideos(ctx, e.ID)

	return &ApiResponseWrapper{Success: true, Data: e}, nil
}

func (s *MyGameService) CreateEvent(ctx context.Context, req CreateEventRequest) (*ApiResponseWrapper, error) {
	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO my_event (my_game_id, title, description, type, start_date, end_date, created_date, last_modified_date)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.GameID, req.Title, req.Description, req.Type, req.StartDate, req.EndDate, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting event: %w", err)
	}
	id, _ := result.LastInsertId()

	e := EventResponse{ID: id, Title: req.Title, Description: req.Description, Type: req.Type, CreatedAt: &now, UpdatedAt: &now}
	e.Images = []EventImageResp{}
	e.Videos = []EventVideoResp{}
	return &ApiResponseWrapper{Success: true, Data: e}, nil
}

// --- Suggestion Methods ---

func (s *MyGameService) CreateSuggestion(ctx context.Context, req CreateSuggestionRequest) (*ApiResponseWrapper, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("내용을 입력해주세요.")
	}
	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO user_suggestion (content, created_date, last_modified_date) VALUES (?, ?, ?)",
		req.Content, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting suggestion: %w", err)
	}
	id, _ := result.LastInsertId()
	return &ApiResponseWrapper{
		Success: true,
		Data:    SuggestionResponse{ID: id, Content: req.Content, CreatedAt: &now},
	}, nil
}

// --- Image Upload ---

type ImageUploadResponse struct {
	URL      string `json:"imageUrl"`
	FileName string `json:"fileName"`
}

func (s *MyGameService) UploadGameImage(_ context.Context, filename string) (*ApiResponseWrapper, error) {
	// Placeholder: actual S3 upload would be done here
	url := fmt.Sprintf("/images/games/%s", filename)
	return &ApiResponseWrapper{
		Success: true,
		Data:    ImageUploadResponse{URL: url, FileName: filename},
	}, nil
}

func (s *MyGameService) UploadEventImage(_ context.Context, filename string) (*ApiResponseWrapper, error) {
	// Placeholder: actual S3 upload would be done here
	url := fmt.Sprintf("/images/events/%s", filename)
	return &ApiResponseWrapper{
		Success: true,
		Data:    ImageUploadResponse{URL: url, FileName: filename},
	}, nil
}

// --- Helpers ---

func (s *MyGameService) loadEventImages(ctx context.Context, eventID int64) []EventImageResp {
	rows, err := s.db.QueryContext(ctx,
		"SELECT my_event_image_id, url, COALESCE(file_name, ''), ordering FROM my_event_image WHERE my_event_id = ? ORDER BY ordering",
		eventID,
	)
	if err != nil {
		return []EventImageResp{}
	}
	defer rows.Close()
	var images []EventImageResp
	for rows.Next() {
		var img EventImageResp
		rows.Scan(&img.ID, &img.URL, &img.FileName, &img.Ordering)
		images = append(images, img)
	}
	if images == nil {
		images = []EventImageResp{}
	}
	return images
}

func (s *MyGameService) loadEventVideos(ctx context.Context, eventID int64) []EventVideoResp {
	rows, err := s.db.QueryContext(ctx,
		"SELECT my_event_video_id, url, ordering FROM my_event_video WHERE my_event_id = ? ORDER BY ordering",
		eventID,
	)
	if err != nil {
		return []EventVideoResp{}
	}
	defer rows.Close()
	var videos []EventVideoResp
	for rows.Next() {
		var v EventVideoResp
		rows.Scan(&v.ID, &v.URL, &v.Ordering)
		videos = append(videos, v)
	}
	if videos == nil {
		videos = []EventVideoResp{}
	}
	return videos
}

func nilIfInvalid(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
