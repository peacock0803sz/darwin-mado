package window_test

import (
	"context"
	"errors"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/window"
)

var moveTestWindows = []ax.Window{
	{AppName: "Terminal", Title: "peacock — zsh", PID: 100, State: ax.StateNormal, Width: 800, Height: 600},
	{AppName: "Safari", Title: "GitHub", PID: 200, State: ax.StateNormal, Width: 1440, Height: 900},
	{AppName: "Safari", Title: "Apple", PID: 200, State: ax.StateNormal, Width: 1200, Height: 800},
	{AppName: "Code", Title: "README.md", PID: 300, State: ax.StateFullscreen, Width: 1440, Height: 900},
}

func TestMove_Position(t *testing.T) {
	svc := &ax.MockWindowService{Windows: moveTestWindows}
	opts := window.MoveOptions{
		AppFilter: "Terminal",
		Position:  &window.Point{X: 0, Y: 0},
	}
	affected, err := window.Move(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(affected) != 1 {
		t.Fatalf("expected 1 affected, got %d", len(affected))
	}
	if affected[0].X != 0 || affected[0].Y != 0 {
		t.Errorf("expected position (0,0), got (%d,%d)", affected[0].X, affected[0].Y)
	}
}

func TestMove_Size(t *testing.T) {
	svc := &ax.MockWindowService{Windows: moveTestWindows}
	opts := window.MoveOptions{
		AppFilter: "Terminal",
		Size:      &window.Size{W: 1024, H: 768},
	}
	affected, err := window.Move(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(affected) != 1 {
		t.Fatalf("expected 1 affected, got %d", len(affected))
	}
	if affected[0].Width != 1024 || affected[0].Height != 768 {
		t.Errorf("expected size (1024,768), got (%d,%d)", affected[0].Width, affected[0].Height)
	}
}

func TestMove_AmbiguousWithoutAll(t *testing.T) {
	svc := &ax.MockWindowService{Windows: moveTestWindows}
	opts := window.MoveOptions{
		AppFilter: "Safari",
		Position:  &window.Point{X: 0, Y: 0},
		All:       false,
	}
	_, err := window.Move(context.Background(), svc, opts)
	if err == nil {
		t.Fatal("expected AmbiguousTargetError, got nil")
	}
	var ambig *ax.AmbiguousTargetError
	if !errors.As(err, &ambig) {
		t.Fatalf("expected *AmbiguousTargetError, got %T: %v", err, err)
	}
	if len(ambig.Candidates) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(ambig.Candidates))
	}
}

func TestMove_AmbiguousWithAll(t *testing.T) {
	svc := &ax.MockWindowService{Windows: moveTestWindows}
	opts := window.MoveOptions{
		AppFilter: "Safari",
		Position:  &window.Point{X: 100, Y: 100},
		All:       true,
	}
	affected, err := window.Move(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(affected) != 2 {
		t.Errorf("expected 2 affected (--all), got %d", len(affected))
	}
}

func TestMove_NotFound(t *testing.T) {
	svc := &ax.MockWindowService{Windows: moveTestWindows}
	opts := window.MoveOptions{
		AppFilter: "NoSuchApp",
		Position:  &window.Point{X: 0, Y: 0},
	}
	_, err := window.Move(context.Background(), svc, opts)
	if err == nil {
		t.Fatal("expected NotFoundError, got nil")
	}
	var notFound *ax.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected *ax.NotFoundError, got %T: %v", err, err)
	}
}

// T035: moving a fullscreen window returns an error equivalent to exit code 5
func TestMove_Fullscreen(t *testing.T) {
	svc := &ax.MockWindowService{Windows: moveTestWindows}
	opts := window.MoveOptions{
		AppFilter: "Code",
		Position:  &window.Point{X: 0, Y: 0},
	}
	_, err := window.Move(context.Background(), svc, opts)
	if err == nil {
		t.Fatal("expected fullscreen error, got nil")
	}
	var fsErr *window.FullscreenError
	if !errors.As(err, &fsErr) {
		t.Fatalf("expected *window.FullscreenError, got %T: %v", err, err)
	}
	if fsErr.Window.AppName == "" {
		t.Error("FullscreenError.Window should be populated")
	}
}

func TestMove_ServiceError(t *testing.T) {
	svc := &ax.MockWindowService{
		Windows: moveTestWindows,
		MoveErr: errors.New("AX error"),
	}
	opts := window.MoveOptions{
		AppFilter: "Terminal",
		Position:  &window.Point{X: 0, Y: 0},
	}
	_, err := window.Move(context.Background(), svc, opts)
	if err == nil {
		t.Fatal("expected service error, got nil")
	}
}

// partialMockService は最初の N 回の MoveWindow を成功させ、それ以降はエラーを返すモック。
type partialMockService struct {
	ax.MockWindowService
	successCount int
	callCount    int
}

func (m *partialMockService) MoveWindow(_ context.Context, _ uint32, _ string, _, _ int) error {
	m.callCount++
	if m.callCount <= m.successCount {
		return nil
	}
	return m.MoveErr
}

func TestMove_PartialSuccess(t *testing.T) {
	svc := &partialMockService{
		MockWindowService: ax.MockWindowService{
			Windows: moveTestWindows,
			MoveErr: errors.New("AX error on second window"),
		},
		successCount: 1,
	}
	opts := window.MoveOptions{
		AppFilter: "Safari",
		Position:  &window.Point{X: 0, Y: 0},
		All:       true,
	}
	affected, err := window.Move(context.Background(), svc, opts)
	if err == nil {
		t.Fatal("expected PartialSuccessError, got nil")
	}
	var partial *ax.PartialSuccessError
	if !errors.As(err, &partial) {
		t.Fatalf("expected *ax.PartialSuccessError, got %T: %v", err, err)
	}
	if len(affected) != 1 {
		t.Errorf("expected 1 affected window in partial success, got %d", len(affected))
	}
	if partial.Cause == nil {
		t.Error("PartialSuccessError.Cause should be non-nil")
	}
}

func TestMove_TitleFilter(t *testing.T) {
	svc := &ax.MockWindowService{Windows: moveTestWindows}
	opts := window.MoveOptions{
		AppFilter:   "Safari",
		TitleFilter: "github", // partial match, case-insensitive
		Position:    &window.Point{X: 50, Y: 50},
	}
	affected, err := window.Move(context.Background(), svc, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(affected) != 1 {
		t.Errorf("expected 1 affected with title filter, got %d", len(affected))
	}
	if affected[0].Title != "GitHub" {
		t.Errorf("expected title 'GitHub', got %q", affected[0].Title)
	}
}
