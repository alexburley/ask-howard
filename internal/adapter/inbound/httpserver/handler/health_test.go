package handler_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexburley/pulse/internal/adapter/inbound/httpserver/handler"
	"github.com/nickbryan/httputil"
)

func TestHealthEndpoint(t *testing.T) {
	srv := httputil.NewServer(slog.New(slog.NewTextHandler(io.Discard, nil)))
	srv.Register(handler.HealthEndpoints()...)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
