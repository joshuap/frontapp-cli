package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)       // Linux
	t.Setenv("HOME", dir)                  // fallback
	// On macOS, UserConfigDir uses ~/Library/Application Support.
	// We override by writing directly to the path returned by Path().

	cfg := &Config{
		TokenCommand: "echo secret",
		User:         "alice@example.com",
	}

	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}

	// Write directly to the path for test portability
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.TokenCommand != "echo secret" {
		t.Errorf("expected token_command 'echo secret', got %q", loaded.TokenCommand)
	}
	if loaded.User != "alice@example.com" {
		t.Errorf("expected user 'alice@example.com', got %q", loaded.User)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TokenCommand != "" || cfg.User != "" {
		t.Error("expected zero-value config for missing file")
	}
}

func TestResolveToken(t *testing.T) {
	cfg := &Config{TokenCommand: "echo test-token"}
	token, err := cfg.ResolveToken()
	if err != nil {
		t.Fatal(err)
	}
	if token != "test-token" {
		t.Errorf("expected 'test-token', got %q", token)
	}
}

func TestResolveToken_Empty(t *testing.T) {
	cfg := &Config{}
	_, err := cfg.ResolveToken()
	if err == nil {
		t.Error("expected error for empty token_command")
	}
}

func TestResolveToken_FailingCommand(t *testing.T) {
	cfg := &Config{TokenCommand: "false"}
	_, err := cfg.ResolveToken()
	if err == nil {
		t.Error("expected error for failing command")
	}
}

func TestResolveUser(t *testing.T) {
	cfg := &Config{User: "bob@example.com"}
	if cfg.ResolveUser() != "bob@example.com" {
		t.Error("expected user")
	}
}
