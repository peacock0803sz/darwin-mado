# mado

A CLI tool for managing macOS windows. Use `mado list` to list open windows and `mado move` to move or resize them.

## Installation

```bash
# Download from the releases page
curl -sL https://github.com/peacock0803sz/mado/releases/latest/download/mado-darwin-universal.tar.gz | tar xz
sudo mv mado /usr/local/bin/

# Or install directly with Go
go install github.com/peacock0803sz/mado/cmd/mado@latest
```

## Accessibility Permission

mado uses the macOS Accessibility API, so permission must be granted.

1. Open **System Settings** → **Privacy & Security** → **Accessibility**
2. Enable the terminal app running mado (Terminal.app / iTerm2 / Warp, etc.)
3. Re-run the command after granting permission

If run without permission, the resolution steps are displayed:

```
Error: Accessibility permission not granted

To grant permission:
  1. Open System Settings → Privacy & Security → Accessibility
  2. Enable mado (or your Terminal app) in the list
  3. Re-run the command
```

## Usage

```bash
# List open windows
mado list

# Filter by app
mado list --app Terminal

# JSON output (for scripting)
mado list --format json | jq '.windows[].app_name'

# Move a window
mado move --app Terminal --position 0,0

# Resize a window
mado move --app Safari --title "GitHub" --size 1440,900

# Move and resize at the same time
mado move --app Terminal --position 0,0 --size 800,600

# Move all windows of an app at once (--all)
mado move --app Safari --all --position 0,0

# List connected displays with their stable UUIDs
mado screen list

# Specify a screen in a multi-display setup — recommended: stable UUID
mado list --screen 37D8832A-2D66-02CA-B9F7-8F30A301B230
# Still supported: localized display name (case-insensitive) or transient numeric ID
mado list --screen "DELL U2720Q"
mado move --app Terminal --screen "Built-in Retina Display" --position 100,100

# Enable shell completion (fish example)
mado completion fish > ~/.config/fish/completions/mado.fish

# Apply a window layout preset
mado preset apply coding

# List available presets
mado preset list

# Show preset details
mado preset show coding

# Validate preset definitions
mado preset validate
```

## Configuration File

Default values can be set in `~/.config/mado/config.yaml`. CLI flags always take precedence over the config file.

```yaml
# yaml-language-server: $schema=https://github.com/peacock0803sz/mado/raw/main/schemas/config.v1.schema.json
timeout: 5s    # AX operation timeout
format: text   # output format: text | json
```

The config file path can be overridden with the `$MADO_CONFIG` environment variable.

### Presets

Define named window layout presets in the same config file and apply them with a single command.

A rule's `screen:` field accepts three forms, resolved in this order:

1. Stable display UUID from `mado screen list` (recommended — survives reboots and reconnection order changes)
2. Human-readable display name (case-insensitive exact match)
3. Legacy transient numeric display ID

`mado preset rec` writes UUIDs into the `screen:` field automatically when the runtime can obtain them.

```yaml
timeout: 5s
format: text
presets:
  - name: coding
    description: "Editor left, terminal right"
    rules:
      - app: Code
        position: [0, 0]
        size: [960, 1080]
      - app: Terminal
        position: [960, 0]
        size: [960, 1080]
  - name: meeting
    description: "Browser center, notes right"
    rules:
      - app: Safari
        title: Zoom
        position: [0, 0]
        size: [1280, 1080]
      - app: Notes
        position: [1280, 0]
        size: [640, 1080]
```

Each rule requires `app` (exact match, case-insensitive) and at least one of `position` or `size`. Optional filters: `title` (partial match) and `screen` (ID or name). Rules are evaluated in order; when multiple rules match the same window, only the first match is applied.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Accessibility permission not granted |
| 3 | Invalid arguments (e.g. bad --position/--size value) |
| 4 | Target window not found or multiple matches without --all |
| 5 | Operation on a fullscreen window |
| 6 | AX operation timed out |
| 7 | Partial success when using --all |
