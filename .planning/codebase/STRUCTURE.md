# STRUCTURE.md — Directory Layout & Structure

> Mapped fresh on 2026-08-04. Refreshed 2026-08-17 after the mo seam (ADR-014), tea-free cleanup stream (ADR-015), and stdio JSON-RPC sidecar (ADR-016) landed.

## Directory Tree

```
.
├── cmd/
│   ├── mole-tui/                   # TUI binary entrypoint
│   │   └── main.go                 # flags, mo.Resolve(), tea.NewProgram (48 ln)
│   └── mole-sidecar/               # JSON-RPC sidecar binary entrypoint (ADR-016)
│       ├── main.go                 # server (jsonrpc.Handler) + lifecycle (199 ln)
│       └── main_test.go            # ping / unknown method / parse error / cleanup e2e (117 ln)
├── internal/
│   ├── mo/                         # mo subprocess boundary seam (ADR-014)
│   │   ├── mo.go                   # Runner, execRunner, Resolve(), NewRunner() (65 ln)
│   │   ├── mo_test.go              # exec runner via fake-mo shell script (158 ln)
│   │   └── motest/
│   │       └── motest.go           # programmable stub mo.Runner (74 ln)
│   ├── scanner/                    # scan side: invoke + parse `mo clean --dry-run`
│   │   ├── scanner.go              # Scan(), Parse(), ParseSize(), types (172 ln)
│   │   ├── scanner_test.go         # table-driven parser tests (193 ln)
│   │   ├── scan_test.go            # Scan() through the motest seam (97 ln)
│   │   └── testdata/
│   │       ├── mo-clean-dryrun-real.txt      # captured real mo dry-run fixture
│   │       ├── mo-clean-dryrun-real.ansi.txt # ANSI-styled variant fixture
│   │       ├── empty.txt                     # synthetic edge case
│   │       └── error-output.txt              # synthetic edge case (mo missing)
│   ├── cleanup/                    # cleanup side: invoke + stream `mo clean`
│   │   ├── cleanup.go              # Session, Run(), Start(), ParseSummary() (208 ln)
│   │   ├── cleanup_test.go         # dry-run short-circuit + summary parsing (47 ln)
│   │   ├── run_test.go             # Run()/Session.Run() via motest (113 ln)
│   │   └── stream_test.go          # Session.Start() event stream tests (150 ln)
│   ├── jsonrpc/                    # newline-delimited JSON-RPC 2.0 transport (ADR-016)
│   │   ├── jsonrpc.go              # Server (Serve + Notify), Handler, Error (127 ln)
│   │   └── jsonrpc_test.go         # framing/encoding/dispatch tests (135 ln)
│   └── ui/                         # Bubble Tea model, views, keybindings, styles
│       ├── model.go                # Model, 5 screens, handlers + views (740 ln)
│       ├── operations.go           # operationController: scan/cleanup lifecycle (95 ln)
│       ├── operations_test.go      # controller cancellation + event translation (83 ln)
│       └── styles.go               # Lip Gloss v2 styles + adaptive palette (101 ln)
├── tasks/
│   └── prd-mole-tui.md             # product requirements (post-grilling revision)
├── ralph/                          # autonomous agent loop (dev workflow)
│   ├── ralph.sh                     # loop driver (314 ln)
│   ├── prompt.md, prd.json          # ralph inputs
│   ├── prompt-{pi,codex,copilot,opencode,cmd,agy,amp,mino}.md  # per-backend prompts
│   └── progress.txt                 # story-by-story progress log
├── .planning/
│   ├── adr/                         # 16 accepted ADRs (001–016)
│   ├── codebase/                    # ← this codebase map (7 docs)
│   └── glossary.md                  # domain glossary
├── .pre-commit-config.yaml          # pre-commit hooks
├── renovate.json                    # dependency bot config
├── justfile                         # primary task runner
├── Makefile                         # equivalent make targets
├── go.mod / go.sum                  # module: github.com/jellydn/mole-tui (Go 1.26.4)
├── AGENTS.md                        # agent-facing project conventions
├── README.md                        # user-facing docs (TUI + sidecar usage)
└── LICENSE                          # MIT
```

## Key Locations

| Concern | Location |
|---------|----------|
| TUI binary entrypoint | `cmd/mole-tui/main.go` |
| Sidecar binary entrypoint | `cmd/mole-sidecar/main.go` |
| mo subprocess seam + resolution | `internal/mo/mo.go` |
| mo stub for tests | `internal/mo/motest/motest.go` |
| Root UI model & all 5 screens | `internal/ui/model.go` |
| Scan/cleanup lifecycle controller | `internal/ui/operations.go` |
| All styling | `internal/ui/styles.go` |
| Dry-run output parser + domain types | `internal/scanner/scanner.go` |
| Cleanup session / stream / summary | `internal/cleanup/cleanup.go` |
| JSON-RPC transport | `internal/jsonrpc/jsonrpc.go` |
| Parser fixtures | `internal/scanner/testdata/` |
| Architecture decisions | `.planning/adr/` (001–016) |
| PRD / requirements | `tasks/prd-mole-tui.md` |
| Build automation | `justfile`, `Makefile` |
| Autonomous dev loop | `ralph/` |

## Naming Conventions

- **Packages**: single-word lowercase — `mo`, `scanner`, `cleanup`, `ui`, `jsonrpc`, `main` (+ `motest` test stub). Domain types live where they're used; no shared types package.
- **Entrypoint layout**: `cmd/<binary>/main.go` (standard Go layout).
- **Seam naming**: the injectable subprocess boundary is `mo.Runner`; the test double lives in a `motest` sub-package (two adapters = the seam is real).
- **Test fixtures**: `internal/<pkg>/testdata/` with descriptive kebab-case names (`mo-clean-dryrun-real.txt`, `error-output.txt`).
- **Tea messages**: `<noun>Msg` suffix — `scanCompleteMsg`, `cleanupStreamMsg`, `cleanupCompleteMsg`, `scanCancelledMsg`.
- **Handlers/views**: `handle<Screen>Key`, `<screen>View` — one of each per screen.
- **Key maps**: per-screen `KeyMap` struct with named `key.Binding` fields, grouped in package-level vars (`dashboardKeys`, `confirmKeys`, `logKeys`, `loadingKeys`).
- **Styles**: `<purpose>Style` vars in `styles.go` (`titleStyle`, `bannerStyle`, `errorBannerStyle`, …) + `*Fg` adaptive color vars (`normalFg`, `accentFg`, …).
- **JSON-RPC**: wire methods/events are lowercase dotted (`scan.start`, `cleanup.line`); error codes are named constants in `internal/jsonrpc` (`CodeParseError`, `CodeMethodNotFound`, …).

## Build Output

- `bin/mole-tui` (gitignored) — always built to `bin/`; `go install` to `$GOBIN` via `just install` / `make install`.
- `bin/mole-sidecar` — built manually via `go build -o ./bin/mole-sidecar ./cmd/mole-sidecar` (no `just`/`make` target).
