# External Integrations

**Analysis Date:** 2026-06-20

## APIs & External Services

**External CLIs:**
- Mole CLI (`mo`) - core backend used to preview and perform cleanup; `cmd/mole-tui/main.go` requires it on `$PATH`, `internal/scanner/scanner.go` runs `mo clean --dry-run`, and `internal/cleanup/cleanup.go` runs `mo clean`.
- `sudo` - optional local privilege escalation for fuller system-cache dry-runs; triggered by the `S` key path and implemented as `sudo sh -c "mo clean --dry-run"` in `internal/scanner/scanner.go`.
- SDK/Client: Go standard library `os/exec` is used for process execution in `cmd/mole-tui/main.go`, `internal/scanner/scanner.go`, and `internal/cleanup/cleanup.go`.
- Auth: no API auth; local sudo authentication may be prompted by the operating system for the optional elevated scan in `internal/scanner/scanner.go`.

**Remote APIs:**
- None found. Source under `cmd/` and `internal/` uses local subprocesses and terminal UI libraries, not HTTP clients or SDKs.

## Data Storage

**Databases:**
- None. No database drivers are declared in `go.mod`, and no persistence layer appears under `cmd/` or `internal/`.
- Connection: None.
- Client: None.

**File Storage:**
- Local filesystem only, indirectly through the external Mole CLI cleanup behavior. The TUI itself does not delete files; it invokes `mo clean` in `internal/cleanup/cleanup.go` and parses dry-run output in `internal/scanner/scanner.go`.
- Test fixtures are local files read by tests under `internal/scanner/testdata/`, as shown in `internal/scanner/scanner_test.go`.

**Caching:**
- None. There is no cache service or persisted local cache; scan results live in the Bubble Tea model in memory in `internal/ui/model.go`.

## Authentication & Identity

**Auth Provider:**
- None.
- Implementation: no users, accounts, tokens, OAuth, or API keys are present in `go.mod`, `cmd/`, or `internal/`.
- Local privilege escalation is limited to optional sudo-based scanning, implemented in `internal/scanner/scanner.go`; this is OS authentication, not app identity.

## Monitoring & Observability

**Error Tracking:**
- None. No error tracking service or telemetry SDK is declared in `go.mod`.

**Logs:**
- Local terminal/log-pane output only. Cleanup stdout/stderr are streamed and captured in `internal/cleanup/cleanup.go`, then displayed in the TUI log/report screen in `internal/ui/model.go`.
- Scanner errors are surfaced as UI error/loading messages through `internal/ui/model.go`.

## CI/CD & Deployment

**Hosting:**
- None. The project builds a local CLI/TUI binary, with install/build flows documented in `README.md`, `justfile`, and `Makefile`.

**CI Pipeline:**
- No `.github` workflow files were present in the repository file listing, so no hosted CI pipeline was found.
- Local CI-equivalent command is `just ci` (`fmt → vet → test → build`) in `justfile`; `Makefile` provides individual `fmt`, `vet`, `test`, and `build` targets.
- pre-commit provides local quality gates via `.pre-commit-config.yaml`.
- Renovate dependency automation is configured by `renovate.json`, but that is dependency management rather than an application deployment pipeline.

## Environment Configuration

**Required env vars:**
- None for application configuration. No `os.Getenv` usage was found in `cmd/` or `internal/` during source review.
- `$PATH` must resolve the `mo` binary; this is validated with `exec.LookPath("mo")` in `cmd/mole-tui/main.go`.
- `$GOBIN`/`$GOPATH` only affect install destination in `Makefile` and normal `go install` behavior; they are not runtime configuration.

**Secrets location:**
- None. No secrets are configured or referenced in source, `go.mod`, `justfile`, `Makefile`, `.pre-commit-config.yaml`, or `renovate.json`.

## Webhooks & Callbacks

**Incoming:**
- None. This is a terminal application with no server, routes, or HTTP listener in `cmd/` or `internal/`.

**Outgoing:**
- None. The app does not call webhooks or external network APIs; it invokes local CLI commands (`mo`, optional `sudo`) via `os/exec` in `internal/scanner/scanner.go` and `internal/cleanup/cleanup.go`.

---

*Integration audit: 2026-06-20*
