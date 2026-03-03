# Contributing Guidelines

## Development Environment

- Recommended: devShell via flake.nix (direnv optional)
- Go 1.25+ (cgo enabled)
- macOS 15 Sequoia+ (darwin/arm64 or darwin/amd64)
- golangci-lint v2 (`brew install golangci-lint`)
- gofumpt (`go install mvdan.cc/gofumpt@latest`)

### Key Dependencies

- Cobra (`github.com/spf13/cobra`) -- CLI framework
- go.yaml.in/yaml/v4 -- YAML config parsing
- encoding/json (stdlib) -- JSON output
- SkyLight.framework (private, via dlsym) -- virtual desktop support
- Nix (flake-parts, home-manager, nix-darwin) -- packaging modules
- Config files: `~/.config/mado/config.yaml` or `$MADO_CONFIG`

## Project Structure

```text
cmd/mado/          -- Entrypoint
internal/ax/       -- AX API adapter (interface + darwin impl + mock)
internal/config/   -- YAML config loader (XDG + $MADO_CONFIG)
internal/window/   -- list/move business logic
internal/output/   -- text/JSON formatter
internal/cli/      -- Cobra subcommand definitions
schemas/           -- JSON Schema (config.v1.schema.json) + example
testdata/golden/   -- Golden files
.github/workflows/ -- CI (lint/test/build)
```

## Running Tests

```bash
# Unit tests (run in CI, no Accessibility permission required)
go test ./...

# Update golden files
go test ./internal/output/... -update

# Integration tests (local only, Accessibility permission required)
go test -tags integration ./...
```

## Build

```bash
# Local build
go build ./cmd/mado

# Build for a specific architecture
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
  SDKROOT=$(xcrun --sdk macosx --show-sdk-path) \
  go build -o mado-arm64 ./cmd/mado

# Universal binary
lipo -create -output mado mado-amd64 mado-arm64
```

## Code Conventions

- **AX API**: access only through the `WindowService` interface in `internal/ax/interface.go`. Direct calls are prohibited.
- **cgo code**: confined to `internal/ax/darwin.go` with a `//go:build darwin` tag.
- **Cobra commands**: constructor pattern using `NewRootCmd()`. Global variables are prohibited.
- **JSON output**: must always include the `schema_version: 1` and `success` fields.
- **AX operations**: must always be wrapped with `context.WithTimeout`.
- **Formatting**: format with `gofumpt`, check with `golangci-lint`.

## Commit Conventions

Use Emoji Prefixes (ref: `.gitmessage`) so the type of change is visible at a glance in `git log --oneline`.

Each commit must leave the build in a passing state. Do not commit when `go test ./...` is failing.

## Creating a Pull Request

1. Branch name: `<issue-number>-<feature-name>` (e.g. `123-add-screen-filter`)
2. PR title: concise description of the change (under 70 characters)
3. Ensure CI (lint + test + build) passes entirely.
4. For changes that require integration tests, verify them manually before opening the PR.
