# Testing Patterns

**Analysis Date:** 2026-06-20

## Test Framework

**Runner:**
- Go's standard `testing` package via `go test ./...`; test files import `testing` directly in `internal/scanner/scanner_test.go` and `internal/cleanup/cleanup_test.go`.
- Module dependencies in `go.mod` are runtime Charm libraries (`charm.land/bubbles/v2`, `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`); no separate assertion framework is listed in `go.mod`.
- Config: no dedicated Go test config file. Commands are defined in `justfile`, `Makefile`, `.pre-commit-config.yaml`, and documented in `AGENTS.md`.

**Assertion Library:**
- Standard library assertions using `if got != want { t.Errorf(...) }`, `t.Fatal(...)`, and `t.Fatalf(...)` in `internal/scanner/scanner_test.go` and `internal/cleanup/cleanup_test.go`.

**Run Commands:**
```bash
go test ./...      # Run all tests, as used by `justfile`, `Makefile`, and `AGENTS.md`
just test          # Run all tests via just
make test          # Run all tests via make
just ci            # fmt → vet → test → build
```

```make
# `justfile`
# Run all unit tests.
test:
    go test ./...

# CI: fmt → vet → test → build.
ci: fmt vet test build
```

## Test File Organization

**Location:**
- Tests are co-located with package code: `internal/scanner/scanner_test.go` tests `internal/scanner/scanner.go`, and `internal/cleanup/cleanup_test.go` tests `internal/cleanup/cleanup.go`.
- Fixtures are package-local under `internal/scanner/testdata/`.
- There are currently no checked-in UI snapshot tests or e2e tests under `internal/ui/` or `cmd/`; `AGENTS.md` and `tasks/prd-mole-tui.md` describe these as expected project patterns.

**Naming:**
- Test functions use `TestXxx` names: `TestParse`, `TestStripANSI`, `TestParseSize`, `TestParseSummary`, and `TestRunDryRun` in `internal/scanner/scanner_test.go` and `internal/cleanup/cleanup_test.go`.
- Table-driven subtests use descriptive `name` fields and `t.Run(tt.name, ...)` in `internal/scanner/scanner_test.go`.

**Structure:**
```
internal/
  cleanup/
    cleanup.go
    cleanup_test.go
  scanner/
    scanner.go
    scanner_test.go
    testdata/
      empty.txt
      error-output.txt
      mo-clean-dryrun-real.txt
```

## Test Structure

**Suite Organization:**
```go
// `internal/scanner/scanner_test.go`
func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		fixture     string
		wantSects   int
		wantSkipped bool
		wantBytes   float64 // approximate — we just check > 0 for real fixture
		checkFunc   func(t *testing.T, result ScanResult)
	}{
		{
			name:      "real fixture has sections",
			fixture:   "testdata/mo-clean-dryrun-real.txt",
			wantSects: 2, // User essentials + App caches (truncated by timeout)
			checkFunc: func(t *testing.T, r ScanResult) {
				if r.Summary.SystemCachesSkipped != true {
					t.Error("expected SystemCachesSkipped=true")
				}
				if r.Summary.FreeSpace == "" {
					t.Error("expected FreeSpace to be parsed")
				}
				t.Logf("FreeSpace: %s, TotalReclaimable: %.0f bytes", r.Summary.FreeSpace, r.Summary.TotalReclaimable)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.fixture)
			if err != nil {
				t.Fatalf("failed to read fixture %s: %v", tt.fixture, err)
			}
			result := Parse(string(data))
			if tt.wantSects >= 0 && len(result.Sections) < tt.wantSects {
				t.Errorf("expected at least %d sections, got %d", tt.wantSects, len(result.Sections))
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}
```

**Patterns:**
- Setup is lightweight and local to each test: `context.Background()` and `bytes.Buffer` in `internal/cleanup/cleanup_test.go`, and `os.ReadFile` for fixtures in `internal/scanner/scanner_test.go`.
- Teardown is not needed in current tests because they do not create persistent files, start servers, or modify process-global state.
- Assertions use direct comparisons plus clear failure messages that include got/want values.

```go
// `internal/cleanup/cleanup_test.go`
func TestRunDryRun(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	result, err := Run(ctx, Options{DryRun: true}, &buf)
	if err != nil {
		t.Fatalf("Run with DryRun failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if !strings.Contains(buf.String(), "Dry run complete") {
		t.Errorf("expected dry-run message in output, got %q", buf.String())
	}
	if result.FreedText == "" {
		t.Errorf("expected FreedText to be set")
	}
}
```

## Mocking

**Framework:** None currently used.

**Patterns:**
```go
// `internal/cleanup/cleanup.go`
if opts.DryRun {
	msg := "Dry run complete — no files were modified\n"
	if _, err := io.WriteString(writer, msg); err != nil {
		return Result{}, fmt.Errorf("write dry-run message: %w", err)
	}
	return Result{
		ExitCode:  0,
		Stdout:    msg,
		FreedText: "Dry run complete — no files were modified",
	}, nil
}
```

**What to Mock:**
- Current unit tests avoid the real `mo` binary by testing pure parser functions (`Parse`, `ParseSize`, `ParseSummary`) and the `cleanup.Run` dry-run branch in `internal/scanner/scanner_test.go` and `internal/cleanup/cleanup_test.go`.
- The planned e2e strategy is to drive the TUI against a stub `mo` shim; this is documented in `AGENTS.md` and `tasks/prd-mole-tui.md`, but no e2e test file or shim is currently checked in.

```markdown
<!-- `AGENTS.md` -->
- One end-to-end test drives the TUI against a stub `mo` shim.
```

**What NOT to Mock:**
- Parser unit tests use real captured Mole output fixtures rather than mocking the parser input shape; see `internal/scanner/testdata/mo-clean-dryrun-real.txt`.
- The actual `scanner.Scan` and non-dry-run `cleanup.Run` subprocess paths are not exercised by the current unit tests, avoiding accidental cleanup or dependency on a local `mo` install.

## Fixtures and Factories

**Test Data:**
```text
# `internal/scanner/testdata/mo-clean-dryrun-real.txt`
Clean Your Mac

Dry Run Mode, Preview only, no deletions

◎ System caches need sudo, run sudo -v && mo clean --dry-run for full preview

⚙ Apple Silicon | Free space: 105.27GB
✓ Whitelist: 21 core patterns active

➤ User essentials
  → User app cache 100 items, 481.6MB dry
  → User app logs 16 items, 4KB dry
  → Darwin user cache files, 241 old items, 331KB dry
  ✓ Trash · already empty

➤ App caches
  → Media analysis temp files, 0B dry
  → Geod temp files, 0B dry
  → Maps geo tile cache 2 items, 1KB dry
  → Apple Media Services cache 2 items, 119KB dry
```

**Location:**
- `internal/scanner/testdata/mo-clean-dryrun-real.txt` is the captured real dry-run fixture.
- `internal/scanner/testdata/empty.txt` covers empty/no-clean output.
- `internal/scanner/testdata/error-output.txt` covers error-shaped output.
- Inline string fixtures cover focused parser cases such as ANSI stripping, reclaimable size aggregation, sudo banner detection, and whitelist sub-items in `internal/scanner/scanner_test.go`.

```go
// `internal/scanner/scanner_test.go`
func TestTotalReclaimable(t *testing.T) {
	input := `➤ A
  → item, 100.0MB dry
  → other, 50.0MB dry
➤ B
  → thing, 2.0GB dry`
	result := Parse(input)
	expected := (100.0 * 1024 * 1024) + (50.0 * 1024 * 1024) + (2.0 * 1024 * 1024 * 1024)
	if result.Summary.TotalReclaimable != expected {
		t.Errorf("expected %.0f bytes, got %.0f", expected, result.Summary.TotalReclaimable)
	}
}
```

## Coverage

**Requirements:** None enforced in the current repository. `AGENTS.md`, `justfile`, `Makefile`, and `.pre-commit-config.yaml` require tests to pass, but they do not set a coverage threshold.

**View Coverage:**
```bash
go test -cover ./...          # Ad hoc package coverage summary
go test -coverprofile=cover.out ./... && go tool cover -html=cover.out
```

## Test Types

**Unit Tests:**
- Parser unit tests in `internal/scanner/scanner_test.go` cover ANSI stripping, section naming, total reclaimable size parsing, no-size lines, sudo banner detection, and whitelist sub-items.
- Cleanup unit tests in `internal/cleanup/cleanup_test.go` cover the dry-run branch and summary parsing.

```go
// `internal/cleanup/cleanup_test.go`
func TestParseSummary(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Total freed: 22.8 GB", "Total freed: 22.8 GB"},
		{"Cleaned: 1.5GB of data", "Total freed: 1.5GB"},
		{"Saved 500MB", "Total freed: 500MB"},
		{"Nothing to report", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := ParseSummary(tt.input)
		if got != tt.want {
			t.Errorf("ParseSummary(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
```

**Integration Tests:**
- No current checked-in integration tests directly execute `mo clean --dry-run` or `mo clean`. The production integration points are `exec.CommandContext(ctx, "mo", "clean", "--dry-run")` in `internal/scanner/scanner.go` and `exec.CommandContext(ctx, "mo", "clean")` in `internal/cleanup/cleanup.go`.

**E2E Tests:**
- Not currently implemented in checked-in Go test files. The required/desired pattern is documented as an e2e test with a stub `mo` shim in `AGENTS.md` and `tasks/prd-mole-tui.md`.
- Screen golden testing with `teatest` or hand-rolled snapshots is also documented in `AGENTS.md`, but no `internal/ui/*_test.go`, `testdata/*.golden`, or `teatest` import exists in the current `cmd/` and `internal/` Go files.

```markdown
<!-- `AGENTS.md` -->
- Screen rendering tested with golden-file snapshots via `teatest` or hand-rolled.
- One end-to-end test drives the TUI against a stub `mo` shim.
```

## Common Patterns

**Async Testing:**
```go
// `internal/ui/model.go`
// Current async behaviour is implemented through Bubble Tea commands; no async tests are checked in yet.
func cleanupCmd(dryRun bool) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		result, err := cleanup.Run(context.Background(), cleanup.Options{DryRun: dryRun}, &buf)
		if err != nil && result.Stdout == "" && result.Stderr == "" {
			// The error occurred before we got any output
			buf.WriteString(fmt.Sprintf("Error: %s\n", err))
		}
		result.Stdout = buf.String()
		return cleanupCompleteMsg{result: result, err: err}
	}
}
```

**Error Testing:**
```go
// `internal/scanner/scanner_test.go`
{
	name:      "error output produces no sections",
	fixture:   "testdata/error-output.txt",
	wantSects: 0,
	checkFunc: func(t *testing.T, r ScanResult) {
		if len(r.Sections) != 0 {
			t.Errorf("expected 0 sections for error output, got %d", len(r.Sections))
		}
	},
},
```

---

*Testing analysis: 2026-06-20*
