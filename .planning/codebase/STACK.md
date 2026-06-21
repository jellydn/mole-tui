# Technology Stack

**Analysis Date:** 2026-06-20

## Languages

**Primary:**
- Go 1.26.4 module target - application code under `cmd/mole-tui/main.go` and `internal/`, declared by `go.mod`.

**Secondary:**
- POSIX shell - build recipes use `/bin/sh` in `justfile` and `Makefile`.
- YAML/JSON - pre-commit and dependency-update configuration in `.pre-commit-config.yaml` and `renovate.json`.

## Runtime

**Environment:**
- Go toolchain targeting host OS/architecture; `go.mod` declares `go 1.26.4`, while the README prerequisite still says Go 1.22+ in `README.md`.
- Terminal runtime for an interactive Bubble Tea TUI; startup creates a `tea.NewProgram` in `cmd/mole-tui/main.go`.

**Package Manager:**
- Go modules - module path `github.com/jellydn/mole-tui` in `go.mod`.
- Lockfile: present via `go.sum`.

## Frameworks

**Core:**
- Bubble Tea `charm.land/bubbletea/v2` v2.0.7 - TUI runtime/program model used by `cmd/mole-tui/main.go` and `internal/ui/model.go`, declared in `go.mod`.
- Bubbles `charm.land/bubbles/v2` v2.1.0 - TUI widgets for help, key bindings, spinner, and viewport in `internal/ui/model.go`, declared in `go.mod`.
- Lip Gloss `charm.land/lipgloss/v2` v2.0.4 - terminal styling and layout in `internal/ui/styles.go` and `internal/ui/model.go`, declared in `go.mod`.

**Testing:**
- Go standard `testing` package - unit tests in `internal/scanner/scanner_test.go` and `internal/cleanup/cleanup_test.go`.
- Test fixtures under `internal/scanner/testdata/` are read by parser tests in `internal/scanner/scanner_test.go`.
- No third-party test framework is declared in `go.mod`; `go.sum` includes transitive/unused checksum entries but no direct testing dependency.

**Build/Dev:**
- `just` recipes - build, install, test, vet, fmt, dev, and ci targets in `justfile`.
- GNU Make-compatible `make` targets - build, install, test, vet, and fmt targets in `Makefile`.
- `go build` with `-trimpath` and ldflags version injection (`-X main.version=<VERSION>`) in `justfile`, `Makefile`, and `cmd/mole-tui/main.go`.
- `gofmt`, `go vet`, `go test ./...`, and `go build` form the local/CI-style quality flow in `justfile`, `Makefile`, and `AGENTS.md`.
- pre-commit hooks v4.6.0/v0.5.1 enforce whitespace, YAML/JSON checks, LF line endings, go-fmt, go-vet, go-build, and go-unit-tests in `.pre-commit-config.yaml`.
- Renovate uses `config:recommended` in `renovate.json`.

## Key Dependencies

**Critical:**
- `charm.land/bubbletea/v2` v2.0.7 - event loop and model/update/view architecture for the TUI, declared in `go.mod` and used in `cmd/mole-tui/main.go` and `internal/ui/model.go`.
- `charm.land/bubbles/v2` v2.1.0 - built-in TUI components (`help`, `key`, `spinner`, `viewport`) used by `internal/ui/model.go`, declared in `go.mod`.
- `charm.land/lipgloss/v2` v2.0.4 - styling, borders, adaptive colours, and layout used by `internal/ui/styles.go`, declared in `go.mod`.
- External `mo` CLI from Mole - required runtime backend; checked with `exec.LookPath("mo")` in `cmd/mole-tui/main.go`, invoked by `internal/scanner/scanner.go` and `internal/cleanup/cleanup.go`, and documented in `README.md`.

**Infrastructure:**
- `github.com/charmbracelet/colorprofile` v0.4.3, `github.com/charmbracelet/x/ansi` v0.11.7, `github.com/charmbracelet/x/term` v0.2.2, `github.com/charmbracelet/x/termios` v0.1.1, and `github.com/charmbracelet/x/windows` v0.2.2 - terminal/color platform support pulled indirectly in `go.mod`.
- `github.com/mattn/go-runewidth` v0.0.24, `github.com/rivo/uniseg` v0.4.7, `github.com/clipperhouse/displaywidth` v0.11.0, and `github.com/clipperhouse/uax29/v2` v2.7.0 - Unicode/display width support pulled indirectly in `go.mod`.
- `golang.org/x/sys` v0.46.0 and `golang.org/x/sync` v0.21.0 - low-level system/concurrency support pulled indirectly in `go.mod`.

## Configuration

**Environment:**
- Runtime configuration is via CLI flags only: `--dry-run`, `-n`, and `--version` in `cmd/mole-tui/main.go`.
- No application environment variables are read in source under `cmd/` or `internal/`; `$PATH` must contain `mo`, checked in `cmd/mole-tui/main.go`.
- Build/install respects Go's normal `$GOBIN`/`$GOPATH/bin`; `Makefile` resolves `GOBIN` via `go env`.

**Build:**
- Build config files: `go.mod`, `go.sum`, `justfile`, `Makefile`, `.pre-commit-config.yaml`, and `renovate.json`.
- Binary output is `bin/mole-tui` for local builds in `justfile` and `Makefile`.
- Version defaults to `dev` and can be overridden with `VERSION=...` in `justfile` and `Makefile`; `cmd/mole-tui/main.go` exposes it through `--version`.

## Platform Requirements

**Development:**
- Go toolchain compatible with the module target in `go.mod`; README documents Go 1.22+ in `README.md`.
- `mo`/Mole CLI installed on `$PATH`; README suggests Homebrew install in `README.md`, and the binary exits early if `mo` is missing in `cmd/mole-tui/main.go`.
- Optional developer tools: `just`, `make`, `pre-commit`, and Renovate configuration as shown in `justfile`, `Makefile`, `.pre-commit-config.yaml`, and `renovate.json`.

**Production:**
- Distributed as a local terminal binary (`mole-tui`) installed with `go install` or `make install`, documented in `README.md`.
- Targets macOS disk cleanup workflows; README describes the app as orchestrating Mole for macOS disk cleanup and recommends `brew install mole` in `README.md`.
- Requires terminal access, local filesystem permissions, and optional sudo flow for system-cache scanning via `sudo sh -c "mo clean --dry-run"` in `internal/scanner/scanner.go`.

---

*Stack analysis: 2026-06-20*
