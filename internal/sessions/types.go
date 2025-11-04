package sessions

import (
	"context"
	"io"
	"time"
)

// What the session needs from any container backend.
type Container interface {
	Start(ctx context.Context, workdir string, cmd []string) error
	Attach(ctx context.Context) (r io.Reader, w io.Writer, wait func() error, err error)
	ResizePTY(ctx context.Context, rows, cols int) error
	Stop(ctx context.Context) error
	Remove(ctx context.Context) error
}

// Minimal description of a workspace for the backend to mount/configure.
type Workspace struct {
	Path string // absolute base for all workspaces
	Repo string
	Env  map[string]string
	Cmd  []string // default command, e.g. {"nvim"}
}

// Factory pattern, Go-style: one function you inject at startup.
type ContainerFactory func() (Container, error)

// Session states (simple).
type State string

const (
	StateInit     State = "init"
	StateStarting State = "starting"
	StateReady    State = "ready"
	StateFailed   State = "failed"
	StateClosed   State = "closed"
)

type Session struct {
	Token     string
	CreatedAt time.Time

	ws      Workspace
	c       Container
	factory ContainerFactory
	state   State
	lastErr error
}
