# ADR-012: Hybrid Parsing — Structured Sections, Raw Item Lines

## Status
Accepted

## Context
The `mo clean --dry-run` output has inconsistent per-item formats (some have sizes, some don't, sub-items vary). Full structured parsing is fragile. Pure raw rendering loses navigation and summary capabilities.

## Decision
Use a hybrid approach:

1. **Parse section headers** (`➤` lines) to create collapsible groups
2. **Parse aggregate sizes** where possible (sum up any `NNN.NMB dry` patterns for a total reclaimable footer)
3. **Render individual item lines as raw text** within their sections — no per-item domain model, just strings
4. **Parse the header line** (e.g., "Apple Silicon | Free space: 104.65GB") for the dashboard header

## Consequences
- `scanner` returns `[]Section` where `Section` has `Name string` and `Lines []string`
- A separate `ScanSummary` struct holds: `FreeSpace`, `TotalReclaimable` (best-effort sum), `SystemCachesSkipped bool`
- Individual item lines retain their original Unicode markers (`→`, `✓`, `◎`, `☞`) for visual rendering
- Collapsible sections work (section headers are parsed)
- Footer can show total reclaimable (best-effort)
- No per-item cursor movement or state filtering in v1 — cursor moves by line within sections
- Much simpler parser: one regex for sections, one for sizes, one for the header
- Robust against Mole output format changes (raw lines are never rejected)
