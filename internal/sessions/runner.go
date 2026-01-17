package sessions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"nvimanywhere/internal/config"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

var (
	r    *runner
	once sync.Once
)

var containerNotStarted = errors.New("Container is not started")

type runner struct {
	imageName  string
	configPath string
	cli        *client.Client
}

func initRunner(cfg *config.SessionRuntime) error {
	var e error
	once.Do(func() {
		cli, err := client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
		e = err
		r = &runner{imageName: cfg.ImageName, configPath: cfg.NvimConfigPath, cli: cli}
	})
	return e
}

func getRunner() *runner {
	if r != nil {
		return r
	}
	return nil
}

func (runner *runner) buildContainerSpec(workspace string) (*container.Config, *container.HostConfig) {
	env := []string{
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"NVIM_LOG_FILE=/workspace/tmp/nvim.log",
		"NVIM_LOG_LEVEL=debug",
	}
	if runner.imageName == "" {
		runner.imageName = "ghcr.io/neovim/neovim:v0.10.3"
	}

	cfg := &container.Config{
		Image:        runner.imageName,
		Tty:          true,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Env:          env,
		WorkingDir:   "/workspace",
	}

	mounts := []mount.Mount{
		{
			Type:     mount.TypeBind,
			Source:   workspace,
			Target:   "/workspace",
			ReadOnly: false,
		},
	}
	if runner.configPath != "" {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   runner.configPath,
			Target:   "/home/nvim/.config/nvim",
			ReadOnly: false,
		})
	}

	hostCfg := &container.HostConfig{Mounts: mounts, LogConfig: container.LogConfig{Type: "none"}}
	return cfg, hostCfg
}

func (runner *runner) start(ctx context.Context, workspace string) (string, error) {
	cfg, hostCfg := runner.buildContainerSpec(workspace)

	resp, err := runner.cli.ContainerCreate(ctx, cfg, hostCfg, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("Failed to create container: %v", err)
	}

	if err := runner.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (runner *runner) attach(
	ctx context.Context, id string) (
	io.Reader,
	io.Writer,
	func() error,
	error) {
	if id == "" {
		return nil, nil, nil, containerNotStarted
	}

	att, err := runner.cli.ContainerAttach(ctx, id, container.AttachOptions{
		Stream: true, Stdin: true, Stdout: true, Stderr: true, Logs: false,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Attach container: %v", err)
	}
	closeAttach := att.Conn.Close

	r := att.Reader
	w := att.Conn

	return r, w, closeAttach, nil
}

func (r *runner) terminateRuntime(ctx context.Context, id string) error {
	if id == "" {
		return containerNotStarted
	}

	if err := r.cli.ContainerStop(ctx, id, container.StopOptions{}); err != nil {
		return err
	}

	statusCh, errCh := r.cli.ContainerWait(
		ctx,
		id,
		container.WaitConditionNotRunning,
	)

	select {
	case <-statusCh:
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	if err := r.cli.ContainerRemove(ctx, id, container.RemoveOptions{}); err != nil {
		return err
	}

	return nil
}

func (runner *runner) resizePTY(ctx context.Context, cols, rows int, id string) error {
	if id == "" {
		return containerNotStarted
	}
	return runner.cli.ContainerResize(ctx, id, container.ResizeOptions{Width: uint(cols), Height: uint(rows)})
}
