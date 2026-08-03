# TESTING.md — Testing

> Mapped fresh on 2026-08-04.

## Test Runner

- Single entrypoint: **`go test ./...`** — run via `just test` / `make test`.
- Pre-commit runs `go-unit-tests` on every commit; `just ci` runs tests as the third stage (fmt → vet → **test** → build).
- No coverage thresholds configured; `go test -cover` works out of the box if needed.

## What's Tested

### `internal/scanner` — `scanner_test.go` (193 lines)

- **`TestParse`** — table-driven over 3 fixtures:
  - `testdata/mo-clean-dryrun-real.txt` — captured real `mo clean --dry-run` output (expects ≥2 sections, `SystemCachesSkipped=true`, parsed `FreeSpace`, positive reclaimable).
  - `testdata/empty.txt` — empty output (expects 0 sections, parsed `FreeSpace: 200.00GB`, 0 reclaimable).
  - `testdata/error-output.txt` — "mo is not installed" style output (expects 0 sections).
- **`TestStripANSI`** — ANSI escape stripping.
- **`TestParseSize`** — KB/MB/GB/B → bytes table.
- **`TestParseReclaimableSize`** — size regex aggregation across `dry` lines.
- **`TestParseNoSizeLines`** — sections with zero size lines.
- **`TestANSICleanYourMac`** — realistic ANSI-styled mo output end-to-end.
- **`TestSectionName`** — multi-section header parsing.
- **`TestTotalReclaimable`** — exact byte aggregation across sections.
- **`TestSystemCachesSkipped`** — sudo-needed banner detection.
- **`TestWhitelistNotSection`** — `↳` sub-items don't create sections.

### `internal/cleanup` — `cleanup_test.go` (45 lines)

- **`TestRunDryRun`** — `Options{DryRun: true}` short-circuits: exit 0, canned message, `FreedText` set, no subprocess.
- **`TestParseSummary`** — table-driven: "Total freed: 22.8 GB", "Cleaned: 1.5GB of data", "Saved 500MB" → normalized `Total freed: …`; "Nothing to report"/"" → `""`.

## Fixture Strategy

- Fixtures live in `internal/<pkg>/testdata/` (ADR-008).
- One **real captured** fixture (`mo-clean-dryrun-real.txt`) anchors parser realism; synthetic fixtures cover edge cases (empty, error).
- Parser tests use `os.ReadFile` on the fixture path relative to the package.

## Coverage Gaps / Discrepancies

> ⚠️ Docs (AGENTS.md, PRD §7) claim the following tests that **do not exist** in the repo — only 2 test files are present (verified via glob for `**/*_test.go`):

- ❌ No golden-file / `teatest` snapshot tests for screen rendering (PRD US-002, AGENTS.md "Screen rendering tested with golden-file snapshots").
- ❌ No end-to-end test driving the TUI against a stub `mo` shim (AGENTS.md "One end-to-end test…", PRD §7).
- ❌ No `internal/cleanup/testdata/` directory despite ADR-008 mentioning it.
- ❌ No test for `scanner.Scan()`'s subprocess path (context cancellation, sudo variant) — only `Parse` is exercised.
- ❌ No test for `cleanup.Run()`'s real subprocess path (streaming, exit codes, stderr teeing) — only the dry-run branch.
- ❌ No UI package tests at all (`internal/ui` has no `_test.go`).

These are prime candidates for the next testing investment — see CONCERNS.md.
