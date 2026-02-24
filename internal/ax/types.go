package ax

// WindowState represents the display state of a window.
type WindowState string

// Window state constants for the State field of Window.
const (
	StateNormal     WindowState = "normal"
	StateMinimized  WindowState = "minimized"
	StateFullscreen WindowState = "fullscreen"
	StateHidden     WindowState = "hidden"
)

// Window represents an individual application window on macOS.
type Window struct {
	AppName    string      `json:"app_name"`
	Title      string      `json:"title"`
	PID        uint32      `json:"pid"`
	X          int         `json:"x"`
	Y          int         `json:"y"`
	Width      int         `json:"width"`
	Height     int         `json:"height"`
	State      WindowState `json:"state"`
	ScreenID   uint32      `json:"screen_id"`
	ScreenName string      `json:"screen_name"`
}

// Screen represents an individual display on macOS.
type Screen struct {
	ID        uint32 `json:"id"`
	Name      string `json:"name"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	IsPrimary bool   `json:"is_primary"`
}

// Application represents a running application on macOS.
type Application struct {
	Name    string   `json:"name"`
	PID     uint32   `json:"pid"`
	Windows []Window `json:"windows"`
}
