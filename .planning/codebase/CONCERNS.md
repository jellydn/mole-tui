# CONCERNS.md — Technical Debt & Concerns

> Mapped fresh on 2026-08-04. Verified by full-file review (no TODO/FIXME/HACK markers, no `panic(`/`log.Fatal` calls anywhere).

## High Priority

### 1. Docs overpromise testing that doesn't exist
`AGENTS.md` and PRD §7 claim golden-file `teatest` snapshot tests and an end-to-end test against a stub `mo` shim. **Neither exists** — the repo has only 2 test files (`scanner_test.go`, `cleanup_test.go`). `internal/ui` has zero tests, `scanner.Scan()`'s subprocess path is untested, and `cleanup.Run()`'s real subprocess path is untested (only the dry-run branch is).

- **Fix:** either write the tests (teatest snapshots + e2e stub `mo` shim — a shim would also unblock testing `scanner.Scan`/`cleanup.Run` subprocess logic) or update the docs to match reality.

### 2. README stack table is stale
README's Stack table says "Bubble Tea v1 / Bubbles v1 / Lip Gloss v1" with `github.com/charmbracelet` paths. The code is on **v2 via `charm.land`** paths (bubbletea v2.0.7, bubbles v2.1.0, lipgloss v2.0.4) — migrated in commits `8957dd6`/`75595bc`/`2d27bc2`/`0fc466e`. Anyone following the README will read wrong versions.

- **Fix:** update the README Stack table (and its "Go 1.22+"/v1 references elsewhere).

## Medium Priority

### 3. `internal/ui/model.go` is an 830-line monolith
Holds the root model, all 5 screens' state, key handling, views, help entries, viewport logic, and helpers. While internally well-sectioned with banner comments, it is 5× the size of any other file and growing — a single change touches model + handlers + views.

- **Fix candidates:** split per-screen files (`dashboard.go`, `log.go`, `help.go`), extract `startCleanup`/`startScan` cmd builders, or move `styleItemLine`/`formatBytes` helpers out. (Note: `startCleanup()` was already extracted in the latest commit — pattern is established.)

### 4. Fragile regex-based parsing is load-bearing
`TotalReclaimable` depends on `reSize` = `([0-9.]+)\s*(KB|MB|GB|B)\s+dry` — **case-sensitive** on the literal `dry` and matching `0B dry` lines too (e.g. "0B dry" inflates nothing but is parsed). Free-space and freed-text regexes similarly assume current `mo` phrasing. ADR-012 mitigates by keeping raw lines, but the aggregate totals silently degrade (show `—`) if `mo` changes its output.

- **Fix candidates:** track a real `mo clean --dry-run` fixture over time; revisit PRD OQ-1 (`--json` flag) if Mole ships one; extend `reSize` to case-insensitive `(?i)`.

### 5. Sudo intent silently propagates to cleanup
If the user's last scan used `S` (sudo), the confirmation modal shows `[SUDO] — elevated from last scan` and cleanup runs `sudo mo clean` (ADR-005 design, but worth flagging): the modal's `n` cancels the *cleanup*, not the sudo intent — a user who only wanted a sudo *preview* may accidentally run an elevated *cleanup* if they confirm later. There's no way to run cleanup non-elevated after a sudo scan.

- **Fix candidates:** explicit sudo choice in the confirmation modal, or reset `lastScanSudo` on modal open unless re-confirmed.

## Low Priority / By Design

### 6. No UI-level graceful degradation when `mo` is missing
Missing `mo` is a **pre-TUI fatal error** (styled stderr + exit 1) — documented and intentional, but there's no fallback or in-TUI error screen. Fine for v1; noted for v2.

### 7. `mo clean` output contract is unverified
PRD OQ-3 ("What does `mo clean` (actual, non-dry-run) output look like?") is still **open** — the `ParseSummary` regex for "Total freed:" is written from assumptions, not a captured real fixture. If `mo clean` prints something else, the log screen just shows no summary (graceful, but unverified).

### 8. No UI tests → no regression safety net for the biggest file
Consequence of #1: the 5-screen state machine (cursor math, collapse/expand `expanded` map, viewport auto-scroll offsets, double-ctrl+c kill flow, help `prevScreen` restore) is the highest-churn logic in the repo and is completely untested.

### 9. Stateless by design (v1)
No persisted config or history — intentional (PRD §6), but means every launch re-scans and there's no audit trail of past cleanups.

### 10. Minor code nits
- `cleanup.Run` stderr handling uses `io.Pipe` + goroutine; if `cmd.Start()` fails the early return leaves no goroutine (safe today), but the pipe/goroutine dance is more complex than a `bytes.Buffer` for stderr would be — only needed for live stderr streaming, which the UI currently doesn't render (only last-20-lines on failure). Dead complexity.
- `error-output.txt` fixture is tested via `TestParse`, but the *real* missing-`mo` path is handled pre-TUI in `main.go` — the fixture exercises `Parse` on error text, which is fine but documents a path the UI can't actually reach (scanner never parses "mo not installed" stderr; that's a `Scan` error, not `Parse` input).
- `formatBytes` uses binary units (1024) — consistent, but a choice worth stating if `mo` ever reports decimal units.

## Security Notes

- ✅ No network, no telemetry, no user data leaves the machine.
- ✅ No file deletion by the TUI itself — all destructive ops go through `mo` (FR-6).
- ⚠️ `sudo` elevation runs the *resolved* `mo` path (`sudo <moPath> clean`) — `moPath` comes from `LookPath`, so PATH-hijack risk is limited at startup, but the sudo call itself is `sudo mo …` executed from a TUI context; standard sudo password prompting applies.
- ⚠️ `--dry-run` flag is a simulation short-circuit in `cleanup.Run` — safe by construction, but a stale `DryRun` flag in a future non-interactive path would silently skip cleanup (worth a comment/assertion if that path ever appears).

## Performance Notes

- Startup to loading screen is trivially fast (<200ms PRD target) — `Init` only starts the spinner + scan cmd.
- Scan takes as long as `mo` needs (~1–3 min typical); the loading screen with elapsed timer (ADR-004) covers this.
- Cleanup streaming is bounded: channel cap 256, `readCleanupStreamCmd` drains one message per frame — no unbounded buffering.
- `viewport` content is rebuilt each frame in `dashboardView` (full `strings.Builder` re-render + `SetContent`) — fine at dashboard sizes, but the log screen also re-sets full content per stream line; acceptable for v1.

## Ops / Tooling Notes

- `renovate.json` uses `config:recommended` — it has already driven the charm v1→v2 migrations; future major upgrades (bubbletea v2.x) should be reviewed against the `charm.land` import convention.
- Pre-commit runs the full `go-unit-tests` suite on every commit — cheap today (2 test files), will slow down as tests grow.
- `ralph/` contains per-backend prompt files for the autonomous loop — the `prompt.md`/`prd.json` pair is the source of truth for that workflow; keep in sync if the PRD changes.
