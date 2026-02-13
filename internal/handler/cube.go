package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type CubeHandler struct {
	svc *service.CubeService
}

func NewCubeHandler(svc *service.CubeService) *CubeHandler {
	return &CubeHandler{svc: svc}
}

// GetCubeData handles GET /api/v1/cube
func (h *CubeHandler) GetCubeData(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	data, err := h.svc.GetCubeData(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

// CreateOrUpdateCube handles POST /api/v1/cube
func (h *CubeHandler) CreateOrUpdateCube(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateCubeRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	cube, err := h.svc.CreateOrUpdateCube(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cube)
}

// DeleteCube handles DELETE /api/v1/cube/{cubeId}
func (h *CubeHandler) DeleteCube(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	cubeID := parseInt64Param(getPathParam(r, "cubeId"))
	if err := h.svc.DeleteCube(r.Context(), username, cubeID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "cube entry deleted"})
}

// GetCubeContent handles GET /api/v1/cube/content
func (h *CubeHandler) GetCubeContent(w http.ResponseWriter, r *http.Request) {
	content, err := h.svc.GetCubeContent(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, content)
}

// UpdateCube handles PATCH /api/v1/cube
func (h *CubeHandler) UpdateCube(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.UpdateCubeRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateCube(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "cube updated"})
}

// CalculateProfits handles GET /api/v1/cube/calculate
func (h *CubeHandler) CalculateProfits(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	calculations, err := h.svc.CalculateProfits(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, calculations)
}
