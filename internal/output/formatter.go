package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/peacock0803sz/mado/internal/ax"
)

// Format represents the type of output format.
type Format string

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

// truncate truncates s to at most max runes.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
