package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type CustomTodoHandler struct {
	svc *service.CustomTodoService
}

func NewCustomTodoHandler(svc *service.CustomTodoService) *CustomTodoHandler {
	return &CustomTodoHandler{svc: svc}
}

// GetCustomTodos handles GET /api/v1/custom-todo/{characterId}
func (h *CustomTodoHandler) GetCustomTodos(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	characterID := parseInt64Param(getPathParam(r, "characterId"))
	todos, err := h.svc.GetCustomTodos(r.Context(), username, characterID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, todos)
}

// CreateCustomTodo handles POST /api/v1/custom-todo
func (h *CustomTodoHandler) CreateCustomTodo(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateCustomTodoRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	todo, err := h.svc.CreateCustomTodo(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

// ToggleCheck handles POST /api/v1/custom/check (Spring Boot sends customTodoId in body)
func (h *CustomTodoHandler) ToggleCheck(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req struct {
		CustomTodoID int64 `json:"customTodoId"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.ToggleCheck(r.Context(), username, req.CustomTodoID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "toggled"})
}

// UpdateCustomTodo handles PATCH /api/v1/custom/{customTodoId}
func (h *CustomTodoHandler) UpdateCustomTodo(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	id := parseInt64Param(getPathParam(r, "customTodoId"))
	var req service.UpdateCustomTodoRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateCustomTodo(r.Context(), username, id, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "custom todo updated"})
}

// DeleteCustomTodo handles DELETE /api/v1/custom/{customTodoId}
func (h *CustomTodoHandler) DeleteCustomTodo(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	id := parseInt64Param(getPathParam(r, "customTodoId"))
	if err := h.svc.DeleteCustomTodo(r.Context(), username, id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "custom todo deleted"})
}
