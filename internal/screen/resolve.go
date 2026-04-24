package screen

import (
	"strconv"
	"strings"

	"github.com/peacock0803sz/mado/internal/ax"
)

// Resolve resolves the given filter to a single Screen, implementing the
// precedence defined in FR-005a / research.md R-4:
//
//  1. UUID exact match (case-preserving).
//  2. Case-insensitive match against Screen.Name.
//  3. Decimal string match against Screen.ID.
//
// When stage 2 or 3 matches more than one screen, AmbiguousScreenError is
// returned. When no stage matches, ScreenNotFoundError is returned. Empty
// filter returns ScreenNotFoundError and callers must guard filter != "".
func Resolve(filter string, screens []ax.Screen) (ax.Screen, error) {
	if filter == "" {
		return ax.Screen{}, &ScreenNotFoundError{Filter: filter, Available: cloneScreens(screens)}
	}

	// Stage 1: UUID exact match.
	for _, s := range screens {
		if s.UUID != "" && s.UUID == filter {
			return s, nil
		}
	}

	// Stage 2: case-insensitive name match.
	var nameHits []ax.Screen
	for _, s := range screens {
		if strings.EqualFold(s.Name, filter) {
			nameHits = append(nameHits, s)
		}
	}
	switch len(nameHits) {
	case 1:
		return nameHits[0], nil
	case 0:
		// fall through
	default:
		return ax.Screen{}, &AmbiguousScreenError{Filter: filter, Candidates: nameHits}
	}

	// Stage 3: decimal numeric ID match.
	var idHits []ax.Screen
	for _, s := range screens {
		if strconv.FormatUint(uint64(s.ID), 10) == filter {
			idHits = append(idHits, s)
		}
	}
	switch len(idHits) {
	case 1:
		return idHits[0], nil
	case 0:
		return ax.Screen{}, &ScreenNotFoundError{Filter: filter, Available: cloneScreens(screens)}
	default:
		return ax.Screen{}, &AmbiguousScreenError{Filter: filter, Candidates: idHits}
	}
}

func cloneScreens(in []ax.Screen) []ax.Screen {
	out := make([]ax.Screen, len(in))
	copy(out, in)
	return out
}
