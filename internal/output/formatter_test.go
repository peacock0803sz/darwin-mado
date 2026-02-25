package output_test

import (
	"bytes"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/output"
	"github.com/peacock0803sz/mado/internal/preset"
	"github.com/sebdah/goldie/v2"
)

var multiScreenWindows = []ax.Window{
	{
		AppName:    "Terminal",
		Title:      "peacock — zsh — 80×24",
		PID:        1234,
		X:          100,
		Y:          200,
		Width:      800,
		Height:     600,
		State:      ax.StateNormal,
		ScreenID:   69678592,
		ScreenName: "Built-in Retina Display",
	},
	{
		AppName:    "Safari",
		Title:      "GitHub",
		PID:        5678,
		X:          -1920,
		Y:          0,
		Width:      1920,
		Height:     1080,
		State:      ax.StateNormal,
		ScreenID:   12345678,
		ScreenName: "DELL U2720Q",
	},
	{
		AppName:  "Finder",
		Title:    "",
		PID:      300,
		Width:    800,
		Height:   600,
		State:    ax.StateMinimized,
		ScreenID: 0,
	},
}

var sampleWindows = []ax.Window{
	{
		AppName:    "Terminal",
		Title:      "peacock — zsh — 80×24",
		PID:        1234,
		X:          100,
		Y:          200,
		Width:      800,
		Height:     600,
		State:      ax.StateNormal,
		ScreenID:   69678592,
		ScreenName: "Built-in Retina Display",
	},
	{
		AppName:    "Safari",
		Title:      "GitHub",
		PID:        5678,
		X:          0,
		Y:          0,
		Width:      1440,
		Height:     900,
		State:      ax.StateNormal,
		ScreenID:   69678592,
		ScreenName: "Built-in Retina Display",
	},
	{
		AppName:  "Safari",
		Title:    "Apple",
		PID:      5678,
		Width:    1200,
		Height:   800,
		State:    ax.StateMinimized,
		ScreenID: 0,
	},
}

func TestPrintWindowsText(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatText, &buf, &buf)
	if err := f.PrintWindows(sampleWindows); err != nil {
		t.Fatal(err)
	}

	g := goldie.New(t)
	g.Assert(t, "list_text", buf.Bytes())
}

func TestPrintWindowsJSON(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatJSON, &buf, &buf)
	if err := f.PrintWindows(sampleWindows); err != nil {
		t.Fatal(err)
	}

	g := goldie.New(t)
	g.AssertJson(t, "list_json", buf.Bytes())
}

func TestPrintWindowsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		format output.Format
		golden string
	}{
		{"empty text", output.FormatText, "list_empty_text"},
		{"empty json", output.FormatJSON, "list_empty_json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := output.New(tt.format, &buf, &buf)
			if err := f.PrintWindows(nil); err != nil {
				t.Fatal(err)
			}
			g := goldie.New(t)
			if tt.format == output.FormatJSON {
				g.AssertJson(t, tt.golden, buf.Bytes())
			} else {
				g.Assert(t, tt.golden, buf.Bytes())
			}
		})
	}
}

func TestPrintMoveResultText(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatText, &buf, &buf)
	affected := []ax.Window{sampleWindows[0]}
	affected[0].X = 0
	affected[0].Y = 0
	if err := f.PrintMoveResult(affected); err != nil {
		t.Fatal(err)
	}
	g := goldie.New(t)
	g.Assert(t, "move_success_text", buf.Bytes())
}

func TestPrintMoveResultJSON(t *testing.T) {
	var buf bytes.Buffer
	f := output.New(output.FormatJSON, &buf, &buf)
	affected := []ax.Window{sampleWindows[0]}
	affected[0].X = 0
	affected[0].Y = 0
	if err := f.PrintMoveResult(affected); err != nil {
		t.Fatal(err)
	}
	g := goldie.New(t)
	g.AssertJson(t, "move_success_json", buf.Bytes())
}

func TestPrintWindowsMultiScreen(t *testing.T) {
	tests := []struct {
		name   string
		format output.Format
		golden string
	}{
		{"multiscreen text", output.FormatText, "list_multiscreen_text"},
		{"multiscreen json", output.FormatJSON, "list_multiscreen_json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := output.New(tt.format, &buf, &buf)
			if err := f.PrintWindows(multiScreenWindows); err != nil {
				t.Fatal(err)
			}
			g := goldie.New(t)
			if tt.format == output.FormatJSON {
				g.AssertJson(t, tt.golden, buf.Bytes())
			} else {
				g.Assert(t, tt.golden, buf.Bytes())
			}
		})
	}
}

var samplePresets = []preset.Preset{
	{
		Name:        "coding",
		Description: "Editor left, terminal right",
		Rules: []preset.Rule{
			{App: "Code", Position: []int{0, 0}, Size: []int{960, 1080}},
			{App: "Terminal", Position: []int{960, 0}, Size: []int{960, 1080}},
		},
	},
	{
		Name:        "meeting",
		Description: "Browser center, notes right",
		Rules: []preset.Rule{
			{App: "Safari", Title: "Zoom", Position: []int{0, 0}, Size: []int{1280, 1080}},
			{App: "Notes", Position: []int{1280, 0}, Size: []int{640, 1080}},
		},
	},
}

func TestPrintPresetList(t *testing.T) {
	tests := []struct {
		name   string
		format output.Format
		golden string
	}{
		{"text", output.FormatText, "preset_list_text"},
		{"json", output.FormatJSON, "preset_list_json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := output.New(tt.format, &buf, &buf)
			if err := f.PrintPresetList(samplePresets); err != nil {
				t.Fatal(err)
			}
			g := goldie.New(t)
			if tt.format == output.FormatJSON {
				g.AssertJson(t, tt.golden, buf.Bytes())
			} else {
				g.Assert(t, tt.golden, buf.Bytes())
			}
		})
	}
}

func TestPrintPresetShow(t *testing.T) {
	tests := []struct {
		name   string
		format output.Format
		golden string
	}{
		{"text", output.FormatText, "preset_show_text"},
		{"json", output.FormatJSON, "preset_show_json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := output.New(tt.format, &buf, &buf)
			if err := f.PrintPresetShow(samplePresets[0]); err != nil {
				t.Fatal(err)
			}
			g := goldie.New(t)
			if tt.format == output.FormatJSON {
				g.AssertJson(t, tt.golden, buf.Bytes())
			} else {
				g.Assert(t, tt.golden, buf.Bytes())
			}
		})
	}
}

func TestPrintPresetApply(t *testing.T) {
	resp := output.PresetApplyResponse{
		SchemaVersion: 1,
		Success:       true,
		Preset:        "coding",
		Applied: []output.PresetApplyAffected{
			{
				RuleIndex: 0,
				AppFilter: "Code",
				Affected: []ax.Window{
					{AppName: "Code", Title: "main.go", X: 0, Y: 0, Width: 960, Height: 1080},
				},
			},
			{
				RuleIndex: 1,
				AppFilter: "Terminal",
				Affected: []ax.Window{
					{AppName: "Terminal", Title: "zsh", X: 960, Y: 0, Width: 960, Height: 1080},
				},
			},
		},
		Skipped: []output.PresetApplySkipped{},
	}
	tests := []struct {
		name   string
		format output.Format
		golden string
	}{
		{"text", output.FormatText, "preset_apply_text"},
		{"json", output.FormatJSON, "preset_apply_json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := output.New(tt.format, &buf, &buf)
			if err := f.PrintPresetApplyResult(resp); err != nil {
				t.Fatal(err)
			}
			g := goldie.New(t)
			if tt.format == output.FormatJSON {
				g.AssertJson(t, tt.golden, buf.Bytes())
			} else {
				g.Assert(t, tt.golden, buf.Bytes())
			}
		})
	}
}

func TestPrintPresetValidate(t *testing.T) {
	tests := []struct {
		name   string
		format output.Format
		count  int
		errs   []preset.ValidationError
		golden string
	}{
		{"valid text", output.FormatText, 3, nil, "preset_validate_valid_text"},
		{"valid json", output.FormatJSON, 3, nil, "preset_validate_valid_json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := output.New(tt.format, &buf, &buf)
			if err := f.PrintPresetValidateResult(tt.count, tt.errs); err != nil {
				t.Fatal(err)
			}
			g := goldie.New(t)
			if tt.format == output.FormatJSON {
				g.AssertJson(t, tt.golden, buf.Bytes())
			} else {
				g.Assert(t, tt.golden, buf.Bytes())
			}
		})
	}
}

func TestPrintAmbiguousError(t *testing.T) {
	candidates := sampleWindows[1:]
	tests := []struct {
		name   string
		format output.Format
		golden string
	}{
		{"ambiguous text", output.FormatText, "move_ambiguous_text"},
		{"ambiguous json", output.FormatJSON, "move_ambiguous_json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := output.New(tt.format, &buf, &buf)
			if err := f.PrintError(4, `ambiguous target: 2 windows match --app "Safari"`, candidates); err != nil {
				t.Fatal(err)
			}
			g := goldie.New(t)
			if tt.format == output.FormatJSON {
				g.AssertJson(t, tt.golden, buf.Bytes())
			} else {
				g.Assert(t, tt.golden, buf.Bytes())
			}
		})
	}
}
