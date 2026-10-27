package test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/platform"
	"news-portal/apps/api/internal/repository/mysql"
	transporthttp "news-portal/apps/api/internal/transport/http"
)

func BenchmarkPublicAPI_Homepage_CacheHit(b *testing.B) {
	cfg, err := config.Load()
	if err != nil {
		b.Fatalf("failed to load config: %v", err)
	}

	tempDir := b.TempDir()
	cfg.MediaStorageDir = tempDir

	ctx := context.Background()
	db, err := platform.NewMySQLPool(ctx, cfg)
	if err != nil {
		b.Fatalf("failed to connect mysql: %v", err)
	}
	defer db.Close()

	rdb, err := platform.NewRedisClient(ctx, cfg)
	if err != nil {
		b.Fatalf("failed to connect redis: %v", err)
	}
	defer rdb.Close()

	storage, _ := platform.NewLocalStorage(tempDir)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := mysql.NewRepository(db)
	cache := platform.NewRedisCache(rdb, logger)
	viewCounter := platform.NewRedisViewCounter(rdb, repo.Article, logger)

	handler := transporthttp.NewRouter(transporthttp.RouterConfig{
		Config:      cfg,
		Logger:      logger,
		DB:          db,
		Redis:       rdb,
		Storage:     storage,
		Cache:       cache,
		ViewCounter: viewCounter,
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()

	// Warmup cache
	warmResp, err := client.Get(server.URL + "/api/v1/homepage")
	if err != nil || warmResp.StatusCode != http.StatusOK {
		b.Fatalf("warmup failed")
	}
	warmResp.Body.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(server.URL + "/api/v1/homepage")
			if err != nil {
				b.Errorf("request failed: %v", err)
				continue
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}
