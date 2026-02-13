package handler

import (
	"net/http"

	"lostark-todo-backend/internal/service"
)

type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// GetDashboard handles GET /admin/dashboard
func (h *AdminHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.svc.GetDashboard(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}

// ListMembers handles GET /admin/member
func (h *AdminHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	search := getQueryParam(r, "search")
	page := parseIntParam(getQueryParam(r, "page"))
	size := parseIntParam(getQueryParam(r, "size"))
	if size <= 0 {
		size = 20
	}
	members, err := h.svc.ListMembers(r.Context(), search, page, size)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, members)
}

// GetMemberDetail handles GET /admin/member/{memberId}
func (h *AdminHandler) GetMemberDetail(w http.ResponseWriter, r *http.Request) {
	memberID := parseInt64Param(getPathParam(r, "memberId"))
	member, err := h.svc.GetMemberDetail(r.Context(), memberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if member == nil {
		writeError(w, http.StatusNotFound, "member not found")
		return
	}
	writeJSON(w, http.StatusOK, member)
}

// UpdateMember handles PATCH /admin/member/{memberId}
func (h *AdminHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	memberID := parseInt64Param(getPathParam(r, "memberId"))
	var req service.UpdateMemberRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateMember(r.Context(), memberID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "member updated"})
}

// DeleteMember handles DELETE /admin/member/{memberId}
func (h *AdminHandler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	memberID := parseInt64Param(getPathParam(r, "memberId"))
	if err := h.svc.DeleteMember(r.Context(), memberID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "member deleted"})
}

// ListCharacters handles GET /admin/character
func (h *AdminHandler) ListCharacters(w http.ResponseWriter, r *http.Request) {
	page := parseIntParam(getQueryParam(r, "page"))
	size := parseIntParam(getQueryParam(r, "size"))
	if size <= 0 {
		size = 20
	}
	characters, err := h.svc.ListCharacters(r.Context(), page, size)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, characters)
}

// GetCharacterDetail handles GET /admin/character/{characterId}
func (h *AdminHandler) GetCharacterDetail(w http.ResponseWriter, r *http.Request) {
	characterID := parseInt64Param(getPathParam(r, "characterId"))
	character, err := h.svc.GetCharacterDetail(r.Context(), characterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if character == nil {
		writeError(w, http.StatusNotFound, "character not found")
		return
	}
	writeJSON(w, http.StatusOK, character)
}

// UpdateCharacter handles PATCH /admin/character/{characterId}
func (h *AdminHandler) UpdateCharacter(w http.ResponseWriter, r *http.Request) {
	characterID := parseInt64Param(getPathParam(r, "characterId"))
	var req service.AdminUpdateCharacterRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateCharacter(r.Context(), characterID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "character updated"})
}

// ListFriends handles GET /admin/friends
func (h *AdminHandler) ListFriends(w http.ResponseWriter, r *http.Request) {
	friends, err := h.svc.ListFriends(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, friends)
}

// ListComments handles GET /admin/comments
func (h *AdminHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	comments, err := h.svc.ListAllComments(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, comments)
}

// DeleteComment handles DELETE /admin/comments/{commentId}
func (h *AdminHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentID := parseInt64Param(getPathParam(r, "commentId"))
	if err := h.svc.DeleteComment(r.Context(), commentID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "comment deleted"})
}

// ListNotifications handles GET /admin/notification
func (h *AdminHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	notifications, err := h.svc.ListAllNotifications(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, notifications)
}

// CreateNotification handles POST /admin/notification
func (h *AdminHandler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	var req service.CreateNotificationRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.CreateNotification(r.Context(), req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "notification created"})
}

// ListContent handles GET /admin/content
func (h *AdminHandler) ListContent(w http.ResponseWriter, r *http.Request) {
	content, err := h.svc.ListContent(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, content)
}

// UpdateContent handles PATCH /admin/content/{contentId}
func (h *AdminHandler) UpdateContent(w http.ResponseWriter, r *http.Request) {
	contentID := parseInt64Param(getPathParam(r, "contentId"))
	var req service.UpdateContentRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateContent(r.Context(), contentID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "content updated"})
}

// ManageAds handles POST /admin/ads
func (h *AdminHandler) ManageAds(w http.ResponseWriter, r *http.Request) {
	var req service.ManageAdsRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.ManageAds(r.Context(), req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "ads updated"})
}
