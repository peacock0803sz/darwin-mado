package cli

import (
	"context"
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/output"
	"github.com/peacock0803sz/mado/internal/window"
)

// newListCmd creates the list subcommand (T023).
func newListCmd(svc ax.WindowService, root *RootFlags) *cobra.Command {
	var appFilter string
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

			opts := window.ListOptions{
				AppFilter:    appFilter,
				ScreenFilter: screenFilter,
			}
			// When --app is explicitly specified, bypass the ignore list.
			// The user's intent to inspect a specific app takes precedence
			// over the ignore_apps config (FR-006).
			if appFilter == "" {
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

			windows, err := window.List(ctx, svc, opts)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					_ = f.PrintError(6, "AX operation timed out", nil)
					os.Exit(6)
				}
				return err
			}

			return f.PrintWindows(windows)
		},
	}

	cmd.Flags().StringVar(&appFilter, "app", "", "filter by app name (case-insensitive, exact match)")
	cmd.Flags().StringVar(&screenFilter, "screen", "", "filter by screen ID or name (exact match)")
	cmd.Flags().IntVar(&desktopFilter, "desktop", 0, "filter by desktop number (1-based, Mission Control order)")

	return cmd
}
