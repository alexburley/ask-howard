package handler

import (
	"net/http"

	"github.com/nickbryan/httputil"
)

func HealthEndpoints() []httputil.Endpoint {
	return []httputil.Endpoint{
		{
			Method: http.MethodGet,
			Path:   "/health",
			Handler: httputil.NewHandler(func(_ httputil.RequestEmpty) (*httputil.Response, error) {
				return httputil.OK(map[string]string{"status": "ok"})
			}),
		},
	}
}
