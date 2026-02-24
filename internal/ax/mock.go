package ax

import "context"

// MockWindowService is a test implementation of WindowService.
// It can be used from tests on any platform.
type MockWindowService struct {
	Windows    []Window
	Screens    []Screen
	PermErr    error
	MoveErr    error
	ResizeErr  error
	ListErr    error
	ScreensErr error
}

// CheckPermission implements WindowService.CheckPermission.
func (m *MockWindowService) CheckPermission() error {
	return m.PermErr
}

// ListWindows implements WindowService.ListWindows.
func (m *MockWindowService) ListWindows(_ context.Context) ([]Window, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	result := make([]Window, len(m.Windows))
	copy(result, m.Windows)
	return result, nil
}

// ListScreens implements WindowService.ListScreens.
func (m *MockWindowService) ListScreens(_ context.Context) ([]Screen, error) {
	if m.ScreensErr != nil {
		return nil, m.ScreensErr
	}
	result := make([]Screen, len(m.Screens))
	copy(result, m.Screens)
	return result, nil
}

// MoveWindow implements WindowService.MoveWindow.
func (m *MockWindowService) MoveWindow(_ context.Context, _ uint32, _ string, _, _ int) error {
	return m.MoveErr
}

// ResizeWindow implements WindowService.ResizeWindow.
func (m *MockWindowService) ResizeWindow(_ context.Context, _ uint32, _ string, _, _ int) error {
	return m.ResizeErr
}
