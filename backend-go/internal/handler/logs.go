package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type LogsHandler struct {
	svc *service.LogsService
}

func NewLogsHandler(svc *service.LogsService) *LogsHandler {
	return &LogsHandler{svc: svc}
}

// GetLogs handles GET /api/v1/logs
func (h *LogsHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	page := parseIntParam(getQueryParam(r, "page"))
	size := parseIntParam(getQueryParam(r, "size"))
	if size <= 0 {
		size = 20
	}
	logs, err := h.svc.GetLogs(r.Context(), username, page, size)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

// CreateLog handles POST /api/v1/logs
func (h *LogsHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateLogRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	log, err := h.svc.CreateLog(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, log)
}

// DeleteLog handles DELETE /api/v1/logs/{logId}
func (h *LogsHandler) DeleteLog(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	logID := parseInt64Param(getPathParam(r, "logId"))
	if err := h.svc.DeleteLog(r.Context(), username, logID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "log deleted"})
}

// GetGoldSum handles GET /api/v1/logs/gold-sum
func (h *LogsHandler) GetGoldSum(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	sum, err := h.svc.GetGoldSum(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

// GetLogsProfit handles GET /api/v1/logs/profit
func (h *LogsHandler) GetLogsProfit(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	profit, err := h.svc.GetLogsProfit(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, profit)
}
