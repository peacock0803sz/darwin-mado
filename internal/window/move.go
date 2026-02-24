package window

import (
	"context"
	"strings"

	"github.com/peacock0803sz/mado/internal/ax"
)

// Point represents a 2D coordinate.
type Point struct {
	X, Y int
}

// Size represents a 2D size.
type Size struct {
	W, H int
}

// MoveOptions holds the options for the move command.
type MoveOptions struct {
	AppFilter    string
	TitleFilter  string
	ScreenFilter string
	Position     *Point
	Size         *Size
	All          bool
}

// Move moves or resizes the target window(s).
// Returns AmbiguousTargetError when multiple windows match and --all is not set.
func Move(ctx context.Context, svc ax.WindowService, opts MoveOptions) ([]ax.Window, error) {
	windows, err := svc.ListWindows(ctx)
	if err != nil {
		return nil, err
	}

	targets := filterForMove(windows, opts)

	if len(targets) == 0 {
		return nil, &ax.NotFoundError{Query: buildQuery(opts)}
	}

	if len(targets) > 1 && !opts.All {
		return nil, &ax.AmbiguousTargetError{
			Query:      buildQuery(opts),
			Candidates: targets,
		}
	}

	var affected []ax.Window
	for _, w := range targets {
		// fullscreen windows cannot be operated on (exit 5)
		if w.State == ax.StateFullscreen {
			return nil, &FullscreenError{Window: w}
		}

		if opts.Position != nil {
			if err := svc.MoveWindow(ctx, w.PID, w.Title, opts.Position.X, opts.Position.Y); err != nil {
				return affected, err
			}
			w.X = opts.Position.X
			w.Y = opts.Position.Y
		}

		if opts.Size != nil {
			if err := svc.ResizeWindow(ctx, w.PID, w.Title, opts.Size.W, opts.Size.H); err != nil {
				return affected, err
			}
			w.Width = opts.Size.W
			w.Height = opts.Size.H
		}

		affected = append(affected, w)
	}

	return affected, nil
}

// filterForMove filters windows for the move command.
func filterForMove(windows []ax.Window, opts MoveOptions) []ax.Window {
	result := make([]ax.Window, 0)
	for _, w := range windows {
		if opts.AppFilter != "" && !strings.EqualFold(w.AppName, opts.AppFilter) {
			continue
		}
		if opts.TitleFilter != "" && !strings.Contains(strings.ToLower(w.Title), strings.ToLower(opts.TitleFilter)) {
			continue
		}
		if opts.ScreenFilter != "" && !matchScreen(w, opts.ScreenFilter) {
			continue
		}
		result = append(result, w)
	}
	return result
}

func buildQuery(opts MoveOptions) string {
	parts := make([]string, 0)
	if opts.AppFilter != "" {
		parts = append(parts, `--app "`+opts.AppFilter+`"`)
	}
	if opts.TitleFilter != "" {
		parts = append(parts, `--title "`+opts.TitleFilter+`"`)
	}
	if len(parts) == 0 {
		return "(no filter)"
	}
	return strings.Join(parts, " ")
}

// FullscreenError is returned when attempting to operate on a fullscreen window.
type FullscreenError struct {
	Window ax.Window
}

func (e *FullscreenError) Error() string {
	return `cannot move fullscreen window: "` + e.Window.Title + `"`
}
