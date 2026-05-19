//go:build functional

package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexburley/pulse/internal/adapter/inbound/httpserver"
	"github.com/alexburley/pulse/internal/adapter/outbound/postgres"
	"github.com/alexburley/pulse/internal/service"
	"github.com/alexburley/pulse/internal/testutil"
)

func TestHealth_Functional(t *testing.T) {
	pg := testutil.NewPostgresContainer(t)

	pool, err := postgres.NewPool(context.Background(), pg.ConnectionString)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	authSvc := service.NewAuthService(postgres.NewUserRepository(pool))
	srv := httpserver.NewServer(slog.New(slog.NewTextHandler(io.Discard, nil)), pool, authSvc, testJWTSecret)
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)

	resp, err := ts.Client().Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET /api/health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "OK" {
		t.Errorf("body.status = %q, want %q", body["status"], "OK")
	}
}
