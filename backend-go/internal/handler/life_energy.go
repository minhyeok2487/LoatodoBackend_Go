package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type LifeEnergyHandler struct {
	svc *service.LifeEnergyService
}

func NewLifeEnergyHandler(svc *service.LifeEnergyService) *LifeEnergyHandler {
	return &LifeEnergyHandler{svc: svc}
}

// GetLifeEnergies handles GET /api/v1/life-energy
func (h *LifeEnergyHandler) GetLifeEnergies(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	energies, err := h.svc.GetLifeEnergies(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, energies)
}

// CreateLifeEnergy handles POST /api/v1/life-energy
func (h *LifeEnergyHandler) CreateLifeEnergy(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateLifeEnergyRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	energy, err := h.svc.CreateLifeEnergy(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, energy)
}

// UpdateLifeEnergy handles PATCH /api/v1/life-energy/{id}
func (h *LifeEnergyHandler) UpdateLifeEnergy(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	id := parseInt64Param(getPathParam(r, "id"))
	var req service.UpdateLifeEnergyRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateLifeEnergy(r.Context(), username, id, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "life energy updated"})
}

// DeleteLifeEnergy handles DELETE /api/v1/life-energy/{id}
func (h *LifeEnergyHandler) DeleteLifeEnergy(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	id := parseInt64Param(getPathParam(r, "id"))
	if err := h.svc.DeleteLifeEnergy(r.Context(), username, id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "life energy deleted"})
}

// SpendLifeEnergy handles POST /api/v1/life-energy/spend
func (h *LifeEnergyHandler) SpendLifeEnergy(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.SpendLifeEnergyRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.SpendLifeEnergy(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "life energy spent"})
}
