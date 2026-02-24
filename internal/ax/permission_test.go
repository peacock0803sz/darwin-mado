package ax_test

import (
	"context"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
)

func TestCheckPermission_Denied(t *testing.T) {
	svc := &ax.MockWindowService{
		PermErr: &ax.PermissionError{},
	}

	err := svc.CheckPermission()
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}

	permErr, ok := err.(*ax.PermissionError)
	if !ok {
		t.Fatalf("expected *PermissionError, got %T", err)
	}

	if permErr.Error() == "" {
		t.Error("PermissionError.Error() should not be empty")
	}
	if permErr.Resolution() == "" {
		t.Error("PermissionError.Resolution() should contain guidance text")
	}
}

func TestCheckPermission_Granted(t *testing.T) {
	svc := &ax.MockWindowService{
		PermErr: nil, // permission granted
	}

	if err := svc.CheckPermission(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestListWindows_PermissionDenied(t *testing.T) {
	svc := &ax.MockWindowService{
		PermErr: &ax.PermissionError{},
	}

	// CheckPermission returning a PermissionError causes the CLI layer to exit with code 2
	err := svc.CheckPermission()
	if _, ok := err.(*ax.PermissionError); !ok {
		t.Fatalf("expected *PermissionError, got %T", err)
	}
}

func TestAmbiguousTargetError(t *testing.T) {
	candidates := []ax.Window{
		{AppName: "Safari", Title: "GitHub", PID: 5678},
		{AppName: "Safari", Title: "Apple", PID: 5678},
	}
	err := &ax.AmbiguousTargetError{
		Query:      `--app "Safari"`,
		Candidates: candidates,
	}

	if err.Error() == "" {
		t.Error("AmbiguousTargetError.Error() should not be empty")
	}
	if len(err.Candidates) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(err.Candidates))
	}
}

func TestListWindows_ContextCancellation(_ *testing.T) {
	svc := &ax.MockWindowService{
		Windows: []ax.Window{
			{AppName: "Terminal", Title: "test", PID: 1},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// The mock does not check context cancellation, but this confirms
	// that the type signature compiles correctly.
	_, _ = svc.ListWindows(ctx)
}
