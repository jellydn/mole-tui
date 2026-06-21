# Coding Conventions

**Analysis Date:** 2026-06-20

## Naming Patterns

**Files:**
- Command entrypoints live under `cmd/<binary>/main.go`; the binary entrypoint is `cmd/mole-tui/main.go`.
- Package files use short, package-matching names: `internal/scanner/scanner.go`, `internal/cleanup/cleanup.go`, `internal/ui/model.go`, and `internal/ui/styles.go`.
- Tests are co-located and named with `_test.go`, for example `internal/scanner/scanner_test.go` and `internal/cleanup/cleanup_test.go`.
- Fixtures live in `testdata/` below the package they support, for example `internal/scanner/testdata/mo-clean-dryrun-real.txt`.

**Functions:**
- Exported functions use PascalCase when they are part of the package API: `NewModel`, `Parse`, `ParseSize`, `Scan`, `Run`, and `ParseSummary` in `internal/ui/model.go`, `internal/scanner/scanner.go`, and `internal/cleanup/cleanup.go`.
- Unexported helpers use lower camelCase: `stripANSI`, `styleItemLine`, `visibleLineCount`, `handleDashboardKey`, and `formatBytes` in `internal/scanner/scanner.go` and `internal/ui/model.go`.
- Bubble Tea command helpers are named with a `Cmd` suffix: `scanCmd` and `cleanupCmd` in `internal/ui/model.go`.

```go
// `internal/ui/model.go`
func (m *Model) Init() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.scanCancel = cancel
	return tea.Batch(m.spinner.Tick, m.scanCmd(ctx, false))
}

// scanCmd returns a tea.Cmd that runs `mo clean --dry-run` with the given ctx.
func (m *Model) scanCmd(ctx context.Context, sudo bool) tea.Cmd {
	return func() tea.Msg {
		result, err := scanner.Scan(ctx, sudo)
		if errors.Is(err, context.Canceled) {
			return scanCancelledMsg{}
		}
		return scanCompleteMsg{result: result, err: err}
	}
}
```

**Variables:**
- Package-level regular expressions use a `re` prefix: `reANSI`, `reSection`, `reSize`, `reFreeSpace`, `reSystemCachesSkipped`, and `reFreed` in `internal/scanner/scanner.go` and `internal/cleanup/cleanup.go`.
- Short local names are idiomatic for narrow scopes (`m`, `s`, `ctx`, `cmd`, `err`, `tt`, `got`, `want`) in `internal/ui/model.go`, `internal/scanner/scanner_test.go`, and `internal/cleanup/cleanup_test.go`.
- UI styles use lower camelCase names ending in `Style`: `titleStyle`, `bannerStyle`, `errorBannerStyle`, `spinnerStyle`, and `cursorStyle` in `internal/ui/styles.go`.

```go
// `internal/scanner/scanner.go`
var reANSI = regexp.MustCompile(`\033\[[0-9;]*[A-Za-z]`)
var reSection = regexp.MustCompile(`^➤\s+(.+)$`)
var reSize = regexp.MustCompile(`([0-9.]+)\s*(KB|MB|GB|B)\s+dry`)
```

**Types:**
- Exported data carriers use clear noun names with field comments where needed: `Section`, `ScanSummary`, `ScanResult`, `Options`, `Result`, and `Model` in `internal/scanner/scanner.go`, `internal/cleanup/cleanup.go`, and `internal/ui/model.go`.
- Internal enum-like state uses an unexported typed alias plus unexported constants: `Screen`, `screenLoading`, `screenDashboard`, `screenConfirm`, `screenLog`, and `screenHelp` in `internal/ui/model.go`.
- Bubble Tea messages are unexported structs with a `Msg` suffix: `scanCompleteMsg`, `scanCancelledMsg`, and `cleanupCompleteMsg` in `internal/ui/model.go`.

```go
// `internal/ui/model.go`
// Screen represents the current UI screen.
type Screen int

const (
	screenLoading Screen = iota
	screenDashboard
	screenConfirm
	screenLog
	screenHelp
)

type scanCompleteMsg struct {
	result scanner.ScanResult
	err    error
}
```

## Code Style

**Formatting:**
- Go source is expected to be `gofmt -s` clean. The `fmt` target in `justfile` fails if `gofmt -l -s .` prints anything, and `Makefile` uses the same check through `/tmp/mole-tui-fmt.out`.

```make
# `Makefile`
fmt: ## Check gofmt (fails if any file is unformatted).
	@gofmt -l -s . | tee /tmp/mole-tui-fmt.out
	@test ! -s /tmp/mole-tui-fmt.out
	@rm -f /tmp/mole-tui-fmt.out
```

**Linting:**
- `go vet ./...` is the repository lint/typecheck pass, exposed by both `justfile` and `Makefile`.
- CI order is documented as `fmt` → `vet` → `test` → `build` in `AGENTS.md`; `just ci` encodes that order in `justfile`.
- Pre-commit runs whitespace/YAML/JSON/merge-conflict checks plus Go hooks (`go-fmt`, `go-vet`, `go-build`, `go-unit-tests`) in `.pre-commit-config.yaml`.

```yaml
# `.pre-commit-config.yaml`
  - repo: https://github.com/dnephin/pre-commit-golang
    rev: v0.5.1
    hooks:
      - id: go-fmt
      - id: go-vet
      - id: go-build
      - id: go-unit-tests
```

## Import Organization

**Order:**
1. Standard library imports first.
2. Third-party Charm libraries second.
3. Internal module imports last.

```go
// `internal/ui/model.go`
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jellydn/mole-tui/internal/cleanup"
	"github.com/jellydn/mole-tui/internal/scanner"
)
```

**Path Aliases:**
- Bubble Tea is imported as `tea`, matching common Bubble Tea idiom, in `cmd/mole-tui/main.go` and `internal/ui/model.go`.
- There are no TypeScript-style path aliases; internal imports use full module paths such as `github.com/jellydn/mole-tui/internal/ui` in `cmd/mole-tui/main.go`.

## Error Handling

**Patterns:**
- Functions return `(Result, error)` or `(ScanResult, error)` instead of panicking. Examples are `Scan` in `internal/scanner/scanner.go` and `Run` in `internal/cleanup/cleanup.go`.
- Errors from subprocess startup and writer failures are wrapped with `%w` where callers may need the original error; command failures that surface stderr are wrapped with contextual text in `internal/scanner/scanner.go`.
- Context cancellation is checked before wrapping subprocess errors, and the UI converts it into a typed Bubble Tea message in `internal/ui/model.go`.
- `main` handles fatal pre-TUI errors by printing to stderr and exiting non-zero in `cmd/mole-tui/main.go`.

```go
// `internal/scanner/scanner.go`
if err := cmd.Run(); err != nil {
	// Check for context cancellation before wrapping
	if ctx.Err() != nil {
		return ScanResult{}, ctx.Err()
	}
	errOutput := strings.TrimSpace(stderr.String())
	if errOutput == "" {
		errOutput = err.Error()
	}
	return ScanResult{}, fmt.Errorf("mo clean --dry-run failed: %s", errOutput)
}
```

```go
// `cmd/mole-tui/main.go`
if _, err := exec.LookPath("mo"); err != nil {
	fmt.Fprintf(os.Stderr, "\033[1;31mError:\033[0m mo is not on $PATH\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "mole-tui requires the \033[1mmo\033[0m CLI from Mole:\n")
	fmt.Fprintf(os.Stderr, "  brew install mo\n")
	fmt.Fprintf(os.Stderr, "  or visit https://github.com/tw93/mole\n")
	os.Exit(1)
}
```

## Logging

**Framework:** None.

**Patterns:**
- The CLI writes user-facing fatal messages to `os.Stderr` with `fmt.Fprintf` in `cmd/mole-tui/main.go`.
- The cleanup package streams subprocess stdout/stderr to an `io.Writer` supplied by the UI, rather than using a logging package, in `internal/cleanup/cleanup.go`.
- Tests may use `t.Logf` for diagnostic context, as in `internal/scanner/scanner_test.go`.

```go
// `internal/cleanup/cleanup.go`
stdout := io.MultiWriter(writer, &stdoutBuf)
cmd.Stdout = stdout
cmd.Stderr = &stderrBuf

// Also capture stderr to writer for live streaming
// exec.Cmd only supports one Stderr writer, so we tee manually
pr, pw := io.Pipe()
cmd.Stderr = pw
```

## Comments

**When to Comment:**
- Exported packages, types, and functions generally have Go doc comments in `internal/scanner/scanner.go`, `internal/cleanup/cleanup.go`, and `internal/ui/model.go`.
- Comments are used to explain non-obvious subprocess and UI behaviour, such as sudo command composition and stderr teeing.
- Visual section separators are used in the large UI model to group screen-specific code in `internal/ui/model.go`.

**JSDoc/TSDoc:**
- Not applicable. Use Go doc comments for exported Go symbols.

```go
// `internal/cleanup/cleanup.go`
// Run executes `mo clean` with the given options. Output is written to writer
// as it arrives (for live streaming). Returns a Result with the full output.
//
// When DryRun is true, no command is executed — the writer receives a canned
// message and the result indicates success.
func Run(ctx context.Context, opts Options, writer io.Writer) (Result, error) {
```

## Function Design

**Size:** Functions are small in parser/cleanup packages and larger in UI view/update handlers. `internal/scanner/scanner.go` keeps parsing helpers (`stripANSI`, `ParseSize`, `Parse`) separate from subprocess execution (`Scan`), while `internal/ui/model.go` groups one handler/view pair per screen.

**Parameters:** Context is the first parameter for subprocess-bound operations (`Scan(ctx context.Context, sudo bool)` and `Run(ctx context.Context, opts Options, writer io.Writer)` in `internal/scanner/scanner.go` and `internal/cleanup/cleanup.go`). Option structs are used once behaviour grows beyond a scalar (`cleanup.Options`).

**Return Values:** Operations return concrete result structs plus `error`; parser-only helpers return concrete values without error when best-effort parsing is acceptable.

```go
// `internal/cleanup/cleanup.go`
type Options struct {
	// DryRun short-circuits execution with a canned success message.
	DryRun bool
}

type Result struct {
	ExitCode  int
	Stdout    string
	Stderr    string
	FreedText string // best-effort summary, e.g. "Total freed: 22.8 GB" or ""
}

func Run(ctx context.Context, opts Options, writer io.Writer) (Result, error) {
```

## Module Design

**Exports:** Packages expose only the API needed across package boundaries. `internal/scanner/scanner.go` exports `Section`, `ScanSummary`, `ScanResult`, `ParseSize`, `Parse`, and `Scan`; `internal/cleanup/cleanup.go` exports `Options`, `Result`, `Run`, and `ParseSummary`; `internal/ui/model.go` exports `Model` and `NewModel`.

**Barrel Files:** Not used. Go packages are imported directly via module paths; there are no re-export aggregator files.

```go
// `AGENTS.md`
cmd/mole-tui/main.go          # binary entrypoint
internal/
  scanner/     # parse `mo clean --dry-run` → []Section & ScanSummary (ADR-001, ADR-012)
  cleanup/     # shell out to `mo clean`, stream output (ADR-002, ADR-007)
  ui/          # Bubble Tea models, views, keybindings (ADR-010)
```

---

*Convention analysis: 2026-06-20*
