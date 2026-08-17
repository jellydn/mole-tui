# TESTING.md — Testing

> Mapped fresh on 2026-08-04. Rewritten 2026-08-17 — the test suite grew substantially with the mo seam (ADR-014), tea-free cleanup stream (ADR-015), UI operation controller, and JSON-RPC transport (ADR-016).

## Test Runner

- Single entrypoint: **`go test ./...`** — run via `just test` / `make test`.
- Pre-commit runs `go-unit-tests` on every commit; `just ci` runs tests as the third stage (fmt → vet → **test** → build).
- No coverage thresholds configured; `go test -cover` and `go test -race ./...` work out of the box. The concurrency-heavy packages (`cleanup` stream, `jsonrpc` notify) are exercised under `-race`.

## The Test-Double Strategy

Two complementary seams make the subprocess and transport paths testable without a real `mo` binary:

1. **`internal/mo/motest`** — a pure-Go stub `mo.Runner` (canned stdout/stderr, exit codes, `Sleep` delays, invocation counters). Used by `scanner`, `cleanup`, and `ui` tests.
2. **A fake `mo` shell script** — written into a `t.TempDir()` by `internal/mo/mo_test.go` to exercise the **production** `execRunner` end-to-end (real `os/exec`, exit-code mapping, cancellation, missing binary).

## What's Tested

### `internal/mo` — `mo_test.go` (158 lines)

Exercises the production `execRunner` against a fake `mo` shell script:
- `TestExecRunnerScan` — dry-run scan streams stdout, exit 0.
- `TestExecRunnerCleanup` — cleanup streams stdout, exit 0.
- `TestExecRunnerNonZeroExit` — non-zero exit returns the code with a **nil** error (the exit-code-vs-error contract).
- `TestExecRunnerCancellation` — pre-cancelled context surfaces `context.Canceled` (deterministic, orphan-free).
- `TestExecRunnerMidRunCancellation` — cancel during a running command surfaces the context error and is bounded by `commandWaitDelay` (no inherited-pipe hang).
- `TestExecRunnerMissingBinary` — invocation failure returns `-1` + error.

### `internal/scanner` — `scanner_test.go` (193 lines) + `scan_test.go` (97 lines)

- **Parser (table-driven over fixtures):** `TestParse` (3 fixtures: real captured output, empty, error-output), `TestStripANSI`, `TestParseSize`, `TestParseReclaimableSize`, `TestParseNoSizeLines`, `TestANSICleanYourMac`, `TestSectionName`, `TestTotalReclaimable`, `TestSystemCachesSkipped`, `TestWhitelistNotSection`.
- **Subprocess path via `motest`:** `TestScanParsesOutput` (end-to-end parse), `TestScanSudoVariant` (sudo flag reaches the seam, ADR-005), `TestScanNoSudoByDefault` (elevation is opt-in), `TestScanError` (non-zero exit surfaces stderr text), `TestScanCancellation` (cancelled context → `context.Canceled`).

### `internal/cleanup` — `cleanup_test.go` (47) + `run_test.go` (113) + `stream_test.go` (150)

- **`cleanup_test.go`:** `TestRunDryRun` (short-circuit: exit 0, canned message, no subprocess), `TestParseSummary` (table-driven over freed/saved/cleaned phrasing).
- **`run_test.go` (synchronous `Run`/`Session.Run` via `motest`):** `TestRunStreamsOutput` (live streaming + freed summary), `TestSessionRunDirect` (deep `Session.Run` API), `TestRunExitCode` (non-zero exit on `Result`, not error), `TestRunStreamsStderr` (stderr streamed live too), `TestRunSudo` (sudo intent reaches the seam), `TestRunCancellation` (cancelled context → error).
- **`stream_test.go` (asynchronous `Session.Start` event stream):** `TestSessionStartStreamsLinesBeforeCompletion` (line events then a completion event), `TestSessionStartDryRunProducesCompletion`, `TestSessionStartCancellationCompletesWithoutLeaking` (cancelled run still delivers completion, then closes), `TestStartFullBufferStopsAfterCancellation` (bounded 256-cap buffer; abandoned full stream unblocks on cancel — no producer leak), `TestSessionStartAbandonedConsumerStopsAfterCancellation` (no consumer → producer still finishes on cancel).

### `internal/ui` — `operations_test.go` (83 lines)

Tests the `operationController` (the Bubble Tea-facing lifecycle):
- `TestOperationControllerScanCancellation` — cancel returns true, command yields `scanCancelledMsg`, clear disables further cancel.
- `TestOperationControllerCleanupTranslatesEvents` — stream lines become `cleanupStreamMsg`, completion becomes `cleanupCompleteMsg`.
- `TestOperationControllerCleanupCancellation` — cancelled cleanup yields `cleanupCompleteMsg` with `context.Canceled`.

### `internal/jsonrpc` — `jsonrpc_test.go` (135 lines)

Tests the transport independently of the domain:
- `TestServeDispatchesAndWritesResponse`, `TestServePropagatesHandlerError`, `TestServeParseError` (`-32700`), `TestServeSkipsBlankLines`, `TestNotifyWritesNewlineDelimitedEvents`, `TestNotifyConcurrentIsSafe` (concurrent `Notify` under `-race`).

### `cmd/mole-sidecar` — `main_test.go` (117 lines)

End-to-end through the transport + server:
- `TestServePing` (round-trip without a `mo` binary), `TestServeUnknownMethod` (`-32601`), `TestServeMalformedJSON` (`-32700`), `TestServeCleanupDryRun` (async cleanup: accepted response + `cleanup.line` + `cleanup.done` events).

## Fixture Strategy

- Fixtures live in `internal/<pkg>/testdata/` — currently `internal/scanner/testdata/` (real captured dry-run + ANSI variant + empty + error).
- One **real captured** fixture (`mo-clean-dryrun-real.txt`) anchors parser realism; synthetic fixtures cover edge cases.
- Parser tests use `os.ReadFile` on the fixture path relative to the package.

## Remaining Coverage Gaps

- ❌ No golden-file / `teatest` snapshot tests for screen rendering (AGENTS.md still claims these; the 5-screen state machine is only covered indirectly via the `operationController` tests).
- ❌ No full TUI end-to-end test driving `tea.NewProgram(ui.NewModel(...))` against a stub runner (the seam makes this possible — `motest` can be injected directly).
- ❌ No `internal/cleanup/testdata/` directory (cleanup is tested via `motest`/fake script instead of fixtures).
- ❌ No explicit test asserting `ParseSummary` against a real captured `mo clean` (non-dry-run) output — the freed-text regex is still written from assumptions (see CONCERNS.md).
- ❌ The sidecar's `scan.start`/`scan.cancel` async paths and the `cleanup.line` chunking across many writes aren't asserted (only the dry-run cleanup path is end-to-end).
