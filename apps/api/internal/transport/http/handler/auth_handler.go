package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"news-portal/apps/api/internal/pkg/response"
	"news-portal/apps/api/internal/transport/http/dto"
	"news-portal/apps/api/internal/transport/http/middleware"
	"news-portal/apps/api/internal/usecase"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
	validator   *validator.Validate
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase, v *validator.Validate) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		validator:   v,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON_BODY", "Request body must be valid JSON", nil)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	authResp, refreshToken, err := h.authUsecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "LOGIN_FAILED", err.Error(), nil)
		return
	}

	// Set HttpOnly refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	response.Success(w, http.StatusOK, authResp, nil)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		response.Error(w, http.StatusUnauthorized, "MISSING_REFRESH_TOKEN", "Refresh token cookie is missing", nil)
		return
	}

	authResp, err := h.authUsecase.Refresh(r.Context(), cookie.Value)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, authResp, nil)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	response.Success(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	}, nil)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated", nil)
		return
	}

	profile, err := h.authUsecase.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, profile, nil)
}
