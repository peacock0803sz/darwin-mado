// Package screen provides screen-level resolution for CLI and preset selectors.
package screen

import (
	"fmt"
	"strings"

	"github.com/peacock0803sz/mado/internal/ax"
)

// ScreenNotFoundError is returned when no connected display matches the filter.
// The name intentionally includes the "Screen" prefix (despite package stutter)
// so external consumers that import the package as `screenpkg` / `_ "screen"`
// or use the CLI contract documentation can reference the exact type from
// specs/008-stable-screen-id/contracts/cli-schema.md §6.
//
//nolint:revive // contract-mandated name, see contracts/cli-schema.md §6
type ScreenNotFoundError struct {
	Filter    string
	Available []ax.Screen
}

func (e *ScreenNotFoundError) Error() string {
	return fmt.Sprintf("no display matched %q. connected: [%s]", e.Filter, formatScreens(e.Available))
}

// AmbiguousScreenError is returned when a filter matches more than one display.
// See ScreenNotFoundError for the naming rationale.
//
//nolint:revive // contract-mandated name, see contracts/cli-schema.md §6
type AmbiguousScreenError struct {
	Filter     string
	Candidates []ax.Screen
}

func (e *AmbiguousScreenError) Error() string {
	return fmt.Sprintf("%q is ambiguous. matching: [%s]", e.Filter, formatScreens(e.Candidates))
}

// formatScreens renders each screen as "<uuid> <name> <id>" joined by ", ".
// Empty UUIDs are rendered as "-" so the list stays readable.
func formatScreens(screens []ax.Screen) string {
	if len(screens) == 0 {
		return ""
	}
	parts := make([]string, len(screens))
	for i, s := range screens {
		uuid := s.UUID
		if uuid == "" {
			uuid = "-"
		}
		parts[i] = fmt.Sprintf("%s %s %d", uuid, s.Name, s.ID)
	}
	return strings.Join(parts, ", ")
}
