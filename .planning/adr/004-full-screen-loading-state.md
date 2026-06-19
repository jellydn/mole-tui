# ADR-004: Full-Screen Loading State During Initial Scan

## Status
Accepted

## Context
With discovery and scan merged (ADR-001), the TUI must run `mo clean --dry-run` before it can populate the dashboard. On the user's machine, this takes ~3 minutes. The PRD targets < 200ms startup to first paint.

Three options were considered:
1. Empty dashboard + spinner (confusing — user sees an empty list)
2. Full-screen loading state (clear expectation setting)
3. Streaming parse (complex, fragile against output format changes)

## Decision
Show a full-screen loading state with a spinner and status text (e.g., "Scanning disk… this may take a few minutes") until the scan completes. Then transition to the populated dashboard.

The < 200ms target is reinterpreted as: the TUI binary starts and renders the loading screen in < 200ms. The scan itself takes as long as Mole needs.

## Consequences
- The TUI has two distinct initial states: `Loading` and `Dashboard`
- The loading screen should show: app title, spinner, "Scanning…" message, and elapsed time
- No items or sizes are visible during loading
- Users see immediate feedback that the app is working
- The Bubble Tea model starts in a `Loading` state and transitions to `Dashboard` on scan completion
- If the scan fails, the loading screen transitions to an error state (not the dashboard)
