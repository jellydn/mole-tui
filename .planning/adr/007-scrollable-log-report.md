# ADR-007: Scrollable Log Report with Parsed Summary

## Status
Accepted

## Context
The PRD (US-007) specified a before/after table per item. Computing "after" sizes requires re-running `mo clean --dry-run` (~3 min), which is too expensive as a post-cleanup step.

## Decision
The completion report shows:
1. The `mo clean` stdout/stderr as a **scrollable log pane** (streamed live during cleanup, then scrollable after)
2. A **summary line** at the bottom if a total freed amount can be parsed from `mo clean` output (e.g., "Total freed: 22.8 GB")
3. If no total can be parsed, show "Cleanup complete" without a bytes-freed number

No re-scan is performed automatically. The user can trigger a re-scan from the dashboard after returning.

## Consequences
- The `cleanup` package streams stdout/stderr to the UI in real-time
- The report screen is essentially the same live log pane, frozen after the process exits
- A lightweight parser attempts to extract a "total freed" number from the final output
- Per-item before/after comparison is dropped from v1
- Simpler implementation, no second scan required
- US-007 acceptance criteria simplified: log + optional total, not a per-item table
