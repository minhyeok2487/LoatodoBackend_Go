package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

// MemberHandler handles member-related endpoints.
type MemberHandler struct {
	memberService *service.MemberService
}

// NewMemberHandler creates a new MemberHandler.
func NewMemberHandler(ms *service.MemberService) *MemberHandler {
	return &MemberHandler{memberService: ms}
}

// GetMember handles GET /api/v1/member
func (h *MemberHandler) GetMember(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	resp, err := h.memberService.GetMember(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// SaveCharacter handles POST /api/v1/member/character
func (h *MemberHandler) SaveCharacter(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req struct {
		CharacterName string `json:"characterName"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}
	if req.CharacterName == "" {
		writeError(w, http.StatusBadRequest, "캐릭터명을 입력해주세요.")
		return
	}

	if err := h.memberService.SaveCharacter(r.Context(), username, req.CharacterName); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// UpdatePassword handles POST /api/v1/member/password
func (h *MemberHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.UpdatePasswordRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.memberService.UpdatePassword(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// EditMainCharacter handles PATCH /api/v1/member/main-character
func (h *MemberHandler) EditMainCharacter(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.EditMainCharacterRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.memberService.EditMainCharacter(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// ChangeProvider handles PATCH /api/v1/member/provider
func (h *MemberHandler) ChangeProvider(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.ChangeProviderRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.memberService.ChangeProvider(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// SaveAds handles POST /api/v1/member/ads
func (h *MemberHandler) SaveAds(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	if err := h.memberService.SaveAds(r.Context(), username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// DeleteAllCharacters handles DELETE /api/v1/member/characters
func (h *MemberHandler) DeleteAllCharacters(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	if err := h.memberService.DeleteAllCharacters(r.Context(), username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// UpdateAPIKey handles PATCH /api/v1/member/api-key
func (h *MemberHandler) UpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if username == "" {
		writeError(w, http.StatusUnauthorized, "인증이 필요합니다.")
		return
	}

	var req service.UpdateAPIKeyRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청입니다.")
		return
	}

	if err := h.memberService.UpdateAPIKey(r.Context(), username, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}
