package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"nvimanywhere/internal/config"
	s "nvimanywhere/internal/sessions"
	"nvimanywhere/internal/templates"
	"sync"

	"github.com/gorilla/websocket"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex

	templates templates.TemplateCache
	cfg       *config.Config
	log       *slog.Logger
	sessions  map[string]*s.Session
	upgrader  websocket.Upgrader
}

func InitApp(cfg *config.Config, log *slog.Logger, t templates.TemplateCache, ctx context.Context) *App {
	return &App{
		ctx:       ctx,
		mu:        sync.Mutex{},
		templates: t,
		cfg:       cfg,
		log:       log,
		sessions:  make(map[string]*s.Session),
		upgrader:  websocket.Upgrader{},
	}
}

func (h *App) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte("ok")); err != nil {
		h.log.Error(err.Error())
	}
}

func (h *App) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowd", http.StatusMethodNotAllowed)
		h.log.Error("Wrong Method: handleIndex")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	temp := h.templates["index"]
	if temp == nil {
		h.respondError(w, http.StatusServiceUnavailable, "Unknown template", nil)
		return
	}
	if err := temp.Execute(w, struct {
		Title string
	}{Title: "NvimAnywhere"}); err != nil {
		h.respondError(w, http.StatusServiceUnavailable, "Failed to make a response", nil)
		return
	}
}

func (h *App) respondError(w http.ResponseWriter, status int, reason string, err error) {
	http.Error(w, reason, status)

	if err != nil {
		h.log.Error(reason,
			"err", err,
			"status", status,
		)
		return
	}

	h.log.Error(reason,
		"status", status,
	)
}
