package httpserver

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver/handler"
	"github.com/alexburley/ask-howard/internal/port/inbound"
	"github.com/alexburley/ask-howard/web"
	"github.com/nickbryan/httputil"
)

// HealthChecker is satisfied by any type that can report database liveness,
// e.g. *pgxpool.Pool.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

func NewServer(logger *slog.Logger, db HealthChecker, authSvc inbound.AuthService, jwtSecret string) *httputil.Server {
	srv := httputil.NewServer(logger)

	srv.Register(httputil.EndpointGroup(handler.HealthEndpoints(db)).WithPrefix("/api")...)
	srv.Register(httputil.EndpointGroup(handler.AuthEndpoints(authSvc, jwtSecret)).WithPrefix("/api")...)
	srv.Register(httputil.Endpoint{
		Method:  http.MethodGet,
		Path:    "/",
		Handler: web.Handler(),
	})

	return srv
}
