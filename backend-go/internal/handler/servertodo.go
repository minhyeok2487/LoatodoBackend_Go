package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type ServerTodoHandler struct {
	svc *service.ServerTodoService
}

func NewServerTodoHandler(svc *service.ServerTodoService) *ServerTodoHandler {
	return &ServerTodoHandler{svc: svc}
}

// GetServerTodos handles GET /api/v1/server-todo
func (h *ServerTodoHandler) GetServerTodos(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	todos, err := h.svc.GetServerTodos(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, todos)
}

// ToggleCheck handles POST /api/v1/server-todo/state
func (h *ServerTodoHandler) ToggleCheck(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.ToggleServerTodoRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.ToggleCheck(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "toggled"})
}

// CreateServerTodo handles POST /api/v1/server-todo (admin)
func (h *ServerTodoHandler) CreateServerTodo(w http.ResponseWriter, r *http.Request) {
	var req service.CreateServerTodoRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	todo, err := h.svc.CreateServerTodo(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, todo)
}

// UpdateServerTodo handles PATCH /api/v1/server-todo/{id}
func (h *ServerTodoHandler) UpdateServerTodo(w http.ResponseWriter, r *http.Request) {
	id := parseInt64Param(getPathParam(r, "id"))
	var req service.UpdateServerTodoRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateServerTodo(r.Context(), id, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "server todo updated"})
}

// DeleteServerTodo handles DELETE /api/v1/server-todo/{id}
func (h *ServerTodoHandler) DeleteServerTodo(w http.ResponseWriter, r *http.Request) {
	id := parseInt64Param(getPathParam(r, "id"))
	if err := h.svc.DeleteServerTodo(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "server todo deleted"})
}
