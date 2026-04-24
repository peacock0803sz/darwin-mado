package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/cli"
	"github.com/peacock0803sz/mado/internal/screen"
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

func TestListCmd_ScreenFilterUUID(t *testing.T) {
	uuidA := "37D8832A-2D66-02CA-B9F7-8F30A301B230"
	uuidB := "12345678-90AB-CDEF-1234-567890ABCDEF"
	windows := []ax.Window{
		{AppName: "Code", Title: "main", PID: 1, State: ax.StateNormal, ScreenID: 1, ScreenName: "Built-in", ScreenUUID: uuidA},
		{AppName: "Safari", Title: "GitHub", PID: 2, State: ax.StateNormal, ScreenID: 2, ScreenName: "DELL U2720Q", ScreenUUID: uuidB},
	}
	screens := []ax.Screen{
		{ID: 1, Name: "Built-in", UUID: uuidA, IsPrimary: true},
		{ID: 2, Name: "DELL U2720Q", UUID: uuidB},
	}
	svc := &ax.MockWindowService{Windows: windows, Screens: screens}
	stdout, err := executeListCmdCapture(t, svc, "format: text\n", "list", "--screen", uuidB, "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "GitHub") {
		t.Errorf("expected Safari/GitHub (on DELL) in output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, `"title": "main"`) {
		t.Errorf("Code (on Built-in) leaked into --screen=%s output", uuidB)
	}
}

func TestListCmd_ScreenFilterByName(t *testing.T) {
	windows := []ax.Window{
		{AppName: "Code", Title: "main", PID: 1, State: ax.StateNormal, ScreenID: 1, ScreenName: "Built-in", ScreenUUID: "AAAAAAAA-0000-0000-0000-000000000000"},
		{AppName: "Safari", Title: "GitHub", PID: 2, State: ax.StateNormal, ScreenID: 2, ScreenName: "DELL U2720Q", ScreenUUID: "BBBBBBBB-0000-0000-0000-000000000000"},
	}
	screens := []ax.Screen{
		{ID: 1, Name: "Built-in", UUID: "AAAAAAAA-0000-0000-0000-000000000000", IsPrimary: true},
		{ID: 2, Name: "DELL U2720Q", UUID: "BBBBBBBB-0000-0000-0000-000000000000"},
	}
	svc := &ax.MockWindowService{Windows: windows, Screens: screens}
	stdout, err := executeListCmdCapture(t, svc, "format: text\n", "list", "--screen", "dell u2720q")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Safari") {
		t.Errorf("expected Safari in output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "Code") {
		t.Errorf("Code (on Built-in) leaked into --screen=dell u2720q output")
	}
}

func TestResolveScreenFilterErr_NotFound(t *testing.T) {
	// Error-path tests go through the unit-testable resolveScreenFilterErr
	// seam. The CLI layer wraps these errors with PrintError(4,...) + os.Exit(4)
	// per the contract; that wrapping is trivial enough that we accept the
	// seam boundary.
	screens := []ax.Screen{
		{ID: 1, Name: "Built-in", UUID: "AAAA0000-0000-0000-0000-000000000000", IsPrimary: true},
	}
	svc := &ax.MockWindowService{Screens: screens}
	_, err := cli.ResolveScreenFilterForTest(context.Background(), svc, "NonExistent")
	var notFound *screen.ScreenNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("got err=%v, want ScreenNotFoundError", err)
	}
	if !strings.Contains(notFound.Error(), "no display matched") {
		t.Errorf("error message missing expected text: %q", notFound.Error())
	}
	if !strings.Contains(notFound.Error(), "AAAA0000") {
		t.Errorf("error message should list Available UUIDs: %q", notFound.Error())
	}
}

func TestResolveScreenFilterErr_Ambiguous(t *testing.T) {
	screens := []ax.Screen{
		{ID: 101, Name: "DELL U2720Q", UUID: "AAAA0000-0000-0000-0000-000000000000"},
		{ID: 102, Name: "DELL U2720Q", UUID: "BBBB0000-0000-0000-0000-000000000000"},
	}
	svc := &ax.MockWindowService{Screens: screens}
	_, err := cli.ResolveScreenFilterForTest(context.Background(), svc, "DELL U2720Q")
	var ambiguous *screen.AmbiguousScreenError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("got err=%v, want AmbiguousScreenError", err)
	}
	if !strings.Contains(ambiguous.Error(), "AAAA0000") || !strings.Contains(ambiguous.Error(), "BBBB0000") {
		t.Errorf("error message should list candidate UUIDs: %q", ambiguous.Error())
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
