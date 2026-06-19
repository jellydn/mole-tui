# mole-tui

A keyboard-driven TUI that orchestrates [`mole`](https://github.com/tw93/mole) CLI (`mo`) for disk cleanup.

## Stack

- Go 1.22+, Bubble Tea v1, Lip Gloss v1, Bubbles v1. No CGO.
- Module: `github.com/tw93/mole-tui`

## Entrypoint & Layout (PRD-locked)

```
cmd/mole-tui/main.go          # binary entrypoint
internal/
  scanner/     # parse `mo clean --dry-run` → []Section & ScanSummary (ADR-001, ADR-012)
  cleanup/     # shell out to `mo clean`, stream output (ADR-002, ADR-007)
  ui/          # Bubble Tea models, views, keybindings (ADR-010)
```

## Commands

Use `just` for shorter aliases; `make` works identically.

| Action                  | `just`                      | `make`                      |
| ----------------------- | --------------------------- | --------------------------- |
| Build                   | `just build`                | `make build`                |
| Install to $GOBIN       | `just install`              | `make install`              |
| All tests               | `just test`                 | `make test`                 |
| Format check            | `just fmt`                  | `make fmt`                  |
| Vet                     | `just vet`                  | `make vet`                  |
| Build + run             | `just dev`                  | —                           |
| CI (fmt→vet→test→build) | `just ci`                   | —                           |
| Version override        | `just build VERSION=v0.1.0` | `make build VERSION=v0.1.0` |

## Lint / typecheck / test order

`fmt` → `vet` → `test` → `build`. This is the CI order (`just ci`). Pre-commit hooks (`.pre-commit-config.yaml`) run all four automatically: go-fmt, go-vet, go-build, go-unit-tests.

## Testing

- `go test ./...` — runs everything. No focused test runner shortcuts.
- Parsers (`scanner`, `cleanup`) use table-driven tests with fixtures under `internal/<pkg>/testdata/`.
- Screen rendering tested with golden-file snapshots via `teatest` or hand-rolled.
- One end-to-end test drives the TUI against a stub `mo` shim.

## Pre-commit

`.pre-commit-config.yaml` enforces: trailing-whitespace, end-of-file-fixer, check-yaml, check-json, check-merge-conflict, mixed-line-ending→LF, go-fmt, go-vet, go-build, go-unit-tests.

Run manually: `pre-commit run --all-files`

## Ralph (autonomous agent loop)

`./ralph/ralph.sh [iterations] [tool] [model]` — iterative agent that reads `prompt.md` + `prd.json`, implements one user story per loop, commits, and updates `progress.txt`. Supports multiple CLI backends (opencode, amp, claude, etc.).

## Build quirks

- Version injected via `-ldflags '-X main.version=<VERSION>'`.
- Binary: `./bin/mole-tui` (always to `bin/`, gitignored? — no `.gitignore` exists yet).
- TUI relies on `mo` CLI being on `$PATH` at runtime. No fallback.

## Constraints

- TUI never deletes files — only shells out to `mo clean` (no selective deletion args in v1).
- No state persisted across runs (v1).
- Supports optional `--dry-run` flag on the TUI itself to run simulated cleanup for testing (ADR-011).
