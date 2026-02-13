package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type CommentsHandler struct {
	svc *service.CommentsService
}

func NewCommentsHandler(svc *service.CommentsService) *CommentsHandler {
	return &CommentsHandler{svc: svc}
}

// GetComments handles GET /api/v1/comments?communityId=X
func (h *CommentsHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	communityID := parseInt64Param(getQueryParam(r, "communityId"))
	if communityID == 0 {
		writeError(w, http.StatusBadRequest, "communityId is required")
		return
	}
	comments, err := h.svc.GetComments(r.Context(), communityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, comments)
}

// CreateComment handles POST /api/v1/comments
func (h *CommentsHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateCommentRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	comment, err := h.svc.CreateComment(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, comment)
}

// UpdateComment handles PATCH /api/v1/comments/{commentId}
func (h *CommentsHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	commentID := parseInt64Param(getPathParam(r, "commentId"))
	var req service.UpdateCommentRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateComment(r.Context(), username, commentID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "comment updated"})
}

// DeleteComment handles DELETE /api/v1/comments/{commentId}
func (h *CommentsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	commentID := parseInt64Param(getPathParam(r, "commentId"))
	if err := h.svc.DeleteComment(r.Context(), username, commentID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "comment deleted"})
}
