package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"news-portal/apps/api/internal/pkg/response"
)

// HealthHandler manages liveness and readiness probe checks.
type HealthHandler struct {
	db  *sql.DB
	rdb *redis.Client
}

// NewHealthHandler constructs a HealthHandler instance.
func NewHealthHandler(db *sql.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:  db,
		rdb: rdb,
	}
}

// Healthz handles liveness probe requests.
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	}, nil)
}

// Readyz handles readiness probe requests by verifying MySQL and Redis health.
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{
		"database": "ok",
		"redis":    "ok",
	}

	allHealthy := true

	if h.db != nil {
		if err := h.db.PingContext(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			allHealthy = false
		}
	} else {
		checks["database"] = "not_configured"
	}

	if h.rdb != nil {
		if err := h.rdb.Ping(ctx).Err(); err != nil {
			checks["redis"] = "unhealthy: " + err.Error()
			allHealthy = false
		}
	} else {
		checks["redis"] = "not_configured"
	}

	statusCode := http.StatusOK
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	response.Success(w, statusCode, map[string]interface{}{
		"status": checks,
		"time":   time.Now().UTC().Format(time.RFC3339),
	}, nil)
}
