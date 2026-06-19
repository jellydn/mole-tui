# ADR-008: Simplified Three-Package Architecture

## Status
Accepted

## Context
The PRD specified 5 internal packages: `discovery`, `scanner`, `preview`, `cleanup`, `ui`. After grilling decisions:
- `discovery` merged into `scanner` (ADR-001)
- `preview` dropped (ADR-006)

## Decision
The internal package layout is:

```
cmd/
  └── mole-tui/main.go     # binary entrypoint
internal/
  ├── scanner/              # invoke `mo clean --dry-run`, parse output → []Section
  ├── cleanup/              # invoke `mo clean`, stream stdout/stderr, parse summary
  └── ui/                   # Bubble Tea models, views, keybindings, all screens
```

## Consequences
- 3 packages instead of 5 — less surface area, simpler dependency graph
- `scanner` owns: subprocess invocation, output parsing, domain types (`Section`, `Item`, `ItemState`, `ItemSize`)
- `cleanup` owns: subprocess invocation, stdout/stderr streaming, optional summary parsing
- `ui` owns: all Bubble Tea models (loading, dashboard, confirmation, log/report), views, keybindings
- Shared types (if any) can live in `scanner` since `cleanup` doesn't need the scan domain model
- Test fixtures live in `internal/scanner/testdata/` and `internal/cleanup/testdata/`
