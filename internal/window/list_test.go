package window_test

import (
	"context"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/window"
)

var testWindows = []ax.Window{
	{AppName: "Terminal", Title: "peacock — zsh", PID: 100, State: ax.StateNormal, ScreenID: 42, ScreenName: "Built-in Retina Display"},
	{AppName: "Safari", Title: "GitHub", PID: 200, State: ax.StateNormal, ScreenID: 42, ScreenName: "Built-in Retina Display"},
	{AppName: "Safari", Title: "Apple", PID: 200, State: ax.StateMinimized},
	{AppName: "Finder", Title: "", PID: 300, State: ax.StateHidden},
}

func TestList_NoFilter(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	windows, err := window.List(context.Background(), svc, window.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(windows) != len(testWindows) {
		t.Errorf("expected %d windows, got %d", len(testWindows), len(windows))
	}
}

func TestList_AppFilter(t *testing.T) {
	tests := []struct {
		name      string
		filter    string
		wantCount int
	}{
		{"exact match", "Safari", 2},
		{"case insensitive lower", "safari", 2},
		{"case insensitive upper", "SAFARI", 2},
		{"no match", "NoSuchApp", 0},
		{"single result", "Terminal", 1},
	}

	svc := &ax.MockWindowService{Windows: testWindows}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := window.ListOptions{AppFilter: tt.filter}
			windows, err := window.List(context.Background(), svc, opts)
			if err != nil {
				t.Fatal(err)
			}
			if len(windows) != tt.wantCount {
				t.Errorf("filter=%q: expected %d windows, got %d", tt.filter, tt.wantCount, len(windows))
			}
		})
	}
}

func TestList_ScreenFilter(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	opts := window.ListOptions{ScreenFilter: "Built-in Retina Display"}
	windows, err := window.List(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	// Terminal + Safari GitHub (minimized/hidden windows have an empty ScreenName)
	if len(windows) != 2 {
		t.Errorf("expected 2 windows on screen, got %d", len(windows))
	}
}

func TestList_ScreenFilterByID(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	// ScreenID 42 を数値文字列で指定
	opts := window.ListOptions{ScreenFilter: "42"}
	windows, err := window.List(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(windows) != 2 {
		t.Errorf("expected 2 windows with screen ID 42, got %d", len(windows))
	}
}

func TestList_ServiceError(t *testing.T) {
	svc := &ax.MockWindowService{
		ListErr: &ax.PermissionError{},
	}
	_, err := window.List(context.Background(), svc, window.ListOptions{})
	if err == nil {
		t.Fatal("expected error from service, got nil")
	}
}

func TestList_EmptyResult(t *testing.T) {
	svc := &ax.MockWindowService{Windows: []ax.Window{}}
	windows, err := window.List(context.Background(), svc, window.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(windows) != 0 {
		t.Errorf("expected empty list, got %d windows", len(windows))
	}
}

func TestList_IgnoreApps(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	opts := window.ListOptions{IgnoreApps: []string{"Safari"}}
	windows, err := window.List(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	// testWindows has 2 Safari windows; remaining: Terminal + Finder = 2
	if len(windows) != 2 {
		t.Errorf("expected 2 windows (Safari excluded), got %d", len(windows))
	}
	for _, w := range windows {
		if w.AppName == "Safari" {
			t.Error("Safari window should be excluded by IgnoreApps")
		}
	}
}

func TestList_IgnoreAppsCaseInsensitive(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	opts := window.ListOptions{IgnoreApps: []string{"safari"}}
	windows, err := window.List(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(windows) != 2 {
		t.Errorf("expected 2 windows (safari case-insensitive), got %d", len(windows))
	}
}

func TestList_IgnoreAppsEmpty(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	opts := window.ListOptions{IgnoreApps: nil}
	windows, err := window.List(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(windows) != len(testWindows) {
		t.Errorf("expected %d windows (no ignore), got %d", len(testWindows), len(windows))
	}
}

func TestList_IgnoreAppsNonExistent(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	opts := window.ListOptions{IgnoreApps: []string{"NoSuchApp"}}
	windows, err := window.List(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(windows) != len(testWindows) {
		t.Errorf("expected %d windows (non-existent ignored app), got %d", len(testWindows), len(windows))
	}
}
