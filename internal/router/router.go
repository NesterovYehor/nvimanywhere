package router

import (
	"io/fs"
	"log/slog"
	"net/http"
	webfs "nvimanywhere"

	"nvimanywhere/internal/handlers"
)

func AddRoutes(mux *http.ServeMux, h *handlers.Handler) {
	staticRoot, err := fs.Sub(webfs.StaticFS, "web/static")
	if err != nil {
		slog.Error(err.Error())
		return
	}
	static := http.StripPrefix("/static/",
		http.FileServer(http.FS(staticRoot)),
	)

	// Static files
	mux.Handle("/static/", static)
	mux.HandleFunc("/health", h.HandleHealth)
	mux.HandleFunc("/", h.HandleIndex)
	mux.HandleFunc("/sessions", h.HandleStartSession)
	mux.HandleFunc("/sessions/", h.HandleSessionPage)
	mux.HandleFunc("/pty", h.HandlePTY)
}
