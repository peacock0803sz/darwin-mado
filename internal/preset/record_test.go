package preset

import (
	"context"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
)

func TestRecord_TwoWindows(t *testing.T) {
	svc := &ax.MockWindowService{
		Windows: []ax.Window{
			{AppName: "Code", Title: "main.go", PID: 1, X: 0, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal},
			{AppName: "Terminal", Title: "zsh", PID: 2, X: 960, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal},
		},
	}

	p, err := Record(context.Background(), svc, "coding", RecordOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Name != "coding" {
		t.Errorf("name = %q, want %q", p.Name, "coding")
	}
	if len(p.Rules) != 2 {
		t.Fatalf("len(rules) = %d, want 2", len(p.Rules))
	}

	// Different apps → no title
	for i, r := range p.Rules {
		if r.Title != "" {
			t.Errorf("rules[%d].Title = %q, want empty (different apps)", i, r.Title)
		}
	}

	r0 := p.Rules[0]
	if r0.App != "Code" {
		t.Errorf("rules[0].App = %q, want %q", r0.App, "Code")
	}
	if r0.Position[0] != 0 || r0.Position[1] != 0 {
		t.Errorf("rules[0].Position = %v, want [0, 0]", r0.Position)
	}
	if r0.Size[0] != 960 || r0.Size[1] != 1080 {
		t.Errorf("rules[0].Size = %v, want [960, 1080]", r0.Size)
	}
}

func TestRecord_SameAppMultipleWindows(t *testing.T) {
	svc := &ax.MockWindowService{
		Windows: []ax.Window{
			{AppName: "Code", Title: "main.go", PID: 1, X: 0, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal},
			{AppName: "Code", Title: "test.go", PID: 1, X: 960, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal},
		},
	}

	p, err := Record(context.Background(), svc, "dual-editor", RecordOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Rules) != 2 {
		t.Fatalf("len(rules) = %d, want 2", len(p.Rules))
	}

	// Same app → title should be populated
	if p.Rules[0].Title != "main.go" {
		t.Errorf("rules[0].Title = %q, want %q", p.Rules[0].Title, "main.go")
	}
	if p.Rules[1].Title != "test.go" {
		t.Errorf("rules[1].Title = %q, want %q", p.Rules[1].Title, "test.go")
	}
}

func TestRecord_FiltersNonNormal(t *testing.T) {
	svc := &ax.MockWindowService{
		Windows: []ax.Window{
			{AppName: "Code", Title: "main.go", PID: 1, X: 0, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal},
			{AppName: "Finder", Title: "Downloads", PID: 3, State: ax.StateMinimized},
			{AppName: "Safari", Title: "Google", PID: 4, State: ax.StateFullscreen},
			{AppName: "Mail", Title: "Inbox", PID: 5, State: ax.StateHidden},
		},
	}

	p, err := Record(context.Background(), svc, "filtered", RecordOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Rules) != 1 {
		t.Fatalf("len(rules) = %d, want 1 (only normal windows)", len(p.Rules))
	}
	if p.Rules[0].App != "Code" {
		t.Errorf("rules[0].App = %q, want %q", p.Rules[0].App, "Code")
	}
}

func TestRecord_NoWindows(t *testing.T) {
	svc := &ax.MockWindowService{}

	p, err := Record(context.Background(), svc, "empty", RecordOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Rules) != 0 {
		t.Errorf("len(rules) = %d, want 0", len(p.Rules))
	}
}

func TestRecord_InvalidName(t *testing.T) {
	svc := &ax.MockWindowService{}

	_, err := Record(context.Background(), svc, "invalid name!", RecordOptions{})
	if err == nil {
		t.Fatal("expected error for invalid preset name, got nil")
	}
}

func TestRecord_ScreenFilter(t *testing.T) {
	svc := &ax.MockWindowService{
		Windows: []ax.Window{
			{AppName: "Code", Title: "main.go", PID: 1, X: 0, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal, ScreenID: 1, ScreenName: "Built-in Retina Display"},
			{AppName: "Terminal", Title: "zsh", PID: 2, X: 960, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal, ScreenID: 2, ScreenName: "DELL U2720Q"},
			{AppName: "Safari", Title: "Google", PID: 3, X: 0, Y: 0, Width: 1920, Height: 1080, State: ax.StateNormal, ScreenID: 2, ScreenName: "DELL U2720Q"},
		},
	}

	// Filter by screen name
	p, err := Record(context.Background(), svc, "external", RecordOptions{Screen: "DELL U2720Q"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Rules) != 2 {
		t.Fatalf("len(rules) = %d, want 2", len(p.Rules))
	}
	if p.Rules[0].App != "Terminal" {
		t.Errorf("rules[0].App = %q, want %q", p.Rules[0].App, "Terminal")
	}
	if p.Rules[1].App != "Safari" {
		t.Errorf("rules[1].App = %q, want %q", p.Rules[1].App, "Safari")
	}
}

func TestRecord_ScreenFilterByID(t *testing.T) {
	svc := &ax.MockWindowService{
		Windows: []ax.Window{
			{AppName: "Code", Title: "main.go", PID: 1, X: 0, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal, ScreenID: 1, ScreenName: "Built-in"},
			{AppName: "Terminal", Title: "zsh", PID: 2, X: 960, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal, ScreenID: 2, ScreenName: "External"},
		},
	}

	// Filter by screen ID
	p, err := Record(context.Background(), svc, "main-only", RecordOptions{Screen: "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Rules) != 1 {
		t.Fatalf("len(rules) = %d, want 1", len(p.Rules))
	}
	if p.Rules[0].App != "Code" {
		t.Errorf("rules[0].App = %q, want %q", p.Rules[0].App, "Code")
	}
}

func TestRecord_ScreenFilterNoMatch(t *testing.T) {
	svc := &ax.MockWindowService{
		Windows: []ax.Window{
			{AppName: "Code", Title: "main.go", PID: 1, X: 0, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal, ScreenID: 1, ScreenName: "Built-in"},
		},
	}

	p, err := Record(context.Background(), svc, "nomatch", RecordOptions{Screen: "NonExistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Rules) != 0 {
		t.Errorf("len(rules) = %d, want 0", len(p.Rules))
	}
}

func intPtr(v int) *int { return &v }

func TestRecord_DesktopField(t *testing.T) {
	cases := []struct {
		desktop int
		want    *int
	}{
		{-1, nil},      // unknown → no constraint
		{0, intPtr(0)}, // all desktops
		{1, intPtr(1)}, // desktop 1
		{3, intPtr(3)}, // desktop 3
	}

	for _, tc := range cases {
		svc := &ax.MockWindowService{
			Windows: []ax.Window{
				{AppName: "Code", Title: "main.go", PID: 1, X: 0, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal, Desktop: tc.desktop},
			},
		}
		p, err := Record(context.Background(), svc, "desktop-test", RecordOptions{})
		if err != nil {
			t.Fatalf("desktop=%d: unexpected error: %v", tc.desktop, err)
		}
		if len(p.Rules) != 1 {
			t.Fatalf("desktop=%d: len(rules) = %d, want 1", tc.desktop, len(p.Rules))
		}
		got := p.Rules[0].Desktop
		if tc.want == nil {
			if got != nil {
				t.Errorf("desktop=%d: Rule.Desktop = %v, want nil", tc.desktop, *got)
			}
		} else {
			if got == nil {
				t.Errorf("desktop=%d: Rule.Desktop = nil, want %d", tc.desktop, *tc.want)
			} else if *got != *tc.want {
				t.Errorf("desktop=%d: Rule.Desktop = %d, want %d", tc.desktop, *got, *tc.want)
			}
		}
	}
}

func TestRecord_DesktopUnknownOmitted(t *testing.T) {
	svc := &ax.MockWindowService{
		Windows: []ax.Window{
			{AppName: "Finder", Title: "Downloads", PID: 1, X: 0, Y: 0, Width: 800, Height: 600, State: ax.StateNormal, Desktop: -1},
		},
	}
	p, err := Record(context.Background(), svc, "minimized", RecordOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Rules) != 1 {
		t.Fatalf("len(rules) = %d, want 1", len(p.Rules))
	}
	if p.Rules[0].Desktop != nil {
		t.Errorf("Rule.Desktop = %v, want nil for Desktop=-1", *p.Rules[0].Desktop)
	}
}

func TestRecord_MixedDesktops(t *testing.T) {
	svc := &ax.MockWindowService{
		Windows: []ax.Window{
			{AppName: "Code", Title: "main.go", PID: 1, X: 0, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal, Desktop: 1},
			{AppName: "Code", Title: "test.go", PID: 1, X: 0, Y: 0, Width: 960, Height: 1080, State: ax.StateNormal, Desktop: 2},
		},
	}
	p, err := Record(context.Background(), svc, "mixed", RecordOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Rules) != 2 {
		t.Fatalf("len(rules) = %d, want 2", len(p.Rules))
	}
	if p.Rules[0].Desktop == nil || *p.Rules[0].Desktop != 1 {
		t.Errorf("rules[0].Desktop = %v, want 1", p.Rules[0].Desktop)
	}
	if p.Rules[1].Desktop == nil || *p.Rules[1].Desktop != 2 {
		t.Errorf("rules[1].Desktop = %v, want 2", p.Rules[1].Desktop)
	}
}
