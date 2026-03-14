package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TokenCommand string `yaml:"token_command,omitempty"`
	User         string `yaml:"user,omitempty"`
}

// Path returns the config file path under os.UserConfigDir().
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return filepath.Join(dir, "front", "config.yaml"), nil
}

// Load reads the config file. Returns zero-value Config if the file doesn't exist.
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return &Config{}, nil
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to disk, creating the directory (0700) if needed.
func Save(cfg *Config) error {
	p, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	return os.WriteFile(p, data, 0600)
}

// ResolveToken runs token_command via sh -c and returns trimmed stdout.
// Returns an error if the command exits non-zero or token_command is empty.
func (c *Config) ResolveToken() (string, error) {
	if c.TokenCommand == "" {
		return "", fmt.Errorf("no token_command configured")
	}

	cmd := exec.Command("sh", "-c", c.TokenCommand)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("token_command failed: %w", err)
	}

	token := strings.TrimSpace(string(out))
	if token == "" {
		return "", fmt.Errorf("token_command returned empty output")
	}
	return token, nil
}

// ResolveUser returns the user field.
func (c *Config) ResolveUser() string {
	return c.User
}
