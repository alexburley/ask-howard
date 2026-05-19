package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver"
	"github.com/alexburley/ask-howard/internal/adapter/outbound/postgres"
	"github.com/alexburley/ask-howard/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://ask-howard:ask-howard@localhost:5432/ask-howard?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Error("JWT_SECRET environment variable is not set")
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, dbURL)
	if err != nil {
		logger.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo)

	srv := httpserver.NewServer(logger, pool, authSvc, jwtSecret)
	srv.Serve(ctx)
}
