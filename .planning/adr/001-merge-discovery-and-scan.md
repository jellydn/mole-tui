# ADR-001: Merge Discovery and Scan into a Single Phase

## Status
Accepted

## Context
The PRD assumed `mo --help` lists per-item cleanup targets (e.g., `downloads`, `trash`, `browser`), allowing a two-phase flow: (1) discover items from `--help`, (2) scan sizes via `mo clean --dry-run`.

Testing against the actual `mo` CLI (Mole v-current) revealed that `mo --help` only lists top-level commands (`clean`, `uninstall`, `optimize`, etc.) and `mo clean --help` only lists flags. Individual cleanup items are **only** exposed in `mo clean --dry-run` output.

## Decision
Merge discovery and scan into a single phase. On startup (or on `s`/`r` keypress), the TUI invokes `mo clean --dry-run` once and parses both item names and their sizes from the output.

## Consequences
- The `discovery` package is either removed or becomes a thin wrapper around the scan parser that returns only item names.
- There is no "fast startup" that shows items without sizes — the TUI must wait for the dry-run to complete before populating the dashboard.
- US-001's acceptance criteria about parsing `mo --help` must be rewritten to reference `mo clean --dry-run`.
- FR-2 (`discovery` must invoke `mo --help`) is superseded.
- The architecture simplifies: one parser, one subprocess invocation, one source of truth.
