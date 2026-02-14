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

	resp, err := h.charService.UpdateSettingsAndReturn(r.Context(), username, friendUsername, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ToggleGoldCharacter handles PATCH /api/v1/character/gold-character
func (h *CharacterHandler) ToggleGoldCharacter(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	friendUsername := getQueryParam(r, "friendUsername")

	var req service.GoldCharacterRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	resp, err := h.charService.ToggleGoldCharacterAndReturn(r.Context(), username, friendUsername, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateMemo handles POST /api/v1/character/memo
func (h *CharacterHandler) UpdateMemo(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	friendUsername := getQueryParam(r, "friendUsername")

	var req service.UpdateMemoRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	resp, err := h.charService.UpdateMemoAndReturn(r.Context(), username, friendUsername, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ToggleDeleted handles PATCH /api/v1/character/deleted
func (h *CharacterHandler) ToggleDeleted(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	friendUsername := getQueryParam(r, "friendUsername")

	var req service.DeleteCharacterRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.charService.ToggleDeletedWithFriend(r.Context(), username, friendUsername, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Spring returns empty 200 OK
	w.WriteHeader(http.StatusOK)
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

	friendUsername := getQueryParam(r, "friendUsername")

	var req service.ChangeCharacterNameRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	resp, err := h.charService.ChangeCharacterNameAndReturn(r.Context(), username, friendUsername, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
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
