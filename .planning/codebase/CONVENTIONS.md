# CONVENTIONS.md — Coding Conventions

> Mapped fresh on 2026-08-04.

## Formatting & Static Analysis (enforced)

- **gofmt -s** must pass — `just fmt` / `make fmt` fail on any unformatted file.
- **go vet ./...** must pass — `just vet`.
- Enforced automatically by pre-commit (`dnephin/pre-commit-golang`: go-fmt, go-vet, go-build, go-unit-tests) plus the generic hooks (trailing whitespace, EOF fixer, YAML/JSON validity, merge-conflict check, LF line endings).
- CI order: **fmt → vet → test → build** (`just ci`).

## Go Style

- Standard library first; Charm TUI stack only for UI. No other third-party runtime deps.
- **Imports from `charm.land/...` vanity paths** — bubbletea, bubbles, lipgloss are all v2 and imported as `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`. Do not revert to `github.com/charmbracelet` paths.
- Doc comments on exported identifiers and package declarations (`// Package scanner invokes …`).
- Small unexported helpers for shared logic (`formatBytes`, `max`, `stripANSI`, `ParseSize`).
- Contexts are `context.Background()` at the root and `context.WithCancel` per subprocess; never `context.TODO()` in production paths.
- No panic/log.Fatal in library or UI code — the codebase currently contains **zero** `panic(`/`log.Fatal` calls. Startup failures print to stderr and `os.Exit(1)` from `main` only.

## Architecture Conventions

- **Three internal packages only** (ADR-008): `scanner`, `cleanup`, `ui`. New domain logic belongs in one of these; no new packages without an ADR.
- **UI never touches the filesystem or `os/exec` directly** — all subprocess work goes through `scanner`/`cleanup`. The TUI never deletes files itself (FR-6).
- **Bubble Tea protocol**: subprocess work is wrapped in `tea.Cmd`; results flow back as `tea.Msg`. No goroutine writes to model state directly — messages are delivered via the channel/pipe pump pattern.
- **Per-screen key handling**: every screen gets a `<screen>Keys KeyMap` + `handle<Screen>Key` + `<screen>View` trio.
- **Cancellation**: every subprocess takes a `context.Context`; UI escape hatches call the stored cancel func (`scanCancel`, `cleanupCancel`).
- **Sudo intent** is threaded from scan → confirmation → cleanup via `lastScanSudo` (ADR-005).

## Error Handling

- **Error model**: errors flow through Bubble Tea as `tea.Msg` (e.g. `scanCompleteMsg{err}`) and render in banners — never panics in normal use.
- `fmt.Errorf("…: %w", err)` wrapping with lowercase message, no trailing punctuation.
- Missing `mo` binary is the only pre-TUI fatal error: styled stderr (`\033[1;31mError:\033[0m …`) + `os.Exit(1)` in `main.go`.
- `cleanup.Run` maps `exec.ExitError` → `ExitCode`; non-`ExitError` → `-1`. Non-zero exits render the code + last 20 stderr lines on the log screen.
- Context cancellation is detected **before** generic error wrapping (`if ctx.Err() != nil { return …, ctx.Err() }` in `scanner.Scan`).

## Tolerant Parsing (ADR-012)

- Parser never rejects unrecognized output — unknown lines are kept raw in the current section.
- Item lines are rendered as raw text preserving Mole's Unicode markers (`→` actionable, `✓` clean, `◎` skipped, `☞` hint, `↳` sub-item) — `styleItemLine()` in `internal/ui/model.go` maps marker → style.
- Size aggregation is best-effort (`reSize` regex); if nothing parses, totals show `—`.

## UI Style Conventions

- All styling lives in `internal/ui/styles.go` — no raw ANSI in views except the deliberate stderr banner in `main.go`.
- **Adaptive dark/light palette** via `compat.AdaptiveColor` (Lip Gloss v2 `color.Color` interface).
- Semantic colors: `accentFg` (navigation/highlights), `warningFg` (actionable/`→` + banners), `successFg` (clean/`✓` + reclaim totals), `errorFg` (errors), `titleFg` (titles), `dimFg`/`subtleFg` (de-emphasis).
- Bordered panels use `lipgloss.RoundedBorder()`; footer gets a `BorderTop` rule.

## Git / Commit Conventions

- **Conventional Commits** (from history): `feat(ui)`, `fix(cleanup)`, `refactor(model)`, `chore(deps)`, `docs(readme)`, `docs:`, `chore:`. Scope = package where applicable.
- Working tree is kept clean; changes land as focused commits (28 commits total).
- Renovate handles dependency bumps with automated PRs.

## Testing Conventions

- `go test ./...` is the single test entrypoint (no focused runner shortcuts).
- Parser tests are **table-driven** with real/synthetic fixtures in `internal/<pkg>/testdata/`.
- See TESTING.md for the full picture.
