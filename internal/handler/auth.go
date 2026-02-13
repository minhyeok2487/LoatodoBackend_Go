package handler

import (
	"net/http"

	"lostark-todo-backend/internal/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(as *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: as}
}

// SignUp handles POST /api/v1/auth/signup
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req service.SignUpRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	resp, err := h.authService.SignUp(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginMemberRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	resp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Logout handles GET /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}
