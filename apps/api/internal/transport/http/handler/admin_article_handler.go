package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"news-portal/apps/api/internal/pkg/response"
	"news-portal/apps/api/internal/transport/http/dto"
	"news-portal/apps/api/internal/transport/http/middleware"
	"news-portal/apps/api/internal/usecase"
)

type AdminArticleHandler struct {
	articleUsecase *usecase.ArticleAdminUsecase
	validator      *validator.Validate
}

func NewAdminArticleHandler(articleUsecase *usecase.ArticleAdminUsecase, v *validator.Validate) *AdminArticleHandler {
	return &AdminArticleHandler{
		articleUsecase: articleUsecase,
		validator:      v,
	}
}

func (h *AdminArticleHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var req dto.ArticleUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON_BODY", err.Error(), nil)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	resp, err := h.articleUsecase.Create(r.Context(), &req, claims.UserID, r.RemoteAddr, r.UserAgent())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusCreated, resp, nil)
}

func (h *AdminArticleHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Article ID must be a valid number", nil)
		return
	}

	var req dto.ArticleUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON_BODY", err.Error(), nil)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	resp, err := h.articleUsecase.Update(r.Context(), id, &req, claims.UserID, r.RemoteAddr, r.UserAgent())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, resp, nil)
}

func (h *AdminArticleHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Article ID must be a valid number", nil)
		return
	}

	resp, err := h.articleUsecase.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, resp, nil)
}

func (h *AdminArticleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Article ID must be a valid number", nil)
		return
	}

	if err := h.articleUsecase.SoftDelete(r.Context(), id, claims.UserID, r.RemoteAddr, r.UserAgent()); err != nil {
		response.Error(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "article deleted successfully"}, nil)
}

func (h *AdminArticleHandler) ListRevisions(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Article ID must be a valid number", nil)
		return
	}

	revisions, err := h.articleUsecase.ListRevisions(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_REVISIONS_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, revisions, nil)
}

func (h *AdminArticleHandler) RestoreRevision(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Article ID must be a valid number", nil)
		return
	}

	revIDStr := chi.URLParam(r, "revision_id")
	revID, err := strconv.ParseUint(revIDStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REVISION_ID", "Revision ID must be a valid number", nil)
		return
	}

	resp, err := h.articleUsecase.RestoreRevision(r.Context(), id, revID, claims.UserID, r.RemoteAddr, r.UserAgent())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "RESTORE_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, resp, nil)
}

func (h *AdminArticleHandler) PreviewToken(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Article ID must be a valid number", nil)
		return
	}

	token, err := h.articleUsecase.GeneratePreviewToken(id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "PREVIEW_TOKEN_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"token": token,
		"url":   "/preview?token=" + token,
	}, nil)
}

func (h *AdminArticleHandler) List(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := int32(1)
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = int32(p)
	}

	limit := int32(20)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = int32(l)
	}

	articles, err := h.articleUsecase.List(r.Context(), page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_ARTICLES_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, articles, map[string]interface{}{
		"page":  page,
		"limit": limit,
	})
}

