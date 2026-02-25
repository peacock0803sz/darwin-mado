// Package output formats mado command results as text or JSON.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/preset"
)

// Format represents the type of output format.
type Format string

// Output format constants.
const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// ListResponse is the JSON output schema for the list command.
type ListResponse struct {
	SchemaVersion int         `json:"schema_version"`
	Success       bool        `json:"success"`
	Windows       []ax.Window `json:"windows"`
}

// MoveResponse is the JSON output schema for the move command.
type MoveResponse struct {
	SchemaVersion int         `json:"schema_version"`
	Success       bool        `json:"success"`
	Affected      []ax.Window `json:"affected"`
}

// ErrorResponse is the JSON output schema for error responses.
type ErrorResponse struct {
	SchemaVersion int          `json:"schema_version"`
	Success       bool         `json:"success"`
	Error         *ErrorDetail `json:"error"`
}

// ErrorDetail contains the details of an error response.
type ErrorDetail struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Candidates []ax.Window `json:"candidates,omitempty"`
}

// Formatter writes output in either text or JSON format.
type Formatter struct {
	format Format
	out    io.Writer
	errOut io.Writer
}

// New creates a new Formatter.
func New(format Format, out, errOut io.Writer) *Formatter {
	return &Formatter{format: format, out: out, errOut: errOut}
}

// IsTerminal reports whether stdout is connected to a TTY.
func IsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// PrintWindows outputs the list of windows.
func (f *Formatter) PrintWindows(windows []ax.Window) error {
	if f.format == FormatJSON {
		return f.printJSON(ListResponse{
			SchemaVersion: 1,
			Success:       true,
			Windows:       windows,
		})
	}
	return f.printWindowsText(windows)
}

// PrintMoveResult outputs the result of a move operation.
func (f *Formatter) PrintMoveResult(affected []ax.Window) error {
	if f.format == FormatJSON {
		return f.printJSON(MoveResponse{
			SchemaVersion: 1,
			Success:       true,
			Affected:      affected,
		})
	}
	for _, w := range affected {
		if _, err := fmt.Fprintf(f.out, "Moved: %s %q → (%d, %d)\n", w.AppName, w.Title, w.X, w.Y); err != nil {
			return err
		}
	}
	return nil
}

// PrintError formats and outputs an error message.
func (f *Formatter) PrintError(code int, message string, candidates []ax.Window) error {
	if f.format == FormatJSON {
		return f.printJSON(ErrorResponse{
			SchemaVersion: 1,
			Success:       false,
			Error: &ErrorDetail{
				Code:       code,
				Message:    message,
				Candidates: candidates,
			},
		})
	}
	return f.printErrorText(code, message, candidates)
}

// --- Preset response types ---

// PresetApplyAffected represents a rule's affected windows in apply output.
type PresetApplyAffected struct {
	RuleIndex int         `json:"rule_index"`
	AppFilter string      `json:"app_filter"`
	Affected  []ax.Window `json:"affected"`
}

// PresetApplySkipped represents a skipped rule in apply output.
type PresetApplySkipped struct {
	RuleIndex int    `json:"rule_index"`
	AppFilter string `json:"app_filter"`
	Reason    string `json:"reason"`
}

// PresetApplyResponse is the JSON output for preset apply.
type PresetApplyResponse struct {
	SchemaVersion int                   `json:"schema_version"`
	Success       bool                  `json:"success"`
	Preset        string                `json:"preset"`
	Applied       []PresetApplyAffected `json:"applied"`
	Skipped       []PresetApplySkipped  `json:"skipped"`
	Error         *ErrorDetail          `json:"error,omitempty"`
}

// PresetListItem represents a single preset in list output.
type PresetListItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RuleCount   int    `json:"rule_count"`
}

// PresetListResponse is the JSON output for preset list.
type PresetListResponse struct {
	SchemaVersion int              `json:"schema_version"`
	Success       bool             `json:"success"`
	Presets       []PresetListItem `json:"presets"`
}

// PresetShowResponse is the JSON output for preset show.
type PresetShowResponse struct {
	SchemaVersion int           `json:"schema_version"`
	Success       bool          `json:"success"`
	Preset        preset.Preset `json:"preset"`
}

// PresetValidateResponse is the JSON output for preset validate.
type PresetValidateResponse struct {
	SchemaVersion    int                      `json:"schema_version"`
	Success          bool                     `json:"success"`
	PresetsValidated int                      `json:"presets_validated"`
	Errors           []preset.ValidationError `json:"errors"`
}

// PrintPresetApplyResult outputs the result of a preset apply operation.
func (f *Formatter) PrintPresetApplyResult(resp PresetApplyResponse) error {
	if f.format == FormatJSON {
		return f.printJSON(resp)
	}
	return f.printPresetApplyText(resp)
}

func (f *Formatter) printPresetApplyText(resp PresetApplyResponse) error {
	fmt.Fprintf(f.out, "Preset %q applied:\n", resp.Preset) //nolint:errcheck
	for _, a := range resp.Applied {
		for _, w := range a.Affected {
			fmt.Fprintf(f.out, "  %s %q → (%d, %d) %dx%d\n", w.AppName, w.Title, w.X, w.Y, w.Width, w.Height) //nolint:errcheck
		}
	}
	for _, s := range resp.Skipped {
		fmt.Fprintf(f.out, "Skipped (%s): %s\n", s.Reason, s.AppFilter) //nolint:errcheck
	}
	return nil
}

// PrintPresetList outputs the list of presets.
func (f *Formatter) PrintPresetList(presets []preset.Preset) error {
	if f.format == FormatJSON {
		items := make([]PresetListItem, len(presets))
		for i, p := range presets {
			items[i] = PresetListItem{
				Name:        p.Name,
				Description: p.Description,
				RuleCount:   len(p.Rules),
			}
		}
		return f.printJSON(PresetListResponse{
			SchemaVersion: 1,
			Success:       true,
			Presets:       items,
		})
	}
	return f.printPresetListText(presets)
}

func (f *Formatter) printPresetListText(presets []preset.Preset) error {
	if len(presets) == 0 {
		_, err := fmt.Fprintln(f.out, "(no presets)")
		return err
	}
	tw := tabwriter.NewWriter(f.out, 8, 1, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tRULES\tDESCRIPTION") //nolint:errcheck
	for _, p := range presets {
		fmt.Fprintf(tw, "%s\t%d\t%s\n", p.Name, len(p.Rules), p.Description) //nolint:errcheck
	}
	return tw.Flush()
}

// PrintPresetShow outputs the details of a single preset.
func (f *Formatter) PrintPresetShow(p preset.Preset) error {
	if f.format == FormatJSON {
		return f.printJSON(PresetShowResponse{
			SchemaVersion: 1,
			Success:       true,
			Preset:        p,
		})
	}
	return f.printPresetShowText(p)
}

func (f *Formatter) printPresetShowText(p preset.Preset) error {
	fmt.Fprintf(f.out, "Preset: %s\n", p.Name) //nolint:errcheck
	if p.Description != "" {
		fmt.Fprintf(f.out, "Description: %s\n", p.Description) //nolint:errcheck
	}
	fmt.Fprintln(f.out, "Rules:") //nolint:errcheck
	for i, r := range p.Rules {
		line := fmt.Sprintf("  [%d] app=%s", i, r.App)
		if r.Title != "" {
			line += fmt.Sprintf(" title=%q", r.Title)
		}
		if r.Screen != "" {
			line += fmt.Sprintf(" screen=%s", r.Screen)
		}
		if len(r.Position) == 2 {
			line += fmt.Sprintf(" position=(%d,%d)", r.Position[0], r.Position[1])
		}
		if len(r.Size) == 2 {
			line += fmt.Sprintf(" size=%dx%d", r.Size[0], r.Size[1])
		}
		fmt.Fprintln(f.out, line) //nolint:errcheck
	}
	return nil
}

// PrintPresetValidateResult outputs the result of preset validation.
func (f *Formatter) PrintPresetValidateResult(count int, errs []preset.ValidationError) error {
	if f.format == FormatJSON {
		if errs == nil {
			errs = []preset.ValidationError{}
		}
		return f.printJSON(PresetValidateResponse{
			SchemaVersion:    1,
			Success:          len(errs) == 0,
			PresetsValidated: count,
			Errors:           errs,
		})
	}
	return f.printPresetValidateText(count, errs)
}

func (f *Formatter) printPresetValidateText(count int, errs []preset.ValidationError) error {
	if len(errs) == 0 {
		_, err := fmt.Fprintf(f.out, "All %d presets are valid.\n", count)
		return err
	}
	fmt.Fprintln(f.errOut, "Validation errors:") //nolint:errcheck
	for _, e := range errs {
		fmt.Fprintf(f.errOut, "  preset %q, %s: %s\n", e.Preset, e.Field, e.Message) //nolint:errcheck
	}
	return nil
}

func (f *Formatter) printWindowsText(windows []ax.Window) error {
	if len(windows) == 0 {
		_, err := fmt.Fprintln(f.out, "(no windows)")
		return err
	}

	// align columns with tabwriter (min width 8, tab width 1, padding 2)
	tw := tabwriter.NewWriter(f.out, 8, 1, 2, ' ', 0)
	fmt.Fprintln(tw, "APP_NAME\tTITLE\tX\tY\tWIDTH\tHEIGHT\tSTATE\tSCREEN") //nolint:errcheck // tabwriter defers errors to Flush()

	for _, w := range windows {
		screenName := truncate(w.ScreenName, 20)
		if w.State == ax.StateMinimized || w.State == ax.StateHidden {
			screenName = "-"
		}
		title := truncate(w.Title, 32)
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\t%d\t%s\t%s\n", //nolint:errcheck // tabwriter defers errors to Flush()
			w.AppName, title, w.X, w.Y, w.Width, w.Height, w.State, screenName)
	}
	return tw.Flush()
}

func (f *Formatter) printErrorText(code int, message string, candidates []ax.Window) error {
	fmt.Fprintf(f.errOut, "Error: %s\n", message) //nolint:errcheck // stderr write errors are not actionable

	if len(candidates) > 0 {
		fmt.Fprintln(f.errOut, "\nCandidates:") //nolint:errcheck
		for i, c := range candidates {
			fmt.Fprintf(f.errOut, "  %d. %s %q\tpid=%d\n", i+1, c.AppName, c.Title, c.PID) //nolint:errcheck
		}
		fmt.Fprintln(f.errOut, "\nHint: use --title to narrow down, or --all to move all") //nolint:errcheck
	}

	_ = code
	return nil
}

func (f *Formatter) printJSON(v any) error {
	enc := json.NewEncoder(f.out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// truncate truncates s to at most limit runes.
func truncate(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit-1]) + "…"
}
