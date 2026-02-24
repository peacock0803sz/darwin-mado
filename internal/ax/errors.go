package ax

import "fmt"

// PermissionError is returned when Accessibility permission is not granted.
// The error message includes resolution steps for macOS.
type PermissionError struct{}

func (e *PermissionError) Error() string {
	return "Accessibility permission not granted"
}

// Resolution returns instructions for resolving the permission issue.
func (e *PermissionError) Resolution() string {
	return `To grant permission:
  1. Open System Settings → Privacy & Security → Accessibility
  2. Enable mado (or your Terminal app) in the list
  3. Re-run the command`
}

// NotFoundError is returned when no windows match the given query.
type NotFoundError struct {
	Query string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("no windows match %q", e.Query)
}

// AmbiguousTargetError is returned when multiple windows match the given query.
type AmbiguousTargetError struct {
	Query      string
	Candidates []Window
}

func (e *AmbiguousTargetError) Error() string {
	return fmt.Sprintf("ambiguous target: %d windows match %q", len(e.Candidates), e.Query)
}

// TimeoutError is returned when an AX operation exceeds the allowed duration.
type TimeoutError struct {
	Op string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("AX operation timed out: %s", e.Op)
}
