package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/alexburley/pulse/internal/adapter/inbound/httpserver/handler"
	"github.com/alexburley/pulse/web"
	"github.com/nickbryan/httputil"
)

func NewServer(logger *slog.Logger) *httputil.Server {
	srv := httputil.NewServer(logger)

	srv.Register(httputil.EndpointGroup(handler.HealthEndpoints()).WithPrefix("/api")...)
	srv.Register(httputil.Endpoint{
		Method:  http.MethodGet,
		Path:    "/",
		Handler: web.Handler(),
	})

	return srv
}
