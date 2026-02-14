package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type AnalysisHandler struct {
	svc *service.AnalysisService
}

func NewAnalysisHandler(svc *service.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{svc: svc}
}

func (h *AnalysisHandler) GetAnalyses(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	analyses, err := h.svc.GetAnalyses(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, analyses)
}

func (h *AnalysisHandler) CreateAnalysis(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateAnalysisRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.svc.CreateAnalysis(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}
