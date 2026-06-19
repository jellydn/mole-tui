# ADR-010: Five TUI Screens

## Status
Accepted

## Context
After removing preview and simplifying the flow, the screen inventory needed confirmation.

## Decision
v1 has 5 screens:

1. **Loading** — Full-screen spinner with "Scanning…" message and elapsed timer. Shown on startup and re-scan.
2. **Dashboard** — Sectioned, scrollable list of items with states (actionable/clean/skipped/hint), sizes, and key hints footer. Main screen.
3. **Confirmation** — Modal overlay on the dashboard showing the literal `mo clean` command, total reclaimable, and y/n/esc prompt. Default focus: No.
4. **Log/Report** — Full-screen scrollable log streaming `mo clean` stdout/stderr. After process exits, shows a summary line (if parseable) and `enter` to return / `q` to quit.
5. **Help** — Triggered by `?`. Shows all keybindings for the current screen. `?` or `esc` dismisses.

## Consequences
- 5 distinct Bubble Tea model states
- Help screen is context-aware (shows bindings relevant to the current screen)
- `?` key binding added to all screens
- Errors are rendered as banners within the relevant screen (not a separate screen)
- Screen flow: Loading → Dashboard ↔ Confirmation → Log/Report → Dashboard
