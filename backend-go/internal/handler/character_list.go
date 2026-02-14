package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

// CharacterListHandler handles character-list endpoints.
type CharacterListHandler struct {
	charService *service.CharacterService
}

// NewCharacterListHandler creates a new CharacterListHandler.
func NewCharacterListHandler(cs *service.CharacterService) *CharacterListHandler {
	return &CharacterListHandler{charService: cs}
}

// GetCharacterList handles GET /api/v1/character-list
func (h *CharacterListHandler) GetCharacterList(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	characters, err := h.charService.GetCharacterList(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, characters)
}

// UpdateSorting handles PATCH /api/v1/character-list/sorting
func (h *CharacterListHandler) UpdateSorting(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var sortList []service.SortCharacterRequest
	if err := readJSON(r, &sortList); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.charService.UpdateCharacterSorting(r.Context(), username, sortList); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return updated character list (matches Spring)
	characters, err := h.charService.GetCharacterList(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, characters)
}

// GetDeletedCharacters handles GET /api/v1/character-list/deleted
func (h *CharacterListHandler) GetDeletedCharacters(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	characters, err := h.charService.GetDeletedCharacters(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, characters)
}
