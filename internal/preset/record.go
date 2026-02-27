package preset

import (
	"context"
	"fmt"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/window"
)

// RecordOptions holds optional parameters for Record.
type RecordOptions struct {
	Screen string // filter by screen name or ID (empty = all screens)
}

// Record captures the current window layout and returns it as a Preset.
// Only windows with StateNormal are included. When multiple windows belong to
// the same application, the title field is populated for disambiguation.
func Record(ctx context.Context, svc ax.WindowService, name string, opts RecordOptions) (*Preset, error) {
	if !namePattern.MatchString(name) {
		return nil, fmt.Errorf("invalid preset name %q: must match %s", name, namePattern.String())
	}

	windows, err := svc.ListWindows(ctx)
	if err != nil {
		return nil, err
	}

	// Count normal windows per application name to decide title inclusion.
	appCount := make(map[string]int)
	var normal []ax.Window
	for _, w := range windows {
		if w.State != ax.StateNormal {
			continue
		}
		if opts.Screen != "" && !window.MatchScreen(w, opts.Screen) {
			continue
		}
		normal = append(normal, w)
		appCount[w.AppName]++
	}

	rules := make([]Rule, 0, len(normal))
	for _, w := range normal {
		r := Rule{
			App:      w.AppName,
			Position: []int{w.X, w.Y},
			Size:     []int{w.Width, w.Height},
		}
		if appCount[w.AppName] > 1 {
			r.Title = w.Title
		}
		// Capture desktop number; skip desktop=-1 (unknown) so the rule matches any desktop.
		if w.Desktop >= 0 {
			d := w.Desktop
			r.Desktop = &d
		}
		rules = append(rules, r)
	}

	return &Preset{
		Name:  name,
		Rules: rules,
	}, nil
}
