package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	domainerrors "news-portal/apps/api/internal/pkg/errors"
	"news-portal/apps/api/internal/pkg/response"
	"news-portal/apps/api/internal/usecase"
)

type PublicHandler struct {
	publicUsecase *usecase.PublicUsecase
}

func NewPublicHandler(publicUsecase *usecase.PublicUsecase) *PublicHandler {
	return &PublicHandler{
		publicUsecase: publicUsecase,
	}
}

// GetHomepage handles GET /api/v1/homepage
func (h *PublicHandler) GetHomepage(w http.ResponseWriter, r *http.Request) {
	data, err := h.publicUsecase.GetHomepage(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "HOMEPAGE_FETCH_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, data, nil)
}

// ListArticles handles GET /api/v1/articles
func (h *PublicHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	limit := int32(12)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
		limit = int32(l)
	}

	data, err := h.publicUsecase.ListArticles(r.Context(), cursorStr, limit)
	if err != nil {
		if errors.Is(err, domainerrors.ErrBadRequest) {
			response.Error(w, http.StatusBadRequest, "INVALID_CURSOR", err.Error(), nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "ARTICLES_FETCH_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, data, nil)
}

// GetArticleBySlug handles GET /api/v1/articles/{slug}
func (h *PublicHandler) GetArticleBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	previewToken := r.URL.Query().Get("token")

	data, err := h.publicUsecase.GetArticleBySlug(r.Context(), slug, previewToken)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Article not found", nil)
			return
		}
		if errors.Is(err, domainerrors.ErrUnauthorized) {
			response.Error(w, http.StatusUnauthorized, "INVALID_PREVIEW_TOKEN", "Preview token is invalid or expired", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "ARTICLE_FETCH_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, data, nil)
}

// ListCategories handles GET /api/v1/categories
func (h *PublicHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	data, err := h.publicUsecase.ListCategories(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "CATEGORIES_FETCH_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, data, nil)
}

// ListArticlesByCategory handles GET /api/v1/categories/{slug}/articles
func (h *PublicHandler) ListArticlesByCategory(w http.ResponseWriter, r *http.Request) {
	catSlug := chi.URLParam(r, "slug")
	cursorStr := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	limit := int32(12)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
		limit = int32(l)
	}

	data, err := h.publicUsecase.ListArticlesByCategory(r.Context(), catSlug, cursorStr, limit)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "CATEGORY_ARTICLES_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, data, nil)
}

// ListArticlesByTag handles GET /api/v1/tags/{slug}/articles
func (h *PublicHandler) ListArticlesByTag(w http.ResponseWriter, r *http.Request) {
	tagSlug := chi.URLParam(r, "slug")
	cursorStr := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	limit := int32(12)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
		limit = int32(l)
	}

	data, err := h.publicUsecase.ListArticlesByTag(r.Context(), tagSlug, cursorStr, limit)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "TAG_NOT_FOUND", "Tag not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "TAG_ARTICLES_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, data, nil)
}

// GetAuthor handles GET /api/v1/authors/{slug}
func (h *PublicHandler) GetAuthor(w http.ResponseWriter, r *http.Request) {
	authorSlug := chi.URLParam(r, "slug")
	cursorStr := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	limit := int32(12)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
		limit = int32(l)
	}

	data, err := h.publicUsecase.GetAuthorProfileAndArticles(r.Context(), authorSlug, cursorStr, limit)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "AUTHOR_NOT_FOUND", "Author not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "AUTHOR_FETCH_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, data, nil)
}

// Search handles GET /api/v1/search?q=...
func (h *PublicHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")

	limit := int32(15)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
		limit = int32(l)
	}

	data, err := h.publicUsecase.SearchArticles(r.Context(), q, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "SEARCH_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, data, nil)
}

// GetTrending handles GET /api/v1/trending?window=today|week
func (h *PublicHandler) GetTrending(w http.ResponseWriter, r *http.Request) {
	window := r.URL.Query().Get("window")
	limitStr := r.URL.Query().Get("limit")

	limit := int32(10)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 20 {
		limit = int32(l)
	}

	data, err := h.publicUsecase.GetTrending(r.Context(), window, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "TRENDING_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, data, nil)
}

// GetPage handles GET /api/v1/pages/{slug}
func (h *PublicHandler) GetPage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	data, err := h.publicUsecase.GetPageBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "PAGE_NOT_FOUND", "Page not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "PAGE_FETCH_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, data, nil)
}

// RecordView handles POST /api/v1/articles/{id}/view
func (h *PublicHandler) RecordView(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	articleID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ARTICLE_ID", "Article ID must be a positive integer", nil)
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()

	res, err := h.publicUsecase.RecordView(r.Context(), articleID, ip, ua)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "RECORD_VIEW_FAILED", err.Error(), nil)
		return
	}
	response.Success(w, http.StatusOK, res, nil)
}
