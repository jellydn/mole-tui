# ADR-006: Drop Preview Pane in v1

## Status
Accepted

## Context
The PRD (US-005) specified a preview pane showing per-item file trees, target paths, and potential savings. This assumed the dry-run output or a separate command would expose granular file listings.

In practice, `mo clean --dry-run` only provides summary lines per item (e.g., "91 items, 272.8MB dry"). There is no per-item file tree available without poking at the filesystem directly — which would violate the TUI's orchestrator-only constraint.

The dashboard (with full-fidelity parsing per ADR-003) already shows all information Mole exposes: section, item name, state, size, and file count.

## Decision
Drop the preview pane entirely in v1. The dashboard provides sufficient information. `p` key binding is freed for future use. US-005 is deferred to v2, contingent on Mole exposing richer per-item detail.

## Consequences
- The `preview` internal package is removed from the v1 architecture
- Repo layout simplifies: `internal/{scanner, cleanup, ui}` (no `discovery`, no `preview`)
- US-005 moves to non-goals / v2 backlog
- One fewer screen/overlay to implement and test
- The `p` key binding can be repurposed (e.g., toggle between section-collapsed/expanded views)
