package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/peacock0803sz/mado/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	// verify that default values are returned when the file does not exist
	t.Setenv("MADO_CONFIG", "/nonexistent/config.yaml")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	def := config.Default()
	if cfg.Timeout != def.Timeout {
		t.Errorf("expected timeout %v, got %v", def.Timeout, cfg.Timeout)
	}
	if cfg.Format != def.Format {
		t.Errorf("expected format %q, got %q", def.Format, cfg.Format)
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := "timeout: 10s\nformat: json\n"
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", cfg.Timeout)
	}
	if cfg.Format != "json" {
		t.Errorf("expected format json, got %q", cfg.Format)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte("{ invalid: yaml: ["), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoad_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte("format: xml\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte("format: json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MADO_CONFIG", cfgFile)

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Format != "json" {
		t.Errorf("expected format json from file, got %q", cfg.Format)
	}
}
