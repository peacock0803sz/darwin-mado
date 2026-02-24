// Package ax provides the macOS Accessibility API adapter for window management.
package ax

import "context"

// WindowService is the interface that abstracts AX API operations.
// It separates cgo-dependent code from business logic, enabling unit tests via mocks.
type WindowService interface {
	// ListWindows returns all currently open windows.
	// Menu-bar-only apps (with no standard windows) are excluded.
	ListWindows(ctx context.Context) ([]Window, error)

	// ListScreens returns all connected displays.
	ListScreens(ctx context.Context) ([]Screen, error)

	// MoveWindow moves the window identified by the given process and title.
	MoveWindow(ctx context.Context, pid uint32, title string, x, y int) error

	// ResizeWindow resizes the window identified by the given process and title.
	ResizeWindow(ctx context.Context, pid uint32, title string, w, h int) error

	// CheckPermission verifies that Accessibility permission is granted.
	// Returns a PermissionError if permission is not available.
	CheckPermission() error
}
