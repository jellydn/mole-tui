# Codebase Concerns

**Analysis Date:** 2026-06-20

## Tech Debt

**Cleanup streaming is only implemented inside the subprocess wrapper, not in Bubble Tea:**
- Issue: `cleanup.Run` writes subprocess output to an `io.Writer` as it arrives, but `internal/ui` passes a local `bytes.Buffer` and emits only one `cleanupCompleteMsg` after `mo clean` exits. The log screen therefore shows `Running mo clean…` until completion instead of the live stream promised by the PRD.
- Files: `internal/cleanup/cleanup.go`, `internal/ui/model.go`, `tasks/prd-mole-tui.md`
- Impact: Long cleanups give no progress feedback, making the app look hung and failing FR-3 / ADR-007 expectations.
- Fix approach: Introduce incremental Bubble Tea messages for output chunks, keep a cleanup context/cancel function on the model, and update `logVP` as stdout/stderr arrives.

**Repeated scan setup logic in the dashboard key handler:**
- Issue: Rescan and sudo-rescan setup is duplicated in both the empty-results branch and the normal dashboard branch.
- Files: `internal/ui/model.go`
- Impact: Small behavior changes, timeouts, or cancellation handling must be updated in multiple places; one path can drift from the other.
- Fix approach: Extract `startScan(sudo bool, message string) tea.Cmd` on `Model` to set screen, loading text, timestamp, context/cancel, and return `scanCmd`.

**Parse model mixes UI-ready raw strings with summary extraction:**
- Issue: `scanner.Parse` stores raw item lines while simultaneously extracting section headers, free-space text, sudo state, and aggregate sizes with package-level regexes.
- Files: `internal/scanner/scanner.go`, `.planning/adr/012-hybrid-parsing.md`
- Impact: This is pragmatic for v1 but makes later behavior such as item filtering, richer display, or JSON parser migration harder because parsed data and presentational raw text are coupled.
- Fix approach: Keep raw lines for fidelity, but add typed metadata per line (`State`, `SizeBytes`, `IsHint`) behind the current UI API before adding more dashboard features.

## Known Bugs

**Sudo scan does not run the command described in the UI and ADR:**
- Symptoms: The code runs `sudo sh -c "mo clean --dry-run"`; the comment says `sudo -v` refreshes credentials, but `sudo -v` is not actually executed. This runs the whole dry-run as root instead of validating credentials then running `mo` with the user's environment.
- Files: `internal/scanner/scanner.go`, `.planning/adr/005-non-sudo-by-default.md`
- Trigger: Press `S` in the dashboard.
- Workaround: Avoid the sudo scan path until it is corrected, or run `sudo -v` manually before launching if Mole itself relies on cached credentials.

**Cleanup cancellation warning is misleading and does not control the subprocess:**
- Symptoms: During cleanup, the UI says “Press ctrl+c again to kill the cleanup process”, but `cleanupCmd` calls `cleanup.Run(context.Background(), ...)` and the model stores no cleanup cancel function. The second `ctrl+c` quits Bubble Tea rather than cancelling the `mo clean` process through its context.
- Files: `internal/ui/model.go`, `internal/cleanup/cleanup.go`, `tasks/prd-mole-tui.md`
- Trigger: Confirm cleanup, press `ctrl+c`, then press `ctrl+c` again while `mo clean` is still running.
- Workaround: Let cleanup finish, or terminate the process externally from the shell/process manager if it hangs.

**Failed cleanup exits are reported as generic UI errors:**
- Symptoms: `cleanup.Run` returns a non-zero `ExitCode` in `Result` but returns `nil` error for normal process exit failures. The UI can show the exit code after completion, but error-specific handling depends on stderr parsing and there is no typed error for non-zero exits.
- Files: `internal/cleanup/cleanup.go`, `internal/ui/model.go`
- Trigger: `mo clean` exits non-zero after starting successfully.
- Workaround: Inspect the captured log/stderr in the log screen.

## Security Considerations

**PATH trust boundary for destructive command execution:**
- Risk: The program checks `exec.LookPath("mo")` and later invokes `mo` by name. A malicious or unexpected `mo` earlier on `$PATH` would be executed for both scan and cleanup.
- Files: `cmd/mole-tui/main.go`, `internal/scanner/scanner.go`, `internal/cleanup/cleanup.go`, `AGENTS.md`
- Current mitigation: There is a startup check that `mo` exists on `$PATH`; cleanup uses `exec.CommandContext("mo", "clean")` with fixed arguments, so there is no user-input argument injection in the normal cleanup path.
- Recommendations: Resolve and store the absolute `mo` path once at startup, display it in confirmation/help, and pass that path into scanner/cleanup. Optionally warn if the path is outside expected install locations.

**Shell usage in sudo scan expands the attack surface:**
- Risk: `exec.CommandContext(ctx, "sudo", "sh", "-c", "mo clean --dry-run")` invokes a shell and relies on root/sudo environment path resolution. The command string is static, so user-input injection is not present, but shell invocation is unnecessary and increases surprises around aliases, functions, and PATH.
- Files: `internal/scanner/scanner.go`
- Current mitigation: The command string is hard-coded; no user input is interpolated.
- Recommendations: Avoid `sh -c`; use `sudo -v` as a separate command, then invoke the resolved `mo` path with `clean --dry-run`, or explicitly run `sudo` with an argument vector if root execution is intended.

**Confirmation does not disclose the resolved executable path:**
- Risk: The modal shows only `mo clean`, so users cannot verify which binary will run before approving a destructive operation delegated to Mole.
- Files: `internal/ui/model.go`, `cmd/mole-tui/main.go`
- Current mitigation: The command has no item arguments and the TUI never deletes files directly.
- Recommendations: Show the absolute path found by `exec.LookPath`, and consider a short “using Mole at …” line on startup or the confirmation modal.

## Performance Bottlenecks

**Initial scan blocks useful dashboard rendering for minutes:**
- Problem: The app cannot render scan results until `mo clean --dry-run` completes; ADR-004 notes this can take about 3 minutes.
- Files: `internal/ui/model.go`, `internal/scanner/scanner.go`, `.planning/adr/004-full-screen-loading-state.md`
- Cause: Discovery and scan are intentionally merged because Mole exposes items only through dry-run output.
- Improvement path: If Mole adds a machine-readable or faster discovery mode, split discovery from size scan. In the meantime, keep cancellation reliable and consider showing captured stderr/stdout progress if Mole emits any.

**Dashboard content is rebuilt on every view render:**
- Problem: `dashboardView` rebuilds the full string, calls `SetContent`, and recalculates scroll positioning every render.
- Files: `internal/ui/model.go`
- Cause: Rendering is simple and state-derived; there is no dirty flag despite a `dashboardContent` cache field.
- Improvement path: Fine for current small output, but if Mole output grows substantially, use a dirty flag keyed on scan result/cursor/expanded state or remove the unused cache field to avoid implying caching exists.

**Cleanup output is fully buffered in memory:**
- Problem: Cleanup stdout and stderr are captured into buffers, then copied into model strings and viewport content.
- Files: `internal/cleanup/cleanup.go`, `internal/ui/model.go`
- Cause: The current result model returns the complete log at the end.
- Improvement path: Keep a bounded ring buffer for very large logs, stream chunks to the viewport, and write full logs to a temporary file only if a “save log” feature is added.

## Fragile Areas

**Mole dry-run parsing depends on human-formatted text:**
- Files: `internal/scanner/scanner.go`, `internal/scanner/testdata/mo-clean-dryrun-real.txt`, `.planning/adr/012-hybrid-parsing.md`
- Why fragile: Section headers must start with `➤`, sizes must look like `NUMBER UNIT dry`, free space must match `Free space: NUMBERUNIT`, and sudo detection must contain `System caches need sudo`. Mole output changes can silently degrade the dashboard or totals.
- Safe modification: Add fixture-driven tests before changing regexes; preserve raw unrecognized lines in sections as required by FR-7.
- Test coverage: Parser tests cover a truncated real fixture, ANSI, empty/error output, sizes, and sudo banner, but not lowercase units, TB sizes, commas, locale formatting, multiple real Mole versions, or full non-truncated dry-run output.

**ANSI stripping handles only a narrow CSI pattern:**
- Files: `internal/scanner/scanner.go`, `internal/scanner/scanner_test.go`
- Why fragile: `reANSI` covers simple `ESC[` sequences ending in a letter. Other terminal control sequences or OSC hyperlinks would remain in parsed lines and could affect display/search.
- Safe modification: Use a well-tested ANSI stripping helper or broaden tests with real captured output from supported terminals.
- Test coverage: Only basic SGR color sequences are tested.

**Sudo behavior crosses UI, process, and Mole semantics:**
- Files: `internal/ui/model.go`, `internal/scanner/scanner.go`, `.planning/adr/005-non-sudo-by-default.md`
- Why fragile: ADR-005 says actual cleanup should inherit the sudo state after an elevated scan, but the current model records only `sudoBanner` and never records that the last scan was elevated or passes sudo intent into `cleanup.Run`.
- Safe modification: Add explicit state such as `lastScanSudo bool`, display it in confirmation, and decide whether cleanup should validate/elevate similarly.
- Test coverage: No tests exercise `S`, sudo scan command construction, or the transition from sudo scan to cleanup.

**Root TUI model is untested:**
- Files: `internal/ui/model.go`
- Why fragile: Screen transitions, keybindings, empty scans, failed rescans with previous results, help overlay behavior, and log state are all encoded in one large model file.
- Safe modification: Add focused model-update tests that drive `tea.KeyPressMsg` and synthetic scan/cleanup messages before refactoring.
- Test coverage: There are no UI model tests or golden/snapshot tests in `internal/ui` despite PRD expectations.

## Scaling Limits

**Single-host, single-process orchestration only:**
- Current capacity: One local `mo clean --dry-run` scan and one local `mo clean` run at a time.
- Files: `internal/ui/model.go`, `internal/scanner/scanner.go`, `internal/cleanup/cleanup.go`, `tasks/prd-mole-tui.md`
- Limit: No cross-host orchestration, scheduling, persistence, or background monitoring by design.
- Scaling path: Keep this as a local wrapper for v1; if scope expands, introduce a command abstraction and a persistent job/history model rather than extending the current Bubble Tea model directly.

**No persisted state or history:**
- Current capacity: The app keeps scan results and logs only in memory for the current run.
- Files: `internal/ui/model.go`, `tasks/prd-mole-tui.md`, `AGENTS.md`
- Limit: Users cannot compare scans, audit past cleanups, or recover logs after quitting.
- Scaling path: If auditability becomes important, add opt-in log/history persistence with clear privacy controls because cleanup output can include local paths.

## Dependencies at Risk

**External `mo` CLI output and behavior:**
- Risk: The project depends on `mo clean --dry-run` and `mo clean` behavior but has no version negotiation, JSON mode, or compatibility matrix.
- Files: `internal/scanner/scanner.go`, `internal/cleanup/cleanup.go`, `.planning/adr/001-merge-discovery-and-scan.md`, `.planning/adr/012-hybrid-parsing.md`
- Impact: Changes in Mole output can break sections, totals, sudo hints, or summary parsing without compile-time failures.
- Migration plan: Track supported Mole versions, capture fixtures per version, prefer a future `--json`/machine-readable output if Mole exposes it, and keep raw-line fallback for display.

**Bubble Tea / Bubbles v2 API churn vs project docs:**
- Risk: The project imports `charm.land/bubbletea/v2`, `bubbles/v2`, and `lipgloss/v2`, while `AGENTS.md` and PRD text mention Bubble Tea v1 / Lip Gloss v1 / Bubbles v1.
- Files: `go.mod`, `internal/ui/model.go`, `internal/ui/styles.go`, `AGENTS.md`, `tasks/prd-mole-tui.md`
- Impact: Documentation drift can lead contributors to use outdated examples or APIs.
- Migration plan: Update project guidance to v2 or intentionally downgrade; pin examples and tests to the actual module versions in `go.mod`.

## Missing Critical Features

**End-to-end safety test with a stub `mo`:**
- Problem: The code relies on strict “TUI never deletes files itself” behavior, but there is no E2E test proving that dry-run mode does not execute `mo clean` and normal mode invokes exactly `mo clean` without extra arguments.
- Files: `internal/ui/model.go`, `internal/cleanup/cleanup.go`, `tasks/prd-mole-tui.md`
- Blocks: Confident refactors of process invocation and UI confirmation flow.

**Executable path visibility and command provenance:**
- Problem: Users are asked to confirm `mo clean` without seeing the resolved binary path.
- Files: `cmd/mole-tui/main.go`, `internal/ui/model.go`
- Blocks: Stronger safety posture for a destructive wrapper that intentionally delegates deletion to another executable.

**Machine-readable parser path:**
- Problem: The parser is intentionally hybrid because Mole currently emits human text, but there is no abstraction boundary for swapping to JSON if Mole adds it.
- Files: `internal/scanner/scanner.go`, `.planning/adr/012-hybrid-parsing.md`, `tasks/prd-mole-tui.md`
- Blocks: Low-risk migration away from regex parsing when a stable upstream format appears.

## Test Coverage Gaps

**Scanner subprocess behavior:**
- What's not tested: `scanner.Scan`, command construction, context cancellation, missing `mo`, non-zero exit stderr propagation, and sudo scan behavior.
- Files: `internal/scanner/scanner.go`, `internal/scanner/scanner_test.go`
- Risk: Cancellation and error handling can regress unnoticed; the sudo command bug is not covered.
- Priority: High

**Cleanup subprocess behavior:**
- What's not tested: Non-dry-run `cleanup.Run`, stdout/stderr teeing, non-zero exit code handling, context cancellation, large output, and actual `ParseSummary` examples from real `mo clean` output.
- Files: `internal/cleanup/cleanup.go`, `internal/cleanup/cleanup_test.go`, `tasks/prd-mole-tui.md`
- Risk: Cleanup can fail to stream, misreport errors, or leave a child process running without tests catching it.
- Priority: High

**Bubble Tea UI flow:**
- What's not tested: Initial loading, dashboard navigation/collapse, rescan failure with previous results, confirmation, dry-run cleanup path, help overlay, log scrolling, and quit/cancel semantics.
- Files: `internal/ui/model.go`
- Risk: Most user-visible behavior can regress while unit tests still pass.
- Priority: High

**Build/version behavior:**
- What's not tested: `--version` output with ldflags, missing `mo` startup failure, `go install` entrypoint behavior, and Make/just build artifacts.
- Files: `cmd/mole-tui/main.go`, `Makefile`, `justfile`, `AGENTS.md`
- Risk: Release/install workflows can break unnoticed; the docs already have stale wording around `.gitignore` even though `bin/` is ignored.
- Priority: Medium

---

*Concerns audit: 2026-06-20*
