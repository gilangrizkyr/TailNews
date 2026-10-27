package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"news-portal/apps/api/internal/pkg/response"
	"news-portal/apps/api/internal/transport/http/dto"
	"news-portal/apps/api/internal/usecase"
)

type AdminRedirectHandler struct {
	redirectUsecase *usecase.RedirectAdminUsecase
	validator       *validator.Validate
}

func NewAdminRedirectHandler(u *usecase.RedirectAdminUsecase, v *validator.Validate) *AdminRedirectHandler {
	return &AdminRedirectHandler{redirectUsecase: u, validator: v}
}

func (h *AdminRedirectHandler) List(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := int32(50)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = int32(l)
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := int32(0)
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = int32(o)
	}

	items, err := h.redirectUsecase.List(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_REDIRECTS_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, items, nil)
}

func (h *AdminRedirectHandler) CreateOrUpdate(w http.ResponseWriter, r *http.Request) {
	var req dto.RedirectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON_BODY", err.Error(), nil)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := h.redirectUsecase.CreateOrUpdate(r.Context(), &req); err != nil {
		response.Error(w, http.StatusInternalServerError, "SAVE_REDIRECT_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "redirect saved successfully"}, nil)
}

// ==========================================
// Dashboard Handler
// ==========================================
type AdminDashboardHandler struct {
	dashboardUsecase *usecase.DashboardAdminUsecase
}

func NewAdminDashboardHandler(dashboardUsecase *usecase.DashboardAdminUsecase) *AdminDashboardHandler {
	return &AdminDashboardHandler{dashboardUsecase: dashboardUsecase}
}

func (h *AdminDashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.dashboardUsecase.GetStats(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_STATS_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, stats, nil)
}
