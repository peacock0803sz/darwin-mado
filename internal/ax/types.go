package ax

// WindowState はウィンドウの表示状態を表す。
type WindowState string

const (
	StateNormal     WindowState = "normal"
	StateMinimized  WindowState = "minimized"
	StateFullscreen WindowState = "fullscreen"
	StateHidden     WindowState = "hidden"
)

// Window はmacOS上の個別アプリケーションウィンドウを表す。
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

// Screen はmacOS上の個別ディスプレイを表す。
type Screen struct {
	ID        uint32 `json:"id"`
	Name      string `json:"name"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	IsPrimary bool   `json:"is_primary"`
}

// Application はmacOS上で実行中のアプリケーションを表す。
type Application struct {
	Name    string   `json:"name"`
	PID     uint32   `json:"pid"`
	Windows []Window `json:"windows"`
}
