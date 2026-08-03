# INTEGRATIONS.md — External Integrations

> Mapped fresh on 2026-08-04.

## Runtime Integrations

### Mole CLI (`mo`) — the only runtime integration

The TUI is a pure **orchestrator** of the [`mole`](https://github.com/tw93/mole) CLI (`mo`) by tw93. All destructive actions go through `mo` — the TUI never deletes files itself (FR-6).

Invocation sites:

| Call site | Command | Purpose |
|-----------|---------|---------|
| `internal/scanner/scanner.go` — `Scan()` | `mo clean --dry-run` | Discovery + scan; parse output into `[]Section` + `ScanSummary` |
| `internal/scanner/scanner.go` — `Scan()` (sudo) | `sudo mo clean --dry-run` | Elevated re-scan for system caches (ADR-005) |
| `internal/cleanup/cleanup.go` — `Run()` | `mo clean` | All-or-nothing cleanup, no item args (ADR-002) |
| `internal/cleanup/cleanup.go` — `Run()` (sudo) | `sudo mo clean` | Elevated cleanup inheriting sudo intent from last scan |
| `cmd/mole-tui/main.go` | `exec.LookPath("mo")` | Resolve absolute `mo` path at startup; fatal error if missing |

Details:
- **No `--json` / machine-readable mode** — output is parsed heuristically (hybrid parsing, ADR-012). PRD open question OQ-1 tracks a future `--json` swap.
- **Cleanup contract**: `mo clean` takes no per-item arguments, so cleanup is all-or-nothing.
- **Dry-run simulation**: when the TUI is launched with `--dry-run`, `cleanup.Run` short-circuits — no subprocess, canned "Dry run complete — no files were modified" output (ADR-011).

### sudo (system binary)

- Used to elevate `mo` invocations (`sudo mo clean --dry-run`, `sudo mo clean`).
- **Non-root by default** (ADR-005): elevation is opt-in via the `S` keybinding; the sudo intent from the last scan propagates into cleanup (`lastScanSudo`).
- May prompt for a password in the terminal during invocation.

## What's NOT integrated

- ❌ No databases, storage, or state persistence (stateless across runs, v1)
- ❌ No auth providers or accounts
- ❌ No webhooks / network APIs / telemetry
- ❌ No cloud services
- ❌ No third-party Go libraries beyond the Charm TUI stack

## Developer / CI Integrations

| Service | Purpose | Where |
|---------|---------|-------|
| pre-commit | Local hook framework (lint/format/test gates) | `.pre-commit-config.yaml` |
| Renovate | Automated dependency PRs | `renovate.json` |
| GitHub (implicit) | Repo hosting, PRs, license badges in README | README |
| Ralph loop (`./ralph/ralph.sh`) | Autonomous agent loop driving CLI backends (opencode, amp, claude, codex, etc.) for story-by-story implementation | `ralph/` — internal dev workflow, not a runtime dep |

## Data Flow Across the Boundary

```
[mo clean --dry-run] --stdout--> scanner.Parse() --> ScanResult{Sections, Summary}
[mo clean]           --stdout/stderr--> cleanup.Run() --> streamed lines + Result{ExitCode, Stdout, Stderr, FreedText}
```
