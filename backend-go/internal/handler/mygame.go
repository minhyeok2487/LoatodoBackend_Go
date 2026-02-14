package handler

import (
	"net/http"
	"os"

	"lostark-todo-backend/internal/service"
)

type MyGameHandler struct {
	svc *service.MyGameService
}

func NewMyGameHandler(svc *service.MyGameService) *MyGameHandler {
	return &MyGameHandler{svc: svc}
}

// validateAPIKey checks the X-API-Key header matches the configured API key
func (h *MyGameHandler) validateAPIKey(r *http.Request) bool {
	apiKey := r.Header.Get("X-API-Key")
	expectedKey := os.Getenv("MYGAME_API_KEY")
	if expectedKey == "" {
		expectedKey = "mygame-secret-key" // default for development
	}
	return apiKey != "" && apiKey == expectedKey
}

// ListGames handles GET /api/v1/games (with search, page, limit query params)
func (h *MyGameHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	if !h.validateAPIKey(r) {
		writeJSON(w, http.StatusUnauthorized, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "유효하지 않은 API Key입니다."}})
		return
	}
	search := getQueryParam(r, "search")
	page := parseIntParam(getQueryParam(r, "page"))
	limit := parseIntParam(getQueryParam(r, "limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	result, err := h.svc.SearchGames(r.Context(), search, page, limit)
	if err != nil {
		writeJSON(w, http.StatusOK, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: err.Error()}})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GetGame handles GET /api/v1/games/{id}
func (h *MyGameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	if !h.validateAPIKey(r) {
		writeJSON(w, http.StatusUnauthorized, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "유효하지 않은 API Key입니다."}})
		return
	}
	gameID := parseInt64Param(getPathParam(r, "id"))
	result, err := h.svc.GetGameById(r.Context(), gameID)
	if err != nil {
		writeJSON(w, http.StatusOK, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: err.Error()}})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GetAllGames handles GET /api/v1/games/all
func (h *MyGameHandler) GetAllGames(w http.ResponseWriter, r *http.Request) {
	if !h.validateAPIKey(r) {
		writeJSON(w, http.StatusUnauthorized, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "유효하지 않은 API Key입니다."}})
		return
	}
	result, err := h.svc.GetAllGames(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: err.Error()}})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// CreateGame handles POST /api/v1/games
func (h *MyGameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req service.CreateGameRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "invalid request body"}})
		return
	}
	result, err := h.svc.CreateGame(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: err.Error()}})
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

// ListEvents handles GET /api/v1/events (with filters)
func (h *MyGameHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	if !h.validateAPIKey(r) {
		writeJSON(w, http.StatusUnauthorized, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "유효하지 않은 API Key입니다."}})
		return
	}
	gameIds := getQueryParam(r, "gameIds")
	startDate := getQueryParam(r, "startDate")
	endDate := getQueryParam(r, "endDate")
	eventType := getQueryParam(r, "type")
	page := parseIntParam(getQueryParam(r, "page"))
	limit := parseIntParam(getQueryParam(r, "limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 100
	}

	result, err := h.svc.SearchEvents(r.Context(), gameIds, startDate, endDate, eventType, page, limit)
	if err != nil {
		writeJSON(w, http.StatusOK, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: err.Error()}})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GetEvent handles GET /api/v1/events/{id}
func (h *MyGameHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	if !h.validateAPIKey(r) {
		writeJSON(w, http.StatusUnauthorized, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "유효하지 않은 API Key입니다."}})
		return
	}
	eventID := parseInt64Param(getPathParam(r, "id"))
	result, err := h.svc.GetEventById(r.Context(), eventID)
	if err != nil {
		writeJSON(w, http.StatusOK, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: err.Error()}})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// CreateEvent handles POST /api/v1/events
func (h *MyGameHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req service.CreateEventRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "invalid request body"}})
		return
	}
	result, err := h.svc.CreateEvent(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: err.Error()}})
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

// CreateSuggestion handles POST /api/v1/suggestions
func (h *MyGameHandler) CreateSuggestion(w http.ResponseWriter, r *http.Request) {
	var req service.CreateSuggestionRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "invalid request body"}})
		return
	}
	result, err := h.svc.CreateSuggestion(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: err.Error()}})
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

// UploadGameImage handles POST /api/v1/games/images
func (h *MyGameHandler) UploadGameImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "file too large"}})
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "missing image file"}})
		return
	}
	defer file.Close()

	result, err := h.svc.UploadGameImage(r.Context(), header.Filename)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: err.Error()}})
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

// UploadEventImage handles POST /api/v1/events/images
func (h *MyGameHandler) UploadEventImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "file too large"}})
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: "missing image file"}})
		return
	}
	defer file.Close()

	result, err := h.svc.UploadEventImage(r.Context(), header.Filename)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &service.ApiResponseWrapper{Success: false, Error: &service.ErrorWrapper{Message: err.Error()}})
		return
	}
	writeJSON(w, http.StatusCreated, result)
}
