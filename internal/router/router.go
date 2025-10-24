package router

import (
	"net/http"
	"nvimanywhere/internal/handlers"
	"path/filepath"
	"runtime"
)

func AddRoutes(mux *http.ServeMux, h *handlers.Handler) {
	_, thisFile, _, ok := runtime.Caller(0) // .../internal/router/router.go
	if !ok {
		panic("runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	staticDir := filepath.Join(repoRoot, "web", "static")

	// Static files
	mux.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir(staticDir)),
		),
	)
	mux.HandleFunc("/health", h.HandleHealth)
	mux.HandleFunc("/", h.HandleIndex)
	mux.HandleFunc("/sessions", h.HandleStartSession)
	mux.HandleFunc("/sessions/", h.HandleSessionPage)
	mux.HandleFunc("/pty", h.HandlePTY)
}
