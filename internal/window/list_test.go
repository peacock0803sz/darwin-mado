package window_test

import (
	"context"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/window"
)

var testWindows = []ax.Window{
	{AppName: "Terminal", AppID: "com.apple.Terminal", Title: "peacock — zsh", PID: 100, State: ax.StateNormal, ScreenID: 42, ScreenName: "Built-in Retina Display"},
	{AppName: "Safari", AppID: "com.apple.Safari", Title: "GitHub", PID: 200, State: ax.StateNormal, ScreenID: 42, ScreenName: "Built-in Retina Display"},
	{AppName: "Safari", AppID: "com.apple.Safari", Title: "Apple", PID: 200, State: ax.StateMinimized},
	{AppName: "Finder", AppID: "com.apple.finder", Title: "", PID: 300, State: ax.StateHidden},
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

func TestList_AppIDFilter(t *testing.T) {
	tests := []struct {
		name      string
		filter    string
		wantCount int
	}{
		{"exact match", "com.apple.Safari", 2},
		{"case insensitive", "COM.APPLE.SAFARI", 2},
		{"no match", "com.example.NoApp", 0},
		{"single result", "com.apple.Terminal", 1},
	}

	svc := &ax.MockWindowService{Windows: testWindows}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := window.ListOptions{AppIDFilter: tt.filter}
			windows, err := window.List(context.Background(), svc, opts)
			if err != nil {
				t.Fatal(err)
			}
			if len(windows) != tt.wantCount {
				t.Errorf("AppIDFilter=%q: expected %d windows, got %d", tt.filter, tt.wantCount, len(windows))
			}
		})
	}
}

func TestList_AppFilterAndAppIDFilterAND(t *testing.T) {
	// Both AppFilter and AppIDFilter must match (AND logic)
	svc := &ax.MockWindowService{Windows: testWindows}
	opts := window.ListOptions{
		AppFilter:   "Safari",
		AppIDFilter: "com.apple.Safari",
	}
	windows, err := window.List(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(windows) != 2 {
		t.Errorf("expected 2 windows (Safari AND com.apple.Safari), got %d", len(windows))
	}
}

func TestList_AppFilterAndAppIDFilterMismatch(t *testing.T) {
	// AppFilter matches but AppIDFilter does not → zero results
	svc := &ax.MockWindowService{Windows: testWindows}
	opts := window.ListOptions{
		AppFilter:   "Safari",
		AppIDFilter: "com.example.Other",
	}
	windows, err := window.List(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(windows) != 0 {
		t.Errorf("expected 0 windows (mismatched AND), got %d", len(windows))
	}
}

func TestIsIgnoredApp_NameEntry(t *testing.T) {
	// Entries without a dot match against AppName, not AppID
	if !window.IsIgnoredApp("Safari", "com.apple.Safari", []string{"Safari"}) {
		t.Error("expected Safari (name entry) to match AppName=Safari")
	}
	// Name entry must not accidentally match AppID
	if window.IsIgnoredApp("Other", "com.apple.Safari", []string{"Safari"}) {
		t.Error("name entry 'Safari' should not match AppName=Other")
	}
}

func TestIsIgnoredApp_BundleIDEntry(t *testing.T) {
	// Entries containing a dot match against AppID (bundle identifier)
	if !window.IsIgnoredApp("Safari", "com.apple.Safari", []string{"com.apple.Safari"}) {
		t.Error("expected com.apple.Safari (bundle-ID entry) to match AppID=com.apple.Safari")
	}
	// Bundle-ID entry must not accidentally match AppName
	if window.IsIgnoredApp("com.apple.Safari", "other.bundle", []string{"com.apple.Safari"}) {
		t.Error("bundle-ID entry should not match AppName even if AppName looks like a bundle ID")
	}
}

func TestIsIgnoredApp_MixedList(t *testing.T) {
	// Mixed list: one name entry, one bundle-ID entry
	ignoreApps := []string{"Dock", "com.apple.Safari"}
	// Dock matched by name entry
	if !window.IsIgnoredApp("Dock", "com.apple.Dock", ignoreApps) {
		t.Error("expected Dock to be ignored via name entry")
	}
	// Safari matched by bundle-ID entry
	if !window.IsIgnoredApp("Safari", "com.apple.Safari", ignoreApps) {
		t.Error("expected Safari to be ignored via bundle-ID entry")
	}
	// Terminal matched by neither entry
	if window.IsIgnoredApp("Terminal", "com.apple.Terminal", ignoreApps) {
		t.Error("Terminal should not be ignored")
	}
}

func TestIsIgnoredApp_CaseInsensitive(t *testing.T) {
	// Both name and bundle-ID matching must be case-insensitive
	if !window.IsIgnoredApp("safari", "COM.APPLE.SAFARI", []string{"com.apple.Safari"}) {
		t.Error("bundle-ID match should be case-insensitive")
	}
	if !window.IsIgnoredApp("SAFARI", "com.apple.Safari", []string{"safari"}) {
		t.Error("name match should be case-insensitive")
	}
}

func TestMatchScreen_Precedence(t *testing.T) {
	w := ax.Window{
		ScreenID:   42,
		ScreenName: "DELL U2720Q",
		ScreenUUID: "37D8832A-2D66-02CA-B9F7-8F30A301B230",
	}
	cases := []struct {
		name   string
		filter string
		want   bool
	}{
		{"UUID exact match", "37D8832A-2D66-02CA-B9F7-8F30A301B230", true},
		{"case-insensitive name match", "dell u2720q", true},
		{"decimal numeric ID match", "42", true},
		{"UUID mismatch falls through — neither name nor id match", "99999999-0000-0000-0000-000000000000", false},
		{"completely unknown value", "NotHere", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := window.MatchScreen(w, tc.filter); got != tc.want {
				t.Errorf("MatchScreen(w, %q) = %v, want %v", tc.filter, got, tc.want)
			}
		})
	}
}

func TestMatchScreen_UUIDWinsOverName(t *testing.T) {
	// Window whose ScreenName coincidentally equals the ScreenUUID string form
	// should still stage-1 hit on UUID; this verifies the ordering, not behavior difference.
	w := ax.Window{
		ScreenID:   1,
		ScreenName: "37D8832A-2D66-02CA-B9F7-8F30A301B230",
		ScreenUUID: "37D8832A-2D66-02CA-B9F7-8F30A301B230",
	}
	if !window.MatchScreen(w, "37D8832A-2D66-02CA-B9F7-8F30A301B230") {
		t.Error("expected UUID-form filter to match when UUID equals name")
	}
}

func TestMatchScreen_EmptyUUID_NameAndIDStillWork(t *testing.T) {
	// Back-compat: windows whose UUID is empty still match by name and id.
	w := ax.Window{
		ScreenID:   42,
		ScreenName: "Built-in",
		ScreenUUID: "",
	}
	if !window.MatchScreen(w, "Built-in") {
		t.Error("expected name match when UUID empty")
	}
	if !window.MatchScreen(w, "42") {
		t.Error("expected id match when UUID empty")
	}
}

func TestIsIgnoredApp_EmptyList(t *testing.T) {
	if window.IsIgnoredApp("Safari", "com.apple.Safari", nil) {
		t.Error("empty ignoreApps should never match")
	}
	if window.IsIgnoredApp("Safari", "com.apple.Safari", []string{}) {
		t.Error("empty ignoreApps slice should never match")
	}
}
