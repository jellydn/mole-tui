# ARCHITECTURE.md — System Architecture

> Mapped fresh on 2026-08-04. Refreshed 2026-08-17 after the mo boundary seam (ADR-014), tea-free cleanup stream (ADR-015), and stdio JSON-RPC sidecar (ADR-016) landed.

## Overview

`mole-tui` is an **orchestrator**: an Elm-architecture (Bubble Tea v2) TUI that shells out to the `mo` CLI through an injectable subprocess seam, plus a stdio JSON-RPC sidecar that reuses the same engine for external clients. Nothing performs filesystem deletion itself and nothing persists state.

```
┌────────────────────────────────────────────────────────────────────┐
│ cmd/mole-tui/main.go — TUI entrypoint                              │
│   flags → mo.Resolve() → tea.NewProgram(ui.NewModel(dryRun, runner))│
└──────────────────────────┬─────────────────────────────────────────┘
                           │ Cmd/Msg (Bubble Tea)
┌──────────────────────────▼─────────────────────────────────────────┐
│ internal/ui — Model + operationController (5 screens)              │
│   Loading → Dashboard → Confirmation → Log/Report → Help           │
│   operationController: scan/cleanup lifecycle + event translation  │
└───────┬───────────────────────────────┬────────────────────────────┘
        │ operationController           │ operationController
┌───────▼───────────┐          ┌────────▼──────────────┐
│ internal/scanner  │          │ internal/cleanup      │
│ Scan() + Parse()  │          │ Session / Run() /     │
│ → []Section       │          │ Start() → Event stream│
└───────┬───────────┘          └────────┬──────────────┘
        │                              │
        └───────────► internal/mo (Runner seam) ◄──────────┘
                        │ mo.NewRunner(moPath) / motest stub
                        ▼
                     [mo] CLI (subprocess)

┌────────────────────────────────────────────────────────────────────┐
│ cmd/mole-sidecar/main.go — sidecar entrypoint (ADR-016)            │
│   server implements jsonrpc.Handler; owns scan/cleanup lifecycle   │
│   internal/jsonrpc: Serve loop + Notify (framing/encoding)         │
│   ── newline-delimited JSON-RPC 2.0 over stdin/stdout ──► clients  │
└────────────────────────────────────────────────────────────────────┘
```

## Layers

### 1. Entrypoints

- **`cmd/mole-tui/main.go`** (48 lines) — parses `--dry-run` / `-n` / `--version` (stdlib `flag`), resolves `mo` via `mo.Resolve()` (pre-TUI fatal error if missing), builds `ui.NewModel(dryRunMode, mo.NewRunner(moPath))` and runs the program.
- **`cmd/mole-sidecar/main.go`** (199 lines) — resolves `mo`, wires `server` to `jsonrpc.NewServer(os.Stdin, os.Stdout, s)` and runs the serve loop. The `server` type implements `jsonrpc.Handler` and owns the scan/cleanup lifecycle; all framing/encoding lives in `internal/jsonrpc`.

### 2. `internal/mo` — the mo boundary seam (ADR-014)

- **`mo.Runner`** — the seam: `Run(ctx, sudo, stdoutW, stderrW, args...) (exitCode, err)`. Streams stdout/stderr to the writers, wraps `sudo`, maps cancellations to the context error, and maps non-zero exits to the exit code (not an error).
- **`mo.Resolve()` / `mo.NewRunner()`** — binary resolution lives at the boundary where it is consumed; `main` and the UI carry a `mo.Runner`, not a bare path.
- **`internal/mo/motest`** — a pure-Go stub runner (the second adapter that makes the seam real): canned stdout/stderr, exit codes, delays, invocation counters.

### 3. `internal/scanner` — scan side of the `mo` boundary

- `Scan(ctx, runner, sudo)` invokes `mo clean --dry-run` through the runner, bounded by `ctx` for cancellation.
- `Parse(output)` — hybrid parser (ADR-012): regexes for section headers (`➤`), sizes (`NNN.NMB dry`), free space, and the sudo banner; unrecognized lines are kept raw. ANSI sequences stripped first.
- Domain types: `Section{Name, Lines}`, `ScanSummary{Header, FreeSpace, TotalReclaimable, SystemCachesSkipped}`, `ScanResult{Sections, Summary, Raw}`.

### 4. `internal/cleanup` — cleanup side of the `mo` boundary

- **`Session`** (ADR-015) — a deep lifecycle owner: holds the operation context, options, runner, synchronous execution (`Session.Run`), and asynchronous event stream (`Session.Start`). Convenience `Run(ctx, opts, writer, runner)` remains as a compatibility wrapper.
- `Run(ctx, opts, writer, runner)` invokes `mo clean`, teeing stdout/stderr live to a writer while buffering for the final `Result`. The pipe/goroutine plumbing for the stream lives in `Session.Start`, which emits plain Go `Event`s (`EventLine` / `EventDone`) on a bounded channel so UI adapters translate rather than own subprocess plumbing.
- `DryRun` option short-circuits with canned success output (ADR-011).
- `ParseSummary(output)` extracts a "Total freed: …" line via `reFreed`.

### 5. `internal/ui` — Bubble Tea layer

- **`Model`** (740 lines) — root Bubble Tea model holding screen state, scan results, log buffers, viewports, and the `operationController`. Five screens per ADR-010.
- **`operationController`** (95 lines, `operations.go`) — owns the UI-facing lifecycle of Scan and Cleanup sessions: start/cancel, and translation of plain package events into Bubble Tea messages. `Model` keeps screen state and rendering.
- `styles.go` — all Lip Gloss v2 styles, `compat.AdaptiveColor` dark/light palette.
- Communication is purely message-driven: `tea.Cmd` and `tea.Msg` (`scanCompleteMsg`, `scanCancelledMsg`, `cleanupStreamMsg`, `cleanupCompleteMsg`).

### 6. `internal/jsonrpc` — JSON-RPC transport (ADR-016)

- Owns newline-delimited JSON-RPC 2.0 framing, wire encoding, the `Serve` read/dispatch/write loop, and a goroutine-safe `Notify` for server-initiated events.
- Narrow `Handler` interface (`Handle(method, params) (any, *Error)`) — the sidecar server implements it; protocol behaviour is tested independently of the domain.

## Data Flow — Scan

```
Init/startScan → ctx+sudo → operationController.startScan → scanner.Scan(ctx, runner, sudo)
   → (ctx cancelled) scanCancelledMsg → back to dashboard / quit
   → ScanResult → scanCompleteMsg → Update:
       - error + no prev scan → loading screen shows "Scan failed: …"
       - error + prev scan  → dashboard keeps previous results
       - success → dashboard, sections collapsed, sudoBanner set
```

## Data Flow — Cleanup

```
Dashboard enter → screenConfirm → y → Model.startCleanup:
   - operationController.startCleanup(sudo) → cleanup.NewSession(ctx, opts, runner).Start()
       → bounded Event channel (cap 256), worker goroutine owns pipe + scanner
   UI chaining: readCleanupStreamCmd() reads one Event per Update tick;
   EventLine → cleanupStreamMsg{line}; EventDone → cleanupCompleteMsg{result, err};
   returns nil when stream is nil (completion / not started).
Ctrl+C during run: first press → quitConfirm warning; second → cancelCleanup() + quit.
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

- **Injectable subprocess seam (ADR-014)**: every `mo` invocation goes through `mo.Runner`; `motest` is the second adapter. `scanner` and `cleanup` never touch `os/exec` directly.
- **Tea-free event stream (ADR-015)**: `cleanup` owns the pipe/scanner/goroutines and emits plain `Event`s; UI and sidecar adapters translate events into their own framework instead of owning subprocess plumbing.
- **Message-driven async I/O**: every subprocess is wrapped in `tea.Cmd` with `context.Context` cancellation; `operationController` owns the cancel funcs.
- **Sudo intent propagation** (ADR-005): `lastScanSudo` recorded at scan start, inherited by the confirmation modal and `cleanup.Run` — a sudo scan leads to a sudo cleanup.
- **Tolerant parsing** (ADR-012): raw lines are never rejected — unknown content is kept as text within the current section.
- **Viewport pattern**: `viewport.Model` used for both dashboard and log; log screen is scrollable only after completion (`logDone`).

## Architecture Decisions (referenced ADRs)

- ADR-001 merge discovery+scan, ADR-002 all-or-nothing cleanup, ADR-003 full-fidelity dry-run parsing, ADR-004 full-screen loading state, ADR-005 non-sudo by default, ADR-006 no preview pane, ADR-007 scrollable log/report, ADR-008 three-package layout (extended by 014/016), ADR-009 minimal keybindings, ADR-010 five screens, ADR-011 TUI `--dry-run` flag, ADR-012 hybrid parsing, ADR-013 stay on Bubble Tea, ADR-014 mo boundary seam, ADR-015 tea-free cleanup stream, ADR-016 stdio JSON-RPC sidecar.

## Dependency Direction

```
cmd/mole-tui    → ui → {scanner, cleanup} → mo
cmd/mole-sidecar → {scanner, cleanup} → mo ; jsonrpc (standalone leaf)
```
One-way, no cycles. `mo`, `scanner`, `cleanup`, `jsonrpc` are leaf packages; `ui` is the only Bubble Tea-coupled package.
