package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/alexburley/pulse/internal/adapter/inbound/httpserver"
	"github.com/alexburley/pulse/internal/adapter/outbound/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://pulse:pulse@localhost:5432/pulse?sslmode=disable"
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, dbURL)
	if err != nil {
		logger.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	srv := httpserver.NewServer(logger, pool)
	srv.Serve(ctx)
}
