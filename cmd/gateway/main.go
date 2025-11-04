package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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
	"path/filepath"
	"time"
)

func NewHTTPServer(cfg *config.Config, h *handlers.Handler) *http.Server {
	mux := http.NewServeMux()
	router.AddRoutes(mux, h)

	return &http.Server{
		Addr:              net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port),
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
	cfgPath, err := GetConfigPath("../../config.yaml")
	if err != nil {
		return err
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	tc, err := templates.NewTemplateCache()
	if err != nil {
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

	srv := NewHTTPServer(cfg, h)

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
func GetConfigPath(defaultPath string) (string, error) {
	var cfgFlag string
	flag.StringVar(&cfgFlag, "config", "", "path to config file (YAML)")
	flag.StringVar(&cfgFlag, "c", "", "path to config file (YAML)") // short alias
	flag.Parse()

	cfg := cfgFlag
	if cfg == "" {
		if env := os.Getenv("NVA_CONFIG"); env != "" {
			cfg = env
		} else {
			cfg = defaultPath
		}
	}

	abs, err := filepath.Abs(cfg)
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}
	st, err := os.Stat(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("config file not found: %s", abs)
		}
		return "", fmt.Errorf("stat config: %w", err)
	}
	if st.IsDir() {
		return "", fmt.Errorf("config path is a directory, want file: %s", abs)
	}
	return abs, nil
}
