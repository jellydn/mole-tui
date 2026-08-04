# ADR-015: Tea-Free Cleanup Stream

## Status

Accepted

## Context

The cleanup stream needs to run `mo clean` asynchronously while preserving the live log required by ADR-007. Before this change, `internal/ui` owned the `io.Pipe`, scanner, output buffering, and two goroutines that bridged cleanup output into Bubble Tea messages. That made the UI responsible for subprocess-stream lifecycle and made the cleanup behavior difficult to exercise without Bubble Tea.

Candidate 1 asked whether the deepened stream should be tea-free or tea-coupled. The cleanup domain already has a plain-Go seam at `mo.Runner` (ADR-014), so coupling its stream to Bubble Tea would add framework knowledge to a package that does not need it.

## Decision

Keep the cleanup stream tea-free.

`cleanup.Start` starts an asynchronous cleanup and returns a `cleanup.Stream` containing a channel of plain `cleanup.Event` values:

- line events contain one output line without its trailing newline;
- one completion event contains the final `cleanup.Result` and error;
- the event channel then closes. If a consumer abandons a full stream and cancels
  the context, the stream may close without a completion event to avoid leaking
  its producer.

`internal/cleanup` owns the pipe, scanner, buffering, cancellation-aware sends, and goroutine lifecycle. The Bubble Tea UI only translates events into its existing `tea.Msg` values and remains responsible for rendering and navigation.

## Consequences

### Positive

- The cleanup stream has a small framework-independent interface and can be tested without a Bubble Tea program.
- Subprocess, output ordering, completion, dry-run, and cancellation behavior have direct package tests.
- The UI loses its pipe and goroutine plumbing and only handles presentation concerns.
- Backpressure is explicit: the bounded event channel prevents unbounded buffering while preserving live output.

### Negative

- `internal/cleanup` owns concurrency and must keep its event channel cancellation-safe.
- The UI still needs one Bubble Tea command per event to remain responsive and preserve message ordering.
- The stream currently emits combined stdout/stderr lines in write order, without labeling their source; the final `Result` retains separate buffers for the report.
