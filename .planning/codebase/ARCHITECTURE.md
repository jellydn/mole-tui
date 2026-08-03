# ARCHITECTURE.md — System Architecture

> Mapped fresh on 2026-08-04. PRD-locked layout (ADR-008).

## Overview

`mole-tui` is an **orchestrator TUI**: an Elm-architecture (Bubble Tea v2) front-end that shells out to the `mo` CLI and renders its output. It performs no filesystem deletion itself and persists no state.

```
┌─────────────────────────────────────────────────────────────┐
│ cmd/mole-tui/main.go — entrypoint                           │
│   flags → LookPath("mo") → tea.NewProgram(ui.NewModel)      │
└──────────────────────────┬──────────────────────────────────┘
                           │ Cmd/Msg (Bubble Tea)
┌──────────────────────────▼──────────────────────────────────┐
│ internal/ui — Model (830 lines): 5-screen state machine     │
│   Loading → Dashboard → Confirmation → Log/Report → Help    │
│   views: loadingView, dashboardView, confirmView,           │
│          logView, helpView                                  │
└───────┬──────────────────────────────┬──────────────────────┘
        │ scanner.Scan()               │ cleanup.Run()
┌───────▼───────────┐          ┌───────▼──────────────┐
│ internal/scanner  │          │ internal/cleanup     │
│ invoke mo clean   │          │ invoke mo clean      │
│ --dry-run, parse  │          │ stream stdout/stderr │
│ → []Section       │          │ → Result + FreedText │
└───────────────────┘          └──────────────────────┘
```

## Layers

1. **Entrypoint** (`cmd/mole-tui/main.go`, 48 lines)
   - Parses `--dry-run` / `-n` / `--version` flags (stdlib `flag`).
   - Resolves `mo` via `exec.LookPath` — pre-TUI fatal error with styled stderr + exit 1 if missing.
   - Builds `ui.NewModel(dryRunMode, moPath)` and runs the program.

2. **UI layer** (`internal/ui`) — the only package with side effects on screen state
   - `Model` — root Bubble Tea model holding all state (screen, cursor, scan results, log buffers, viewports, cancellable contexts).
   - `styles.go` — all Lip Gloss v2 styles, `compat.AdaptiveColor` dark/light palette.
   - Communication is purely message-driven: `tea.Cmd` (functions returning `tea.Msg`) and `tea.Msg` types (`scanCompleteMsg`, `scanCancelledMsg`, `cleanupStreamMsg`, `cleanupCompleteMsg`).

3. **Scanner** (`internal/scanner`) — scan side of the `mo` boundary
   - `Scan(ctx, moPath, sudo)` invokes `mo clean --dry-run` (optionally via `sudo`), bounded by `ctx` for cancellation.
   - `Parse(output)` — hybrid parser (ADR-012): regexes for section headers (`➤`), sizes (`NNN.NMB dry`), free space, and the sudo banner; unrecognized lines are kept raw. ANSI sequences stripped first.
   - Domain types: `Section{Name, Lines}`, `ScanSummary{Header, FreeSpace, TotalReclaimable, SystemCachesSkipped}`, `ScanResult{Sections, Summary, Raw}`.

4. **Cleanup** (`internal/cleanup`) — cleanup side of the `mo` boundary
   - `Run(ctx, opts, writer, moPath)` invokes `mo clean`, teeing stdout live to a writer while buffering it; stderr is piped and also teed (one pipe + goroutine).
   - `DryRun` option short-circuits with canned success output (ADR-011).
   - `ParseSummary(output)` extracts a "Total freed: …" line via `reFreed`.
   - Returns `Result{ExitCode, Stdout, Stderr, FreedText}`.

## Data Flow — Scan

```
Init/startScan → ctx+sudo → scanCmd → scanner.Scan
   → (ctx cancelled) scanCancelledMsg → back to dashboard / quit
   → ScanResult → scanCompleteMsg → Update:
       - error + no prev scan → loading screen shows "Scan failed: …"
       - error + prev scan  → dashboard keeps previous results
       - success → dashboard, sections collapsed, sudoBanner set
```

## Data Flow — Cleanup

```
Dashboard enter → screenConfirm → y → startCleanup:
   - DryRun mode → canned cleanupCompleteMsg (no subprocess)
   - else io.Pipe + buffered channel (cap 256) + worker goroutine
       worker: cleanup.Run (tee → pipe + buffer) → pw.Close → cleanupCompleteMsg
       pump:   bufio.Scan(pr) → cleanupStreamMsg{line} (select on ctx.Done)
   UI chaining: readCleanupStreamCmd() reads one msg per Update tick;
   returns nil when stream is nil (completion / not started).
Ctrl+C during run: first press → quitConfirm warning; second → cleanupCancel() + quit.
```

## UI State Machine (ADR-010, five screens)

```
             Init/scan      enter (y)          done / scrollable
 Loading ─────────────► Dashboard ────────► Confirm ────────► Log/Report ──┐
    ▲                      │  ?                 │  ?              │        │
    │ r / S re-scan        ▼                    ▼                ▼ enter   │
    └──────────────── Help (context-aware, prevScreen restore) ◄───────────┘
```

- `Screen` enum: `screenLoading`, `screenDashboard`, `screenConfirm`, `screenLog`, `screenHelp`.
- `prevScreen` + `helpScreen` track where Help opened from.
- Per-screen key handling via `handleLoadingKey` / `handleDashboardKey` / `handleConfirmKey` / `handleLogKey` / `handleHelpKey`, gated by `KeyMap` bindings (`key.Matches`).
- Cursor is a flat index over visible lines (section headers + expanded items); `visibleLineCount()` and `expanded map[int]bool` drive navigation; the dashboard viewport auto-scrolls to keep the cursor visible.

## Key Abstractions & Patterns

- **Message-driven async I/O**: every subprocess is wrapped in `tea.Cmd` with `context.Context` cancellation (`scanCancel`, `cleanupCancel`).
- **Sudo intent propagation** (ADR-005): `lastScanSudo` recorded at scan start, inherited by the confirmation modal and `cleanup.Run` — a sudo scan leads to a sudo cleanup.
- **Tolerant parsing** (ADR-012): raw lines are never rejected — unknown content is kept as text within the current section.
- **Viewport pattern**: `viewport.Model` used for both dashboard and log; log screen is scrollable only after completion (`logDone`).

## Architecture Decisions (referenced ADRs)

- ADR-001 merge discovery+scan, ADR-002 all-or-nothing cleanup, ADR-003 full-fidelity dry-run parsing, ADR-004 full-screen loading state, ADR-005 non-sudo by default, ADR-006 no preview pane, ADR-007 scrollable log/report, ADR-008 three-package layout, ADR-009 minimal keybindings, ADR-010 five screens, ADR-011 TUI `--dry-run` flag, ADR-012 hybrid parsing.

## Dependency Direction

`cmd → ui → {scanner, cleanup}` — one-way, no cycles. `scanner` and `cleanup` are independent leaf packages (no shared domain types between them).
