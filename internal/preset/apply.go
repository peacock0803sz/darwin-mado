package preset

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/screen"
	"github.com/peacock0803sz/mado/internal/window"
)

// NotFoundError is returned when the specified preset name does not exist.
type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("preset %q not found", e.Name)
}

// AllFullscreenError is returned when all matched windows are fullscreen.
type AllFullscreenError struct {
	Skipped int
}

func (e *AllFullscreenError) Error() string {
	return fmt.Sprintf("all %d matched windows are fullscreen", e.Skipped)
}

// ApplyResult holds the outcome of applying a single rule.
type ApplyResult struct {
	RuleIndex     int
	SelectorKind  string // "app" or "app_id"
	SelectorValue string // the app name or bundle identifier
	Affected      []ax.Window
	Skipped       bool
	Reason        string
	Err           error
}

// selectorOf returns the selector kind and value for a rule.
// It returns ("app_id", rule.AppID) when AppID is set, else ("app", rule.App).
func selectorOf(r Rule) (kind, value string) {
	if r.AppID != "" {
		return "app_id", r.AppID
	}
	return "app", r.App
}

// ApplyOutcome holds the aggregate result of applying a preset.
type ApplyOutcome struct {
	PresetName string
	Results    []ApplyResult
}

// Apply applies the named preset to matching windows.
// ignoreApps contains app names to skip (case-insensitive). Rules targeting ignored apps
// are skipped with reason "ignored".
func Apply(ctx context.Context, svc ax.WindowService, presets []Preset, name string, ignoreApps []string) (*ApplyOutcome, error) {
	var target *Preset
	for i := range presets {
		if presets[i].Name == name {
			target = &presets[i]
			break
		}
	}
	if target == nil {
		return nil, &NotFoundError{Name: name}
	}

	windows, err := svc.ListWindows(ctx)
	if err != nil {
		return nil, err
	}

	// Screens are fetched lazily: only when at least one rule has a non-empty
	// `screen:` field. Cache so we don't pay the AX roundtrip per rule.
	var screensCached []ax.Screen
	var screensErr error
	var screensLoaded bool
	loadScreens := func() ([]ax.Screen, error) {
		if screensLoaded {
			return screensCached, screensErr
		}
		screensCached, screensErr = svc.ListScreens(ctx)
		screensLoaded = true
		return screensCached, screensErr
	}

	// 適用済みウィンドウの追跡 (PID+Title で一意に識別)
	type winKey struct {
		PID   uint32
		Title string
	}
	applied := make(map[winKey]bool)

	outcome := &ApplyOutcome{PresetName: name}
	totalMatched := 0
	totalFullscreen := 0

	for i, rule := range target.Rules {
		kind, value := selectorOf(rule)

		// Resolve screen filter before anything else so unresolvable or ambiguous
		// selectors short-circuit to a Skip result without hitting AX.
		if rule.Screen != "" {
			screens, err := loadScreens()
			if err != nil {
				return outcome, err
			}
			if _, resolveErr := screen.Resolve(rule.Screen, screens); resolveErr != nil {
				reason := ""
				var notFound *screen.ScreenNotFoundError
				var ambiguous *screen.AmbiguousScreenError
				switch {
				case errors.As(resolveErr, &notFound):
					reason = "screen_not_found"
				case errors.As(resolveErr, &ambiguous):
					reason = "screen_ambiguous"
				default:
					return outcome, resolveErr
				}
				outcome.Results = append(outcome.Results, ApplyResult{
					RuleIndex:     i,
					SelectorKind:  kind,
					SelectorValue: value,
					Skipped:       true,
					Reason:        reason,
				})
				continue
			}
		}

		// ルールに基づいてウィンドウをフィルタリング
		matches := filterForRule(windows, rule)

		// 適用済みウィンドウを除外 (first match wins)
		var candidates []ax.Window
		for _, w := range matches {
			key := winKey{PID: w.PID, Title: w.Title}
			if !applied[key] {
				candidates = append(candidates, w)
			}
		}

		if len(candidates) == 0 {
			outcome.Results = append(outcome.Results, ApplyResult{
				RuleIndex:     i,
				SelectorKind:  kind,
				SelectorValue: value,
				Skipped:       true,
				Reason:        "no_match",
			})
			continue
		}

		// Filter out windows whose app is in the ignore list.
		// Check against actual window fields (AppName + AppID) so that
		// ignore entries match regardless of whether the rule uses app or app_id.
		var nonIgnored []ax.Window
		for _, w := range candidates {
			if !window.IsIgnoredApp(w.AppName, w.AppID, ignoreApps) {
				nonIgnored = append(nonIgnored, w)
			}
		}
		if len(nonIgnored) == 0 {
			outcome.Results = append(outcome.Results, ApplyResult{
				RuleIndex:     i,
				SelectorKind:  kind,
				SelectorValue: value,
				Skipped:       true,
				Reason:        "ignored",
			})
			continue
		}
		candidates = nonIgnored

		// フルスクリーンウィンドウの除外
		var normal []ax.Window
		fullscreenCount := 0
		for _, w := range candidates {
			if w.State == ax.StateFullscreen {
				fullscreenCount++
				totalFullscreen++
				continue
			}
			normal = append(normal, w)
		}
		totalMatched += len(candidates)

		if len(normal) == 0 && fullscreenCount > 0 {
			outcome.Results = append(outcome.Results, ApplyResult{
				RuleIndex:     i,
				SelectorKind:  kind,
				SelectorValue: value,
				Skipped:       true,
				Reason:        "fullscreen",
			})
			// フルスクリーンでもappliedセットに追加
			for _, w := range candidates {
				applied[winKey{PID: w.PID, Title: w.Title}] = true
			}
			continue
		}

		// マッチした全ウィンドウに対してMoveWindow/ResizeWindowを実行
		var ruleAffected []ax.Window
		var ruleErr error

		for _, w := range normal {
			if len(rule.Position) == 2 {
				if err := svc.MoveWindow(ctx, w.PID, w.Title, rule.Position[0], rule.Position[1]); err != nil {
					ruleErr = err
					break
				}
				w.X = rule.Position[0]
				w.Y = rule.Position[1]
			}
			if len(rule.Size) == 2 {
				if err := svc.ResizeWindow(ctx, w.PID, w.Title, rule.Size[0], rule.Size[1]); err != nil {
					ruleErr = err
					break
				}
				w.Width = rule.Size[0]
				w.Height = rule.Size[1]
			}
			ruleAffected = append(ruleAffected, w)
			applied[winKey{PID: w.PID, Title: w.Title}] = true
		}

		outcome.Results = append(outcome.Results, ApplyResult{
			RuleIndex:     i,
			SelectorKind:  kind,
			SelectorValue: value,
			Affected:      ruleAffected,
			Err:           ruleErr,
		})
	}

	// 全マッチがフルスクリーンの場合
	if totalMatched > 0 && totalMatched == totalFullscreen {
		return outcome, &AllFullscreenError{Skipped: totalFullscreen}
	}

	// 部分成功の確認
	var successCount, failCount int
	for _, r := range outcome.Results {
		if r.Err != nil {
			failCount++
		} else if !r.Skipped {
			successCount++
		}
	}
	if failCount > 0 && successCount > 0 {
		var allAffected []ax.Window
		for _, r := range outcome.Results {
			allAffected = append(allAffected, r.Affected...)
		}
		return outcome, &ax.PartialSuccessError{
			Affected: allAffected,
			Cause:    fmt.Errorf("partial success: %d rules applied, %d failed", successCount, failCount),
		}
	}
	if failCount > 0 && successCount == 0 {
		// 全失敗の場合は最初のエラーを返す
		for _, r := range outcome.Results {
			if r.Err != nil {
				return outcome, r.Err
			}
		}
	}

	return outcome, nil
}

// filterForRule はルールの条件に基づいてウィンドウを絞り込む
func filterForRule(windows []ax.Window, rule Rule) []ax.Window {
	var result []ax.Window
	lowerRuleTitle := strings.ToLower(rule.Title)

	for _, w := range windows {
		// app / app_id: case-insensitive exact match (mutually exclusive)
		if rule.AppID != "" {
			if !strings.EqualFold(w.AppID, rule.AppID) {
				continue
			}
		} else {
			if !strings.EqualFold(w.AppName, rule.App) {
				continue
			}
		}
		// title: case-insensitive partial match
		if rule.Title != "" && !strings.Contains(strings.ToLower(w.Title), lowerRuleTitle) {
			continue
		}
		// screen: reuse MatchScreen
		if rule.Screen != "" && !window.MatchScreen(w, rule.Screen) {
			continue
		}
		// desktop: nil = no filter; *Desktop=0 = only all-desktops windows; *Desktop=N = N or all-desktops
		if rule.Desktop != nil {
			switch *rule.Desktop {
			case 0:
				if w.Desktop != 0 {
					continue
				}
			default:
				if w.Desktop != 0 && w.Desktop != *rule.Desktop {
					continue
				}
			}
		}
		result = append(result, w)
	}
	return result
}
