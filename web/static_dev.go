//go:build dev

package web

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func Handler() http.Handler {
	viteURL := os.Getenv("VITE_URL")
	if viteURL == "" {
		viteURL = "http://localhost:5173"
	}
	target, _ := url.Parse(viteURL)
	return httputil.NewSingleHostReverseProxy(target)
}
