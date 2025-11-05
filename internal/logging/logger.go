package logging

import (
	"fmt"
	"io"
	"log/slog"
	"nvimanywhere/internal/config"
	"os"
)

func NewLogger(cfg *config.Config) (*slog.Logger, func() error, error) {
	if cfg.LogFilePath == "" {
		return nil, nil, fmt.Errorf("File path for logs is empty")
	}

	f, err := os.OpenFile(cfg.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to open log file: %w", err)
	}
	closeFn := func() error { return f.Close() }
	w := io.MultiWriter(os.Stdout, f)
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:     slog.LevelDebug.Level(),
		AddSource: false,
	})

	hostName, err := os.Hostname()
	if err != nil {
		return nil, nil, err
	}
	log := slog.New(h).With(
		"service", "nvimanywhere",
		"env", cfg.Env,
		"instance", hostName,
		"pid", os.Getpid(),
		"listen_addr", cfg.HTTP.Host+":"+cfg.HTTP.Port,
		"workspaces_dir", cfg.Container.Workdir,
		"container_runtime", "docker",
	)

	return log, closeFn, nil

}
