package ax

import "context"

// WindowService はAX API操作を抽象化するインターフェース。
// cgo依存コードとビジネスロジックを分離し、ユニットテストをモックで実行可能にする。
type WindowService interface {
	// ListWindows は現在開いているすべてのウィンドウを返す。
	// メニューバーのみのアプリ（標準ウィンドウなし）は除外する。
	ListWindows(ctx context.Context) ([]Window, error)

	// ListScreens は接続中のすべてのディスプレイを返す。
	ListScreens(ctx context.Context) ([]Screen, error)

	// MoveWindow は指定プロセス・タイトルのウィンドウを移動する。
	MoveWindow(ctx context.Context, pid uint32, title string, x, y int) error

	// ResizeWindow は指定プロセス・タイトルのウィンドウをリサイズする。
	ResizeWindow(ctx context.Context, pid uint32, title string, w, h int) error

	// CheckPermission はAccessibility権限の有無を確認する。
	// 権限がない場合はPermissionErrorを返す。
	CheckPermission() error
}
