//go:build !darwin

package ax

// NewWindowService returns a mock WindowService for testing.
// This file is compiled on non-darwin platforms (e.g., Ubuntu CI).
func NewWindowService() WindowService {
	return &MockWindowService{}
}
