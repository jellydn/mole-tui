# ADR-005: Non-Sudo by Default with Opt-In Elevation

## Status
Accepted

## Context
`mo clean --dry-run` reports partial results without `sudo` — system caches are skipped with a message:
```
◎ System caches need sudo, run sudo -v && mo clean --dry-run for full preview
```

The PRD states the TUI must not require root (FR-10). But a full scan needs `sudo -v` first.

## Decision
Run without `sudo` by default. Show a persistent banner in the dashboard when system caches were skipped. Offer a key binding (e.g., `S` for "sudo re-scan") that runs `sudo -v && mo clean --dry-run` to get the full picture.

## Consequences
- The initial scan is fast and non-privileged
- A banner/notice in the dashboard informs the user: "System caches skipped — press S to re-scan with sudo"
- The `S` keybinding triggers `sudo -v` (which prompts in the user's terminal auth flow — Touch ID, password, etc.) followed by a full re-scan
- The parser must detect the `◎ System caches need sudo` line and set a flag on the scan result
- The TUI model gains a `SudoAvailable` / `SystemCachesSkipped` boolean
- `mo clean` (actual cleanup) inherits the same sudo state — if the user elevated for the scan, cleanup should also run elevated
