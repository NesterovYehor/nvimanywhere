package container

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"nvimanywhere/internal/config"
	"nvimanywhere/internal/errors"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type Container struct {
	cli       *client.Client
	imagePath string
	id        string
}

func NewFactory(cfg config.Container) func() (*Container, error) {
	return func() (*Container, error) {
		{
			cli, err := client.NewClientWithOpts(
				client.FromEnv,
				client.WithAPIVersionNegotiation(),
			)
			if err != nil {
				return nil, err
			}
			return &Container{
				cli:       cli,
				imagePath: cfg.Image,
			}, nil
		}
	}
}

func (c *Container) Start(ctx context.Context, workdir string, cmd []string) error {
	since := time.Now()
	if len(cmd) == 0 {
		cmd = []string{"nvim"}
	}

	cfg := &container.Config{
		Image:        c.imagePath,
		Cmd:          cmd,
		Tty:          true, // PTY for terminal apps
		OpenStdin:    true, // keep stdin open
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Env:          []string{"TERM=xterm-256color", "COLORTERM=truecolor"},
		WorkingDir:   "/workspace",
	}

	hostCfg := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   workdir,      // ← HOST absolute repo path
				Target:   "/workspace", // ← CONTAINER mount point (matches WorkingDir)
				ReadOnly: false,
			},
		},
	}
	netCfg := &network.NetworkingConfig{}

	resp, err := c.cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, nil, "")
	if err != nil {
		return err
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}
	slog.Info("Container is starting ")
	c.id = resp.ID

	slog.Info(fmt.Sprintf("Finish to prepare container:%v", time.Since(since)))
	return nil
}

func (c *Container) Attach(ctx context.Context) (r io.Reader, w io.Writer, wait func() error, err error) {
	if c.id == "" {
		return nil, nil, nil, errors.ContainerNotStarted
	}

	att, err := c.cli.ContainerAttach(ctx, c.id, container.AttachOptions{
		Stream: true, Stdin: true, Stdout: true, Stderr: true, Logs: false,
	})

	if err != nil {
		return nil, nil, nil, fmt.Errorf("Attach container: %v", err)
	}

	r = att.Reader
	w = att.Conn

	wait = func() error {
		ch, errCh := c.cli.ContainerWait(ctx, c.id, container.WaitConditionNotRunning)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		case <-ch:
			return nil
		}
	}
	slog.Info("Container Attached")
	return r, w, wait, nil
}
func (c *Container) Stop(ctx context.Context) error {
	if c.id == "" {
		return errors.ContainerNotStarted
	}
	return c.cli.ContainerStop(ctx, c.id, container.StopOptions{})
}

func (c *Container) Remove(ctx context.Context) error {
	if c.id == "" {
		return errors.ContainerNotStarted
	}
	err := c.cli.ContainerRemove(ctx, c.id, container.RemoveOptions{})
	if err == nil {
		c.id = "" // mark as gone
	}
	return err
}

func (c *Container) ResizePTY(ctx context.Context, cols, rows int) error {
	if c.id == "" {
		return errors.ContainerNotStarted
	}

	return c.cli.ContainerResize(ctx, c.id, container.ResizeOptions{Width: uint(cols), Height: uint(rows)})
}
