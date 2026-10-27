package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"news-portal/apps/api/internal/pkg/response"
)

// Recoverer intercepts panics in handlers and writes a standardized 500 error response.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					reqID := GetRequestID(r.Context())
					stack := string(debug.Stack())

					logger.Error("panic_recovered",
						slog.String("request_id", reqID),
						slog.Any("panic", rvr),
						slog.String("stack", stack),
					)

					response.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred.", nil)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
