//go:build dev

package web

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Handler proxies requests to the local Vite dev server.
// Run with: go run -tags dev ./cmd/server
func Handler() http.Handler {
	target, _ := url.Parse("http://localhost:5173")
	return httputil.NewSingleHostReverseProxy(target)
}
