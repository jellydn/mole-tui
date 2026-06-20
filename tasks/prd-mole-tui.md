# PRD: Mole TUI (v1 — Post-Grilling Revision)

> **Revision history:** Original PRD grilled on 2026-06-19. This revision incorporates 12 ADRs from the grilling session. See [.planning/adr/](file:///Users/huynhdung/src/tries/2026-06-16-mole-tui/.planning/adr/) for decision records.

## 1. Introduction / Overview

Mole TUI is a keyboard-driven terminal interface that orchestrates the existing [`mole`](https://github.com/tw93/mole) CLI (`mo`) for disk cleanup. It does **not** reimplement Mole's cleanup engine — it scans what Mole would clean via `mo clean --dry-run`, renders the results in a structured dashboard, and executes `mo clean` with a confirmation gate.

The goal is to give terminal-first power users (macOS, Linux, SSH) a discoverable, preview-first, 100% keyboard workflow on top of the Mole engine they already trust.

**v1 scope:** Orchestrate `mo clean` only. Discover items by parsing `mo clean --dry-run` output (ADR-001). Cleanup is all-or-nothing — `mo clean` takes no per-item arguments (ADR-002). Distribution via `go install` (no brew tap / release binaries in v1).

## 2. Goals

**Primary**

- Scan and display everything `mo clean --dry-run` reports — sections, item states, sizes — in a structured dashboard (ADR-003).
- Show total reclaimable space aggregated from dry-run output.
- Provide a confirmation gate before running `mo clean`.
- Stream `mo clean` stdout/stderr live during cleanup and display a summary on completion (ADR-007).
- Support optional `sudo` elevation for system-level caches (ADR-005).

**Secondary**

- Sub-200ms startup to loading screen render (actual scan takes longer) (ADR-004).
- Zero-mouse workflow (keyboard shortcuts cover every action).
- `--dry-run` flag for safe development and demos (ADR-011).
- Architecture leaves room for `mo analyze` / `mo status` and per-item selection in later phases.

## 3. User Stories

### US-001: Scan and discover cleanup items

**Description:** As a user, I want the TUI to scan what Mole would clean and show me the results so I understand what's reclaimable.

**Acceptance Criteria:**

- [ ] On launch, the TUI shows a full-screen loading state with a spinner and elapsed timer (ADR-004).
- [ ] The TUI runs `mo clean --dry-run` and parses: section headers (`➤`), item lines (raw text), and a best-effort total reclaimable size (ADR-012).
- [ ] Parsed sections and their items populate the dashboard when the scan completes.
- [ ] If `mo` is not on `$PATH`, the TUI prints a styled error to stderr and exits with code 1 (no TUI renders).
- [ ] If `mo clean --dry-run` exits non-zero, the loading screen transitions to an error state with the stderr output.
- [ ] If the output contains the `◎ System caches need sudo` line, a banner is set for the dashboard (ADR-005).
- [ ] `go build` and `go test ./...` pass.
- [ ] Parser is unit-tested with a captured real `mo clean --dry-run` fixture + 2 synthetic edge cases (empty output, error output).

### US-002: Render the dashboard view

**Description:** As a user, I want to see all cleanup sections and items in one screen so I can understand what is using my disk.

**Acceptance Criteria:**

- [ ] Dashboard shows: header (title, free disk space), sectioned item list, total reclaimable footer, key-hint footer.
- [ ] Sections (`➤`) are collapsible via `tab` key.
- [ ] Item lines are rendered as raw text within their sections, preserving Mole's original Unicode markers (`→`, `✓`, `◎`, `☞`) (ADR-012).
- [ ] If system caches were skipped, a persistent banner shows: "System caches skipped — press S to re-scan with sudo" (ADR-005).
- [ ] Layout adapts to terminal widths from 60 to 200 columns without clipping.
- [ ] `go build` passes.
- [ ] Verified visually with a terminal capture (80×24, 120×40).

### US-003: Navigate and browse with the keyboard

**Description:** As a power user, I want full keyboard control of the dashboard so I never need the mouse.

**Acceptance Criteria:**

- [ ] `↑` / `↓` (and `k` / `j`) move the cursor through items and section headers.
- [ ] `tab` toggles collapse/expand of the focused section.
- [ ] `?` opens a context-aware help overlay showing keybindings for the current screen (ADR-010).
- [ ] `q` and `ctrl+c` exit cleanly (no TTY corruption on exit).
- [ ] `go build` passes.
- [ ] Verified visually with a terminal capture.

### US-004: Re-scan and sudo elevation

**Description:** As a user, I want to re-scan on demand and optionally elevate to see system-level caches.

**Acceptance Criteria:**

- [ ] `r` re-runs `mo clean --dry-run` and refreshes the dashboard.
- [ ] `S` (shift-s) runs `sudo -v && mo clean --dry-run` for a full scan including system caches (ADR-005).
- [ ] During any scan, the TUI transitions to the loading screen with a spinner.
- [ ] `esc` cancels a running scan; previous results are preserved.
- [ ] If `mo clean --dry-run` exits non-zero during re-scan, the TUI shows the error as a banner and keeps previous results.
- [ ] `go test ./...` passes.

### US-005: Confirm and execute cleanup

**Description:** As a user, I want to clean my disk with a clear confirmation gate before any destructive action.

**Acceptance Criteria:**

- [ ] `enter` opens a confirmation modal showing: the literal command (`mo clean`), total reclaimable space, and y/n prompt.
- [ ] If `--dry-run` flag was passed to the TUI, the modal shows `[DRY RUN]` in the command preview (ADR-011).
- [ ] Default focus is **No**; `y` confirms, `n` / `esc` cancels.
- [ ] On confirm, the TUI transitions to a full-screen log pane streaming `mo clean` stdout/stderr live.
- [ ] In TUI `--dry-run` mode, the cleanup is simulated: "Dry run complete — no files were modified" (ADR-011).
- [ ] When the child exits 0, a summary line is shown if a total freed amount can be parsed from the output (ADR-007).
- [ ] When the child exits non-zero, the error and last 20 lines of stderr are shown with a `Press any key to return` prompt.
- [ ] The TUI never deletes files itself — it only invokes `mo clean` (FR-6).
- [ ] `go test ./...` passes.
- [ ] Verified visually with a terminal capture of the success and failure paths.

### US-006: View cleanup log and return

**Description:** As a user, I want to review the cleanup output and return to the dashboard.

**Acceptance Criteria:**

- [ ] After cleanup completes, the log pane becomes scrollable (`j/k` or `↑/↓`).
- [ ] A summary line at the bottom shows total freed if parseable, otherwise "Cleanup complete" (ADR-007).
- [ ] `enter` returns to the dashboard (previous scan results are stale — user can `r` to re-scan).
- [ ] `q` exits the TUI.
- [ ] `go build` passes.
- [ ] Verified visually with a terminal capture.

### US-007: Exit cleanly from any screen

**Description:** As a user, I want to quit the TUI from any view without leaving my terminal in a broken state.

**Acceptance Criteria:**

- [ ] `q` and `ctrl+c` exit from all 5 screens: loading, dashboard, confirmation, log/report, help.
- [ ] During a running cleanup, the first `ctrl+c` stops the streaming log and asks for a second `ctrl+c` to kill the child process.
- [ ] Terminal cursor and alt-screen are restored on every exit path (verified by running in tmux and attaching afterward).
- [ ] Exit code is 0 on normal quit, 1 on errors before TUI starts.
- [ ] `go test ./...` passes.

### US-008: Local install via `go install` / `make install`

**Description:** As a developer, I want a one-line install so I can try the TUI immediately.

**Acceptance Criteria:**

- [ ] `go install github.com/jellydn/mole-tui/cmd/mole-tui@latest` produces a working `mole-tui` binary on `$GOBIN`.
- [ ] `make install` does the same for users who prefer Make.
- [ ] `mole-tui --version` prints the build version.
- [ ] `mole-tui --dry-run` runs the full UI without executing `mo clean` (ADR-011).
- [ ] README documents both install paths and the prerequisite that `mo` be on `$PATH`.
- [ ] `go build ./...` passes.

## 4. Functional Requirements

- **FR-1:** The TUI must be a single Go module at `github.com/jellydn/mole-tui` with a `cmd/mole-tui` entrypoint and 3 internal packages: `scanner`, `cleanup`, `ui` (ADR-008).
- **FR-2:** `scanner` must invoke `mo clean --dry-run`, parse section headers (`➤`) into `[]Section{Name, Lines}`, aggregate a best-effort total reclaimable size, and detect the sudo-needed banner. A `context.Context` must bound the run; cancellation must stop the subprocess within 1s (ADR-001, ADR-003, ADR-012).
- **FR-3:** `cleanup` must invoke `mo clean` (no item arguments — all-or-nothing) via `os/exec`, streaming stdout/stderr to the UI. It must support a `DryRun` mode that short-circuits execution (ADR-002, ADR-011).
- **FR-4:** The UI must be implemented with Bubble Tea + Lip Gloss + Bubbles. 5 screens: Loading, Dashboard, Confirmation (modal), Log/Report, Help (ADR-010).
- **FR-5:** The TUI must show a full-screen loading state with spinner + elapsed timer whenever a `mo` subprocess is running and disable conflicting shortcuts (ADR-004).
- **FR-6:** The TUI must never delete files itself. All destructive actions go through `mo clean`.
- **FR-7:** Output parsing must be tolerant: unrecognized lines are included as raw text within the current section, not rejected (ADR-012).
- **FR-8:** The TUI must not require root by default. Sudo elevation is opt-in via `S` keybinding (ADR-005).

## 5. Non-Goals (Out of Scope for v1)

- **NG-1:** Orchestrating any command other than `mo clean` (`mo analyze`, `mo status`, `mo uninstall`, `mo optimize`, `mo purge`, `mo history`, `mo installer`).
- **NG-2:** Homebrew formula, GitHub Releases binaries, or any package-manager publication.
- **NG-3:** Cross-host / SSH / multi-server dashboards.
- **NG-4:** Scheduled cleanup (`mole-tui schedule`).
- **NG-5:** Safe-mode age filters (30/60/90 days) — would need Mole support that may not exist.
- **NG-6:** Continuous monitoring / dashboard-agent mode.
- **NG-7:** A new cleanup engine or any direct filesystem deletion. TUI is strictly an orchestrator.
- **NG-8:** A graphical / web frontend.
- **NG-9:** Windows support (Mole is macOS-first; Linux is opportunistic via `mo` builds).
- **NG-10:** Per-item selective cleanup — `mo clean` takes no item arguments. Multi-select is deferred until Mole exposes per-item selection (ADR-002).
- **NG-11:** Preview pane with per-item file tree — `mo clean --dry-run` does not expose file listings (ADR-006).
- **NG-12:** Per-item before/after size comparison in the completion report (ADR-007).

## 6. Design Considerations

- **UI style:** Match the Mole CLI aesthetic — bordered panels, monospaced, minimal color, dark/light aware. Lip Gloss styles only; no raw ANSI.
- **Layout:** Single-pane dashboard with collapsible sections. Confirmation is a modal overlay. Log/report and help are full-screen.
- **Defaults:** First section header is focused on launch. Cursor wraps at list ends.
- **Persistence:** None in v1 — no config file, no history file. The TUI is stateless across runs.
- **Loading:** Full-screen loading state with spinner and elapsed timer. Shown on startup and re-scan (ADR-004).

## 7. Technical Considerations

- **Language / Stack:** Go 1.22+, Bubble Tea v1.x, Lip Gloss v1.x, Bubbles v1.x. No CGO.
- **Process invocation:** `os/exec` with `context.Context`. Stdout/stderr captured into `io.Pipe` and pumped into Bubble Tea via `tea.Cmd`.
- **Parsing strategy:** Hybrid approach (ADR-012). Section headers (`➤`) are parsed structurally. Individual item lines are stored as raw strings. Size values are extracted via regex for aggregate totals. Unrecognized lines are kept, never rejected.
- **Cancellation:** All subprocesses take `ctx`. UI-initiated cancel uses `esc`; program quit uses `ctrl+c` and triggers `cancel()` on the active context.
- **Error model:** Errors flow through Bubble Tea as a `tea.Msg` (e.g. `errMsg{err}`). The UI renders them in a banner; the program never panics in normal use. Missing `mo` binary is a pre-TUI fatal error (stderr + exit 1).
- **Performance:** TUI binary must reach the loading screen in < 200ms. The scan itself takes as long as Mole needs (~1–3 min typical).
- **Testing:** Unit tests for `scanner` parser with a captured real fixture + 2 synthetic edge cases. Golden-file tests for screen renders using `teatest` (or hand-rolled snapshot). One end-to-end test that drives the TUI against a stub `mo` shim returning canned output.
- **CI:** `go vet ./...`, `gofmt -l` clean, `go test ./...`, and a smoke build for `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `linux/arm64`.
- **Repo layout** (revised per ADR-008):
  ```
  cmd/
   └── mole-tui/main.go
  internal/
   ├── scanner/       # invoke `mo clean --dry-run`, parse output → []Section
   ├── cleanup/       # invoke `mo clean`, stream stdout/stderr, parse summary
   └── ui/            # Bubble Tea models, views, keybindings, all 5 screens
  ```

## 8. Success Metrics

- `mole-tui` installs via `go install` in a single command and runs against an existing `mo` install with no extra setup.
- A new user can launch the TUI, review the scan, and trigger cleanup in **≤ 30 seconds** of interaction (after scan completes).
- **100% keyboard-driven** — every action in the UI has a documented key binding (`?` help); the mouse is not required.
- **100% coverage of Mole's dry-run output** — every section and item Mole reports appears in the dashboard without code changes in `mole-tui`.
- Startup to loading screen **< 200ms** on the target machines.
- No P0 bugs in the first 30 days; the log/report screen always reflects the actual `mo clean` output.

## 9. Resolved Questions

> Resolved during the grilling session on 2026-06-19. See ADRs for full context.

- **RQ-1 (was OQ-5):** `--dry-run` flag on the TUI itself? **Yes — added to v1** (ADR-011).
- **RQ-2:** Does `mo --help` list per-item cleanup targets? **No** — items only appear in `mo clean --dry-run` output. Discovery and scan merged (ADR-001).
- **RQ-3:** Does `mo clean` accept per-item arguments? **No** — cleanup is all-or-nothing in v1 (ADR-002).
- **RQ-4:** What format does `mo clean --dry-run` produce? **Sections (`➤`) with 4 item states (`→`, `✓`, `◎`, `☞`), inconsistent sizes** — hybrid parsing adopted (ADR-003, ADR-012).

## 10. Open Questions

- **OQ-1:** Does Mole emit a stable, machine-parseable `--json` flag for `mo clean --dry-run`? If yes in a future Mole release, swap the hybrid parser for a JSON one (out of v1 scope).
- **OQ-2:** Should the confirmation modal require a typed phrase ("yes") for large cleanups (> 50 GB)? **Defer to v2**; v1 uses a single `y`.
- **OQ-3:** What does `mo clean` (actual, non-dry-run) output look like? Need to capture a fixture for the summary parser. **Action:** Run `mo clean` once and capture the output format.
