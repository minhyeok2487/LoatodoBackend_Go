package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

// CharacterHandler handles character-related endpoints.
type CharacterHandler struct {
	charService *service.CharacterService
}

// NewCharacterHandler creates a new CharacterHandler.
func NewCharacterHandler(cs *service.CharacterService) *CharacterHandler {
	return &CharacterHandler{charService: cs}
}

// UpdateSettings handles PATCH /api/v1/character/settings
func (h *CharacterHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	friendUsername := getQueryParam(r, "friendUsername")

	var req service.UpdateSettingsRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.charService.UpdateSettings(r.Context(), username, friendUsername, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// ToggleGoldCharacter handles PATCH /api/v1/character/gold-character
func (h *CharacterHandler) ToggleGoldCharacter(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.GoldCharacterRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.charService.ToggleGoldCharacter(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// UpdateMemo handles POST /api/v1/character/memo
func (h *CharacterHandler) UpdateMemo(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.UpdateMemoRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.charService.UpdateMemo(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// ToggleDeleted handles PATCH /api/v1/character/deleted
func (h *CharacterHandler) ToggleDeleted(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.DeleteCharacterRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.charService.ToggleDeleted(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// UpdateCharacter handles PUT /api/v1/character
func (h *CharacterHandler) UpdateCharacter(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.UpdateCharacterRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	resp, err := h.charService.UpdateSingleCharacter(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ChangeCharacterName handles PATCH /api/v1/character/name
func (h *CharacterHandler) ChangeCharacterName(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.ChangeCharacterNameRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.charService.ChangeCharacterName(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// AddCharacter handles POST /api/v1/character
func (h *CharacterHandler) AddCharacter(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.AddCharacterRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	resp, err := h.charService.AddCharacter(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
