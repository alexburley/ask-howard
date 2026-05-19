package httpserver

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/alexburley/pulse/internal/adapter/inbound/httpserver/handler"
	"github.com/alexburley/pulse/web"
	"github.com/nickbryan/httputil"
)

// HealthChecker is satisfied by any type that can report database liveness,
// e.g. *pgxpool.Pool.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

func NewServer(logger *slog.Logger, db HealthChecker) *httputil.Server {
	srv := httputil.NewServer(logger)

	srv.Register(httputil.EndpointGroup(handler.HealthEndpoints(db)).WithPrefix("/api")...)
	srv.Register(httputil.Endpoint{
		Method:  http.MethodGet,
		Path:    "/",
		Handler: web.Handler(),
	})

	return srv
}
