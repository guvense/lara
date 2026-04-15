package lara

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareConfigDefaultsWithoutConfigFile(t *testing.T) {
	cfg, err := PrepareConfig("localhost", 8898, false, "/mocs", "")
	if err != nil {
		t.Fatalf("PrepareConfig returned error: %v", err)
	}

	if cfg.ServerConfig.Host != "localhost" {
		t.Fatalf("expected host localhost, got %s", cfg.ServerConfig.Host)
	}
	if cfg.ServerConfig.Port != 8898 {
		t.Fatalf("expected port 8898, got %d", cfg.ServerConfig.Port)
	}
	if cfg.MocksPath != "/mocs" {
		t.Fatalf("expected mocks path /mocs, got %s", cfg.MocksPath)
	}
	if cfg.Watcher {
		t.Fatalf("expected watcher false")
	}
}

func TestPrepareConfigMergesFileAndFlags(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.yml")

	content := []byte("server:\n host: 0.0.0.0\n port: 8899\nmocks-path: ./mocks\nwatcher: true\n")
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("failed writing config file: %v", err)
	}

	cfg, err := PrepareConfig("127.0.0.1", 9000, false, "./my-mocks", configPath)
	if err != nil {
		t.Fatalf("PrepareConfig returned error: %v", err)
	}

	if cfg.ServerConfig.Host != "127.0.0.1" {
		t.Fatalf("expected host from flag, got %s", cfg.ServerConfig.Host)
	}
	if cfg.ServerConfig.Port != 9000 {
		t.Fatalf("expected port from flag, got %d", cfg.ServerConfig.Port)
	}
	if cfg.MocksPath != "./my-mocks" {
		t.Fatalf("expected mocks path from flag, got %s", cfg.MocksPath)
	}
	if !cfg.Watcher {
		t.Fatalf("expected watcher true from config file")
	}
}

func TestParseConfigReturnsErrorOnMissingFile(t *testing.T) {
	_, err := parseConfig("/definitely/not/found/config.yml")
	if err == nil {
		t.Fatalf("expected error for missing config file")
	}
}
