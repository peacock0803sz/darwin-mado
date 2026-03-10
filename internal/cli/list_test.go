package cli_test

import (
	"bytes"
	"encoding/json"
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

// executeListCmdCapture executes a command and captures os.Stdout output.
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

	// output.Formatter writes directly to os.Stdout, so we capture it via pipe
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

// executeListCmdCaptureStdoutStderr captures stdout and stderr separately.
func executeListCmdCaptureStdoutStderr(t *testing.T, svc ax.WindowService, configContent string, args ...string) (stdout, stderr string, execErr error) {
	t.Helper()

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte(configContent), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MADO_CONFIG", cfgFile)

	cmd := cli.NewRootCmd(svc)
	cmd.SetArgs(args)

	var stderrBuf bytes.Buffer
	cmd.SetErr(&stderrBuf)

	// output.Formatter writes directly to os.Stdout, so we capture it via pipe
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	execErr = cmd.Execute()

	_ = w.Close()
	captured, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	return string(captured), stderrBuf.String(), execErr
}

func TestListCmd_JSONVerbose(t *testing.T) {
	// --format json --verbose produces valid JSON on stdout, verbose on stderr
	svc := &ax.MockWindowService{Windows: listTestWindows}

	stdout, stderr, err := executeListCmdCaptureStdoutStderr(t, svc, "format: text\n", "list", "--format", "json", "--verbose")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// stdout must be valid JSON
	if !json.Valid([]byte(stdout)) {
		t.Errorf("stdout is not valid JSON:\n%s", stdout)
	}

	// stderr must contain verbose prefix
	if !strings.Contains(stderr, "verbose: ") {
		t.Errorf("stderr should contain verbose prefix, got:\n%s", stderr)
	}

	// stdout must be identical with and without --verbose
	stdoutNoVerbose, _, err := executeListCmdCaptureStdoutStderr(t, svc, "format: text\n", "list", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != stdoutNoVerbose {
		t.Errorf("stdout should be identical with/without --verbose\nwith:    %s\nwithout: %s", stdout, stdoutNoVerbose)
	}
}

func TestListCmd_VerboseFlagOverConfig(t *testing.T) {
	// config verbose: true + --verbose=false suppresses verbose output
	svc := &ax.MockWindowService{Windows: listTestWindows}

	_, stderr, err := executeListCmdCaptureStdoutStderr(t, svc, "verbose: true\n", "list", "--verbose=false")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(stderr, "verbose: ") {
		t.Errorf("--verbose=false should suppress config verbose: true, got stderr:\n%s", stderr)
	}

	// config verbose: true without flag enables verbose output
	_, stderr, err = executeListCmdCaptureStdoutStderr(t, svc, "verbose: true\n", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "verbose: ") {
		t.Errorf("config verbose: true should enable verbose output, got stderr:\n%s", stderr)
	}
}
