package http

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/transport/http/handler"
	"news-portal/apps/api/internal/transport/http/middleware"
)

// RouterConfig defines the dependencies required to construct the HTTP router.
type RouterConfig struct {
	Config *config.Config
	Logger *slog.Logger
	DB     *sql.DB
	Redis  *redis.Client
}

// NewRouter builds and returns the fully wired Chi HTTP handler.
func NewRouter(rc RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global Core Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.StructuredLogger(rc.Logger))
	r.Use(middleware.Recoverer(rc.Logger))
	r.Use(middleware.CORS(rc.Config.CORSAllowedOrigins))
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.CleanPath)

	// Liveness & Readiness Probes
	healthHandler := handler.NewHealthHandler(rc.DB, rc.Redis)
	r.Get("/healthz", healthHandler.Healthz)
	r.Get("/readyz", healthHandler.Readyz)

	// API Namespace
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"message":"pong"}}`))
		})
	})

	return r
}
