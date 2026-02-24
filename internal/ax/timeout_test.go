package ax_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/peacock0803sz/mado/internal/ax"
)

func TestTimeoutError_Message(t *testing.T) {
	err := &ax.TimeoutError{Op: "ListWindows"}
	if err.Error() == "" {
		t.Error("TimeoutError.Error() should not be empty")
	}
	if !strings.Contains(err.Error(), "ListWindows") {
		t.Errorf("TimeoutError.Error() should contain op name, got %q", err.Error())
	}
}

func TestMockWindowService_ListWindows_ContextTimeout(t *testing.T) {
	svc := &ax.MockWindowService{
		Windows: []ax.Window{
			{AppName: "Terminal", Title: "test", PID: 1},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// 文脈タイムアウトが発火するまで待機
	time.Sleep(5 * time.Millisecond)

	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", ctx.Err())
	}

	// モックはコンテキストキャンセルを伝播しないが、タイムアウト後も安全に呼び出せること
	_, err := svc.ListWindows(ctx)
	if err != nil {
		t.Errorf("mock should not propagate ctx cancellation, got: %v", err)
	}
}

func TestMockWindowService_ListScreens_Error(t *testing.T) {
	svc := &ax.MockWindowService{
		ScreensErr: context.DeadlineExceeded,
	}

	_, err := svc.ListScreens(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded propagated from mock, got %v", err)
	}
}

func TestMockWindowService_MoveWindow_Timeout(t *testing.T) {
	svc := &ax.MockWindowService{
		Windows: []ax.Window{
			{AppName: "Terminal", Title: "test", PID: 1, State: ax.StateNormal},
		},
		MoveErr: context.DeadlineExceeded,
	}

	err := svc.MoveWindow(context.Background(), 1, "test", 0, 0)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded from mock MoveWindow, got %v", err)
	}
}
