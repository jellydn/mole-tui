# Mole TUI — Domain Glossary

| Term | Definition |
|------|-----------|
| **Mole** | The underlying CLI tool (`mo`) by tw93 for macOS/Linux disk cleanup. Mole TUI orchestrates it. |
| **Item** | A line of output from `mo clean --dry-run`, prefixed by a state marker (`→`, `✓`, `◎`, `☞`). Rendered as raw text within its section. |
| **Scan** | A single invocation of `mo clean --dry-run` that both discovers items and reports their reclaimable sizes. In v1, scan = discovery (ADR-001). |
| **Dry-run** | A non-destructive preview mode of `mo clean` that reports what *would* be deleted without modifying the filesystem. |
| **Dashboard** | The main TUI screen showing all sections and their items with a total reclaimable footer. |
| **Confirmation gate** | A minimal modal showing the `mo clean` command and total reclaimable, requiring `y` to proceed. |
| **Cleanup session** | One invocation of `mo clean` (no item arguments in v1 — all-or-nothing, ADR-002). |
| **Log/Report** | Post-cleanup screen showing a scrollable log of `mo clean` stdout/stderr with an optional summary line (ADR-007). |
| **Loading screen** | Full-screen spinner shown during scan with elapsed time (ADR-004). |
| **Help screen** | Context-aware keybinding overlay triggered by `?` (ADR-010). |
| **Orchestrator** | The TUI's role — it never deletes files itself; it only invokes `mo clean`. |
| **Section** | A named group of related items in `mo clean --dry-run` output, prefixed with `➤` (e.g., "User essentials", "Browsers", "Developer tools"). Collapsible in the TUI. |
| **Actionable** (`→`) | An item that will be cleaned by `mo clean`. May include size and file count. |
| **Clean** (`✓`) | An item that is already empty / nothing to do (e.g., "Trash · already empty"). |
| **Skipped** (`◎`) | An item that was skipped, typically because an app is running or it requires opt-in (e.g., Docker). |
| **Hint** (`☞`) | A recommendation for manual action, not automatically cleaned (e.g., "Review: docker system df"). |
| **Whitelist** | Mole's mechanism for protecting paths from cleanup. Managed via `mo clean --whitelist`. The TUI does not manipulate it. |
| **Sudo elevation** | Optional re-scan with `sudo -v` to include system-level caches. Triggered by `S` keybinding (ADR-005). |
| **Hybrid parsing** | v1 parsing strategy: structured section headers + raw item lines + best-effort size aggregation (ADR-012). |
