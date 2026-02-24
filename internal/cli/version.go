package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newVersionCmd creates the version subcommand (T050).
// Does not require Accessibility permission.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, _ []string) {
			// cmd.Root().Version is injected by main.go via ldflags
			fmt.Fprintf(cmd.OutOrStdout(), "mado version %s\n", cmd.Root().Version) //nolint:errcheck
		},
	}
}
