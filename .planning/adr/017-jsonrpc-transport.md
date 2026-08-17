# ADR-017: Extract the JSON-RPC Transport from the Sidecar Server

## Status

Accepted

## Context

`cmd/mole-sidecar` exposes the scan/cleanup engine over newline-delimited JSON-RPC 2.0 on stdio (ADR-016). As first shipped, the `server` type mixed several concerns into one shallow type:

- JSON-RPC wire types (`request`, `response`, `event`, `rpcError`);
- newline-delimited framing — the `serve` loop reading request lines, skipping blanks, and mapping malformed input to a `-32700` parse error;
- wire encoding and synchronized writes (`respond`, `emit` guarded by a mutex);
- method dispatch (`handle`), parameter decoding, and the scan/cleanup lifecycle (`startScan`, `startCleanup`, cancellation).

Because framing, encoding, and lifecycle were entangled, protocol behaviour (parse errors, notification framing, concurrent event writes) could only be exercised through the domain server, and the transport logic could not be reused or tested independently. The interface was nearly as complex as the implementation.

## Decision

Extract a focused **`internal/jsonrpc`** transport module that owns the JSON-RPC 2.0 protocol, leaving `cmd/mole-sidecar`'s `server` responsible for domain operations and lifecycle only.

The transport provides:

- **`Server`** — `NewServer(in, out, handler)` with `Serve()` (the newline-delimited request loop: read a line, skip blanks, decode, dispatch, write the response; malformed lines get a `-32700` parse error) and `Notify(method, params)` (goroutine-safe server-initiated events). All writes are serialized by one internal mutex.
- **`Handler`** — a narrow interface, `Handle(method string, params json.RawMessage) (any, *Error)`, implemented by the sidecar server. The transport owns framing and encoding; the handler owns parameter decoding and lifecycle.
- **`Error`** and named error-code constants (`CodeParseError`, `CodeMethodNotFound`, `CodeInvalidParams`, `CodeServerError`, …).

The wire contract is unchanged: the sidecar `server` still implements the ADR-016 methods (`ping`, `scan.start`, `scan.cancel`, `cleanup.start`, `cleanup.cancel`) and emits the same notifications (`scan.done`, `scan.error`, `scan.cancelled`, `cleanup.line`, `cleanup.done`, `cleanup.error`). Parse errors still carry `"id": null`, matching the JSON-RPC spec.

## Consequences

### Positive

- Protocol behaviour is independently testable: framing, parse errors, blank-line handling, notification encoding, and concurrent `Notify` are covered in `internal/jsonrpc` without a domain server.
- The transport is a deep module — its interface (Serve + Notify + Handler) is much smaller than its implementation, and it is reusable by any future client or server.
- The sidecar server shrinks to a single responsibility: the scan/cleanup lifecycle and event translation.
- Concurrent write safety is centralized in one place rather than repeated per message type.
- The existing end-to-end sidecar tests keep exercising the full path through the transport.

### Negative

- One more internal package, extending the ADR-008 layout alongside `mo` (ADR-014) and the sidecar (ADR-016).
- The transport is deliberately generic and does not know the mole domain; method names, parameter shapes, and event names remain a compatibility surface versioned at the sidecar layer.
- Domain errors must be translated to JSON-RPC error codes by the handler; the transport only encodes what the handler returns.
- Faithful to the original wire behaviour, the transport does not validate the `jsonrpc` version field, does not implement batched requests, and responds to every request line (including ones with no id) — acceptable for the single-client stdio use case, but worth revisiting if the protocol grows.

## Alternatives Considered

### Keep framing and encoding in the sidecar server

Rejected because it leaves the shallow type that motivated the change: protocol logic stays entangled with lifecycle, untestable in isolation and unreusable.

### A transport library of pure encode/decode functions

Considered — the server would still own the read/write loop and write-serialization. Rejected in favour of a `Server` that owns the loop, because it yields a deeper module and centralizes the concurrency concerns that caused the original goroutine-safety review.
