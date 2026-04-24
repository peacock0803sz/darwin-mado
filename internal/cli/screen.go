package cli

import (
	"context"
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/output"
)

// newScreenCmd creates the `screen` command group.
// `mado screen list` is safe to run without Accessibility permission — it only
// queries CoreGraphics/NSScreen, which do not require TCC access (R-7).
func newScreenCmd(svc ax.WindowService, flags *RootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screen",
		Short: "Manage display information",
		Long:  "Inspect connected displays. Subcommands are non-AX and do not require Accessibility permission.",
	}
	cmd.AddCommand(newScreenListCmd(svc, flags))
	return cmd
}

func newScreenListCmd(svc ax.WindowService, flags *RootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List connected displays with stable UUIDs",
		Long: `List all connected displays with their stable UUIDs, localized names,
transient numeric IDs, geometry, and primary flag. UUIDs survive reboots and
reconnection order changes — copy them into presets and --screen flags.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			f := output.New(newOutputFormat(flags.Format), os.Stdout, os.Stderr)

			ctx, cancel := context.WithTimeout(cmd.Context(), flags.Timeout)
			defer cancel()

			screens, err := svc.ListScreens(ctx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					_ = f.PrintError(6, "AX operation timed out", nil)
					os.Exit(6)
				}
				_ = f.PrintError(1, err.Error(), nil)
				os.Exit(1)
			}

			Verbosef(flags.Verbose, cmd.ErrOrStderr(), "found %d screen(s)", len(screens))
			return f.PrintScreenList(screens)
		},
	}
}
