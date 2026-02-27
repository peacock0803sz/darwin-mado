// Package preset implements YAML preset management for window layouts.
package preset

// Preset is a named window layout definition loaded from the config file.
type Preset struct {
	Name        string `json:"name"        yaml:"name"`
	Description string `json:"description" yaml:"description,omitempty"`
	Rules       []Rule `json:"rules"       yaml:"rules"`
}

// Rule is a single window operation instruction within a preset.
type Rule struct {
	App    string `json:"app"                yaml:"app"`
	Title  string `json:"title,omitempty"     yaml:"title,omitempty"`
	Screen string `json:"screen,omitempty"    yaml:"screen,omitempty"`
	// Desktop scopes this rule to a specific desktop (1-based Mission Control order).
	// nil = no filter (matches all desktops); *Desktop=0 = match only all-desktops windows.
	Desktop  *int  `json:"desktop,omitempty"   yaml:"desktop,omitempty"`
	Position []int `json:"position,omitempty"  yaml:"position,omitempty,flow"`
	Size     []int `json:"size,omitempty"      yaml:"size,omitempty,flow"`
}
