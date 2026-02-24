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

	return cmd
}
