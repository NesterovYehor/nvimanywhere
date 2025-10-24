package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"nvimanywhere/internal/config"
	"nvimanywhere/internal/container"
	"nvimanywhere/internal/handlers"
	"nvimanywhere/internal/router"
	"nvimanywhere/internal/sessions"
	"nvimanywhere/internal/templates"
	"os"
	"os/signal"
	"time"
)

func NewHTTPServer(cfg *config.HTTP, h *handlers.Handler) *http.Server {
	mux := http.NewServeMux()
	router.AddRoutes(mux, h)

	return &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("server error", "err", err)
	}
}

func run(ctx context.Context) error {
	cfg := &config.Config{
		Htpp:      &config.HTTP{Host: "", Port: "8080"},
		Container: &config.Container{Image: "nvimanywhere-session:dev"},
		BaseDir:   "/Users/yehornesterov/dev/Go/nvimanywhere/data/workspaces",
	}

	tc, err := templates.NewTemplateCache()
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	raw := container.NewFactory(*cfg.Container)
	factory := func() (sessions.Container, error) {
		c, err := raw()
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	h := handlers.NewHandler(tc, cfg, factory)

	srv := NewHTTPServer(cfg.Htpp, h)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}
