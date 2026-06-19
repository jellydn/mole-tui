# ADR-003: Full-Fidelity Dry-Run Parsing with Sections and Item States

## Status
Accepted

## Context
The PRD modeled scan results as a flat `map[string]ItemSize{Bytes, FileCount, LastUpdated}`. The actual `mo clean --dry-run` output is richer:

1. **Sections** (`➤`): Named groups like "User essentials", "Browsers", "Developer tools"
2. **Item states**:
   - `→` — Actionable: will be cleaned, may include size/count
   - `✓` — Already clean: nothing to do
   - `◎` — Skipped: app running, or opt-in required (e.g., Docker)
   - `☞` — Hint: recommendation for manual action
3. **No `LastUpdated` field** in any item
4. **Inconsistent size data**: some items report `272.8MB dry`, others say `would clean` with no size

## Decision
Parse and display the full structure:
- Model **sections** as first-class groups (collapsible in the TUI)
- Parse all **4 item states** and render with distinct visual treatments (color/icon)
- Drop `LastUpdated` from the domain model entirely
- Items without sizes display `—` or "unknown"

## Consequences
- The `scanner` domain model becomes `[]Section` where each `Section` has a `Name` and `[]Item`
- `Item` gains a `State` enum: `Actionable | Clean | Skipped | Hint`
- `ItemSize` becomes optional (nullable/zero-value) since not all items report sizes
- The dashboard UI needs grouped rendering (not a flat list)
- Parser complexity increases but produces a faithful representation of Mole's output
- `FileCount` is kept where available; `LastUpdated` is removed
