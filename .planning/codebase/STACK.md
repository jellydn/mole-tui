# STACK.md — Technology Stack

> Mapped fresh on 2026-08-04. Refreshed 2026-08-17 after the mo seam (ADR-014), tea-free cleanup stream (ADR-015), and stdio JSON-RPC sidecar (ADR-016) landed.

## Languages & Runtime

| Layer | Choice | Notes |
|-------|--------|-------|
| Language | Go 1.26.4 | `go 1.26.4` in `go.mod`; requires Go 1.26+ toolchain |
| Platform | macOS (primary), Linux (opportunistic) | Mole is macOS-first; Linux works where `mo` builds |
| CGO | None | Pure Go, statically buildable with `-trimpath` |

## Module

- Module path: `github.com/jellydn/mole-tui`
- Binary entrypoints: `cmd/mole-tui` (TUI), `cmd/mole-sidecar` (JSON-RPC sidecar, ADR-016)
- Internal packages: `mo` (ADR-014), `mo/motest` (test stub), `scanner`, `cleanup`, `ui`, `jsonrpc` (ADR-016)

## Direct Dependencies (`go.mod`)

| Package | Version | Purpose |
|---------|---------|---------|
| `charm.land/bubbletea/v2` | v2.0.8 | TUI runtime — Elm-style Cmd/Msg model |
| `charm.land/bubbles/v2` | v2.1.1 | Widgets: `spinner`, `viewport`, `help`, `key` |
| `charm.land/lipgloss/v2` | v2.0.5 | Styling — styles, adaptive colors, borders |

> **Note:** Charm libs are imported from the `charm.land` vanity paths (not `github.com/charmbracelet`). Keep it that way — see CONVENTIONS.md.

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
- **Version injection** — `main.version` set via `-ldflags '-X main.version=<VERSION>'`; defaults to `"dev"`. Both binaries use the same pattern.
- The sidecar has no `just`/`make` target; build it with `go build -o ./bin/mole-sidecar ./cmd/mole-sidecar` (documented in README).

## CI / Quality Automation

- **`.pre-commit-config.yaml`** — pre-commit framework:
  - `pre-commit-hooks` v4.6.0: trailing-whitespace, end-of-file-fixer, check-yaml, check-json, check-merge-conflict, mixed-line-ending→LF
  - `dnephin/pre-commit-golang` v0.5.1: go-fmt, go-build, go-unit-tests
  - local `go-vet` hook (`pass_filenames: false`) — runs `go vet ./...` on the whole module
- **`renovate.json`** — Renovate bot with `config:recommended` (dependency updates; it drives the charm v1→v2 migration and subsequent point bumps).
- No GitHub Actions workflow — quality gates run locally via pre-commit and `just ci`; GitHub only runs third-party security bots (GitGuardian, Socket) and CodeRabbit review.

## Configuration

- No runtime config files — the TUI and sidecar are stateless across runs by design (PRD §6).
- TUI CLI flags: `--dry-run` / `-n` (simulated cleanup, ADR-011), `--version`.
- Sidecar: no flags — it reads newline-delimited JSON-RPC on stdin and writes to stdout (ADR-016).
- `.gitignore`: `bin/`, `*.exe`, `*.test`, `*.out`, `*.ansi.txt`.

## External Runtime Dependency

- **`mo` CLI (Mole, tw93)** must be on `$PATH` — resolved once at startup via `mo.Resolve()` (which wraps `exec.LookPath`) in both `cmd/mole-tui/main.go` and `cmd/mole-sidecar/main.go`; absence is a fatal error (styled stderr + exit 1) in the TUI, and a stderr message + exit 1 in the sidecar. No fallback.

## Key Source Sizes

| File | Lines |
|------|-------|
| `internal/ui/model.go` | 740 |
| `internal/cleanup/cleanup.go` | 208 |
| `cmd/mole-sidecar/main.go` | 199 |
| `internal/scanner/scanner_test.go` | 193 |
| `internal/scanner/scanner.go` | 172 |
| `internal/mo/mo_test.go` | 158 |
| `internal/cleanup/stream_test.go` | 150 |
| `internal/jsonrpc/jsonrpc_test.go` | 135 |
| `internal/jsonrpc/jsonrpc.go` | 127 |
| `cmd/mole-sidecar/main_test.go` | 117 |
| `internal/cleanup/run_test.go` | 113 |
| `internal/ui/styles.go` | 101 |
| `internal/scanner/scan_test.go` | 97 |
| `internal/ui/operations.go` | 95 |
| `internal/ui/operations_test.go` | 83 |
| `internal/mo/motest/motest.go` | 74 |
| `internal/mo/mo.go` | 65 |
| `cmd/mole-tui/main.go` | 48 |
| `internal/cleanup/cleanup_test.go` | 47 |

> Total: ~2,900 lines of Go across 19 files (9 production, 9 test, 1 test-support stub).
