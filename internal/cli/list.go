package cli

import (
	"context"
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/output"
	"github.com/peacock0803sz/mado/internal/screen"
	"github.com/peacock0803sz/mado/internal/window"
)

// resolveScreenFilterErr converts a user-provided --screen value into a
// canonical filter string that window.MatchScreen stage-1 hits. Returns:
//   - filter == ""       → ("", nil)
//   - resolvable          → (uuid | name, nil)
//   - unresolved/ambiguous → ("", error from screen package)
//   - AX list failure     → ("", wrapped error)
func resolveScreenFilterErr(ctx context.Context, svc ax.WindowService, filter string) (string, error) {
	if filter == "" {
		return "", nil
	}
	screens, err := svc.ListScreens(ctx)
	if err != nil {
		return "", err
	}
	s, err := screen.Resolve(filter, screens)
	if err != nil {
		return "", err
	}
	if s.UUID != "" {
		return s.UUID, nil
	}
	// Fallback: resolved screen without UUID — use the name for stage-2 match.
	return s.Name, nil
}

// resolveScreenFilter calls resolveScreenFilterErr and translates errors
// into the contract's exit codes: 1 for AX failure, 4 for resolve failure.
func resolveScreenFilter(ctx context.Context, svc ax.WindowService, f *output.Formatter, filter string) string {
	resolved, err := resolveScreenFilterErr(ctx, svc, filter)
	if err == nil {
		return resolved
	}
	var notFound *screen.ScreenNotFoundError
	var ambiguous *screen.AmbiguousScreenError
	if errors.As(err, &notFound) || errors.As(err, &ambiguous) {
		_ = f.PrintError(4, err.Error(), nil)
		os.Exit(4)
	}
	_ = f.PrintError(1, err.Error(), nil)
	os.Exit(1)
	return "" // unreachable
}

// newListCmd creates the list subcommand (T023).
func newListCmd(svc ax.WindowService, root *RootFlags) *cobra.Command {
	var appFilter string
	var appIDFilter string
	var screenFilter string
	var desktopFilter int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List currently open windows",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), root.Timeout)
			defer cancel()

			f := output.New(newOutputFormat(root.Format), os.Stdout, os.Stderr)

			if err := svc.CheckPermission(); err != nil {
				msg := err.Error()
				if permErr, ok := err.(*ax.PermissionError); ok {
					msg = permErr.Error() + "\n\n" + permErr.Resolution()
				}
				_ = f.PrintError(2, msg, nil)
				os.Exit(2)
			}

			resolvedScreen := resolveScreenFilter(ctx, svc, f, screenFilter)
			opts := window.ListOptions{
				AppFilter:    appFilter,
				AppIDFilter:  appIDFilter,
				ScreenFilter: resolvedScreen,
			}
			// When --app or --app-id is explicitly specified, bypass the ignore list.
			// The user's intent to inspect a specific app takes precedence
			// over the ignore_apps config (FR-006).
			if appFilter == "" && appIDFilter == "" {
				opts.IgnoreApps = root.IgnoreApps
			}
			// Only apply desktop filter when explicitly specified.
			if cmd.Flags().Changed("desktop") {
				if desktopFilter < 1 {
					_ = f.PrintError(3, "invalid --desktop value: must be a positive integer", nil)
					os.Exit(3)
				}
				opts.DesktopFilter = desktopFilter
			}

			// verbose: log filter options
			stderr := cmd.ErrOrStderr()
			if appFilter != "" {
				Verbosef(root.Verbose, stderr, "filter: app=%q", appFilter)
			}
			if appIDFilter != "" {
				Verbosef(root.Verbose, stderr, "filter: app-id=%q", appIDFilter)
			}
			if screenFilter != "" {
				Verbosef(root.Verbose, stderr, "filter: screen=%q", screenFilter)
			}
			if len(opts.IgnoreApps) > 0 {
				Verbosef(root.Verbose, stderr, "ignore_apps=%v", opts.IgnoreApps)
			}

			windows, err := window.List(ctx, svc, opts)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					_ = f.PrintError(6, "AX operation timed out", nil)
					os.Exit(6)
				}
				return err
			}

			Verbosef(root.Verbose, stderr, "found %d windows", len(windows))
			return f.PrintWindows(windows)
		},
	}

	cmd.Flags().StringVar(&appFilter, "app", "", "filter by app name (case-insensitive, exact match)")
	cmd.Flags().StringVar(&appIDFilter, "app-id", "", "filter by bundle identifier (case-insensitive, exact match)")
	cmd.Flags().StringVar(&screenFilter, "screen", "", "filter by stable UUID, localized name, or transient numeric ID (see `mado screen list`)")
	cmd.Flags().IntVar(&desktopFilter, "desktop", 0, "filter by desktop number (1-based, Mission Control order)")

	return cmd
}
