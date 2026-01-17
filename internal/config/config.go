package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Http struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type WS struct {
	MaxMessageSize int64         `yaml:"max_message_size"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	PingInterval   time.Duration `yaml:"ping_interval"`
}

type SessionRuntime struct {
	ImageName      string `yaml:"image_name"`
	BasePath       string `yaml:"base_path"`
	NvimConfigPath string `yaml:"nvim_config_path"`
	WS             *WS    `yaml:"ws"`
}

type Config struct {
	HTTP           *Http           `yaml:"http"`
	SessionRuntime *SessionRuntime `yaml:"session_runtime"`
	LogFilePath    string          `yaml:"log_file_path"`
	Env            string          `yaml:"env"`
}

// Load reads YAML config, applies env overrides, fills defaults,
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

	// ---------------------------------------------------------------------
	// Defaults
	// ---------------------------------------------------------------------

	if c.HTTP == nil {
		c.HTTP = &Http{}
	}
	if c.HTTP.Host == "" {
		c.HTTP.Host = "0.0.0.0"
	}
	if c.HTTP.Port == "" {
		c.HTTP.Port = "8080"
	}

	if c.SessionRuntime == nil {
		c.SessionRuntime = &SessionRuntime{}
	}
	if c.SessionRuntime.BasePath == "" {
		c.SessionRuntime.BasePath = "/workspaces"
	}

	if c.LogFilePath == "" {
		c.LogFilePath = "/logs"
	}

	if c.Env == "" {
		c.Env = "dev"
	}

	// ---------------------------------------------------------------------
	// ENV overrides (deployment-level)
	// ---------------------------------------------------------------------

	if v := os.Getenv("NVA_HTTP_HOST"); v != "" {
		c.HTTP.Host = v
	}
	if v := os.Getenv("NVA_HTTP_PORT"); v != "" {
		c.HTTP.Port = v
	}
	if v := os.Getenv("NVA_ENV"); v != "" {
		c.Env = v
	}
	if v := os.Getenv("NVA_NVIM_CONFIG_PATH"); v != "" {
		c.SessionRuntime.NvimConfigPath = v
	}

	// ---------------------------------------------------------------------
	// Validation
	// ---------------------------------------------------------------------

	if strings.TrimSpace(c.SessionRuntime.ImageName) == "" {
		return nil, errors.New("session_runtime.image_name is required")
	}

	if !isAbsolute(c.SessionRuntime.BasePath) {
		return nil, errors.New("session_runtime.base_path must be absolute")
	}

	if c.SessionRuntime.WS == nil {
		return nil, errors.New("session_runtime.ws is required")
	}

	ws := c.SessionRuntime.WS

	if ws.ReadTimeout <= 0 {
		return nil, errors.New("ws.read_timeout must be > 0")
	}
	if ws.WriteTimeout <= 0 {
		return nil, errors.New("ws.write_timeout must be > 0")
	}
	if ws.PingInterval <= 0 {
		return nil, errors.New("ws.ping_interval must be > 0")
	}
	if ws.PingInterval >= ws.ReadTimeout {
		return nil, errors.New("ws.ping_interval must be < ws.read_timeout")
	}
	if ws.MaxMessageSize <= 0 {
		return nil, errors.New("ws.max_message_size must be > 0")
	}

	if _, err := strconv.Atoi(c.HTTP.Port); err != nil {
		return nil, errors.New("http.port must be numeric")
	}

	return &c, nil
}

func isAbsolute(p string) bool {
	return strings.HasPrefix(p, "/")
}

