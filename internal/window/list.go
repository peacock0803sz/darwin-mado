// Package window implements the business logic for listing and moving macOS windows.
package window

import (
	"context"
	"strconv"
	"strings"

	"github.com/peacock0803sz/mado/internal/ax"
)

// ListOptions holds filter options for the list command.
type ListOptions struct {
	AppFilter     string
	ScreenFilter  string
	IgnoreApps    []string
	DesktopFilter int // 0 = no filter; N = only windows on desktop N (plus desktop=0 windows)
}

// List retrieves all windows and returns them after applying filters.
func List(ctx context.Context, svc ax.WindowService, opts ListOptions) ([]ax.Window, error) {
	windows, err := svc.ListWindows(ctx)
	if err != nil {
		return nil, err
	}

	return filterWindows(windows, opts), nil
}

// filterWindows narrows down the window list based on filter options.
func filterWindows(windows []ax.Window, opts ListOptions) []ax.Window {
	result := make([]ax.Window, 0, len(windows))
	for _, w := range windows {
		if opts.AppFilter != "" && !strings.EqualFold(w.AppName, opts.AppFilter) {
			continue
		}
		if IsIgnoredApp(w.AppName, opts.IgnoreApps) {
			continue
		}
		if opts.ScreenFilter != "" && !MatchScreen(w, opts.ScreenFilter) {
			continue
		}
		if !MatchDesktop(w, opts.DesktopFilter) {
			continue
		}
		result = append(result, w)
	}
	return result
}

// MatchDesktop reports whether w should pass a desktop filter.
// filter=0 passes all windows. Windows with Desktop=0 (all desktops) always pass.
func MatchDesktop(w ax.Window, filter int) bool {
	if filter == 0 {
		return true
	}
	if w.Desktop == 0 {
		return true
	}
	return w.Desktop == filter
}

// IsIgnoredApp returns true if appName matches any entry in ignoreApps (case-insensitive).
func IsIgnoredApp(appName string, ignoreApps []string) bool {
	for _, ignored := range ignoreApps {
		if strings.EqualFold(appName, ignored) {
			return true
		}
	}
	return false
}

// MatchScreen filters a window by screen ID (numeric string) or screen name (case-insensitive).
func MatchScreen(w ax.Window, filter string) bool {
	if strings.EqualFold(w.ScreenName, filter) {
		return true
	}
	return strconv.FormatUint(uint64(w.ScreenID), 10) == filter
}
