// package: config
// go: 1.22+
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type HTTP struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Container struct {
	Image   string `yaml:"image"`
	Workdir string `yaml:"workdir"`
}

type Config struct {
	HTTP      *HTTP      `yaml:"http"`
	Container *Container `yaml:"container"`
	BaseDir   string     `yaml:"basedir"`
}

// Load reads YAML at path, applies env overrides, fills defaults,
// validates, and returns a ready-to-use Config.
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// --- defaults -----------------------------------------------------------
	if c.HTTP == nil {
		c.HTTP = &HTTP{}
	}
	if c.HTTP.Host == "" {
		c.HTTP.Host = "0.0.0.0"
	}
	if c.HTTP.Port == "" {
		c.HTTP.Port = "8080"
	}
	if c.Container == nil {
		c.Container = &Container{}
	}
	if c.Container.Workdir == "" {
		c.Container.Workdir = "/workspace"
	}
	if strings.TrimSpace(c.BaseDir) == "" {
		c.BaseDir = "/srv/nvimanywhere/data/workspaces"
	}

	// --- env overrides (optional, handy in prod) ----------------------------
	// NVA_IMAGE, NVA_WORKDIR, NVA_BASEDIR can override yaml.
	if v := os.Getenv("NVA_IMAGE"); v != "" {
		c.Container.Image = v
	}
	if v := os.Getenv("NVA_WORKDIR"); v != "" {
		c.Container.Workdir = v
	}
	if v := os.Getenv("NVA_BASEDIR"); v != "" {
		c.BaseDir = v
	}
	if v := os.Getenv("NVA_HOST"); v != "" {
		c.HTTP.Host = v
	}
	if v := os.Getenv("NVA_PORT"); v != "" {
		c.HTTP.Port = v
	}

	// --- normalize/validate -------------------------------------------------
	// absolutize basedir
	absBase, err := filepath.Abs(c.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("normalize basedir: %w", err)
	}
	c.BaseDir = absBase

	// ensure basedir exists & writable
	if err := os.MkdirAll(c.BaseDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir basedir: %w", err)
	}
	if err := writeTest(c.BaseDir); err != nil {
		return nil, fmt.Errorf("basedir not writable: %w", err)
	}

	if strings.TrimSpace(c.Container.Image) == "" {
		return nil, errors.New("container.image is required")
	}
	if !strings.HasPrefix(c.Container.Workdir, "/") {
		return nil, fmt.Errorf("container.workdir must be absolute, got %q", c.Container.Workdir)
	}

	return &c, nil
}

func writeTest(dir string) error {
	f, err := os.CreateTemp(dir, ".wtest-*")
	if err != nil {
		return err
	}
	name := f.Name()
	_ = f.Close()
	return os.Remove(name)
}

