package cli_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/cli"
)

var listTestWindows = []ax.Window{
	{AppName: "Terminal", Title: "zsh", PID: 100, State: ax.StateNormal, ScreenID: 1, ScreenName: "Built-in"},
	{AppName: "Safari", Title: "GitHub", PID: 200, State: ax.StateNormal, ScreenID: 1, ScreenName: "Built-in"},
	{AppName: "Finder", Title: "Home", PID: 300, State: ax.StateNormal, ScreenID: 1, ScreenName: "Built-in"},
}

// executeListCmdCapture はコマンド実行し、os.Stdout出力をキャプチャする
func executeListCmdCapture(t *testing.T, svc ax.WindowService, configContent string, args ...string) (string, error) {
	t.Helper()

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte(configContent), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MADO_CONFIG", cfgFile)

	cmd := cli.NewRootCmd(svc)
	cmd.SetArgs(args)

	// output.Formatterはos.Stdoutに直接書くためos.Stdoutをキャプチャ
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	execErr := cmd.Execute()

	_ = w.Close()
	captured, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	return string(captured), execErr
}

func TestListCmd_IgnoreOverride(t *testing.T) {
	// --app Safari overrides ignore_apps: ["Safari"]
	svc := &ax.MockWindowService{Windows: listTestWindows}
	configContent := "ignore_apps:\n  - Safari\n"
	output, err := executeListCmdCapture(t, svc, configContent, "list", "--app", "Safari")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Safari") {
		t.Errorf("expected Safari in output (--app overrides ignore), got:\n%s", output)
	}
}

func TestListCmd_IgnoreNoAppFlag(t *testing.T) {
	// Without --app, ignore_apps: ["Safari"] excludes Safari
	svc := &ax.MockWindowService{Windows: listTestWindows}
	configContent := "ignore_apps:\n  - Safari\n"
	output, err := executeListCmdCapture(t, svc, configContent, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(output, "Safari") {
		t.Errorf("expected Safari to be excluded by ignore_apps, got:\n%s", output)
	}
	if !strings.Contains(output, "Terminal") {
		t.Errorf("expected Terminal in output, got:\n%s", output)
	}
}
