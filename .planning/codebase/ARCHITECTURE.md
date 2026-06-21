# Architecture

**Analysis Date:** 2026-06-20

## Pattern Overview

**Overall:** Elm Architecture / Bubble Tea MVU orchestrator over the external Mole CLI.

**Key Characteristics:**
- The TUI is a Bubble Tea model-view-update application: `internal/ui/model.go` defines the root `Model`, `Init`, `Update`, and `View` methods.
- Side effects are isolated behind `tea.Cmd` functions that invoke package APIs: scan commands call `scanner.Scan` in `internal/ui/model.go`, and cleanup commands call `cleanup.Run` in `internal/ui/model.go`.
- Domain/process packages do not import UI code: `internal/scanner/scanner.go` parses `mo clean --dry-run`, `internal/cleanup/cleanup.go` executes `mo clean`, and `internal/ui/model.go` composes them.
- The application is an orchestrator, not a cleanup engine: destructive work is delegated to `mo clean` in `internal/cleanup/cleanup.go`, matching the PRD in `tasks/prd-mole-tui.md` and ADR-002 in `.planning/adr/002-all-or-nothing-cleanup.md`.

## Layers

**Command Entrypoint:**
- Purpose: Parse CLI flags, validate the `mo` prerequisite, create and run the Bubble Tea program.
- Location: `cmd/mole-tui/main.go`
- Contains: `main`, build-time `version`, flag handling for `--dry-run`, `-n`, and `--version`, preflight `exec.LookPath("mo")`.
- Depends on: Go stdlib `flag`, `os`, `os/exec`, Bubble Tea, and `internal/ui` via `cmd/mole-tui/main.go`.
- Used by: The compiled `mole-tui` binary built from `cmd/mole-tui/main.go`.

**Scanner:**
- Purpose: Invoke `mo clean --dry-run` and parse Mole dry-run text into a sectioned preview plus summary metadata.
- Location: `internal/scanner/`
- Contains: `Section`, `ScanSummary`, `ScanResult`, ANSI stripping, regex-based section/header/size/sudo parsing, and subprocess execution in `internal/scanner/scanner.go`.
- Depends on: Go stdlib `context`, `os/exec`, `regexp`, string parsing, and fixtures/tests in `internal/scanner/scanner_test.go`.
- Used by: `internal/ui/model.go` through `scanner.Scan` and `scanner.ScanResult`.

**Cleanup:**
- Purpose: Execute or simulate `mo clean`, stream output to a writer, capture stdout/stderr, and parse a best-effort freed-space summary.
- Location: `internal/cleanup/`
- Contains: `Options`, `Result`, `Run`, `ParseSummary`, stderr pipe teeing, dry-run short-circuiting, and summary regex in `internal/cleanup/cleanup.go`.
- Depends on: Go stdlib `context`, `io`, `os/exec`, `regexp`, `bytes`, and unit tests in `internal/cleanup/cleanup_test.go`.
- Used by: `internal/ui/model.go` through `cleanup.Run`, `cleanup.Options`, `cleanup.Result`, and `cleanup.ParseSummary`.

**UI:**
- Purpose: Own all Bubble Tea state, screen transitions, keybindings, views, and styling.
- Location: `internal/ui/`
- Contains: Root `Model`, `Screen` enum, key maps, private `tea.Msg` types, command constructors, view renderers, and screen-specific key handlers in `internal/ui/model.go`; Lip Gloss styles in `internal/ui/styles.go`.
- Depends on: Bubble Tea, Bubbles spinner/help/viewport/key packages, Lip Gloss, `internal/scanner`, and `internal/cleanup`.
- Used by: `cmd/mole-tui/main.go` via `ui.NewModel`.

**Planning / Decision Records:**
- Purpose: Record the product and architecture decisions that explain the current simplified shape.
- Location: `.planning/`, `tasks/`
- Contains: ADRs such as `.planning/adr/001-merge-discovery-and-scan.md`, `.planning/adr/008-three-package-architecture.md`, `.planning/adr/010-five-screens.md`, glossary in `.planning/glossary.md`, and PRD in `tasks/prd-mole-tui.md`.
- Depends on: Markdown only.
- Used by: Humans and agents maintaining the codebase.

## Data Flow

**Startup Scan → Dashboard:**
1. `cmd/mole-tui/main.go` parses flags, checks `mo` with `exec.LookPath`, then starts `tea.NewProgram(ui.NewModel(dryRunMode))`.
2. `Model.Init` in `internal/ui/model.go` creates a cancellable context and returns `tea.Batch(m.spinner.Tick, m.scanCmd(ctx, false))`.
3. `scanCmd` in `internal/ui/model.go` calls `scanner.Scan(ctx, false)`, which runs `mo clean --dry-run` in `internal/scanner/scanner.go`.
4. `scanner.Parse` in `internal/scanner/scanner.go` strips ANSI, finds `➤` section headers, keeps raw item lines, extracts free space and total reclaimable size, and detects the sudo-skipped banner.
5. `scanCompleteMsg` is handled by `Model.Update` in `internal/ui/model.go`, storing `scanner.ScanResult`, setting `sudoBanner`, marking `hasPrevScan`, and moving from `screenLoading` to `screenDashboard`.
6. `dashboardView` in `internal/ui/model.go` renders the sectioned preview, total reclaimable footer, optional sudo/error banners, and key-hint footer.

**Dashboard Re-scan / Sudo Re-scan:**
1. `handleDashboardKey` in `internal/ui/model.go` handles `r` by launching `m.scanCmd(ctx, false)` and `S` by launching `m.scanCmd(ctx, true)`.
2. `scanner.Scan` in `internal/scanner/scanner.go` runs `mo clean --dry-run` for normal scans or `sudo sh -c "mo clean --dry-run"` for sudo scans.
3. Successful `scanCompleteMsg` replaces the current `scanResult`; failed re-scans keep the previous dashboard when `hasPrevScan` is true in `internal/ui/model.go`.

**Dashboard → Confirm → Clean → Log/Report:**
1. `handleDashboardKey` in `internal/ui/model.go` handles `enter` by moving to `screenConfirm` and setting `confirmDryRun` from `Model.DryRun`.
2. `confirmView` in `internal/ui/model.go` shows the literal `mo clean` command, dry-run notice when relevant, and the total from `scanner.ScanSummary`.
3. `handleConfirmKey` in `internal/ui/model.go` handles `y` by moving to `screenLog` and returning `cleanupCmd(m.DryRun)`.
4. `cleanupCmd` in `internal/ui/model.go` calls `cleanup.Run(context.Background(), cleanup.Options{DryRun: dryRun}, &buf)`.
5. `cleanup.Run` in `internal/cleanup/cleanup.go` either writes a canned dry-run success message or starts `mo clean`, teeing stdout/stderr into the writer and buffers.
6. `cleanupCompleteMsg` is handled by `Model.Update` in `internal/ui/model.go`, filling `logContent`, `logStderr`, `logExit`, and `logSummary`, then setting the viewport content for `logView`.

**State Management:**
- UI state is centralized in the root `Model` struct in `internal/ui/model.go`; there are no sub-model packages.
- Screen state is a private `Screen` enum with `screenLoading`, `screenDashboard`, `screenConfirm`, `screenLog`, and `screenHelp` in `internal/ui/model.go`, reflecting ADR-010 in `.planning/adr/010-five-screens.md`.
- Process completion and cancellation are represented as private message types: `scanCompleteMsg`, `scanCancelledMsg`, and `cleanupCompleteMsg` in `internal/ui/model.go`.
- Scan cancellation state is stored as `scanCancel context.CancelFunc` in `internal/ui/model.go`; cleanup currently uses `context.Background()` in `cleanupCmd`, so the UI can warn/quit but does not pass a cancellation context to `cleanup.Run`.
- Parsed scan data is stored as `scanner.ScanResult` in `Model.scanResult`; runtime log/report data is stored as strings and viewport state in `Model.logContent`, `Model.logStderr`, `Model.logVP`, and related fields in `internal/ui/model.go`.

## Key Abstractions

**`scanner.Section`:**
- Purpose: A named group from Mole dry-run output, with raw item lines preserved under it.
- Examples: `internal/scanner/scanner.go`, `internal/scanner/scanner_test.go`, `internal/scanner/testdata/mo-clean-dryrun-real.txt`.
- Pattern: Hybrid parsing from ADR-012 in `.planning/adr/012-hybrid-parsing.md`: structured section headers plus raw item lines.

**`scanner.ScanSummary`:**
- Purpose: Metadata extracted from dry-run output: header, free space, best-effort total reclaimable bytes, and system-cache sudo skip detection.
- Examples: `internal/scanner/scanner.go`, `internal/ui/model.go`.
- Pattern: Lightweight aggregate DTO produced by parser and consumed by UI.

**`scanner.ScanResult`:**
- Purpose: The complete parsed scan result, combining `[]Section`, `ScanSummary`, and raw ANSI-stripped output.
- Examples: `internal/scanner/scanner.go`, `internal/ui/model.go`.
- Pattern: Boundary object between the scanner/process layer and Bubble Tea UI layer.

**`cleanup.Options` and `cleanup.Result`:**
- Purpose: Control cleanup dry-run behavior and return process exit/output/summary data.
- Examples: `internal/cleanup/cleanup.go`, `internal/cleanup/cleanup_test.go`, `internal/ui/model.go`.
- Pattern: Command execution boundary around the external `mo clean` process.

**Bubble Tea `Model`:**
- Purpose: Root UI state and behavior for all screens.
- Examples: `internal/ui/model.go`.
- Pattern: Bubble Tea MVU: `Init` emits commands, `Update` transforms state in response to messages, and `View` renders from state.

**Bubble Tea `tea.Msg` / `tea.Cmd`:**
- Purpose: Asynchronous process integration and UI event flow.
- Examples: `scanCompleteMsg`, `scanCancelledMsg`, `cleanupCompleteMsg`, `scanCmd`, and `cleanupCmd` in `internal/ui/model.go`.
- Pattern: Side-effect functions return messages; `Update` handles messages to mutate state and trigger transitions.

**Screen-specific key maps:**
- Purpose: Keep keyboard behavior explicit for each screen.
- Examples: `dashboardKeys`, `confirmKeys`, `logKeys`, and `loadingKeys` in `internal/ui/model.go`.
- Pattern: Bubbles `key.Binding` maps dispatched through screen-specific handlers such as `handleDashboardKey` in `internal/ui/model.go`.

## Entry Points

**Binary entrypoint:**
- Location: `cmd/mole-tui/main.go`
- Triggers: Running the compiled `mole-tui` binary or `go run ./cmd/mole-tui`.
- Responsibilities: Parse `--dry-run` / `-n` / `--version`, validate `mo` is on `$PATH`, instantiate `ui.NewModel`, run Bubble Tea, and print fatal errors to stderr.

**Initial Bubble Tea command:**
- Location: `internal/ui/model.go`
- Triggers: Bubble Tea calls `Model.Init` after program creation in `cmd/mole-tui/main.go`.
- Responsibilities: Start spinner ticks and launch the initial non-sudo scan via `m.scanCmd(ctx, false)`.

**Scanner API:**
- Location: `internal/scanner/scanner.go`
- Triggers: `scanCmd` in `internal/ui/model.go`.
- Responsibilities: Run `mo clean --dry-run` or sudo dry-run and return `ScanResult` or an error.

**Cleanup API:**
- Location: `internal/cleanup/cleanup.go`
- Triggers: `cleanupCmd` in `internal/ui/model.go` after confirmation.
- Responsibilities: Simulate cleanup in TUI dry-run mode or execute `mo clean`, stream output, capture process results, and parse a summary.

## Error Handling

**Strategy:** Fatal preconditions fail before Bubble Tea starts; runtime scan and cleanup errors are converted into return errors or Bubble Tea messages and rendered in the current screen.

**Patterns:**
- Missing `mo` binary is checked in `cmd/mole-tui/main.go` with `exec.LookPath`; failure prints styled installation guidance to stderr and exits with code 1.
- `scanner.Scan` in `internal/scanner/scanner.go` wraps non-zero `mo clean --dry-run` failures as `mo clean --dry-run failed: ...`, preferring stderr over the generic process error.
- Context cancellation from a scan is detected with `ctx.Err()` in `internal/scanner/scanner.go` and converted to `scanCancelledMsg` by `scanCmd` in `internal/ui/model.go`.
- Initial scan failure leaves the loading screen with `loadingMsg = "Scan failed: ..."`; re-scan failure keeps prior dashboard data and displays `errorMsg` as a banner in `internal/ui/model.go`.
- `cleanup.Run` in `internal/cleanup/cleanup.go` returns a `Result` with `ExitCode` for process failures and only returns a Go error for start/write-level failures.
- `cleanupCompleteMsg` handling in `internal/ui/model.go` turns cleanup errors into a `logSummary` error and non-zero log exit state; `logView` renders an error banner and last stderr lines.

## Cross-Cutting Concerns

**Logging:** Runtime logging is user-facing command output. `cleanup.Run` streams stdout/stderr into a writer and buffers in `internal/cleanup/cleanup.go`; `logView` displays that output through a viewport in `internal/ui/model.go`. There is no separate structured logger.

**Validation:** Input validation is minimal and boundary-focused: `cmd/mole-tui/main.go` validates `mo` is available; `scanner.Parse` tolerates malformed or unknown dry-run lines by preserving raw item text in `internal/scanner/scanner.go`; `cleanup.ParseSummary` returns an empty summary when no freed amount matches in `internal/cleanup/cleanup.go`.

**Authentication:** There is no application authentication. Privilege elevation is delegated to system `sudo` for a re-scan path in `scanner.Scan` via `sudo sh -c "mo clean --dry-run"` in `internal/scanner/scanner.go`; ADR-005 documents the non-sudo default in `.planning/adr/005-non-sudo-by-default.md`.

**Styling / Accessibility:** Visual styling is centralized in `internal/ui/styles.go` with adaptive light/dark Lip Gloss colors; screen rendering and key-hint footers live in `internal/ui/model.go`.

**Testing:** Parser and process-boundary behavior is covered by table tests in `internal/scanner/scanner_test.go` and `internal/cleanup/cleanup_test.go`; real and edge-case scanner fixtures live in `internal/scanner/testdata/`.

---

*Architecture analysis: 2026-06-20*
