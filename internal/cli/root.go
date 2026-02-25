// Package cli defines the Cobra subcommands for the mado CLI.
package cli

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/config"
	"github.com/peacock0803sz/mado/internal/output"
	"github.com/peacock0803sz/mado/internal/preset"
)

// RootFlags holds the global flags for the root command.
type RootFlags struct {
	Format     string
	Timeout    time.Duration
	Presets    []preset.Preset
	IgnoreApps []string
}

// NewRootCmd creates the root command.
// Uses a constructor pattern without global variables to keep the command testable.
// Loads the config file and implements CLI-flag-over-file priority (T042).
func NewRootCmd(svc ax.WindowService) *cobra.Command {
	def := config.Default()

	flags := &RootFlags{
		Format:  def.Format,
		Timeout: def.Timeout,
	}

	root := &cobra.Command{
		Use:   "mado",
		Short: "macOS window management CLI",
		Long: `mado — a CLI tool for managing macOS windows.

Commands that require Accessibility permission: list, move, preset apply, preset rec
Commands that do not require permission: help, version, completion, preset list, preset show, preset validate`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if skipConfigLoad(cmd) {
				return nil
			}
			cfg, err := config.Load()
			if err != nil {
				f := output.New(newOutputFormat(flags.Format), os.Stdout, os.Stderr)
				_ = f.PrintError(3, err.Error(), nil)
				os.Exit(3)
			}
			if !cmd.Root().PersistentFlags().Changed("format") {
				flags.Format = cfg.Format
			}
			if !cmd.Root().PersistentFlags().Changed("timeout") {
				flags.Timeout = cfg.Timeout
			}
			flags.Presets = cfg.Presets
			flags.IgnoreApps = cfg.IgnoreApps
			return nil
		},
	}

	// global flags (CLI flags override config file values)
	root.PersistentFlags().StringVar(&flags.Format, "format", def.Format, "output format (text|json)")
	root.PersistentFlags().DurationVar(&flags.Timeout, "timeout", def.Timeout, "AX operation timeout")

	root.AddCommand(newListCmd(svc, flags))
	root.AddCommand(newMoveCmd(svc, flags))
	root.AddCommand(newPresetCmd(svc, flags))
	root.AddCommand(newVersionCmd())
	root.AddCommand(newCompletionCmd(root))

	return root
}

// skipConfigLoad returns true for commands that don't need config loading.
func skipConfigLoad(cmd *cobra.Command) bool {
	if !cmd.HasParent() {
		return true
	}
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "version", "completion", "help":
			return true
		}
	}
	return false
}

// newOutputFormat converts a flag string to an output.Format value.
func newOutputFormat(s string) output.Format {
	if s == "json" {
		return output.FormatJSON
	}
	return output.FormatText
}
