package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/alexburley/pulse/internal/adapter/inbound/httpserver"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	srv := httpserver.NewServer(logger)
	srv.Serve(context.Background())
}
