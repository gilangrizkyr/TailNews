package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"news-portal/apps/api/internal/pkg/response"
	"news-portal/apps/api/internal/transport/http/dto"
	"news-portal/apps/api/internal/usecase"
)

type AdminTaxonomyHandler struct {
	usecase   *usecase.TaxonomyAdminUsecase
	validator *validator.Validate
}

func NewAdminTaxonomyHandler(u *usecase.TaxonomyAdminUsecase, v *validator.Validate) *AdminTaxonomyHandler {
	return &AdminTaxonomyHandler{usecase: u, validator: v}
}

// Categories
func (h *AdminTaxonomyHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.usecase.ListCategories(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_CATEGORIES_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, cats, nil)
}

func (h *AdminTaxonomyHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req dto.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON_BODY", err.Error(), nil)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}
	resp, err := h.usecase.CreateCategory(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "CREATE_CATEGORY_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusCreated, resp, nil)
}

func (h *AdminTaxonomyHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Category ID must be a number", nil)
		return
	}

	var req dto.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON_BODY", err.Error(), nil)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := h.usecase.UpdateCategory(r.Context(), id, &req); err != nil {
		response.Error(w, http.StatusInternalServerError, "UPDATE_CATEGORY_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"message": "category updated successfully"}, nil)
}

func (h *AdminTaxonomyHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Category ID must be a number", nil)
		return
	}

	if err := h.usecase.DeleteCategory(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "DELETE_CATEGORY_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"message": "category deleted successfully"}, nil)
}

// Tags
func (h *AdminTaxonomyHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.usecase.ListTags(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_TAGS_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, tags, nil)
}

func (h *AdminTaxonomyHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req dto.TagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON_BODY", err.Error(), nil)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}
	resp, err := h.usecase.CreateTag(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "CREATE_TAG_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusCreated, resp, nil)
}

func (h *AdminTaxonomyHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Tag ID must be a number", nil)
		return
	}

	if err := h.usecase.DeleteTag(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "DELETE_TAG_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"message": "tag deleted successfully"}, nil)
}

// Authors
func (h *AdminTaxonomyHandler) ListAuthors(w http.ResponseWriter, r *http.Request) {
	authors, err := h.usecase.ListAuthors(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_AUTHORS_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, authors, nil)
}

func (h *AdminTaxonomyHandler) CreateAuthor(w http.ResponseWriter, r *http.Request) {
	var req dto.AuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON_BODY", err.Error(), nil)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}
	resp, err := h.usecase.CreateAuthor(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "CREATE_AUTHOR_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusCreated, resp, nil)
}

func (h *AdminTaxonomyHandler) UpdateAuthor(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Author ID must be a number", nil)
		return
	}

	var req dto.AuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON_BODY", err.Error(), nil)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := h.usecase.UpdateAuthor(r.Context(), id, &req); err != nil {
		response.Error(w, http.StatusInternalServerError, "UPDATE_AUTHOR_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"message": "author updated successfully"}, nil)
}
