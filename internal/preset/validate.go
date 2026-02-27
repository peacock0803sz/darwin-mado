package preset

import (
	"fmt"
	"regexp"
)

// namePattern validates preset names: starts with alphanumeric, then alphanumeric/hyphen/underscore.
var namePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// ValidationError represents a single validation failure with context.
type ValidationError struct {
	Preset  string `json:"preset"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("preset %q, %s: %s", e.Preset, e.Field, e.Message)
}

// ValidatePresets checks all presets for structural validity.
// Returns nil when all presets are valid.
func ValidatePresets(presets []Preset) []ValidationError {
	var errs []ValidationError
	seen := make(map[string]bool)

	for i, p := range presets {
		prefix := fmt.Sprintf("presets[%d]", i)

		// Validate preset name
		if p.Name == "" {
			errs = append(errs, ValidationError{
				Preset:  prefix,
				Field:   "name",
				Message: "name is required",
			})
		} else if !namePattern.MatchString(p.Name) {
			errs = append(errs, ValidationError{
				Preset:  p.Name,
				Field:   "name",
				Message: fmt.Sprintf("invalid name %q: must match %s", p.Name, namePattern.String()),
			})
		} else if seen[p.Name] {
			errs = append(errs, ValidationError{
				Preset:  p.Name,
				Field:   "name",
				Message: "duplicate preset name",
			})
		} else {
			seen[p.Name] = true
		}

		name := p.Name
		if name == "" {
			name = prefix
		}

		// Validate rules
		if len(p.Rules) == 0 {
			errs = append(errs, ValidationError{
				Preset:  name,
				Field:   "rules",
				Message: "at least one rule is required",
			})
		}

		for j, r := range p.Rules {
			ruleField := fmt.Sprintf("rules[%d]", j)

			if r.App == "" {
				errs = append(errs, ValidationError{
					Preset:  name,
					Field:   ruleField,
					Message: "app is required",
				})
			}

			// Desktop value -1 is an internal runtime sentinel, not valid in config.
			if r.Desktop != nil && *r.Desktop < 0 {
				errs = append(errs, ValidationError{
					Preset:  name,
					Field:   ruleField + ".desktop",
					Message: "desktop must be >= 0 (0 = all desktops, 1+ = specific desktop)",
				})
			}

			hasPosition := len(r.Position) > 0
			hasSize := len(r.Size) > 0

			if !hasPosition && !hasSize {
				errs = append(errs, ValidationError{
					Preset:  name,
					Field:   ruleField,
					Message: "position or size is required",
				})
			}

			if hasPosition && len(r.Position) != 2 {
				errs = append(errs, ValidationError{
					Preset:  name,
					Field:   ruleField + ".position",
					Message: "position must have exactly 2 values [x, y]",
				})
			}

			if hasSize {
				if len(r.Size) != 2 {
					errs = append(errs, ValidationError{
						Preset:  name,
						Field:   ruleField + ".size",
						Message: "size must have exactly 2 values [width, height]",
					})
				} else {
					if r.Size[0] <= 0 {
						errs = append(errs, ValidationError{
							Preset:  name,
							Field:   ruleField + ".size",
							Message: "width must be positive",
						})
					}
					if r.Size[1] <= 0 {
						errs = append(errs, ValidationError{
							Preset:  name,
							Field:   ruleField + ".size",
							Message: "height must be positive",
						})
					}
				}
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}
