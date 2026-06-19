# ADR-009: Minimal Keybinding Set for v1

## Status
Accepted

## Context
With multi-select (ADR-002) and preview (ADR-006) removed, many PRD keybindings are orphaned. The keybinding set should reflect the actual v1 feature set.

## Decision

### Dashboard screen
| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `tab` | Collapse / expand focused section |
| `enter` | Open confirmation modal (to run `mo clean`) |
| `r` | Re-scan (re-run `mo clean --dry-run`) |
| `S` (shift-s) | Re-scan with sudo (`sudo -v && mo clean --dry-run`) |
| `q` / `ctrl+c` | Quit |

### Confirmation modal
| Key | Action |
|-----|--------|
| `y` | Confirm and execute `mo clean` |
| `n` / `esc` | Cancel and return to dashboard |

### Log/Report screen (during and after cleanup)
| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll log down |
| `k` / `↑` | Scroll log up |
| `enter` | Return to dashboard (after cleanup completes) |
| `q` / `ctrl+c` | Quit (first press shows warning during active cleanup) |

### Loading screen
| Key | Action |
|-----|--------|
| `esc` | Cancel scan |
| `q` / `ctrl+c` | Quit |

## Consequences
- 7 unique actions across 4 screens (much simpler than the PRD's ~10)
- All vim-like navigation preserved (`j/k`)
- `space`, `a`, `c`, `p`, `s` are freed for v2 features
- Key hint footer updates per screen
- Every screen has `q/ctrl+c` as an exit path (US-008)
