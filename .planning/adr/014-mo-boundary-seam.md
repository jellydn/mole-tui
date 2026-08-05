# ADR-014: mo Boundary Seam — Injectable Subprocess Runner

## Status
Accepted

## Context
Candidate 2 from the 2026-08-04 architecture review: `scanner` and `cleanup` both shelled out to the `mo` binary directly via `os/exec`, with the resolved `moPath` string threaded from `main` through the UI model. The subprocess behaviour (sudo wrapping, `exec.CommandContext` cancellation, exit-code mapping, streaming) was **untested** — the seam was the OS process, and there was no second adapter to inject (one adapter = hypothetical, two = real). `AGENTS.md` and PRD §7 already promised an end-to-end test against a stub `mo` shim that did not exist.

## Decision
Introduce a small `internal/mo` package owning the mo boundary:

- **`mo.Runner`** — the seam: `Run(ctx, sudo, stdoutW, stderrW, args...) (exitCode, err)`. It streams stdout/stderr to the given writers, owns the subprocess plumbing, and wraps `sudo` when elevation is requested. Non-zero exits surface as the exit code, not an error; errors are reserved for invocation failures (missing binary, cancellation).
- **`mo.Resolve()` / `mo.NewRunner()`** — binary resolution lives at the boundary where it is consumed; `main` and the UI stop carrying the bare path.
- **`internal/mo/motest`** — a pure-Go stub runner (the second adapter): canned stdout/stderr, exit codes, delays, and invocation counters.
- **`scanner.Scan(ctx, mo.Runner, sudo)`** and **`cleanup.Run(ctx, opts, writer, mo.Runner)`** consume the seam; the UI model threads the runner instead of `MoPath string`.
- `cleanup`'s `io.Pipe`/goroutine stderr plumbing (the `14e391f` goroutine-leak fix) is absorbed by the runner, deleting that whole bug class.

This adds a fourth internal package, extending ADR-008's three-package layout. The seam is shared by both boundary consumers, so it lives in its own package rather than in either one.

## Consequences

### Positive
- Subprocess paths are now testable: scan (parse, sudo variant, error, cancellation) and cleanup (streaming, exit codes, stderr, cancellation) via the stub.
- The promised TUI end-to-end test can inject a stub runner directly into `ui.NewModel` — no stub executable required.
- Resolution and sudo wrapping live in one place; `main`/UI stop carrying `moPath`.
- Simpler `cleanup.Run`: the pipe/goroutine bug class is gone.

### Negative
- One more package (four internal packages plus the `motest` test-support package).
- `cmd/mole-sidecar` (feat/mole-sidecar stack) still uses the pre-seam signatures; it must be updated when that stack rebases onto main after this lands.
- Real `sudo` remains a single-adapter concern: the stub records elevation intent but does not exercise actual `sudo` (covered by ADR-005's opt-in elevation and manual verification).
