package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/platform"
	transporthttp "news-portal/apps/api/internal/transport/http"
)

func main() {
	// Initialize default logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Adjust log level based on config
	if cfg.LogLevel == "debug" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	logger.Info("starting news portal api server",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.Port),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize MySQL Connection Pool
	db, err := platform.NewMySQLPool(ctx, cfg)
	if err != nil {
		logger.Error("failed to connect to mysql", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("error closing mysql connection", slog.String("error", closeErr.Error()))
		}
	}()
	logger.Info("connected to mysql database successfully")

	// Initialize Redis Client
	rdb, err := platform.NewRedisClient(ctx, cfg)
	if err != nil {
		logger.Error("failed to connect to redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if closeErr := rdb.Close(); closeErr != nil {
			logger.Error("error closing redis connection", slog.String("error", closeErr.Error()))
		}
	}()
	logger.Info("connected to redis cluster successfully")

	// Assemble HTTP Router
	router := transporthttp.NewRouter(transporthttp.RouterConfig{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Redis:  rdb,
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server runner in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("http server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Graceful Shutdown Listening
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("server encountered fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	case sig := <-shutdown:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed; forcing close", slog.String("error", err.Error()))
			_ = server.Close()
		}
		logger.Info("server gracefully stopped")
	}
}
