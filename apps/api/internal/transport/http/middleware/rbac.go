package middleware

import (
	"net/http"

	"news-portal/apps/api/internal/pkg/response"
)

// RequireRoles enforces that the authenticated user possesses one of the allowed roles.
func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	allowedMap := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		allowedMap[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
				return
			}

			// Admin always has bypass access
			if claims.RoleName == "admin" || allowedMap[claims.RoleName] {
				next.ServeHTTP(w, r)
				return
			}

			response.Error(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource", nil)
		})
	}
}
