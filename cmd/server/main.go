package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver"
	"github.com/alexburley/ask-howard/internal/adapter/outbound/postgres"
	"github.com/alexburley/ask-howard/internal/auth"
	"github.com/alexburley/ask-howard/internal/service"
)

type config struct {
	DatabaseURL string
	JWTSecret   auth.JWTSecret
}

func loadConfig() config {
	return config{
		DatabaseURL: envOr("DATABASE_URL", "postgres://ask-howard:ask-howard@localhost:5432/ask-howard?sslmode=disable"),
		JWTSecret:   auth.NewJWTSecret(os.Getenv("JWT_SECRET")),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := loadConfig()

	if cfg.JWTSecret.IsZero() {
		cfg.JWTSecret = auth.NewJWTSecret("dev-secret-do-not-use-in-production")
		logger.Warn("JWT_SECRET not set — using insecure default (dev only)")
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo)

	srv := httpserver.NewServer(logger, pool, authSvc, cfg.JWTSecret)
	srv.Serve(ctx)
}
