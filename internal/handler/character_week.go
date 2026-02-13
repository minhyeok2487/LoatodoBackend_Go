package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

// CharacterWeekHandler handles weekly content endpoints.
type CharacterWeekHandler struct {
	weekService *service.CharacterWeekService
}

// NewCharacterWeekHandler creates a new CharacterWeekHandler.
func NewCharacterWeekHandler(ws *service.CharacterWeekService) *CharacterWeekHandler {
	return &CharacterWeekHandler{weekService: ws}
}

// AddRemoveWeekRaid handles POST /api/v1/character/week/raid
func (h *CharacterWeekHandler) AddRemoveWeekRaid(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekRaidRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.AddRemoveWeekRaid(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// GetWeekRaidForm handles GET /api/v1/character/week/raid/form?characterId=X
func (h *CharacterWeekHandler) GetWeekRaidForm(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	characterID := parseInt64Param(getQueryParam(r, "characterId"))
	if characterID == 0 {
		writeError(w, http.StatusBadRequest, "characterId가 필요합니다.")
		return
	}

	resp, err := h.weekService.GetWeekRaidForm(r.Context(), username, characterID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateBusGold handles POST /api/v1/character/week/raid/bus
func (h *CharacterWeekHandler) UpdateBusGold(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekRaidBusRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.UpdateBusGold(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// CheckRaid handles POST /api/v1/character/week/raid/check
func (h *CharacterWeekHandler) CheckRaid(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekRaidCheckRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.CheckRaid(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// UpdateRaidMessage handles POST /api/v1/character/week/raid/message
func (h *CharacterWeekHandler) UpdateRaidMessage(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekRaidMessageRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.UpdateRaidMessage(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// SortRaids handles POST /api/v1/character/week/raid/sort
func (h *CharacterWeekHandler) SortRaids(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekRaidSortRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.SortRaids(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// CheckWeekEpona handles POST /api/v1/character/week/epona
func (h *CharacterWeekHandler) CheckWeekEpona(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekEponaRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.CheckWeekEpona(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// ToggleSilmael handles POST /api/v1/character/week/silmael
func (h *CharacterWeekHandler) ToggleSilmael(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekSilmaelRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.ToggleSilmael(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// UpdateCubeTicket handles POST /api/v1/character/week/cube
func (h *CharacterWeekHandler) UpdateCubeTicket(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekCubeRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.UpdateCubeTicket(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// ToggleGoldCheck handles PATCH /api/v1/character/week/raid/gold-check
func (h *CharacterWeekHandler) ToggleGoldCheck(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekGoldCheckRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.ToggleGoldCheck(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// ToggleGoldCheckVersion handles PATCH /api/v1/character/week/gold-check-version
func (h *CharacterWeekHandler) ToggleGoldCheckVersion(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekGoldCheckVersionRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.ToggleGoldCheckVersion(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// ToggleMoreReward handles POST /api/v1/character/week/raid/more-reward
func (h *CharacterWeekHandler) ToggleMoreReward(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekMoreRewardRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.ToggleMoreReward(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// UpdateElysian handles POST /api/v1/character/week/elysian
func (h *CharacterWeekHandler) UpdateElysian(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekElysianRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.UpdateElysian(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// ToggleElysianAll handles POST /api/v1/character/week/elysian/all
func (h *CharacterWeekHandler) ToggleElysianAll(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.WeekElysianAllRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.weekService.ToggleElysianAll(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}
