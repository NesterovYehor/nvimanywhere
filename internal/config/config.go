package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Http struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Container struct {
	Image   string `yaml:"image"`
	Workdir string `yaml:"workspaces_path"`
}

type Config struct {
	HTTP      *Http      `yaml:"http"`
	Container *Container `yaml:"container"`
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
		c.HTTP = &Http{}
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

	if strings.TrimSpace(c.Container.Image) == "" {
		return nil, errors.New("container.image is required")
	}
	return &c, nil
}
