package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/alexburley/pulse/internal/adapter/inbound/httpserver"
	"github.com/alexburley/pulse/internal/adapter/outbound/postgres"
	"github.com/alexburley/pulse/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://pulse:pulse@localhost:5432/pulse?sslmode=disable"
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
