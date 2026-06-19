# ADR-002: All-or-Nothing Cleanup (No Per-Item Selection)

## Status
Accepted

## Context
The PRD assumed `mo clean <item1> <item2> …` syntax for selective cleanup. Testing reveals `mo clean` only accepts `[OPTIONS]` — no positional item arguments. The CLI cleans everything in its scope.

The `--whitelist` mechanism allows _excluding_ certain paths, but operates on raw filesystem paths, not on the high-level item categories shown in `--dry-run`. Using whitelist manipulation to simulate selective cleanup would be fragile, surprising to users, and risk corrupting user-configured whitelists.

## Decision
Accept that v1 cleanup is all-or-nothing. The TUI runs `mo clean` without item arguments. The multi-select UI shifts from "select what to clean" to "preview what will be cleaned" — the TUI becomes a preview + confirm wrapper.

## Consequences
- US-003 (multi-select) is reduced to a preview/information role — the checkbox model is removed or becomes cosmetic.
- US-006 confirmation modal shows the full scope (`mo clean` cleans everything) rather than a user-selected subset.
- The "Selected: X GB" footer concept becomes "Total reclaimable: X GB".
- FR-5's `mo clean <items…>` shell quoting is simplified to `mo clean` (no arguments).
- Multi-select cleanup becomes a v2 feature, contingent on Mole exposing per-item arguments.
- The TUI's core value proposition pivots from "select + clean" to "preview + understand + confirm."
