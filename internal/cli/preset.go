package cli

import (
	"context"
	"errors"
	"os"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/config"
	"github.com/peacock0803sz/mado/internal/output"
	"github.com/peacock0803sz/mado/internal/preset"
)

// newPresetCmd creates the preset command group with apply/list/show/validate subcommands.
func newPresetCmd(svc ax.WindowService, flags *RootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preset",
		Short: "Manage window layout presets",
		Long:  "Apply, list, show, or validate window layout presets defined in the config file.",
	}

	cmd.AddCommand(newPresetApplyCmd(svc, flags))
	cmd.AddCommand(newPresetListCmd(flags))
	cmd.AddCommand(newPresetRecCmd(svc, flags))
	cmd.AddCommand(newPresetShowCmd(flags))
	cmd.AddCommand(newPresetValidateCmd(flags))

	return cmd
}

func newPresetApplyCmd(svc ax.WindowService, flags *RootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "apply <name>",
		Short: "Apply a preset layout to matching windows",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f := output.New(newOutputFormat(flags.Format), os.Stdout, os.Stderr)
			name := args[0]

			if err := svc.CheckPermission(); err != nil {
				msg := err.Error()
				if permErr, ok := err.(*ax.PermissionError); ok {
					msg = permErr.Error() + "\n\n" + permErr.Resolution()
				}
				_ = f.PrintError(2, msg, nil)
				os.Exit(2)
			}

			cfg, err := config.Load()
			if err != nil {
				_ = f.PrintError(3, err.Error(), nil)
				os.Exit(3)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), flags.Timeout)
			defer cancel()

			outcome, err := preset.Apply(ctx, svc, cfg.Presets, name)
			if err != nil {
				return handleApplyError(f, err, outcome)
			}

			return f.PrintPresetApplyResult(buildApplyResponse(name, outcome, true, nil))
		},
	}
}

func handleApplyError(f *output.Formatter, err error, outcome *preset.ApplyOutcome) error {
	var notFound *preset.NotFoundError
	if errors.As(err, &notFound) {
		_ = f.PrintError(4, notFound.Error(), nil)
		os.Exit(4)
	}

	var allFS *preset.AllFullscreenError
	if errors.As(err, &allFS) {
		if outcome != nil {
			_ = f.PrintPresetApplyResult(buildApplyResponse(outcome.PresetName, outcome, false,
				&output.ErrorDetail{Code: 5, Message: allFS.Error()}))
		} else {
			_ = f.PrintError(5, allFS.Error(), nil)
		}
		os.Exit(5)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		_ = f.PrintError(6, "AX operation timed out", nil)
		os.Exit(6)
	}

	var partialErr *ax.PartialSuccessError
	if errors.As(err, &partialErr) {
		if outcome != nil {
			_ = f.PrintPresetApplyResult(buildApplyResponse(outcome.PresetName, outcome, false,
				&output.ErrorDetail{Code: 7, Message: partialErr.Error()}))
		}
		os.Exit(7)
	}

	_ = f.PrintError(1, err.Error(), nil)
	os.Exit(1)
	return nil
}

func buildApplyResponse(name string, outcome *preset.ApplyOutcome, success bool, errDetail *output.ErrorDetail) output.PresetApplyResponse {
	resp := output.PresetApplyResponse{
		SchemaVersion: 1,
		Success:       success,
		Preset:        name,
		Error:         errDetail,
		Applied:       make([]output.PresetApplyAffected, 0),
		Skipped:       make([]output.PresetApplySkipped, 0),
	}

	for _, r := range outcome.Results {
		if r.Skipped {
			resp.Skipped = append(resp.Skipped, output.PresetApplySkipped{
				RuleIndex: r.RuleIndex,
				AppFilter: r.AppFilter,
				Reason:    r.Reason,
			})
		} else if len(r.Affected) > 0 {
			resp.Applied = append(resp.Applied, output.PresetApplyAffected{
				RuleIndex: r.RuleIndex,
				AppFilter: r.AppFilter,
				Affected:  r.Affected,
			})
		}
	}

	return resp
}

func newPresetRecCmd(svc ax.WindowService, flags *RootFlags) *cobra.Command {
	var screen string

	cmd := &cobra.Command{
		Use:   "rec <name> [output-path]",
		Short: "Record current window layout as a preset",
		Long:  "Capture the current window positions and sizes and output them as a YAML preset definition.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			f := output.New(newOutputFormat(flags.Format), os.Stdout, os.Stderr)
			name := args[0]

			outputPath := "stdout"
			if len(args) >= 2 {
				outputPath = args[1]
			}

			if err := svc.CheckPermission(); err != nil {
				msg := err.Error()
				if permErr, ok := err.(*ax.PermissionError); ok {
					msg = permErr.Error() + "\n\n" + permErr.Resolution()
				}
				_ = f.PrintError(2, msg, nil)
				os.Exit(2)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), flags.Timeout)
			defer cancel()

			p, err := preset.Record(ctx, svc, name, preset.RecordOptions{
				Screen: screen,
			})
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					_ = f.PrintError(6, "AX operation timed out", nil)
					os.Exit(6)
				}
				_ = f.PrintError(3, err.Error(), nil)
				os.Exit(3)
			}

			data, err := yaml.Marshal(p)
			if err != nil {
				_ = f.PrintError(1, err.Error(), nil)
				os.Exit(1)
			}

			if outputPath == "stdout" {
				_, err = cmd.OutOrStdout().Write(data)
				return err
			}

			if err := os.WriteFile(outputPath, data, 0o600); err != nil {
				_ = f.PrintError(1, err.Error(), nil)
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&screen, "screen", "", "record only windows on the specified screen (name or ID)")

	return cmd
}

func newPresetListCmd(flags *RootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all defined presets",
		RunE: func(_ *cobra.Command, _ []string) error {
			f := output.New(newOutputFormat(flags.Format), os.Stdout, os.Stderr)

			cfg, err := config.Load()
			if err != nil {
				_ = f.PrintError(3, err.Error(), nil)
				os.Exit(3)
			}

			return f.PrintPresetList(cfg.Presets)
		},
	}
}

func newPresetShowCmd(flags *RootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show details of a preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			f := output.New(newOutputFormat(flags.Format), os.Stdout, os.Stderr)
			name := args[0]

			cfg, err := config.Load()
			if err != nil {
				_ = f.PrintError(3, err.Error(), nil)
				os.Exit(3)
			}

			for _, p := range cfg.Presets {
				if p.Name == name {
					return f.PrintPresetShow(p)
				}
			}

			_ = f.PrintError(4, (&preset.NotFoundError{Name: name}).Error(), nil)
			os.Exit(4)
			return nil
		},
	}
}

func newPresetValidateCmd(flags *RootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate preset configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			f := output.New(newOutputFormat(flags.Format), os.Stdout, os.Stderr)

			cfg, err := config.Load()
			if err != nil {
				// Catch validation errors from config.Load
				_ = f.PrintError(3, err.Error(), nil)
				os.Exit(3)
			}

			// If Load succeeds, presets are already validated
			return f.PrintPresetValidateResult(len(cfg.Presets), nil)
		},
	}
}
