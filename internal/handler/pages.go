package handler

import (
	"embed"
	"net/http"
)

var staticFS embed.FS

func InitStatic(efs embed.FS) {
	staticFS = efs
}

func StaticHandler() http.Handler {
	return http.FileServer(http.FS(staticFS))
}

func PageHandler(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFS.ReadFile(filename)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}
}
