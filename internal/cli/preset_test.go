package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/cli"
)

// executePresetCmd は preset サブコマンドを実行してエラーを返す
func executePresetCmd(t *testing.T, svc ax.WindowService, configContent string, args ...string) error {
	t.Helper()

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte(configContent), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MADO_CONFIG", cfgFile)

	cmd := cli.NewRootCmd(svc)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)

	return cmd.Execute()
}

var validPresetConfig = `presets:
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

func TestPresetList_Empty(t *testing.T) {
	svc := &ax.MockWindowService{}
	err := executePresetCmd(t, svc, "format: text\n", "preset", "list")
	if err != nil {
		t.Fatalf("expected no error for empty preset list, got: %v", err)
	}
}

func TestPresetList_WithPresets(t *testing.T) {
	svc := &ax.MockWindowService{}
	err := executePresetCmd(t, svc, validPresetConfig, "preset", "list")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPresetShow_MissingArgs(t *testing.T) {
	svc := &ax.MockWindowService{}
	cmd := cli.NewRootCmd(svc)
	cmd.SetArgs([]string{"preset", "show"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	// Cobra handles ExactArgs(1) → error
	if err == nil {
		t.Fatal("expected error for missing args, got nil")
	}
}

func TestPresetApply_MissingArgs(t *testing.T) {
	svc := &ax.MockWindowService{}
	cmd := cli.NewRootCmd(svc)
	cmd.SetArgs([]string{"preset", "apply"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	// Cobra handles ExactArgs(1) → error
	if err == nil {
		t.Fatal("expected error for missing args, got nil")
	}
}
