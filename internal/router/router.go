package router

import (
	"io/fs"
	"net/http"
	webfs "nvimanywhere"

	"nvimanywhere/internal/handlers"
)

func AddRoutes(mux *http.ServeMux, h *handlers.Handler) error{
	staticRoot, err := fs.Sub(webfs.StaticFS, "web/static")
	if err != nil {
		return err
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
	return nil
}
