package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/router"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/cache"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/config"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/database"
	"github.com/ISubamariner/guimba-go/backend/pkg/logger"
)

// @title           Guimba-GO API
// @version         1.0
// @description     Social Protection Management Information System API
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Set up structured logger
	log := logger.New()
	slog.SetDefault(log)

	slog.Info("starting server", "env", cfg.App.Env, "port", cfg.App.Port)

	ctx := context.Background()

	// Connect to PostgreSQL
	pgPool, err := database.NewPostgresPool(ctx, cfg.Postgres.DSN)
	if err != nil {
		slog.Error("failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer database.ClosePostgres(pgPool)

	// Run database migrations
	if err := database.RunMigrations(cfg.Postgres.DSN, "migrations"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Connect to MongoDB
	mongoClient, err := database.NewMongoClient(ctx, cfg.Mongo.URI)
	if err != nil {
		slog.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer database.CloseMongo(ctx, mongoClient)

	// Connect to Redis
	redisClient, err := cache.NewRedisClient(ctx, cfg.Redis.Addr, cfg.Redis.Password)
	if err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer cache.CloseRedis(redisClient)

	// Wire handlers
	healthHandler := handler.NewHealthHandler(pgPool, mongoClient, redisClient)

	// TODO: Wire domain repositories, use cases, and handlers here (Phase 4)

	// Set up router
	r := router.NewRouter(healthHandler, cfg.App.FrontendURL)

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server listening", "port", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("server shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}
	slog.Info("server stopped")
}
