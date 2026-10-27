package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORS creates a strict Cross-Origin Resource Sharing middleware from allowed origins.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "If-None-Match"},
		ExposedHeaders:   []string{"Link", "ETag", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
