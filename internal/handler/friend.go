package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type FriendHandler struct {
	svc *service.FriendService
}

func NewFriendHandler(svc *service.FriendService) *FriendHandler {
	return &FriendHandler{svc: svc}
}

// GetFriends handles GET /api/v1/friend
// Returns the full friends list with character data, settings, and friendship status.
func (h *FriendHandler) GetFriends(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	friends, err := h.svc.GetFriends(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, friends)
}

// SearchCharacter handles GET /api/v1/friend/character/{characterName}
// Searches for characters by name, returning a list of potential friend targets.
// Excludes the requesting user's own characters.
func (h *FriendHandler) SearchCharacter(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	characterName := getPathParam(r, "characterName")
	results, err := h.svc.SearchCharacter(r.Context(), username, characterName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, results)
}

// SendFriendRequest handles POST /api/v1/friend
// Creates a bidirectional friend request (requester=accepted, target=pending).
func (h *FriendHandler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.FriendRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.SendFriendRequest(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// HandleFriendRequest handles POST /api/v1/friend/request
// Accepts, rejects, or deletes a friend request based on category (OK/REJECT/DELETE).
func (h *FriendHandler) HandleFriendRequest(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.FriendRequestAction
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.HandleFriendRequest(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// UpdateSettings handles PATCH /api/v1/friend/settings
// Updates a single friend settings field by name (e.g. "showDayTodo", "checkRaid").
// Returns the updated FriendSettings object.
func (h *FriendHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.FriendSettingsUpdateRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	settings, err := h.svc.UpdateSettings(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

// DeleteFriend handles DELETE /api/v1/friend/{friendId}
// Deletes both sides of the friendship (bidirectional delete).
func (h *FriendHandler) DeleteFriend(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	friendID := parseInt64Param(getPathParam(r, "friendId"))
	if err := h.svc.DeleteFriend(r.Context(), username, friendID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// UpdateSortOrder handles PUT /api/v1/friend/sort
// Updates the display ordering of friends. Items receive ordering 1, 2, 3, ... based on list position.
func (h *FriendHandler) UpdateSortOrder(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.FriendSortRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateSortOrder(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}
