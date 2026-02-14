package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type CommunityHandler struct {
	svc *service.CommunityService
}

func NewCommunityHandler(svc *service.CommunityService) *CommunityHandler {
	return &CommunityHandler{svc: svc}
}

// ListPosts handles GET /api/v1/community
func (h *CommunityHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	category := getQueryParam(r, "category")
	page := parseIntParam(getQueryParam(r, "page"))
	size := parseIntParam(getQueryParam(r, "size"))
	if size <= 0 {
		size = 20
	}
	posts, err := h.svc.ListPosts(r.Context(), username, category, page, size)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, posts)
}

// GetPost handles GET /api/v1/community/{communityId}
func (h *CommunityHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	communityID := parseInt64Param(getPathParam(r, "communityId"))
	post, err := h.svc.GetPost(r.Context(), username, communityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if post == nil {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	writeJSON(w, http.StatusOK, post)
}

// CreatePost handles POST /api/v1/community
func (h *CommunityHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateCommunityRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	post, err := h.svc.CreatePost(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, post)
}

// UpdatePost handles PATCH /api/v1/community (Spring Boot sends communityId in body)
func (h *CommunityHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	// Try path param first (backward compat), then body
	communityID := parseInt64Param(getPathParam(r, "communityId"))
	var req struct {
		CommunityID int64  `json:"communityId"`
		Category    string `json:"category"`
		Title       string `json:"title"`
		Body        string `json:"body"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if communityID == 0 {
		communityID = req.CommunityID
	}
	updateReq := service.UpdateCommunityRequest{
		Category: req.Category,
		Title:    req.Title,
		Body:     req.Body,
	}
	if err := h.svc.UpdatePost(r.Context(), username, communityID, updateReq); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "post updated"})
}

// DeletePost handles DELETE /api/v1/community/{communityId}
func (h *CommunityHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	communityID := parseInt64Param(getPathParam(r, "communityId"))
	if err := h.svc.DeletePost(r.Context(), username, communityID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "post deleted"})
}

// ToggleLike handles POST /api/v1/community/{communityId}/like
func (h *CommunityHandler) ToggleLike(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	communityID := parseInt64Param(getPathParam(r, "communityId"))
	liked, err := h.svc.ToggleLike(r.Context(), username, communityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"liked": liked})
}

// GetImages handles GET /api/v1/community/{communityId}/images
func (h *CommunityHandler) GetImages(w http.ResponseWriter, r *http.Request) {
	communityID := parseInt64Param(getPathParam(r, "communityId"))
	images, err := h.svc.GetImages(r.Context(), communityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, images)
}

// UploadImage handles POST /api/v1/community/{communityId}/images
func (h *CommunityHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	communityID := parseInt64Param(getPathParam(r, "communityId"))

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		writeError(w, http.StatusBadRequest, "file too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing image file")
		return
	}
	defer file.Close()

	image, err := h.svc.UploadImage(r.Context(), username, communityID, header.Filename, file)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, image)
}

// GetCategories handles GET /api/v1/community/category
func (h *CommunityHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories := []string{"일상", "깐부모집", "길드모집", "고정팟모집", "로투두공지", "로투두건의사항"}
	writeJSON(w, http.StatusOK, categories)
}

// UploadCommunityImage handles POST /api/v1/community/image
func (h *CommunityHandler) UploadCommunityImage(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file too large")
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing image file")
		return
	}
	defer file.Close()
	image, err := h.svc.UploadImage(r.Context(), username, 0, header.Filename, file)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, image)
}
