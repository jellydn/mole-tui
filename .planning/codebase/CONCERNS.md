# CONCERNS.md — Technical Debt & Concerns

> Mapped fresh on 2026-08-04. Refreshed 2026-08-17 (verified against current code — no TODO/FIXME/HACK markers, no `panic(`/`log.Fatal` calls anywhere).

## Resolved Since Last Map

- ✅ **Docs overpromised testing that didn't exist.** The repo now has 9 test files: subprocess paths via the `motest` stub (scanner, cleanup) and a fake-`mo` shell script (mo execRunner), cleanup streaming/backpressure, UI `operationController`, the jsonrpc transport, and the sidecar end-to-end. Remaining gaps are narrower — see "Testing" below.
- ✅ **README stack table was stale.** README now lists v2 via `charm.land` paths (bubbletea v2.0.8, bubbles v2.1.1, lipgloss v2.0.5) and documents the sidecar (ADR-016).
- ✅ **`cleanup.Run`'s stderr `io.Pipe` goroutine was "dead complexity".** The pipe/goroutine plumbing was absorbed into `internal/mo`'s runner (ADR-014) and the remaining pipe now lives in `cleanup.Session.Start` where it is load-bearing for the tea-free event stream (ADR-015).
- ✅ **The mo subprocess was untestable.** The `mo.Runner` seam (ADR-014) + `motest` stub made scan/cleanup subprocess paths testable; the production `execRunner` is tested against a fake `mo` script.

## High Priority

### 1. `internal/ui/model.go` is still a 740-line monolith
Down from 830 lines after `operationController` was extracted to `internal/ui/operations.go` (95 lines), but `model.go` still holds the root model, all 5 screens' state, key handling, views, help entries, viewport logic, and helpers. A single change to a screen touches model + handlers + views.

- **Fix candidates:** split per-screen files (`dashboard.go`, `log.go`, `help.go`), or move `styleItemLine`/`formatBytes`/`helpEntriesFor` helpers out.

### 2. Fragile regex-based parsing is load-bearing
`TotalReclaimable` depends on `reSize` = `([0-9.]+)\s*(KB|MB|GB|B)\s+dry` — **case-sensitive** on the literal `dry` and matching `0B dry` lines too. Free-space and freed-text regexes similarly assume current `mo` phrasing. ADR-012 mitigates by keeping raw lines, but the aggregate totals silently degrade (show `—`) if `mo` changes its output.

- **Fix candidates:** track real `mo clean --dry-run` fixtures over time; revisit PRD OQ-1 (`--json` flag) if Mole ships one; extend `reSize` to case-insensitive `(?i)`.

### 3. Sudo intent silently propagates to cleanup
If the user's last scan used `S` (sudo), the confirmation modal shows `[SUDO] — elevated from last scan` and cleanup runs `sudo mo clean` (ADR-005 design, but worth flagging): the modal's `n` cancels the *cleanup*, not the sudo intent — a user who only wanted a sudo *preview* may accidentally run an elevated *cleanup* if they confirm later. There's no way to run cleanup non-elevated after a sudo scan.

- **Fix candidates:** explicit sudo choice in the confirmation modal, or reset `lastScanSudo` on modal open unless re-confirmed.

## Medium Priority

### 4. No UI-level graceful degradation when `mo` is missing
Missing `mo` is a **pre-TUI fatal error** (styled stderr + exit 1) in both binaries — documented and intentional, but there's no fallback or in-TUI error screen. Fine for v1; noted for v2.

### 5. `mo clean` output contract is unverified
PRD OQ-3 ("What does `mo clean` (actual, non-dry-run) output look like?") is still **open** — the `ParseSummary` regex for "Total freed:" is written from assumptions, not a captured real fixture. If `mo clean` prints something else, the log screen just shows no summary (graceful, but unverified).

### 6. The 5-screen state machine still has no rendering tests
`operationController` is now covered, but the screen-level logic (cursor math, collapse/expand `expanded` map, viewport auto-scroll offsets, double-ctrl+c kill flow, help `prevScreen` restore) is only exercised by hand. AGENTS.md still promises golden-file/`teatest` snapshots and a full TUI e2e that don't exist — the `motest` seam makes a `tea.NewProgram(ui.NewModel(..., stub))` e2e feasible.

- **Fix candidates:** add `teatest`-style golden snapshots, or at least a program-level e2e driving the stub runner.

### 7. Duplicated config on `Model` vs `operationController`
`Model` keeps public `DryRun`/`Mo` fields while `operationController` stores its own copies (`dryRun`, `runner`). They're set together in `NewModel` and can't diverge today, but it's ownership debt — either make the controller the sole owner or document the invariant.

## Low Priority / By Design

### 8. Stateless by design (v1)
No persisted config or history — intentional (PRD §6), but means every launch re-scans and there's no audit trail of past cleanups.

### 9. Minor code nits
- `error-output.txt` fixture is tested via `TestParse`, but the *real* missing-`mo` path is handled pre-TUI in `main.go` — the fixture exercises `Parse` on error text, which documents a path the UI can't actually reach (scanner never parses "mo not installed" stderr; that's a `Scan` error, not `Parse` input).
- `formatBytes` uses binary units (1024) — consistent, but a choice worth stating if `mo` ever reports decimal units.
- The sidecar has no `just`/`make` target — it's built via a raw `go build` (documented in README) and is skipped by `just build`.
- `internal/jsonrpc` emits `"id":null` on parse errors (spec-correct) — if a client ever treats `id:null` as a notification and skips reading it, parse errors could be missed; the sidecar's own tests cover it.

## Security Notes

- ✅ No network, no telemetry, no user data leaves the machine (the sidecar is local stdio only, ADR-016).
- ✅ No file deletion by the TUI or sidecar — all destructive ops go through `mo` (FR-6).
- ⚠️ `sudo` elevation runs the *resolved* `mo` path (`sudo <moPath> clean`) — `moPath` comes from `mo.Resolve()`/`LookPath`, so PATH-hijack risk is limited at startup, but the sudo call itself is `sudo mo …` executed from a TUI context; standard sudo password prompting applies.
- ⚠️ `--dry-run` flag is a simulation short-circuit in `cleanup.Run` — safe by construction, but a stale `DryRun` flag in a future non-interactive path would silently skip cleanup (worth a comment/assertion if that path ever appears).

## Performance Notes

- Startup to loading screen is trivially fast (<200ms PRD target) — `Init` only starts the spinner + scan cmd.
- Scan takes as long as `mo` needs (~1–3 min typical); the loading screen with elapsed timer (ADR-004) covers this.
- Cleanup streaming is bounded: `cleanup.Session.Start` uses a channel cap of 256; `operationController.readCleanupStreamCmd` drains one event per frame — no unbounded buffering, and abandoned/cancelled streams close without leaking the producer (tested).
- `viewport` content is rebuilt each frame in `dashboardView` (full `strings.Builder` re-render + `SetContent`) — fine at dashboard sizes, but the log screen also re-sets full content per stream line; acceptable for v1.

## Ops / Tooling Notes

- `renovate.json` uses `config:recommended` — it drives charm point bumps (v2.0.8/v2.1.1/v2.0.5); future major upgrades (bubbletea v2.x) should be reviewed against the `charm.land` import convention.
- Pre-commit runs the full `go-unit-tests` suite on every commit — now 9 test files, still cheap, will slow down as tests grow.
- The local `go-vet` hook uses `language: system` with `pass_filenames: false` — it requires `go` on the invoking shell's `$PATH` (a machine with a broken/missing Go toolchain on PATH will fail commits at the hook).
- `ralph/` contains per-backend prompt files for the autonomous loop — the `prompt.md`/`prd.json` pair is the source of truth for that workflow; keep in sync if the PRD changes.
