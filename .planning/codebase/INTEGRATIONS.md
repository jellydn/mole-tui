# INTEGRATIONS.md — External Integrations

> Mapped fresh on 2026-08-04. Refreshed 2026-08-17 after the mo boundary seam (ADR-014) and stdio JSON-RPC sidecar (ADR-016) landed.

## Runtime Integrations

### Mole CLI (`mo`) — the only subprocess integration

The project is a pure **orchestrator** of the [`mole`](https://github.com/tw93/mole) CLI (`mo`) by tw93. All destructive actions go through `mo` — neither the TUI nor the sidecar ever deletes files itself (FR-6).

Every `mo` invocation flows through the **`internal/mo` boundary seam** (ADR-014): `mo.Runner.Run(ctx, sudo, stdoutW, stderrW, args...)`. Production callers pass `mo.NewRunner(moPath)`; tests pass `internal/mo/motest`'s stub. The resolved path is threaded once from `main` (via `mo.Resolve()`), never carried as a bare string by consumers.

Invocation sites:

| Call site | Command | Purpose |
|-----------|---------|---------|
| `internal/scanner/scanner.go` — `Scan(ctx, runner, sudo)` | `mo clean --dry-run` | Discovery + scan; parse output into `[]Section` + `ScanSummary` |
| `internal/scanner/scanner.go` — `Scan(..., sudo=true)` | `sudo mo clean --dry-run` | Elevated re-scan for system caches (ADR-005) |
| `internal/cleanup/cleanup.go` — `Run(ctx, opts, writer, runner)` | `mo clean` | All-or-nothing cleanup, no item args (ADR-002) |
| `internal/cleanup/cleanup.go` — `Run(..., opts.Sudo=true)` | `sudo mo clean` | Elevated cleanup inheriting sudo intent from last scan |
| `cmd/mole-tui/main.go` / `cmd/mole-sidecar/main.go` | `mo.Resolve()` | Resolve absolute `mo` path at startup; fatal error if missing |

Details:
- **No `--json` / machine-readable mode** — output is parsed heuristically (hybrid parsing, ADR-012). PRD open question OQ-1 tracks a future `--json` swap.
- **Cleanup contract**: `mo clean` takes no per-item arguments, so cleanup is all-or-nothing.
- **Dry-run simulation**: when run with `Options.DryRun`, `cleanup.Run` short-circuits — no subprocess, canned "Dry run complete — no files were modified" output (ADR-011).

### sudo (system binary)

- Used to elevate `mo` invocations (`sudo mo clean --dry-run`, `sudo mo clean`) via `mo.Runner`.
- **Non-root by default** (ADR-005): elevation is opt-in via the `S` keybinding; the sudo intent from the last scan propagates into cleanup (`lastScanSudo`).
- May prompt for a password in the terminal during invocation.

### Sidecar wire protocol (JSON-RPC 2.0 over stdio)

`cmd/mole-sidecar` exposes the scan/cleanup engine over newline-delimited JSON-RPC 2.0 on stdin/stdout (ADR-016). It is a **process-level seam** for desktop GUIs and other clients that don't want to embed the Bubble Tea TUI or own `mo` subprocesses.

- Framing and wire encoding live in `internal/jsonrpc` (a reusable transport: `Serve` loop + `Notify` for async events + narrow `Handler` interface).
- The `server` in `cmd/mole-sidecar/main.go` implements `jsonrpc.Handler` and owns only the scan/cleanup lifecycle.
- Methods: `ping`, `scan.start`, `scan.cancel`, `cleanup.start`, `cleanup.cancel`; notifications: `scan.done`, `scan.error`, `scan.cancelled`, `cleanup.line`, `cleanup.done`, `cleanup.error`.
- The sidecar writes protocol messages to stdout; non-protocol diagnostics go to stderr.

## What's NOT integrated

- ❌ No databases, storage, or state persistence (stateless across runs, v1)
- ❌ No auth providers or accounts
- ❌ No webhooks / network APIs / telemetry (the sidecar is local stdio only — no HTTP, ADR-016)
- ❌ No cloud services
- ❌ No third-party Go libraries beyond the Charm TUI stack

## Developer / CI Integrations

| Service | Purpose | Where |
|---------|---------|-------|
| pre-commit | Local hook framework (lint/format/test gates) | `.pre-commit-config.yaml` |
| Renovate | Automated dependency PRs | `renovate.json` |
| GitHub (implicit) | Repo hosting, PRs, license badges in README; third-party security bots (GitGuardian, Socket) + CodeRabbit review | README |
| Ralph loop (`./ralph/ralph.sh`) | Autonomous agent loop driving CLI backends (opencode, amp, claude, codex, etc.) for story-by-story implementation | `ralph/` — internal dev workflow, not a runtime dep |

## Data Flow Across the Boundary

```
[mo clean --dry-run] --stdout--> scanner.Scan(ctx, runner, sudo) --> Parse() --> ScanResult{Sections, Summary}
[mo clean]           --stdout/stderr--> cleanup.Run(ctx, opts, writer, runner) --> streamed lines + Result{ExitCode, Stdout, Stderr, FreedText}
                                         cleanup.Session.Start() --> tea-free Event stream (EventLine / EventDone)
[sidecar client]     --JSON lines--> internal/jsonrpc (Serve) --> server.Handle --> mo seam --> Notify() --> JSON lines
```
