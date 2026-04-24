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

var screenListTestScreens = []ax.Screen{
	// Deliberately non-sorted input to verify output-time ordering.
	{ID: 12345678, Name: "DELL U2720Q", UUID: "12345678-90AB-CDEF-1234-567890ABCDEF", X: 2560, Y: 0, Width: 3840, Height: 2160, IsPrimary: false},
	{ID: 69678592, Name: "Built-in Retina Display", UUID: "37D8832A-2D66-02CA-B9F7-8F30A301B230", X: 0, Y: 0, Width: 2560, Height: 1600, IsPrimary: true},
	{ID: 42, Name: "Sidecar Display", UUID: "", X: -1920, Y: 0, Width: 1920, Height: 1080, IsPrimary: false},
}

func executeScreenCmdCapture(t *testing.T, svc ax.WindowService, args ...string) (stdout, stderr string, execErr error) {
	t.Helper()

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte("format: text\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MADO_CONFIG", cfgFile)

	cmd := cli.NewRootCmd(svc)
	cmd.SetArgs(args)

	var stderrBuf bytes.Buffer
	cmd.SetErr(&stderrBuf)

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

func TestScreenListCmd_TextOutput(t *testing.T) {
	svc := &ax.MockWindowService{Screens: screenListTestScreens}
	stdout, _, err := executeScreenCmdCapture(t, svc, "screen", "list", "--format", "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Header and geometry rendering.
	if !strings.Contains(stdout, "UUID") || !strings.Contains(stdout, "NAME") {
		t.Errorf("missing header columns:\n%s", stdout)
	}
	// Sort order: primary first, then X asc — Built-in (primary) before Sidecar (X=-1920) before DELL (X=2560).
	iPrimary := strings.Index(stdout, "Built-in Retina Display")
	iSidecar := strings.Index(stdout, "Sidecar Display")
	iDell := strings.Index(stdout, "DELL U2720Q")
	if iPrimary < 0 || iSidecar <= iPrimary || iDell <= iSidecar {
		t.Errorf("sort order wrong: primary=%d sidecar=%d dell=%d\n%s", iPrimary, iSidecar, iDell, stdout)
	}
	// Empty UUID renders as "-".
	if !strings.Contains(stdout, "-  Sidecar Display") && !strings.Contains(stdout, "- ") {
		t.Errorf("expected empty UUID rendered as '-', got:\n%s", stdout)
	}
	// `yes`/`no` primary column.
	if !strings.Contains(stdout, "yes") || !strings.Contains(stdout, "no") {
		t.Errorf("expected yes/no in PRIMARY column, got:\n%s", stdout)
	}
}

func TestScreenListCmd_JSONOutput(t *testing.T) {
	svc := &ax.MockWindowService{Screens: screenListTestScreens}
	stdout, _, err := executeScreenCmdCapture(t, svc, "screen", "list", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !json.Valid([]byte(stdout)) {
		t.Fatalf("stdout is not valid JSON:\n%s", stdout)
	}

	var resp struct {
		SchemaVersion int         `json:"schema_version"`
		Success       bool        `json:"success"`
		Screens       []ax.Screen `json:"screens"`
	}
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if resp.SchemaVersion != 1 || !resp.Success {
		t.Errorf("schema_version=%d success=%v; want 1/true", resp.SchemaVersion, resp.Success)
	}
	if len(resp.Screens) != 3 {
		t.Fatalf("len(screens) = %d, want 3", len(resp.Screens))
	}
	// Primary first after sort.
	if !resp.Screens[0].IsPrimary {
		t.Errorf("screens[0] is not primary: %+v", resp.Screens[0])
	}
	// Empty UUID preserved (not replaced with '-' in JSON).
	sidecarIdx := -1
	for i, s := range resp.Screens {
		if s.Name == "Sidecar Display" {
			sidecarIdx = i
		}
	}
	if sidecarIdx < 0 {
		t.Fatal("Sidecar not present in json output")
	}
	if resp.Screens[sidecarIdx].UUID != "" {
		t.Errorf("Sidecar UUID = %q, want empty", resp.Screens[sidecarIdx].UUID)
	}
}

func TestScreenListCmd_DoesNotRequirePermission(t *testing.T) {
	// Mock with a PermErr set: `screen list` must NOT call CheckPermission.
	svc := &ax.MockWindowService{
		Screens: screenListTestScreens,
		PermErr: &ax.PermissionError{},
	}
	stdout, _, err := executeScreenCmdCapture(t, svc, "screen", "list")
	if err != nil {
		t.Fatalf("screen list failed with permission error set on mock: %v", err)
	}
	if !strings.Contains(stdout, "Built-in Retina Display") {
		t.Errorf("expected listing output, got:\n%s", stdout)
	}
}

func TestListCmd_JSONIncludesScreenUUID(t *testing.T) {
	// T022: behavioral assertion that `mado list --format json` emits
	// screen_uuid equal to the window's ScreenUUID for non-minimized windows.
	uuid := "37D8832A-2D66-02CA-B9F7-8F30A301B230"
	windows := []ax.Window{
		{
			AppName: "Terminal", Title: "zsh", PID: 100, State: ax.StateNormal,
			ScreenID: 42, ScreenName: "Built-in", ScreenUUID: uuid,
		},
		{AppName: "Finder", Title: "home", PID: 200, State: ax.StateMinimized}, // UUID must be empty
	}
	svc := &ax.MockWindowService{Windows: windows}
	stdout, _, err := executeListCmdCaptureStdoutStderr(t, svc, "format: text\n", "list", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var resp struct {
		Windows []ax.Window `json:"windows"`
	}
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if len(resp.Windows) < 1 {
		t.Fatalf("no windows in output:\n%s", stdout)
	}
	var gotUUID string
	for _, w := range resp.Windows {
		if w.AppName == "Terminal" {
			gotUUID = w.ScreenUUID
		}
		if w.State == ax.StateMinimized && w.ScreenUUID != "" {
			t.Errorf("minimized window leaked ScreenUUID=%q", w.ScreenUUID)
		}
	}
	if gotUUID != uuid {
		t.Errorf("Terminal.ScreenUUID = %q, want %q", gotUUID, uuid)
	}
}
