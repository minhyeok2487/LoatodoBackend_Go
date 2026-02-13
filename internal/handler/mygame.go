package handler

import (
	"net/http"

	"lostark-todo-backend/internal/service"
)

type MyGameHandler struct {
	svc *service.MyGameService
}

func NewMyGameHandler(svc *service.MyGameService) *MyGameHandler {
	return &MyGameHandler{svc: svc}
}

// --- Games ---

// ListGames handles GET /api/v1/games
func (h *MyGameHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.svc.ListGames(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, games)
}

// GetGame handles GET /api/v1/games/{gameId}
func (h *MyGameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	gameID := parseInt64Param(getPathParam(r, "gameId"))
	game, err := h.svc.GetGame(r.Context(), gameID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if game == nil {
		writeError(w, http.StatusNotFound, "game not found")
		return
	}
	writeJSON(w, http.StatusOK, game)
}

// CreateGame handles POST /api/v1/games
func (h *MyGameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req service.CreateGameRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	game, err := h.svc.CreateGame(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, game)
}

// UpdateGame handles PATCH /api/v1/games/{gameId}
func (h *MyGameHandler) UpdateGame(w http.ResponseWriter, r *http.Request) {
	gameID := parseInt64Param(getPathParam(r, "gameId"))
	var req service.UpdateGameRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateGame(r.Context(), gameID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "game updated"})
}

// DeleteGame handles DELETE /api/v1/games/{gameId}
func (h *MyGameHandler) DeleteGame(w http.ResponseWriter, r *http.Request) {
	gameID := parseInt64Param(getPathParam(r, "gameId"))
	if err := h.svc.DeleteGame(r.Context(), gameID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "game deleted"})
}

// --- Events ---

// ListEvents handles GET /api/v1/events
func (h *MyGameHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.svc.ListEvents(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}

// CreateEvent handles POST /api/v1/events
func (h *MyGameHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req service.CreateEventRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	event, err := h.svc.CreateEvent(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

// --- Suggestions ---

// ListSuggestions handles GET /api/v1/suggestions
func (h *MyGameHandler) ListSuggestions(w http.ResponseWriter, r *http.Request) {
	suggestions, err := h.svc.ListSuggestions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, suggestions)
}

// CreateSuggestion handles POST /api/v1/suggestions
func (h *MyGameHandler) CreateSuggestion(w http.ResponseWriter, r *http.Request) {
	var req service.CreateSuggestionRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	suggestion, err := h.svc.CreateSuggestion(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, suggestion)
}
