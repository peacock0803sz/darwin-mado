// Package cli defines the Cobra subcommands for the mado CLI.
package cli

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/config"
	"github.com/peacock0803sz/mado/internal/output"
)

// RootFlags holds the global flags for the root command.
type RootFlags struct {
	Format  string
	Timeout time.Duration
}

// NewRootCmd creates the root command.
// Uses a constructor pattern without global variables to keep the command testable.
// Loads the config file and implements CLI-flag-over-file priority (T042).
func NewRootCmd(svc ax.WindowService) *cobra.Command {
	// initialize flag defaults from the config file values
	cfg, _ := config.Load() // ignore errors and fall back to defaults

	flags := &RootFlags{
		Format:  cfg.Format,
		Timeout: cfg.Timeout,
	}

	root := &cobra.Command{
		Use:   "mado",
		Short: "macOS window management CLI",
		Long: `mado — a CLI tool for managing macOS windows.

Commands that require Accessibility permission: list, move
Commands that do not require permission: help, version, completion`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// global flags (CLI flags override config file values)
	root.PersistentFlags().StringVar(&flags.Format, "format", cfg.Format, "output format (text|json)")
	root.PersistentFlags().DurationVar(&flags.Timeout, "timeout", cfg.Timeout, "AX operation timeout")

	root.AddCommand(newListCmd(svc, flags))
	root.AddCommand(newMoveCmd(svc, flags))
	root.AddCommand(newVersionCmd())
	root.AddCommand(newCompletionCmd(root))

	return root
}

// newOutputFormat converts a flag string to an output.Format value.
func newOutputFormat(s string) output.Format {
	if s == "json" {
		return output.FormatJSON
	}
	return output.FormatText
}
