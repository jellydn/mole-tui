# ADR-011: Add `--dry-run` Flag to TUI in v1

## Status
Accepted

## Context
OQ-5 in the PRD deferred a `mole-tui --dry-run` flag to v2. However, during development and testing, the ability to exercise the full UI flow without invoking `mo clean` is essential for safety and iteration speed.

## Decision
Add `--dry-run` (`-n`) flag to `mole-tui` in v1.

When `--dry-run` is active:
- `mo clean --dry-run` still runs normally for scanning (it's already non-destructive)
- The confirmation modal clearly shows "[DRY RUN]" in the command preview
- Instead of running `mo clean`, the TUI simulates a successful cleanup with a canned "Dry run complete — no files were modified" message
- Exit code is 0

## Consequences
- Safe development workflow — developers can test the full flow without risk
- Useful for demos and screenshots
- The `cleanup` package gains a `DryRun bool` option that short-circuits execution
- The confirmation modal renders the command differently in dry-run mode
- OQ-5 is resolved (pulled from v2 to v1)
