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
	result, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	def := config.Default()
	if result.Config.Timeout != def.Timeout {
		t.Errorf("expected timeout %v, got %v", def.Timeout, result.Config.Timeout)
	}
	if result.Config.Format != def.Format {
		t.Errorf("expected format %q, got %q", def.Format, result.Config.Format)
	}
	if result.SourcePath != "" {
		t.Errorf("expected empty SourcePath for missing file, got %q", result.SourcePath)
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
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Config.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", result.Config.Timeout)
	}
	if result.Config.Format != "json" {
		t.Errorf("expected format json, got %q", result.Config.Format)
	}
	if result.SourcePath != cfgFile {
		t.Errorf("expected SourcePath %q, got %q", cfgFile, result.SourcePath)
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

func TestLoad_PresetsValid(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := `presets:
  - name: coding
    description: "Editor left, terminal right"
    rules:
      - app: Code
        position: [0, 0]
        size: [960, 1080]
      - app: Terminal
        position: [960, 0]
        size: [960, 1080]
`
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Config.Presets) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(result.Config.Presets))
	}
	if result.Config.Presets[0].Name != "coding" {
		t.Errorf("expected preset name 'coding', got %q", result.Config.Presets[0].Name)
	}
	if len(result.Config.Presets[0].Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(result.Config.Presets[0].Rules))
	}
}

func TestLoad_PresetsInvalid(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	// ルールにappフィールドがない不正なプリセット
	content := `presets:
  - name: broken
    rules:
      - position: [0, 0]
`
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected validation error for invalid preset, got nil")
	}
}

func TestLoad_PresetsEmpty(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := "presets: []\n"
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Config.Presets) != 0 {
		t.Errorf("expected 0 presets, got %d", len(result.Config.Presets))
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte("format: json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MADO_CONFIG", cfgFile)

	result, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if result.Config.Format != "json" {
		t.Errorf("expected format json from file, got %q", result.Config.Format)
	}
}

func TestLoad_IgnoreAppsValid(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := "ignore_apps:\n  - Safari\n  - Dock\n"
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Config.IgnoreApps) != 2 {
		t.Fatalf("expected 2 ignore_apps, got %d", len(result.Config.IgnoreApps))
	}
	if result.Config.IgnoreApps[0] != "Safari" || result.Config.IgnoreApps[1] != "Dock" {
		t.Errorf("expected [Safari, Dock], got %v", result.Config.IgnoreApps)
	}
}

func TestLoad_IgnoreAppsEmptyString(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := "ignore_apps:\n  - Safari\n  - \"\"\n"
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for empty string in ignore_apps, got nil")
	}
}

func TestLoad_XDGConfigHome(t *testing.T) {
	// Verify that $XDG_CONFIG_HOME/mado/config.yaml is discovered
	// when $MADO_CONFIG is unset.
	dir := t.TempDir()
	madoDir := filepath.Join(dir, "mado")
	if err := os.MkdirAll(madoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgFile := filepath.Join(madoDir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte("format: json\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", "")
	t.Setenv("XDG_CONFIG_HOME", dir)

	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Config.Format != "json" {
		t.Errorf("expected format json via XDG, got %q", result.Config.Format)
	}
}

func TestLoad_XDGFallbackToDefaults(t *testing.T) {
	// When $MADO_CONFIG is unset and no config file exists under
	// $XDG_CONFIG_HOME (or /etc/mado), Load should return defaults.
	dir := t.TempDir()

	t.Setenv("MADO_CONFIG", "")
	t.Setenv("XDG_CONFIG_HOME", dir)

	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def := config.Default()
	if result.Config.Timeout != def.Timeout || result.Config.Format != def.Format {
		t.Errorf("expected defaults, got timeout=%v format=%q", result.Config.Timeout, result.Config.Format)
	}
}

func TestLoad_IgnoreAppsAbsent(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := "format: text\n"
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Config.IgnoreApps) != 0 {
		t.Errorf("expected empty ignore_apps, got %v", result.Config.IgnoreApps)
	}
}

func TestLoad_IgnoreAppsDuplicates(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := "ignore_apps:\n  - Safari\n  - Safari\n"
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Config.IgnoreApps) != 2 {
		t.Fatalf("expected 2 ignore_apps (duplicates accepted), got %d", len(result.Config.IgnoreApps))
	}
}

func TestLoad_PresetAppIDRule(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := "presets:\n  - name: bundle\n    rules:\n      - app_id: com.apple.Safari\n        position: [0, 0]\n"
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error for app_id-only rule: %v", err)
	}
	if len(result.Config.Presets) != 1 || len(result.Config.Presets[0].Rules) != 1 {
		t.Fatalf("expected 1 preset with 1 rule, got %d presets", len(result.Config.Presets))
	}
	r := result.Config.Presets[0].Rules[0]
	if r.AppID != "com.apple.Safari" {
		t.Errorf("rule.AppID = %q, want %q", r.AppID, "com.apple.Safari")
	}
	if r.App != "" {
		t.Errorf("rule.App = %q, want empty for app_id rule", r.App)
	}
}

func TestLoad_IgnoreAppsBundleID(t *testing.T) {
	// Bundle IDs (containing dots) are valid ignore_apps entries
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := "ignore_apps:\n  - com.apple.Safari\n  - Dock\n"
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error for bundle ID in ignore_apps: %v", err)
	}
	if len(result.Config.IgnoreApps) != 2 {
		t.Fatalf("expected 2 ignore_apps, got %d", len(result.Config.IgnoreApps))
	}
	if result.Config.IgnoreApps[0] != "com.apple.Safari" {
		t.Errorf("ignore_apps[0] = %q, want %q", result.Config.IgnoreApps[0], "com.apple.Safari")
	}
}

func TestLoad_VerboseTrue(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte("verbose: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Config.Verbose {
		t.Error("expected Verbose=true, got false")
	}
}

func TestLoad_VerboseFalse(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte("verbose: false\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Config.Verbose {
		t.Error("expected Verbose=false, got true")
	}
}

func TestLoad_VerboseAbsent(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte("format: text\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Config.Verbose {
		t.Error("expected Verbose=false when absent, got true")
	}
}

func TestLoad_SourcePath(t *testing.T) {
	// SourcePath should be the resolved config file path
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte("format: text\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MADO_CONFIG", cfgFile)
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SourcePath != cfgFile {
		t.Errorf("expected SourcePath=%q, got %q", cfgFile, result.SourcePath)
	}
}

func TestLoad_SourcePathEmpty(t *testing.T) {
	// SourcePath should be empty when no config file exists
	t.Setenv("MADO_CONFIG", "/nonexistent/config.yaml")
	result, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SourcePath != "" {
		t.Errorf("expected empty SourcePath for missing file, got %q", result.SourcePath)
	}
}
