package preset_test

import (
	"context"
	"errors"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/preset"
)

var testPresets = []preset.Preset{
	{
		Name:        "coding",
		Description: "Editor left, terminal right",
		Rules: []preset.Rule{
			{App: "Code", Position: []int{0, 0}, Size: []int{960, 1080}},
			{App: "Terminal", Position: []int{960, 0}, Size: []int{960, 1080}},
		},
	},
	{
		Name:        "meeting",
		Description: "Browser and notes",
		Rules: []preset.Rule{
			{App: "Safari", Title: "Zoom", Position: []int{0, 0}, Size: []int{1280, 1080}},
			{App: "Notes", Position: []int{1280, 0}, Size: []int{640, 1080}},
		},
	},
}

var testWindows = []ax.Window{
	{AppName: "Code", Title: "main.go", PID: 100, State: ax.StateNormal, Width: 800, Height: 600},
	{AppName: "Terminal", Title: "zsh", PID: 200, State: ax.StateNormal, Width: 800, Height: 600},
	{AppName: "Safari", Title: "GitHub", PID: 300, State: ax.StateNormal, Width: 1440, Height: 900},
	{AppName: "Safari", Title: "Zoom Meeting", PID: 300, State: ax.StateNormal, Width: 1440, Height: 900},
	{AppName: "Safari", Title: "Apple", PID: 300, State: ax.StateNormal, Width: 1200, Height: 800},
	{AppName: "Notes", Title: "Meeting Notes", PID: 400, State: ax.StateNormal, Width: 640, Height: 480},
}

func TestApply_Success(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	outcome, err := preset.Apply(context.Background(), svc, testPresets, "coding", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var appliedCount int
	for _, r := range outcome.Results {
		if !r.Skipped && len(r.Affected) > 0 {
			appliedCount++
		}
	}
	if appliedCount != 2 {
		t.Errorf("expected 2 rules applied, got %d", appliedCount)
	}
}

func TestApply_SkipNonRunningApp(t *testing.T) {
	// テスト用のウィンドウにSlackがない場合、そのルールはスキップされる
	presets := []preset.Preset{{
		Name: "test",
		Rules: []preset.Rule{
			{App: "Code", Position: []int{0, 0}},
			{App: "Slack", Position: []int{960, 0}}, // Slackは起動していない
		},
	}}
	svc := &ax.MockWindowService{Windows: testWindows}
	outcome, err := preset.Apply(context.Background(), svc, presets, "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var skippedCount int
	for _, r := range outcome.Results {
		if r.Skipped && r.Reason == "no_match" {
			skippedCount++
		}
	}
	if skippedCount != 1 {
		t.Errorf("expected 1 skipped rule (Slack), got %d", skippedCount)
	}
}

func TestApply_SkipFullscreen(t *testing.T) {
	windows := []ax.Window{
		{AppName: "Code", Title: "main.go", PID: 100, State: ax.StateFullscreen, Width: 1440, Height: 900},
		{AppName: "Terminal", Title: "zsh", PID: 200, State: ax.StateNormal, Width: 800, Height: 600},
	}
	svc := &ax.MockWindowService{Windows: windows}
	outcome, err := preset.Apply(context.Background(), svc, testPresets, "coding", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var fullscreenSkipped bool
	for _, r := range outcome.Results {
		if r.Skipped && r.Reason == "fullscreen" {
			fullscreenSkipped = true
		}
	}
	if !fullscreenSkipped {
		t.Error("expected fullscreen window to be skipped")
	}
}

func TestApply_AllFullscreen(t *testing.T) {
	// 全マッチがフルスクリーンの場合はAllFullscreenErrorを返す
	windows := []ax.Window{
		{AppName: "Code", Title: "main.go", PID: 100, State: ax.StateFullscreen, Width: 1440, Height: 900},
		{AppName: "Terminal", Title: "zsh", PID: 200, State: ax.StateFullscreen, Width: 1440, Height: 900},
	}
	svc := &ax.MockWindowService{Windows: windows}
	_, err := preset.Apply(context.Background(), svc, testPresets, "coding", nil)
	if err == nil {
		t.Fatal("expected AllFullscreenError, got nil")
	}
	var allFS *preset.AllFullscreenError
	if !errors.As(err, &allFS) {
		t.Fatalf("expected *preset.AllFullscreenError, got %T: %v", err, err)
	}
}

func TestApply_MultiWindowMatch(t *testing.T) {
	// 1つのルールで複数のSafariウィンドウにマッチする
	presets := []preset.Preset{{
		Name: "browse",
		Rules: []preset.Rule{
			{App: "Safari", Position: []int{0, 0}, Size: []int{1440, 900}},
		},
	}}
	svc := &ax.MockWindowService{Windows: testWindows}
	outcome, err := preset.Apply(context.Background(), svc, presets, "browse", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var totalAffected int
	for _, r := range outcome.Results {
		totalAffected += len(r.Affected)
	}
	// Safari has 3 windows in testWindows
	if totalAffected != 3 {
		t.Errorf("expected 3 affected Safari windows, got %d", totalAffected)
	}
}

func TestApply_FirstMatchDedup(t *testing.T) {
	// 重複するルール: SafariのZoomウィンドウは最初のルールにのみマッチ
	presets := []preset.Preset{{
		Name: "dedup",
		Rules: []preset.Rule{
			{App: "Safari", Title: "Zoom", Position: []int{0, 0}, Size: []int{1280, 1080}},
			{App: "Safari", Position: []int{1280, 0}, Size: []int{640, 1080}},
		},
	}}
	svc := &ax.MockWindowService{Windows: testWindows}
	outcome, err := preset.Apply(context.Background(), svc, presets, "dedup", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 最初のルールはZoomのみ (1ウィンドウ)
	// 2番目のルールはZoomを除くSafari (2ウィンドウ)
	if len(outcome.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(outcome.Results))
	}
	if len(outcome.Results[0].Affected) != 1 {
		t.Errorf("rule 0: expected 1 affected (Zoom), got %d", len(outcome.Results[0].Affected))
	}
	if len(outcome.Results[1].Affected) != 2 {
		t.Errorf("rule 1: expected 2 affected (other Safari), got %d", len(outcome.Results[1].Affected))
	}
}

// partialMockService は最初の N 回の操作を成功させ、それ以降はエラーを返すモック
type partialMockService struct {
	ax.MockWindowService
	moveSuccessCount   int
	moveCallCount      int
	resizeSuccessCount int
	resizeCallCount    int
}

func (m *partialMockService) MoveWindow(_ context.Context, _ uint32, _ string, _, _ int) error {
	m.moveCallCount++
	if m.moveCallCount <= m.moveSuccessCount {
		return nil
	}
	return m.MoveErr
}

func (m *partialMockService) ResizeWindow(_ context.Context, _ uint32, _ string, _, _ int) error {
	m.resizeCallCount++
	if m.resizeCallCount <= m.resizeSuccessCount {
		return nil
	}
	return m.ResizeErr
}

func TestApply_PartialSuccess(t *testing.T) {
	svc := &partialMockService{
		MockWindowService: ax.MockWindowService{
			Windows: testWindows,
			MoveErr: errors.New("AX error"),
		},
		moveSuccessCount:   1,
		resizeSuccessCount: 100,
	}
	_, err := preset.Apply(context.Background(), svc, testPresets, "coding", nil)
	if err == nil {
		// Codeは1つのウィンドウなので成功する。
		// Terminalも1つなので2回目のMoveWindowがエラーになる。
		// 結果: 1 success + 1 fail = partial success
		t.Fatal("expected error for partial success, got nil")
	}
	var partialErr *ax.PartialSuccessError
	if !errors.As(err, &partialErr) {
		t.Fatalf("expected *ax.PartialSuccessError, got %T: %v", err, err)
	}
}

func TestApply_NotFound(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	_, err := preset.Apply(context.Background(), svc, testPresets, "nonexistent", nil)
	if err == nil {
		t.Fatal("expected NotFoundError, got nil")
	}
	var notFound *preset.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected *preset.NotFoundError, got %T: %v", err, err)
	}
}

func TestApply_IgnoredAppSkipped(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	ignoreApps := []string{"Code"}
	outcome, err := preset.Apply(context.Background(), svc, testPresets, "coding", ignoreApps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ignoredCount int
	for _, r := range outcome.Results {
		if r.Skipped && r.Reason == "ignored" {
			ignoredCount++
		}
	}
	if ignoredCount != 1 {
		t.Errorf("expected 1 ignored rule (Code), got %d", ignoredCount)
	}
}

func TestApply_IgnoredAppNonIgnoredStillApplies(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	ignoreApps := []string{"Code"}
	outcome, err := preset.Apply(context.Background(), svc, testPresets, "coding", ignoreApps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Code is ignored, Terminal should still be applied
	var appliedCount int
	for _, r := range outcome.Results {
		if !r.Skipped && len(r.Affected) > 0 {
			appliedCount++
			if r.AppFilter != "Terminal" {
				t.Errorf("expected applied rule for Terminal, got %q", r.AppFilter)
			}
		}
	}
	if appliedCount != 1 {
		t.Errorf("expected 1 applied rule (Terminal), got %d", appliedCount)
	}
}

func TestApply_EmptyIgnoreAppsNoSkipping(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	outcome, err := preset.Apply(context.Background(), svc, testPresets, "coding", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, r := range outcome.Results {
		if r.Reason == "ignored" {
			t.Error("no rules should be ignored with empty ignoreApps")
		}
	}
}

func TestApply_IgnoredAppCaseInsensitive(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	ignoreApps := []string{"code"} // lowercase, rule.App is "Code"
	outcome, err := preset.Apply(context.Background(), svc, testPresets, "coding", ignoreApps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var ignoredCount int
	for _, r := range outcome.Results {
		if r.Skipped && r.Reason == "ignored" {
			ignoredCount++
		}
	}
	if ignoredCount != 1 {
		t.Errorf("expected 1 ignored rule (case-insensitive), got %d", ignoredCount)
	}
}

func TestApply_IgnoredReasonInResult(t *testing.T) {
	svc := &ax.MockWindowService{Windows: testWindows}
	ignoreApps := []string{"Code"}
	outcome, err := preset.Apply(context.Background(), svc, testPresets, "coding", ignoreApps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify first result (Code) has correct fields for JSON output
	found := false
	for _, r := range outcome.Results {
		if r.AppFilter == "Code" {
			found = true
			if !r.Skipped {
				t.Error("expected Skipped=true for ignored app")
			}
			if r.Reason != "ignored" {
				t.Errorf("expected Reason='ignored', got %q", r.Reason)
			}
		}
	}
	if !found {
		t.Error("expected result for Code app")
	}
}

func TestApply_IgnoredPlusPartialFailure(t *testing.T) {
	// Code is ignored, Terminal move fails → error + ignore warning
	svc := &ax.MockWindowService{
		Windows: testWindows,
		MoveErr: errors.New("AX error"),
	}
	ignoreApps := []string{"Code"}
	outcome, err := preset.Apply(context.Background(), svc, testPresets, "coding", ignoreApps)
	// Terminal move fails, so we expect an error
	if err == nil {
		t.Fatal("expected error for Terminal move failure, got nil")
	}

	// But the outcome should still have the ignored result
	var hasIgnored bool
	for _, r := range outcome.Results {
		if r.Reason == "ignored" {
			hasIgnored = true
		}
	}
	if !hasIgnored {
		t.Error("expected ignored result even with partial failure")
	}
}
