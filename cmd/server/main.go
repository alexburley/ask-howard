package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver"
	"github.com/alexburley/ask-howard/internal/adapter/outbound/postgres"
	"github.com/alexburley/ask-howard/internal/adapter/outbound/s3"
	"github.com/alexburley/ask-howard/internal/auth"
	"github.com/alexburley/ask-howard/internal/service"
)

type config struct {
	DatabaseURL string
	JWTSecret   auth.JWTSecret
	S3          s3.Config
}

func loadConfig() config {
	return config{
		DatabaseURL: envOr("DATABASE_URL", "postgres://ask-howard:ask-howard@localhost:5432/ask-howard?sslmode=disable"),
		JWTSecret:   auth.NewJWTSecret(os.Getenv("JWT_SECRET")),
		S3: s3.Config{
			Endpoint:        os.Getenv("S3_ENDPOINT"),
			PresignEndpoint: os.Getenv("S3_PRESIGN_ENDPOINT"),
			Bucket:          envOr("S3_BUCKET", "ask-howard-docs"),
			Region:          envOr("S3_REGION", "us-east-1"),
			AccessKey:       os.Getenv("S3_ACCESS_KEY"),
			SecretKey:       os.Getenv("S3_SECRET_KEY"),
			UsePathStyle:    os.Getenv("S3_USE_PATH_STYLE") == "true",
		},
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
	if err := run(logger); err != nil {
		logger.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := loadConfig()

	if cfg.JWTSecret.IsZero() {
		cfg.JWTSecret = auth.NewJWTSecret("dev-secret-do-not-use-in-production")
		logger.Warn("JWT_SECRET not set — using insecure default (dev only)")
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	objectStore, err := s3.NewStore(ctx, &cfg.S3)
	if err != nil {
		return fmt.Errorf("connect to object store: %w", err)
	}

	userRepo := postgres.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo)

	docRepo := postgres.NewDocumentRepository(pool)
	docSvc := service.NewDocumentService(docRepo, objectStore)

	srv := httpserver.NewServer(logger, pool, authSvc, docSvc, cfg.JWTSecret)
	srv.Serve(ctx)
	return nil
}
