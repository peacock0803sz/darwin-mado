package preset_test

import (
	"testing"

	"github.com/peacock0803sz/mado/internal/preset"
)

func TestValidatePresets_Valid(t *testing.T) {
	presets := []preset.Preset{{
		Name: "coding",
		Rules: []preset.Rule{
			{App: "Code", Position: []int{0, 0}, Size: []int{960, 1080}},
		},
	}}
	if errs := preset.ValidatePresets(presets); errs != nil {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidatePresets_EmptyName(t *testing.T) {
	presets := []preset.Preset{{
		Name: "",
		Rules: []preset.Rule{
			{App: "Code", Position: []int{0, 0}},
		},
	}}
	errs := preset.ValidatePresets(presets)
	if errs == nil {
		t.Fatal("expected validation errors for empty name, got nil")
	}
	found := false
	for _, e := range errs {
		if e.Field == "name" {
			found = true
		}
	}
	if !found {
		t.Error("expected name validation error")
	}
}

func TestValidatePresets_DuplicateNames(t *testing.T) {
	presets := []preset.Preset{
		{Name: "coding", Rules: []preset.Rule{{App: "Code", Position: []int{0, 0}}}},
		{Name: "coding", Rules: []preset.Rule{{App: "Terminal", Position: []int{960, 0}}}},
	}
	errs := preset.ValidatePresets(presets)
	if errs == nil {
		t.Fatal("expected validation errors for duplicate names, got nil")
	}
	found := false
	for _, e := range errs {
		if e.Message == "duplicate preset name" {
			found = true
		}
	}
	if !found {
		t.Error("expected duplicate name error")
	}
}

func TestValidatePresets_MissingApp(t *testing.T) {
	presets := []preset.Preset{{
		Name: "broken",
		Rules: []preset.Rule{
			{Position: []int{0, 0}},
		},
	}}
	errs := preset.ValidatePresets(presets)
	if errs == nil {
		t.Fatal("expected validation errors for missing app, got nil")
	}
	found := false
	for _, e := range errs {
		if e.Message == "exactly one of app or app_id is required" {
			found = true
		}
	}
	if !found {
		t.Error("expected app/app_id required error")
	}
}

func TestValidatePresets_AppIDValid(t *testing.T) {
	presets := []preset.Preset{{
		Name: "bundle-id",
		Rules: []preset.Rule{
			{AppID: "com.apple.Safari", Position: []int{0, 0}},
		},
	}}
	if errs := preset.ValidatePresets(presets); errs != nil {
		t.Errorf("expected no errors for app_id-only rule, got %v", errs)
	}
}

func TestValidatePresets_BothAppAndAppID(t *testing.T) {
	presets := []preset.Preset{{
		Name: "broken",
		Rules: []preset.Rule{
			{App: "Safari", AppID: "com.apple.Safari", Position: []int{0, 0}},
		},
	}}
	errs := preset.ValidatePresets(presets)
	if errs == nil {
		t.Fatal("expected validation error for both app and app_id set, got nil")
	}
	found := false
	for _, e := range errs {
		if e.Message == "app and app_id are mutually exclusive; set exactly one" {
			found = true
		}
	}
	if !found {
		t.Error("expected mutual exclusivity error")
	}
}

func TestValidatePresets_MissingPositionAndSize(t *testing.T) {
	presets := []preset.Preset{{
		Name: "broken",
		Rules: []preset.Rule{
			{App: "Code"},
		},
	}}
	errs := preset.ValidatePresets(presets)
	if errs == nil {
		t.Fatal("expected validation errors for missing position and size, got nil")
	}
	found := false
	for _, e := range errs {
		if e.Message == "position or size is required" {
			found = true
		}
	}
	if !found {
		t.Error("expected position-or-size required error")
	}
}

func TestValidatePresets_NegativeSize(t *testing.T) {
	presets := []preset.Preset{{
		Name: "broken",
		Rules: []preset.Rule{
			{App: "Code", Size: []int{-1, 1080}},
		},
	}}
	errs := preset.ValidatePresets(presets)
	if errs == nil {
		t.Fatal("expected validation errors for negative size, got nil")
	}
	found := false
	for _, e := range errs {
		if e.Message == "width must be positive" {
			found = true
		}
	}
	if !found {
		t.Error("expected width positive error")
	}
}

func TestValidatePresets_InvalidNameChars(t *testing.T) {
	presets := []preset.Preset{{
		Name: "-invalid",
		Rules: []preset.Rule{
			{App: "Code", Position: []int{0, 0}},
		},
	}}
	errs := preset.ValidatePresets(presets)
	if errs == nil {
		t.Fatal("expected validation errors for invalid name chars, got nil")
	}
}

func TestValidatePresets_PositionOnly(t *testing.T) {
	// position のみ指定は有効
	presets := []preset.Preset{{
		Name: "pos-only",
		Rules: []preset.Rule{
			{App: "Code", Position: []int{100, 200}},
		},
	}}
	if errs := preset.ValidatePresets(presets); errs != nil {
		t.Errorf("expected no errors for position-only rule, got %v", errs)
	}
}

func TestValidatePresets_SizeOnly(t *testing.T) {
	// size のみ指定は有効
	presets := []preset.Preset{{
		Name: "size-only",
		Rules: []preset.Rule{
			{App: "Code", Size: []int{960, 1080}},
		},
	}}
	if errs := preset.ValidatePresets(presets); errs != nil {
		t.Errorf("expected no errors for size-only rule, got %v", errs)
	}
}
