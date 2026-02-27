package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/output"
	"github.com/peacock0803sz/mado/internal/window"
)

// newMoveCmd creates the move subcommand (T029).
func newMoveCmd(svc ax.WindowService, root *RootFlags) *cobra.Command {
	var (
		appFilter     string
		titleFilter   string
		screenFilter  string
		desktopFilter int
		positionStr   string
		sizeStr       string
		all           bool
	)

	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move or resize a window",
		RunE: func(cmd *cobra.Command, _ []string) error {
			f := output.New(newOutputFormat(root.Format), os.Stdout, os.Stderr)

			// T030: exit 3 when neither --position nor --size is specified
			if positionStr == "" && sizeStr == "" {
				_ = f.PrintError(3, "--position or --size is required", nil)
				os.Exit(3)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), root.Timeout)
			defer cancel()

			if err := svc.CheckPermission(); err != nil {
				msg := err.Error()
				if permErr, ok := err.(*ax.PermissionError); ok {
					msg = permErr.Error() + "\n\n" + permErr.Resolution()
				}
				_ = f.PrintError(2, msg, nil)
				os.Exit(2)
			}

			opts := window.MoveOptions{
				AppFilter:    appFilter,
				TitleFilter:  titleFilter,
				ScreenFilter: screenFilter,
				All:          all,
			}
			// Only apply desktop filter when explicitly specified.
			if cmd.Flags().Changed("desktop") {
				if desktopFilter < 1 {
					_ = f.PrintError(3, "invalid --desktop value: must be a positive integer", nil)
					os.Exit(3)
				}
				opts.DesktopFilter = desktopFilter
			}

			if positionStr != "" {
				x, y, err := parseCoords(positionStr)
				if err != nil {
					_ = f.PrintError(3, fmt.Sprintf("invalid --position value: %v", err), nil)
					os.Exit(3)
				}
				opts.Position = &window.Point{X: x, Y: y}
			}

			if sizeStr != "" {
				w, h, err := parseCoords(sizeStr)
				if err != nil {
					_ = f.PrintError(3, fmt.Sprintf("invalid --size value: %v", err), nil)
					os.Exit(3)
				}
				if w <= 0 || h <= 0 {
					_ = f.PrintError(3, "--size width and height must be positive integers", nil)
					os.Exit(3)
				}
				opts.Size = &window.Size{W: w, H: h}
			}

			affected, err := window.Move(ctx, svc, opts)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					_ = f.PrintError(6, "AX operation timed out", nil)
					os.Exit(6)
				}
				var fsErr *window.FullscreenError
				if errors.As(err, &fsErr) {
					_ = f.PrintError(5, err.Error(), nil)
					os.Exit(5)
				}
				var partialErr *ax.PartialSuccessError
				if errors.As(err, &partialErr) {
					_ = f.PrintMoveResult(partialErr.Affected)
					_ = f.PrintError(7, partialErr.Cause.Error(), nil)
					os.Exit(7)
				}
				switch e := err.(type) {
				case *ax.NotFoundError:
					_ = f.PrintError(4, e.Error(), nil)
					os.Exit(4)
				case *ax.AmbiguousTargetError:
					_ = f.PrintError(4, e.Error(), e.Candidates)
					os.Exit(4)
				default:
					return e
				}
			}

			return f.PrintMoveResult(affected)
		},
	}

	cmd.Flags().StringVar(&appFilter, "app", "", "filter by app name (case-insensitive, exact match)")
	cmd.Flags().StringVar(&titleFilter, "title", "", "filter by title (case-insensitive, partial match)")
	cmd.Flags().StringVar(&screenFilter, "screen", "", "filter by screen ID or name")
	cmd.Flags().IntVar(&desktopFilter, "desktop", 0, "scope operation to desktop number (1-based, Mission Control order)")
	cmd.Flags().StringVar(&positionStr, "position", "", "target position x,y (global coordinates)")
	cmd.Flags().StringVar(&sizeStr, "size", "", "target size width,height")
	cmd.Flags().BoolVar(&all, "all", false, "apply to all matching windows when multiple match")

	return cmd
}

// parseCoords parses a "x,y" formatted string into two integers.
func parseCoords(s string) (int, int, error) {
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("specify two values separated by a comma (e.g. 100,200)")
	}
	a, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("first value is not an integer: %q", parts[0])
	}
	b, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("second value is not an integer: %q", parts[1])
	}
	return a, b, nil
}
