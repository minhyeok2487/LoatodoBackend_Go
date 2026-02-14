package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// GetNotifications handles GET /api/v1/notification
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	notifications, err := h.svc.GetNotifications(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, notifications)
}

// MarkAsRead handles PATCH /api/v1/notification/{notificationId}
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	notificationID := parseInt64Param(getPathParam(r, "notificationId"))
	if err := h.svc.MarkAsRead(r.Context(), username, notificationID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "notification marked as read"})
}

// DeleteNotification handles DELETE /api/v1/notification/{notificationId}
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	notificationID := parseInt64Param(getPathParam(r, "notificationId"))
	if err := h.svc.DeleteNotification(r.Context(), username, notificationID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "notification deleted"})
}

// DeleteAllRead handles DELETE /api/v1/notification
func (h *NotificationHandler) DeleteAllRead(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if err := h.svc.DeleteAllRead(r.Context(), username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "all read notifications deleted"})
}

// GetRecentNotification handles GET /api/v1/notification/status
func (h *NotificationHandler) GetRecentNotification(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	recent, err := h.svc.GetRecentNotification(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, recent)
}

// MarkAllAsRead handles POST /api/v1/notification/all
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if err := h.svc.MarkAllAsRead(r.Context(), username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "all notifications marked as read"})
}
