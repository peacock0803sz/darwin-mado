// Package config loads and provides mado configuration from YAML files.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v4"

	"github.com/peacock0803sz/mado/internal/preset"
)

// Config is the structure of the mado configuration file.
// Priority order: CLI flags > config file > default values.
type Config struct {
	Timeout time.Duration
	Format  string
	Presets []preset.Preset
}

// rawConfig is an intermediate structure for YAML parsing.
// time.Duration cannot be decoded directly from YAML, so it is received as a string.
type rawConfig struct {
	Timeout string          `yaml:"timeout"`
	Format  string          `yaml:"format"`
	Presets []preset.Preset `yaml:"presets"`
}

// Default returns the default configuration.
func Default() Config {
	return Config{
		Timeout: 5 * time.Second,
		Format:  "text",
	}
}

// Load reads the configuration file and returns the result.
// When the config file does not exist, the default values are returned without an error.
// Search order:
//  1. $MADO_CONFIG environment variable
//  2. $XDG_CONFIG_HOME/mado/config.yaml (defaults to ~/.config/mado/config.yaml)
func Load() (Config, error) {
	cfg := Default()

	path, err := configPath()
	if err != nil {
		return cfg, nil // fall back to defaults if path resolution fails
	}

	data, err := os.ReadFile(path) //nolint:gosec // G304: path is resolved from trusted config locations (XDG or $MADO_CONFIG)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil // no file = use default values
		}
		return cfg, fmt.Errorf("config file read error: %w", err)
	}

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return cfg, fmt.Errorf("config file parse error (%s): %w", path, err)
	}

	// convert timeout string (e.g. "5s") to time.Duration
	if raw.Timeout != "" {
		d, err := time.ParseDuration(raw.Timeout)
		if err != nil {
			return cfg, fmt.Errorf("config: invalid timeout %q: %w", raw.Timeout, err)
		}
		cfg.Timeout = d
	}

	// apply format if specified: "text" | "json"
	if raw.Format != "" {
		switch raw.Format {
		case "text", "json":
			cfg.Format = raw.Format
		default:
			return cfg, fmt.Errorf("config: invalid format %q (must be \"text\" or \"json\")", raw.Format)
		}
	}

	// Validate presets
	if len(raw.Presets) > 0 {
		if verrs := preset.ValidatePresets(raw.Presets); verrs != nil {
			var errMsgs []string
			for _, vErr := range verrs {
				errMsgs = append(errMsgs, vErr.Error())
			}
			return cfg, fmt.Errorf("config (%s): preset validation failed: %s", path, strings.Join(errMsgs, "; "))
		}
		cfg.Presets = raw.Presets
	}

	return cfg, nil
}

// configPath returns the path to the configuration file.
func configPath() (string, error) {
	// 1. $MADO_CONFIG environment variable
	if p := os.Getenv("MADO_CONFIG"); p != "" {
		return p, nil
	}

	// 2. XDG_CONFIG_HOME or ~/.config
	baseDir := os.Getenv("XDG_CONFIG_HOME")
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home dir: %w", err)
		}
		baseDir = filepath.Join(home, ".config")
	}

	return filepath.Join(baseDir, "mado", "config.yaml"), nil
}
