//go:build !dev

package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var assets embed.FS

func Handler() http.Handler {
	dist, err := fs.Sub(assets, "dist")
	if err != nil {
		panic("web: embed dist subtree: " + err.Error())
	}
	return http.FileServerFS(dist)
}
