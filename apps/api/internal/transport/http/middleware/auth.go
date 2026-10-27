package middleware

import (
	"context"
	"net/http"
	"strings"

	"news-portal/apps/api/internal/pkg/auth"
	"news-portal/apps/api/internal/pkg/response"
)

type authContextKey string

const UserClaimsKey authContextKey = "user_claims"

// AuthMiddleware creates a JWT authentication middleware verifying the Bearer token.
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header", nil)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Authorization header must be Bearer token", nil)
				return
			}

			claims, err := auth.ValidateAccessToken(jwtSecret, parts[1])
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "INVALID_OR_EXPIRED_TOKEN", err.Error(), nil)
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserClaims extracts the authenticated user claims from context.
func GetUserClaims(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(UserClaimsKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}
