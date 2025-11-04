package sessions

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"
)

func New(workspacePath, repoUrl string, factory ContainerFactory) (*Session, error) {
	token, err := newToken(8)
	if err != nil {
		return nil, fmt.Errorf("Failed to create token")
	}
	if factory == nil {
		return nil, fmt.Errorf("nil ContainerFactory")
	}
	if workspacePath == "" {
		return nil, fmt.Errorf("Workspace path must be absolute")
	}

	ws := Workspace{
		Repo: repoUrl,
		Path: workspacePath,
		Cmd:  []string{"nvim", "/workspace"},
	}

	return &Session{
		Token:     token,
		CreatedAt: time.Now(),
		ws:        ws,
		factory:   factory,
		state:     StateInit,
	}, nil
}

func (s *Session) Start(ctx context.Context) error {
	since := time.Now()
	slog.Info("Start to preparing session")
	if s.state == StateClosed {
		return fmt.Errorf("session closed")
	}
	s.state = StateStarting
	c, err := s.factory()
	if err != nil {
		return fmt.Errorf("Failed to start container:%w", err)
	}

	s.c = c
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.c.Start(gctx, filepath.Join(s.ws.Path, s.Token), s.ws.Cmd)
	})
	path, err := prepareWorkspaceDir(s.ws.Path, s.Token)
	if err != nil {
		return err
	}
	if s.ws.Repo != "" {
		go cloneWorkspace(ctx, path, s.ws.Repo)
	}

	if err := g.Wait(); err != nil {
		s.state, s.lastErr = StateFailed, err
		return err
	}

	// if s.Repo != "" {
	// 	if err := verifyGitRepo(ctx, s.ws.RepoDir); err != nil {
	// 		_ = s.teardown(ctx)
	// 		s.state, s.lastErr = StateFailed, err
	// 		return err
	// 	}
	// }
	//
	s.state = StateReady
	slog.Info(fmt.Sprintf("Finish to prepare session:%v", time.Since(since)))
	return nil
}

func (s *Session) Close(ctx context.Context) error {
	if err := s.c.Stop(ctx); err != nil {
		return fmt.Errorf("Failed to stop container: %w", err)
	}
	if err := s.c.Remove(ctx); err != nil {
		return fmt.Errorf("Failed to remove container: %w", err)
	}
	if err := os.RemoveAll(filepath.Join(s.ws.Path, s.Token)); err != nil {
		return fmt.Errorf("Failed to remove workspace: %w", err)
	}
	return nil
}

func (s *Session) Attach(ctx context.Context) (r io.Reader, w io.Writer, wait func() error, err error) {
	if s.c == nil {
		return nil, nil, nil, fmt.Errorf("Container is not started:%w", err)
	}
	return s.c.Attach(ctx)

}
func (s *Session) ResizePTY(ctx context.Context, cols, rows int) error {
	if s.c == nil {
		return fmt.Errorf("Container is not started")
	}
	return s.c.ResizePTY(ctx, cols, rows)
}

//	func (s *Session) tokenHash() string {
//		sum := sha256.Sum256([]byte(s.Token))
//		return hex.EncodeToString(sum[:8])
//	}
func newToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("Generating random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
func defaultCmd(cmd []string) []string {
	if len(cmd) == 0 {
		return []string{"nvim"}
	}
	return cmd
}

func cloneWorkspace(ctx context.Context, path, url string) {
	slog.Info("Start clone workspace")
	if url == "" || path == "" {
		slog.Error(fmt.Sprintf("Params are invalid, url:%s path:%s", url, path))
		return
	}
	args := []string{
		"clone",
		"--depth=1",
		"--filter=blob:none",
		"--single-branch",
		"--no-tags",
		url,
		path,
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		slog.Error(fmt.Sprintf("Failded fetching repo: %v, %s", err, stderr.String()))
		return
	}

	if entities, err := os.ReadDir(path); err != nil || len(entities) == 0 {
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to check workspace: %v", err))
			return

		}
		slog.Error("Workspace is empty")
	}
	slog.Info("Finish clone workspace")
}

func prepareWorkspaceDir(base, token string) (string, error) {
	if base == "" || token == "" {
		return "", fmt.Errorf("Failed to prepare Workspace dir. Base: %s, Token: %s", base, token)
	}
	path := filepath.Join(base, token)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", err
	}
	slog.Info(path)
	return path, nil

}
