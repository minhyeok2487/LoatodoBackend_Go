package handler

import (
	"net/http"

	"lostark-todo-backend/internal/service"
)

type EmailHandler struct {
	svc *service.EmailService
}

func NewEmailHandler(svc *service.EmailService) *EmailHandler {
	return &EmailHandler{svc: svc}
}

// SendSignUpMail handles POST /api/v1/mail
func (h *EmailHandler) SendSignUpMail(w http.ResponseWriter, r *http.Request) {
	var req service.SendMailRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.svc.SendSignUpMail(r.Context(), req.Mail)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// SendResetPasswordMail handles POST /api/v1/mail/password
func (h *EmailHandler) SendResetPasswordMail(w http.ResponseWriter, r *http.Request) {
	var req service.SendMailRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.SendResetPasswordMail(r.Context(), req.Mail); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

// VerifyEmailCode handles POST /api/v1/mail/auth
func (h *EmailHandler) VerifyEmailCode(w http.ResponseWriter, r *http.Request) {
	var req service.MailCheckRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.svc.CheckMail(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}
