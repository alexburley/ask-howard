package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/alexburley/ask-howard/internal/adapter/inbound/httpserver/handler"
	"github.com/alexburley/ask-howard/internal/auth"
	"github.com/alexburley/ask-howard/internal/port/inbound"
	"github.com/alexburley/ask-howard/internal/port/outbound"
	"github.com/alexburley/ask-howard/web"
	"github.com/nickbryan/httputil"
)

func NewServer(logger *slog.Logger, db inbound.HealthChecker, authSvc inbound.AuthService, docSvc inbound.DocumentService, store outbound.ObjectStore, jwtSecret auth.JWTSecret) *httputil.Server {
	srv := httputil.NewServer(logger)
	authGuard := handler.NewAuthGuard(jwtSecret)

	srv.Register(httputil.EndpointGroup(handler.HealthEndpoints(db)).WithPrefix("/api")...)
	srv.Register(httputil.EndpointGroup(handler.OpenAuthEndpoints(authSvc, jwtSecret)).WithPrefix("/api")...)
	srv.Register(httputil.EndpointGroup(handler.ProtectedAuthEndpoints(authSvc)).WithGuard(authGuard).WithPrefix("/api")...)
	if docSvc != nil {
		srv.Register(httputil.EndpointGroup(handler.DocumentEndpoints(docSvc, store)).WithGuard(authGuard).WithPrefix("/api")...)
	}
	srv.Register(httputil.Endpoint{
		Method:  http.MethodGet,
		Path:    "/",
		Handler: web.Handler(),
	})

	return srv
}
