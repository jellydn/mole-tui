# ADR-016: Use a stdio JSON-RPC Sidecar for Programmatic Access

Date: 2026-08-04

## Status

Accepted

## Context

The primary product is a Bubble Tea terminal interface, but other clients may need to drive the Mole cleanup workflow without owning the `mo` subprocess or embedding the TUI. In particular, a future desktop GUI can use the existing scan and cleanup behavior while remaining independent of the terminal presentation layer.

Embedding another client directly into `cmd/mole-tui` or `internal/ui` would couple that client to Bubble Tea lifecycle and view state. It would also make subprocess ownership, live cleanup output, cancellation, and completion reporting inconsistent across clients.

The project therefore needs a small, process-level seam that can be launched by an external client and that preserves the existing rule that mole-tui orchestrates `mo` rather than deleting files itself.

## Decision

Provide a separate `mole-sidecar` executable at `cmd/mole-sidecar` that exposes the scan and cleanup engine over JSON-RPC 2.0 using newline-delimited JSON on stdin/stdout.

The sidecar owns the `mo` subprocesses and translates their results into JSON-RPC responses and server-initiated notifications. It does not implement a second cleanup engine and does not delete files directly; destructive work continues to go through `mo clean`.

The initial protocol supports these request methods:

- `ping` — returns the sidecar version and resolved `mo` path.
- `scan.start` — starts `mo clean --dry-run`, with an optional `sudo` field in the request parameters.
- `scan.cancel` — requests cancellation of the active scan.
- `cleanup.start` — starts `mo clean`, with optional `dryRun` and `sudo` fields in the request parameters.
- `cleanup.cancel` — requests cancellation of the active cleanup.

Long-running operations return an immediate acceptance response and report progress or completion as notifications without an `id`:

- Scan notifications: `scan.done`, `scan.error`, `scan.cancelled`.
- Cleanup notifications: `cleanup.line` (which carries an output chunk), `cleanup.done`, `cleanup.error`.

The sidecar writes protocol messages to stdout. Diagnostics that are not protocol messages go to stderr. The protocol is line-oriented so clients can consume responses and events incrementally without a terminal or GUI-specific transport.

## Consequences

### Positive

- External clients can reuse the scan and cleanup workflow without depending on Bubble Tea or `internal/ui`.
- The sidecar provides one explicit process boundary for subprocess ownership, cancellation, streaming output, and completion events.
- JSON-RPC request/response IDs remain available for immediate acknowledgements, while notifications support asynchronous scan and cleanup events.
- Stdio keeps the first integration simple, local, and dependency-free; desktop clients can launch and supervise the sidecar as a child process.
- The existing safety model is preserved: the sidecar orchestrates `mo clean` and never performs file deletion itself.

### Negative

- The JSON-RPC method names, parameters, event names, and result shapes become a compatibility surface that must be versioned carefully.
- Clients must manage a child process and parse newline-delimited JSON, including asynchronous notifications that may arrive independently of request responses. Cleanup output notifications carry arbitrary write chunks rather than guaranteed complete lines.
- Stdio is local and single-client oriented; network access, authentication, and multi-client concurrency are intentionally outside this decision.
- The sidecar currently permits at most one active scan and one active cleanup at a time; richer job management can be added later without coupling it to the TUI.

## Alternatives Considered

### Embed a GUI in the TUI binary

Rejected because it couples a second presentation framework to Bubble Tea and expands the binary's runtime and distribution concerns.

### Expose an HTTP server

Deferred. HTTP would add a network lifecycle, port management, and authentication considerations that are not required for a local desktop client launched alongside the sidecar.

### Let each client invoke `mo` directly

Rejected because it duplicates subprocess invocation, cancellation, output parsing, and safety behavior across clients.
