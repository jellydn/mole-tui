# STRUCTURE.md — Directory Layout & Structure

> Mapped fresh on 2026-08-04.

## Directory Tree

```
.
├── cmd/
│   └── mole-tui/
│       └── main.go                  # binary entrypoint (48 ln) — flags, mo lookup, program run
├── internal/
│   ├── scanner/                     # scan side: invoke + parse `mo clean --dry-run`
│   │   ├── scanner.go               # Scan(), Parse(), ParseSize(), types (175 ln)
│   │   ├── scanner_test.go          # table-driven parser tests (193 ln)
│   │   └── testdata/
│   │       ├── mo-clean-dryrun-real.txt   # captured real mo dry-run fixture
│   │       ├── empty.txt                  # synthetic edge case
│   │       └── error-output.txt           # synthetic edge case (mo missing)
│   ├── cleanup/                     # cleanup side: invoke + stream `mo clean`
│   │   ├── cleanup.go               # Run(), ParseSummary(), Options/Result (121 ln)
│   │   └── cleanup_test.go          # dry-run short-circuit + summary parsing (45 ln)
│   └── ui/                          # Bubble Tea model, views, keybindings, styles
│       ├── model.go                 # Model, 5 screens, all handlers + views (830 ln)
│       └── styles.go                # all Lip Gloss v2 styles + adaptive palette (101 ln)
├── tasks/
│   └── prd-mole-tui.md              # product requirements (post-grilling revision)
├── ralph/                           # autonomous agent loop (dev workflow)
│   ├── ralph.sh                     # loop driver (314 ln)
│   ├── prompt.md, prd.json          # ralph inputs
│   ├── prompt-{pi,codex,copilot,opencode,cmd,agy,amp,mino}.md  # per-backend prompts
│   └── progress.txt                 # story-by-story progress log
├── .planning/
│   ├── adr/                         # 12 accepted ADRs (001–012)
│   ├── codebase/                    # ← this codebase map (7 docs)
│   └── glossary.md                  # domain glossary
├── .pre-commit-config.yaml          # pre-commit hooks
├── renovate.json                    # dependency bot config
├── justfile                         # primary task runner
├── Makefile                         # equivalent make targets
├── go.mod / go.sum                  # module: github.com/jellydn/mole-tui (Go 1.26.4)
├── AGENTS.md                        # agent-facing project conventions
├── README.md                        # user-facing docs (partly stale — see CONCERNS)
└── LICENSE                          # MIT
```

## Key Locations

| Concern | Location |
|---------|----------|
| Binary entrypoint | `cmd/mole-tui/main.go` |
| Root UI model & all 5 screens | `internal/ui/model.go` |
| All styling | `internal/ui/styles.go` |
| Dry-run output parser + domain types | `internal/scanner/scanner.go` |
| Cleanup runner + summary parser | `internal/cleanup/cleanup.go` |
| Parser fixtures | `internal/scanner/testdata/` |
| Architecture decisions | `.planning/adr/` |
| PRD / requirements | `tasks/prd-mole-tui.md` |
| Build automation | `justfile`, `Makefile` |
| Autonomous dev loop | `ralph/` |

## Naming Conventions

- **Packages**: single-word lowercase — `scanner`, `cleanup`, `ui`, `main`. Domain types live where they're used; no shared types package.
- **Entrypoint layout**: `cmd/<binary>/main.go` (standard Go layout).
- **Test fixtures**: `internal/<pkg>/testdata/` with descriptive kebab-case names (`mo-clean-dryrun-real.txt`, `error-output.txt`).
- **Tea messages**: `<noun>Msg` suffix — `scanCompleteMsg`, `cleanupStreamMsg`, `cleanupCompleteMsg`, `scanCancelledMsg`.
- **Handlers/views**: `handle<Screen>Key`, `<screen>View` — one of each per screen.
- **Key maps**: per-screen `KeyMap` struct with named `key.Binding` fields, grouped in package-level vars (`dashboardKeys`, `confirmKeys`, `logKeys`, `loadingKeys`).
- **Styles**: `<purpose>Style` vars in `styles.go` (`titleStyle`, `bannerStyle`, `errorBannerStyle`, …) + `*Fg` adaptive color vars (`normalFg`, `accentFg`, …).

## Build Output

- `bin/mole-tui` (gitignored) — always built to `bin/`; `go install` to `$GOBIN` via `just install` / `make install`.
