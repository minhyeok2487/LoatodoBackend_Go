package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type ScheduleHandler struct {
	svc *service.ScheduleService
}

func NewScheduleHandler(svc *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{svc: svc}
}

// GetSchedules handles GET /api/v1/schedule?startDate=X&endDate=Y
func (h *ScheduleHandler) GetSchedules(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	startDate := getQueryParam(r, "startDate")
	endDate := getQueryParam(r, "endDate")
	schedules, err := h.svc.GetSchedules(r.Context(), username, startDate, endDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, schedules)
}

// CreateSchedule handles POST /api/v1/schedule
func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateScheduleRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	schedule, err := h.svc.CreateSchedule(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, schedule)
}

// UpdateSchedule handles PATCH /api/v1/schedule/{scheduleId}
func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	scheduleID := parseInt64Param(getPathParam(r, "scheduleId"))
	var req service.UpdateScheduleRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateSchedule(r.Context(), username, scheduleID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "schedule updated"})
}

// DeleteSchedule handles DELETE /api/v1/schedule/{scheduleId}
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	scheduleID := parseInt64Param(getPathParam(r, "scheduleId"))
	if err := h.svc.DeleteSchedule(r.Context(), username, scheduleID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "schedule deleted"})
}

// GetRaidCategories handles GET /api/v1/schedule/raid-categories
func (h *ScheduleHandler) GetRaidCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.GetRaidCategories(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

// GetCharacters handles GET /api/v1/schedule/characters
func (h *ScheduleHandler) GetCharacters(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	characters, err := h.svc.GetCharacters(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, characters)
}

// GetWeekSchedule handles GET /api/v1/schedule/week
func (h *ScheduleHandler) GetWeekSchedule(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	week, err := h.svc.GetWeekSchedule(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, week)
}

// EditFriend handles POST /api/v1/schedule/{scheduleId}/friend
func (h *ScheduleHandler) EditFriend(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	scheduleID := parseInt64Param(getPathParam(r, "scheduleId"))
	var req struct {
		FriendUsername string `json:"friendUsername"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.EditFriend(r.Context(), username, scheduleID, req.FriendUsername); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "schedule friend updated"})
}
