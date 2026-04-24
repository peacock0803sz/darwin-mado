// Package cli defines the Cobra subcommands for the mado CLI.
package cli

import (
	"fmt"
	"io"
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
	Verbose    bool
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
Commands that do not require permission: help, version, completion, screen list, preset list, preset show, preset validate`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if skipConfigLoad(cmd) {
				return nil
			}
			result, err := config.Load()
			if err != nil {
				f := output.New(newOutputFormat(flags.Format), os.Stdout, os.Stderr)
				_ = f.PrintError(3, err.Error(), nil)
				os.Exit(3)
			}
			cfg := result.Config
			if !cmd.Root().PersistentFlags().Changed("format") {
				flags.Format = cfg.Format
			}
			if !cmd.Root().PersistentFlags().Changed("timeout") {
				flags.Timeout = cfg.Timeout
			}
			if !cmd.Root().PersistentFlags().Changed("verbose") {
				flags.Verbose = cfg.Verbose
			}
			flags.Presets = cfg.Presets
			flags.IgnoreApps = cfg.IgnoreApps

			// verbose: config loading diagnostics
			stderr := cmd.ErrOrStderr()
			if result.SourcePath != "" {
				Verbosef(flags.Verbose, stderr, "config loaded from %s", result.SourcePath)
			} else {
				Verbosef(flags.Verbose, stderr, "no config file found, using defaults")
			}
			Verbosef(flags.Verbose, stderr, "format=%s timeout=%s verbose=%t", flags.Format, flags.Timeout, flags.Verbose)
			if len(flags.IgnoreApps) > 0 {
				Verbosef(flags.Verbose, stderr, "ignore_apps=%v", flags.IgnoreApps)
			}
			return nil
		},
	}

	// global flags (CLI flags override config file values)
	root.PersistentFlags().StringVar(&flags.Format, "format", def.Format, "output format (text|json)")
	root.PersistentFlags().DurationVar(&flags.Timeout, "timeout", def.Timeout, "AX operation timeout")
	root.PersistentFlags().BoolVar(&flags.Verbose, "verbose", false, "enable verbose diagnostic output to stderr")

	root.AddCommand(newListCmd(svc, flags))
	root.AddCommand(newMoveCmd(svc, flags))
	root.AddCommand(newPresetCmd(svc, flags))
	root.AddCommand(newScreenCmd(svc, flags))
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

// Verbosef writes a formatted diagnostic message to w when verbose is true.
// Messages are prefixed with "verbose: " and terminated with a newline.
// Write failures are silently ignored to avoid affecting exit codes.
func Verbosef(verbose bool, w io.Writer, format string, args ...any) {
	if !verbose {
		return
	}
	_, _ = fmt.Fprintf(w, "verbose: "+format+"\n", args...)
}
