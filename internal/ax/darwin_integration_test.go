//go:build integration && darwin

package ax_test

import (
	"context"
	"strings"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
)

// TestListWindows_AppIDPopulated verifies that ListWindows() populates the
// AppID field for real macOS windows. Run with:
//
//	go test ./internal/ax/... -tags integration
//
// This test requires a running macOS GUI session with at least one visible
// application window. It is excluded from regular `go test ./...` runs.
func TestListWindows_AppIDPopulated(t *testing.T) {
	svc := ax.NewWindowService()
	windows, err := svc.ListWindows(context.Background())
	if err != nil {
		t.Fatalf("ListWindows() error: %v", err)
	}
	if len(windows) == 0 {
		t.Skip("no windows found; ensure at least one app is running")
	}

	// At least one GUI app should have a bundle identifier.
	hasAppID := false
	for _, w := range windows {
		if w.AppID != "" {
			hasAppID = true
			break
		}
	}
	if !hasAppID {
		t.Error("expected at least one window with a non-empty AppID (bundle identifier)")
	}

	// Validate that all AppID values look like reverse-DNS identifiers when set.
	// Some windows (e.g. background helpers) may legitimately have an empty AppID.
	for _, w := range windows {
		if w.AppID == "" {
			continue
		}
		// A bundle ID must contain at least one dot (e.g. "com.apple.Safari").
		if !strings.Contains(w.AppID, ".") {
			t.Errorf("window %q (app=%q) has AppID %q with no dot; expected reverse-DNS format",
				w.Title, w.AppName, w.AppID)
		}
	}
}
