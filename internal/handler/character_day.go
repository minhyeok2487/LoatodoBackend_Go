package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

// CharacterDayHandler handles daily content endpoints.
type CharacterDayHandler struct {
	dayService *service.CharacterDayService
}

// NewCharacterDayHandler creates a new CharacterDayHandler.
func NewCharacterDayHandler(ds *service.CharacterDayService) *CharacterDayHandler {
	return &CharacterDayHandler{dayService: ds}
}

// CheckDayContent handles POST /api/v1/character/day/check
func (h *CharacterDayHandler) CheckDayContent(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.DayCheckRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	resp, err := h.dayService.CheckDayContent(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateGauge handles POST /api/v1/character/day/gauge
func (h *CharacterDayHandler) UpdateGauge(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.DayGaugeRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.dayService.UpdateGauge(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// CheckAllDayContent handles POST /api/v1/character/day/check/all
func (h *CharacterDayHandler) CheckAllDayContent(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.DayCheckAllRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	resp, err := h.dayService.CheckAllDayContent(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// CheckAllDayCharacters handles POST /api/v1/character/day/check/all-characters
func (h *CharacterDayHandler) CheckAllDayCharacters(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.DayCheckAllCharactersRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.dayService.CheckAllDayCharacters(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}
