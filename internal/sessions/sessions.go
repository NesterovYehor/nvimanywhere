package sessions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"nvimanywhere/internal/config"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

var initOnce sync.Once
var initErr error

func Init(cfg *config.Config) error {
	initOnce.Do(func() {
		if err := initRunner(cfg.SessionRuntime); err != nil {
			initErr = fmt.Errorf("init runtime: %w", err)
			return
		}

		if err := os.MkdirAll(cfg.SessionRuntime.BasePath, 0o755); err != nil {
			initErr = fmt.Errorf(
				"create workspaces dir %q: %w",
				cfg.SessionRuntime.BasePath,
				err,
			)
			return
		}
	})

	return initErr
}

func StartNewSession(parentCtx context.Context, cfg *config.SessionRuntime, url, workspaceEndpoint string) (*Session, error) {
	ctx, cancel := context.WithCancel(parentCtx)

	s := &Session{
		ctx:       ctx,
		cancel:    cancel,
		createdAt: time.Now(),
		repoUrl:   url,
		cfg:       cfg,
		rootPath:  filepath.Join(cfg.BasePath, workspaceEndpoint),
	}

	if err := prepareWorkspaceDir(s.rootPath); err != nil {
		cancel()
		return nil, err
	}
	if s.repoUrl != "" {
		go s.cloneWorkspace()
	}
	id, err := getRunner().start(ctx, s.rootPath)
	if err != nil {
		cancel()
		return nil, err
	}
	s.runtimeId = id

	return s, nil
}

func (s *Session) Close() error {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	if err := getRunner().terminateRuntime(ctx, s.runtimeId); err != nil {
		return fmt.Errorf("Failed to terminate Runtime: %w", err)
	}
	if err := os.RemoveAll(s.rootPath); err != nil {
		return fmt.Errorf("Failed to remove workspace: %w", err)
	}
	return nil
}

func (s *Session) Attach(conn *websocket.Conn) error {
	conn.SetReadLimit(s.cfg.WS.MaxMessageSize)
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(s.cfg.WS.ReadTimeout)); return nil })
	defer s.cancel()
	output, input, closeAttach, err := getRunner().attach(s.ctx, s.runtimeId)

	if err != nil {
		return err
	}

	defer closeAttach()

	grp, gctx := errgroup.WithContext(s.ctx)

	grp.Go(func() error { return s.pumpInput(gctx, conn, input) })
	grp.Go(func() error { return s.pumpOutput(gctx, conn, output) })
	grp.Go(func() error { return s.pingConn(gctx, conn) })

	if err := grp.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *Session) pingConn(ctx context.Context, ws *websocket.Conn) error {
	ticker := time.NewTicker(s.cfg.WS.PingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(s.cfg.WS.WriteTimeout)); err != nil {

				s.cancel()
				return err
			}
		}
	}

}

func (s *Session) pumpInput(ctx context.Context, conn *websocket.Conn, input io.Writer) error {
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		t, message, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("Failed to read data from WS conn: %w", err)
		}

		switch t {
		case websocket.BinaryMessage:
			if _, err := input.Write(message); err != nil {
				return fmt.Errorf("Failed to write data to terminal input chan: %w", err)
			}

		case websocket.TextMessage:
			var m struct {
				Cols int `json:"cols"`
				Rows int `json:"rows"`
			}
			if err := json.Unmarshal(message, &m); err == nil && m.Cols > 0 && m.Rows > 0 {
				if err := s.resizePTY(ctx, m.Cols, m.Rows); err != nil {
					return fmt.Errorf("Failed to resize terminal : %w", err)
				}
			}
		}
	}
}

func (s *Session) pumpOutput(ctx context.Context, conn *websocket.Conn, output io.Reader) error {
	buf := make([]byte, 32*1024)

	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		n, err := output.Read(buf)
		if err != nil {
			return fmt.Errorf("Failed to read data from terminal output chan: %w", err)
		}
		if n > 0 {
			conn.SetWriteDeadline(time.Now().Add(s.cfg.WS.WriteTimeout))
			if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				return fmt.Errorf("Failed to write data to WS Conn: %w", err)
			}
		}
	}
}

func (s *Session) resizePTY(ctx context.Context, cols, rows int) error {
	return getRunner().resizePTY(ctx, cols, rows, s.runtimeId)
}

func (s *Session) cloneWorkspace() {
	if s.repoUrl == "" || s.rootPath == "" {
		s.fail(fmt.Errorf("Params are invalid, url:%s path:%s", s.repoUrl, s.rootPath))
		return
	}
	args := []string{
		"clone",
		"--depth=1",
		"--filter=blob:none",
		"--single-branch",
		"--no-tags",
		s.repoUrl,
		s.rootPath,
	}

	cmd := exec.CommandContext(s.ctx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		s.fail(fmt.Errorf("Failded fetching repo: %v, %s", err, stderr.String()))
		return
	}

	if entities, err := os.ReadDir(s.rootPath); err != nil || len(entities) == 0 {
		if err != nil {
			s.fail(fmt.Errorf("Failed to check workspace: %v", err))
			return
		}
		s.fail(fmt.Errorf("Workspace is empty"))
		return
	}
}

func prepareWorkspaceDir(path string) error {
	if path == "" {
		return fmt.Errorf("Failed to prepare Workspace dir: %s", path)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	return nil
}

func (s *Session) fail(err error) {
	s.errOnce.Do(func() {
		s.lastError = err
	})
}
