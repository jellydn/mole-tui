# STACK.md — Technology Stack

> Mapped fresh on 2026-08-04.

## Languages & Runtime

| Layer | Choice | Notes |
|-------|--------|-------|
| Language | Go 1.26.4 | `go 1.26.4` in `go.mod`; requires Go 1.26+ toolchain |
| Platform | macOS (primary), Linux (opportunistic) | Mole is macOS-first; Linux works where `mo` builds |
| CGO | None | Pure Go, statically buildable with `-trimpath` |

## Module

- Module path: `github.com/jellydn/mole-tui`
- Binary entrypoint: `cmd/mole-tui`
- Internal packages: `scanner`, `cleanup`, `ui` (ADR-008)

## Direct Dependencies (`go.mod`)

| Package | Version | Purpose |
|---------|---------|---------|
| `charm.land/bubbletea/v2` | v2.0.7 | TUI runtime — Elm-style Cmd/Msg model |
| `charm.land/bubbles/v2` | v2.1.0 | Widgets: `spinner`, `viewport`, `help`, `key` |
| `charm.land/lipgloss/v2` | v2.0.4 | Styling — styles, adaptive colors, borders |

> **Note:** Charm libs are imported from the `charm.land` vanity paths (not `github.com/charmbracelet`). The README's Stack table still says "v1 / github.com/charmbracelet" — **stale**, see CONCERNS.md.

## Indirect Dependencies

`charmbracelet/colorprofile`, `charmbracelet/ultraviolet`, `charmbracelet/x/{ansi,term,termios,windows}`, `clipperhouse/displaywidth`, `clipperhouse/uax29/v2`, `lucasb-eyer/go-colorful`, `mattn/go-runewidth`, `muesli/cancelreader`, `rivo/uniseg`, `xo/terminfo`, `golang.org/x/{sync,sys}`.

## Build & Task Tooling

- **`justfile` / `just`** — primary task runner (also mirrored in `Makefile`):
  - `just build` → `go build -ldflags '-X main.version={{VERSION}}' -trimpath -o bin/mole-tui ./cmd/mole-tui`
  - `just install` → `go install` to `$GOBIN`
  - `just test` → `go test ./...`
  - `just vet` → `go vet ./...`
  - `just fmt` → `gofmt -l -s .` (fails on unformatted files)
  - `just dev` → build + run `./bin/mole-tui`
  - `just ci` → fmt → vet → test → build (CI order)
- **`Makefile`** — equivalent targets (`build`, `install`, `test`, `vet`, `fmt`, `help`); `VERSION` override via `make build VERSION=v0.1.0`.
- **Version injection** — `main.version` set via `-ldflags '-X main.version=<VERSION>'`; defaults to `"dev"`.

## CI / Quality Automation

- **`.pre-commit-config.yaml`** — pre-commit framework:
  - `pre-commit-hooks` v4.6.0: trailing-whitespace, end-of-file-fixer, check-yaml, check-json, check-merge-conflict, mixed-line-ending→LF
  - `dnephin/pre-commit-golang` v0.5.1: go-fmt, go-vet, go-build, go-unit-tests
- **`renovate.json`** — Renovate bot with `config:recommended` (dependency updates, see commit history for v1→v2 migration commits).

## Configuration

- No runtime config files — the TUI is stateless across runs by design (PRD §6).
- Only CLI flags: `--dry-run` / `-n` (simulated cleanup, ADR-011), `--version`.
- `.gitignore`: `bin/`, `*.exe`, `*.test`, `*.out`, `*.ansi.txt`.

## External Runtime Dependency

- **`mo` CLI (Mole, tw93)** must be on `$PATH` — resolved once at startup via `exec.LookPath("mo")` in `cmd/mole-tui/main.go`; absence is a pre-TUI fatal error (styled stderr + exit 1). No fallback.

## Key Source Sizes

| File | Lines |
|------|-------|
| `internal/ui/model.go` | 830 |
| `internal/scanner/scanner_test.go` | 193 |
| `internal/scanner/scanner.go` | 175 |
| `internal/ui/styles.go` | 101 |
| `internal/cleanup/cleanup.go` | 121 |
| `cmd/mole-tui/main.go` | 48 |
| `internal/cleanup/cleanup_test.go` | 45 |
| `README.md` | 141 |
