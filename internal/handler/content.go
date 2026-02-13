package handler

import (
	"net/http"

	"lostark-todo-backend/internal/service"
)

type ContentHandler struct {
	svc *service.ContentService
}

func NewContentHandler(svc *service.ContentService) *ContentHandler {
	return &ContentHandler{svc: svc}
}

// GetWeekContent handles GET /api/v1/content/week
func (h *ContentHandler) GetWeekContent(w http.ResponseWriter, r *http.Request) {
	content, err := h.svc.GetWeekContent(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, content)
}
