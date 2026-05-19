package handler

import (
	"context"
	"net/http"

	"github.com/nickbryan/httputil"
	"github.com/nickbryan/httputil/problem"
)

type dbPinger interface {
	Ping(ctx context.Context) error
}

func HealthEndpoints(db dbPinger) []httputil.Endpoint {
	return []httputil.Endpoint{
		{
			Method: http.MethodGet,
			Path:   "/health",
			Handler: httputil.NewHandler(func(r httputil.RequestEmpty) (*httputil.Response, error) {
				if err := db.Ping(r.Context()); err != nil {
					return nil, (&problem.DetailedError{
						Type:   "https://ask-howard.io/problems/db-unavailable",
						Title:  "Database Unavailable",
						Status: http.StatusServiceUnavailable,
					}).WithDetail(err.Error())
				}
				return httputil.OK(map[string]string{"status": "OK"})
			}),
		},
	}
}
