# Codebase Structure

**Analysis Date:** 2026-06-20

## Directory Layout

```
mole-tui/
├── AGENTS.md # Project-specific agent/build guidance
├── CLAUDE.md # Claude/project notes
├── LICENSE # Project license
├── Makefile # Build, install, test, fmt, vet targets
├── README.md # User-facing project documentation
├── go.mod # Go module definition for github.com/jellydn/mole-tui
├── go.sum # Go dependency checksums
├── justfile # Just aliases mirroring Makefile tasks
├── renovate.json # Dependency automation config
├── cmd/
│   └── mole-tui/
│       └── main.go # Binary entrypoint
├── internal/
│   ├── cleanup/
│   │   ├── cleanup.go # mo clean execution, streaming, summary parsing
│   │   └── cleanup_test.go # cleanup dry-run and summary parser tests
│   ├── scanner/
│   │   ├── scanner.go # mo clean --dry-run execution and parsing
│   │   ├── scanner_test.go # scanner parser tests
│   │   └── testdata/
│   │       ├── empty.txt # Empty/minimal dry-run fixture
│   │       ├── error-output.txt # Error-output fixture
│   │       └── mo-clean-dryrun-real.txt # Captured real Mole dry-run fixture
│   └── ui/
│       ├── model.go # Bubble Tea Model/Update/View, screens, commands, keys
│       └── styles.go # Lip Gloss styles and adaptive colors
├── .planning/
│   ├── adr/
│   │   ├── 001-merge-discovery-and-scan.md # Discovery merged into scan
│   │   ├── 002-all-or-nothing-cleanup.md # mo clean has no per-item args
│   │   ├── 003-full-fidelity-dry-run-parsing.md # Initial parsing model decision
│   │   ├── 004-full-screen-loading-state.md # Loading-screen decision
│   │   ├── 005-non-sudo-by-default.md # Sudo is opt-in
│   │   ├── 006-drop-preview-pane.md # No v1 preview pane
│   │   ├── 007-scrollable-log-report.md # Log/report completion view
│   │   ├── 008-three-package-architecture.md # internal/scanner, cleanup, ui
│   │   ├── 009-minimal-keybindings.md # v1 keybinding inventory
│   │   ├── 010-five-screens.md # Loading, dashboard, confirm, log, help
│   │   ├── 011-tui-dry-run-flag.md # --dry-run / -n behavior
│   │   └── 012-hybrid-parsing.md # Structured sections plus raw item lines
│   ├── codebase/
│   │   ├── ARCHITECTURE.md # Generated architecture map
│   │   └── STRUCTURE.md # Generated structure map
│   └── glossary.md # Domain glossary
├── ralph/
│   ├── prd.json # Ralph autonomous-agent PRD data
│   ├── progress.txt # Ralph progress tracker
│   ├── prompt.md # Main Ralph prompt
│   ├── prompt-agy.md # Backend-specific Ralph prompt
│   ├── prompt-amp.md # Backend-specific Ralph prompt
│   ├── prompt-cmd.md # Backend-specific Ralph prompt
│   ├── prompt-codex.md # Backend-specific Ralph prompt
│   ├── prompt-copilot.md # Backend-specific Ralph prompt
│   ├── prompt-mino.md # Backend-specific Ralph prompt
│   ├── prompt-opencode.md # Backend-specific Ralph prompt
│   ├── prompt-pi.md # Backend-specific Ralph prompt
│   └── ralph.sh # Ralph iteration runner
└── tasks/
    └── prd-mole-tui.md # Product requirements and functional requirements
```

## Directory Purposes

**`cmd/`:**
- Purpose: Go command entrypoints.
- Contains: One command package under `cmd/mole-tui/`.
- Key files: `cmd/mole-tui/main.go`.

**`cmd/mole-tui/`:**
- Purpose: Buildable `main` package for the `mole-tui` binary.
- Contains: CLI flag parsing, version output, pre-TUI `mo` availability check, and Bubble Tea program startup.
- Key files: `cmd/mole-tui/main.go`.

**`internal/`:**
- Purpose: Private application packages; Go prevents imports from outside this module tree.
- Contains: The PRD/ADR-locked three-package architecture: scanner, cleanup, and UI.
- Key files: `internal/scanner/scanner.go`, `internal/cleanup/cleanup.go`, `internal/ui/model.go`.

**`internal/scanner/`:**
- Purpose: Own `mo clean --dry-run` invocation and dry-run output parsing.
- Contains: Scanner domain types, regex parsers, context-aware subprocess execution, table-driven tests, and fixtures.
- Key files: `internal/scanner/scanner.go`, `internal/scanner/scanner_test.go`, `internal/scanner/testdata/mo-clean-dryrun-real.txt`.

**`internal/scanner/testdata/`:**
- Purpose: Go test fixture directory for scanner parser inputs.
- Contains: Captured real Mole output and synthetic edge cases.
- Key files: `internal/scanner/testdata/mo-clean-dryrun-real.txt`, `internal/scanner/testdata/empty.txt`, `internal/scanner/testdata/error-output.txt`.

**`internal/cleanup/`:**
- Purpose: Own `mo clean` execution, dry-run simulation, stdout/stderr streaming, and cleanup summary parsing.
- Contains: Process boundary code and unit tests.
- Key files: `internal/cleanup/cleanup.go`, `internal/cleanup/cleanup_test.go`.

**`internal/ui/`:**
- Purpose: Own all Bubble Tea screens, state transitions, key handling, rendering, and styling.
- Contains: Root model, `tea.Msg`/`tea.Cmd` flow, five screen renderers/handlers, keybindings, viewports, spinner/help integration, and Lip Gloss styles.
- Key files: `internal/ui/model.go`, `internal/ui/styles.go`.

**`.planning/`:**
- Purpose: Project planning, decision history, generated codebase maps, and domain glossary.
- Contains: ADRs in `.planning/adr/`, generated mapping docs in `.planning/codebase/`, and `.planning/glossary.md`.
- Key files: `.planning/adr/008-three-package-architecture.md`, `.planning/adr/010-five-screens.md`, `.planning/adr/012-hybrid-parsing.md`, `.planning/glossary.md`.

**`.planning/adr/`:**
- Purpose: Architecture Decision Records explaining why the current codebase diverges from the original broader PRD.
- Contains: Twelve accepted ADRs covering scan/discovery merge, all-or-nothing cleanup, no preview pane, five screens, and hybrid parsing.
- Key files: `.planning/adr/001-merge-discovery-and-scan.md`, `.planning/adr/002-all-or-nothing-cleanup.md`, `.planning/adr/008-three-package-architecture.md`, `.planning/adr/012-hybrid-parsing.md`.

**`.planning/codebase/`:**
- Purpose: Generated codebase understanding documents.
- Contains: Architecture and structure maps.
- Key files: `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/STRUCTURE.md`.

**`ralph/`:**
- Purpose: Autonomous agent loop assets for Ralph.
- Contains: Ralph PRD JSON, progress tracking, backend-specific prompts, and runner script.
- Key files: `ralph/ralph.sh`, `ralph/prd.json`, `ralph/progress.txt`, `ralph/prompt.md`.

**`tasks/`:**
- Purpose: Human-readable planning/task documents.
- Contains: The revised Mole TUI PRD.
- Key files: `tasks/prd-mole-tui.md`.

**`bin/`:**
- Purpose: Local build output directory referenced by project guidance.
- Contains: Built `mole-tui` binary when using build tasks.
- Key files: `bin/mole-tui` when generated; no source files are expected here.

## Key File Locations

**Entry Points:**
- `cmd/mole-tui/main.go`: Binary entrypoint; parses flags, validates `mo`, starts the Bubble Tea app.
- `internal/ui/model.go`: Bubble Tea runtime entry through `NewModel`, `Init`, `Update`, and `View`.

**Configuration:**
- `go.mod`: Module path `github.com/jellydn/mole-tui` and Go dependencies.
- `go.sum`: Dependency checksum lockfile.
- `Makefile`: Build/test/install/fmt/vet command definitions.
- `justfile`: Short aliases such as `just build`, `just test`, and `just ci`.
- `renovate.json`: Renovate dependency update configuration.
- `AGENTS.md`: Project-specific agent instructions and required test/build order.

**Core Logic:**
- `internal/scanner/scanner.go`: `mo clean --dry-run` subprocess invocation and hybrid parser.
- `internal/cleanup/cleanup.go`: `mo clean` subprocess invocation, output streaming, dry-run simulation, summary parser.
- `internal/ui/model.go`: TUI state machine, commands, messages, key handlers, screen views, and helper functions.
- `internal/ui/styles.go`: UI color palette and Lip Gloss styles.

**Testing:**
- `internal/scanner/scanner_test.go`: Parser tests for real and synthetic Mole dry-run output.
- `internal/scanner/testdata/`: Scanner fixture inputs.
- `internal/cleanup/cleanup_test.go`: Dry-run and summary parser tests for cleanup package.

**Planning / Requirements:**
- `tasks/prd-mole-tui.md`: Product requirements, functional requirements, non-goals, and technical considerations.
- `.planning/adr/`: Accepted architecture decisions.
- `.planning/glossary.md`: Shared domain vocabulary.

## Naming Conventions

**Files:**
- Go implementation files use concise package-oriented names: `scanner.go` in `internal/scanner/`, `cleanup.go` in `internal/cleanup/`, `model.go` and `styles.go` in `internal/ui/`.
- Go tests use `_test.go` colocated with the package under test: `internal/scanner/scanner_test.go`, `internal/cleanup/cleanup_test.go`.
- Scanner fixtures live under Go's conventional `testdata/` directory: `internal/scanner/testdata/mo-clean-dryrun-real.txt`.
- ADRs use zero-padded numeric prefixes and kebab-case titles: `.planning/adr/008-three-package-architecture.md`.
- Ralph prompts use `prompt-<backend>.md` names: `ralph/prompt-codex.md`, `ralph/prompt-amp.md`.

**Directories:**
- Public command packages live under `cmd/<binary>/`: `cmd/mole-tui/`.
- Private app packages live under `internal/<concern>/`: `internal/scanner/`, `internal/cleanup/`, `internal/ui/`.
- Planning artifacts live under dot-prefixed `.planning/` and task documents under `tasks/`.
- Generated/local build output belongs in `bin/`, not in `cmd/` or `internal/`.

## Where to Add New Code

**New scan parsing behavior:**
- Primary code: `internal/scanner/scanner.go`.
- Tests: `internal/scanner/scanner_test.go` and fixtures in `internal/scanner/testdata/`.

**New cleanup execution behavior:**
- Primary code: `internal/cleanup/cleanup.go`.
- Tests: `internal/cleanup/cleanup_test.go`; add `internal/cleanup/testdata/` if fixture-based cleanup output tests become necessary.

**New TUI screen or key interaction:**
- Primary code: `internal/ui/model.go`.
- Styles: `internal/ui/styles.go`.
- Tests: Add UI tests near `internal/ui/` if screen snapshot or Bubble Tea model tests are introduced.

**New command-line flag or startup validation:**
- Primary code: `cmd/mole-tui/main.go`.
- Tests: Add command-level tests if argument behavior grows beyond simple flags.

**New Mole command integration beyond `mo clean`:**
- Primary code: Add a new package under `internal/<command>/` only if it is a distinct process/domain boundary; otherwise extend `internal/scanner/` or `internal/cleanup/` if it directly belongs to scan/clean.
- UI integration: `internal/ui/model.go`.
- Planning: Update or add ADRs under `.planning/adr/` if it changes the PRD-locked v1 scope from `tasks/prd-mole-tui.md`.

**Utilities:**
- Shared parser helpers for scan output: keep in `internal/scanner/scanner.go` until they need extraction.
- Shared cleanup output helpers: keep in `internal/cleanup/cleanup.go` until they need extraction.
- Shared UI rendering/style helpers: keep in `internal/ui/model.go` or `internal/ui/styles.go`; avoid creating a generic utilities package unless multiple packages need the same code.

## Special Directories

**`.planning/`:**
- Purpose: Durable planning context, ADRs, glossary, and generated codebase maps.
- Generated: Partially; `.planning/codebase/ARCHITECTURE.md` and `.planning/codebase/STRUCTURE.md` are generated analysis docs, while ADRs and glossary are curated.
- Committed: Yes, expected to be committed as project documentation.

**`.planning/adr/`:**
- Purpose: Accepted Architecture Decision Records that constrain and explain implementation choices.
- Generated: No.
- Committed: Yes.

**`.planning/codebase/`:**
- Purpose: Codebase mapping outputs for architecture and structure.
- Generated: Yes.
- Committed: Yes, if the project tracks planning outputs.

**`ralph/`:**
- Purpose: Autonomous implementation loop inputs and progress tracking.
- Generated: Partially; `ralph/progress.txt` is updated by the Ralph loop, prompts and `ralph/prd.json` are inputs.
- Committed: Yes, based on its presence in the repository tree.

**`tasks/`:**
- Purpose: Product and task planning documents.
- Generated: No.
- Committed: Yes.

**`internal/*/testdata/`:**
- Purpose: Test-only fixtures automatically supported by Go tooling.
- Generated: No.
- Committed: Yes.

**`bin/`:**
- Purpose: Local binary output such as `bin/mole-tui` from build tasks.
- Generated: Yes.
- Committed: No; project guidance describes it as build output.

---

*Structure analysis: 2026-06-20*
