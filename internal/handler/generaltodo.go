package handler

import (
	"net/http"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/service"
)

type GeneralTodoHandler struct {
	svc *service.GeneralTodoService
}

func NewGeneralTodoHandler(svc *service.GeneralTodoService) *GeneralTodoHandler {
	return &GeneralTodoHandler{svc: svc}
}

// --- Overview ---

// GetOverview handles GET /api/v1/general-todo
func (h *GeneralTodoHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	overview, err := h.svc.GetOverview(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, overview)
}

// --- Folder ---

// CreateFolder handles POST /api/v1/general-todo/folder
func (h *GeneralTodoHandler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateFolderRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	folder, err := h.svc.CreateFolder(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, folder)
}

// UpdateFolder handles PATCH /api/v1/general-todo/folder/{folderId}
func (h *GeneralTodoHandler) UpdateFolder(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	folderID := parseInt64Param(getPathParam(r, "folderId"))
	var req service.UpdateFolderRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateFolder(r.Context(), username, folderID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "folder updated"})
}

// DeleteFolder handles DELETE /api/v1/general-todo/folder/{folderId}
func (h *GeneralTodoHandler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	folderID := parseInt64Param(getPathParam(r, "folderId"))
	if err := h.svc.DeleteFolder(r.Context(), username, folderID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "folder deleted"})
}

// UpdateFolderSort handles PATCH /api/v1/general-todo/folder/sort
func (h *GeneralTodoHandler) UpdateFolderSort(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var items []service.FolderSortItem
	if err := readJSON(r, &items); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateFolderSort(r.Context(), username, items); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "folder sort updated"})
}

// --- Category ---

// CreateCategory handles POST /api/v1/general-todo/category
func (h *GeneralTodoHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateCategoryRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	category, err := h.svc.CreateCategory(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, category)
}

// UpdateCategory handles PATCH /api/v1/general-todo/category/{categoryId}
func (h *GeneralTodoHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	categoryID := parseInt64Param(getPathParam(r, "categoryId"))
	var req service.UpdateCategoryRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateCategory(r.Context(), username, categoryID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "category updated"})
}

// DeleteCategory handles DELETE /api/v1/general-todo/category/{categoryId}
func (h *GeneralTodoHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	categoryID := parseInt64Param(getPathParam(r, "categoryId"))
	if err := h.svc.DeleteCategory(r.Context(), username, categoryID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "category deleted"})
}

// UpdateCategorySort handles PATCH /api/v1/general-todo/category/sort
func (h *GeneralTodoHandler) UpdateCategorySort(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var items []service.CategorySortItem
	if err := readJSON(r, &items); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateCategorySort(r.Context(), username, items); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "category sort updated"})
}

// --- Item ---

// CreateItem handles POST /api/v1/general-todo/item
func (h *GeneralTodoHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateItemRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.svc.CreateItem(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

// UpdateItem handles PATCH /api/v1/general-todo/item/{itemId}
func (h *GeneralTodoHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	itemID := parseInt64Param(getPathParam(r, "itemId"))
	var req service.UpdateItemRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateItem(r.Context(), username, itemID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "item updated"})
}

// DeleteItem handles DELETE /api/v1/general-todo/item/{itemId}
func (h *GeneralTodoHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	itemID := parseInt64Param(getPathParam(r, "itemId"))
	if err := h.svc.DeleteItem(r.Context(), username, itemID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "item deleted"})
}

// UpdateItemSort handles PATCH /api/v1/general-todo/item/sort
func (h *GeneralTodoHandler) UpdateItemSort(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var items []service.ItemSortItem
	if err := readJSON(r, &items); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateItemSort(r.Context(), username, items); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "item sort updated"})
}

// --- Status ---

// CreateStatus handles POST /api/v1/general-todo/status
func (h *GeneralTodoHandler) CreateStatus(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	var req service.CreateStatusRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	status, err := h.svc.CreateStatus(r.Context(), username, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, status)
}

// ToggleStatus handles PATCH /api/v1/general-todo/status/{statusId}
func (h *GeneralTodoHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	statusID := parseInt64Param(getPathParam(r, "statusId"))
	if err := h.svc.ToggleStatus(r.Context(), username, statusID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "status toggled"})
}

// DeleteStatus handles DELETE /api/v1/general-todo/status/{statusId}
func (h *GeneralTodoHandler) DeleteStatus(w http.ResponseWriter, r *http.Request) {
	username := auth.GetUsername(r)
	statusID := parseInt64Param(getPathParam(r, "statusId"))
	if err := h.svc.DeleteStatus(r.Context(), username, statusID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "status deleted"})
}
