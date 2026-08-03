# ADR-013: Stay on Bubble Tea — Reject the OpenTUI (Zig/TypeScript) Swap

## Status
Accepted

## Context
During the architecture review on 2026-08-04, the question was raised whether the TUI could be rebuilt on **OpenTUI** (`anomalyco/opentui`) — the terminal UI framework that powers OpenCode. Verified facts about OpenTUI:

- It is a **terminal** UI framework, not a windowed GUI.
- Native core written in **Zig**; primary bindings in **TypeScript** (`@opentui/core`) with React/Solid reconcilers and a keymap engine; exposes a C ABI for other languages.
- Requires Zig to build and Bun/Node at runtime.
- Component-tree programming model (Yoga-powered flexbox layout), different from Bubble Tea's Elm-style Model-View-Update loop.

The project is a pure Go module (Go 1.26.4, `github.com/jellydn/mole-tui`) with a **no-CGO constraint** (AGENTS.md), distributed via `go install`, and its UI is built on Bubble Tea v2 + Bubbles v2 + Lip Gloss v2 (charm.land paths). PRD **NG-8** lists a graphical/web frontend as an explicit v1 non-goal; the product is a keyboard-driven terminal orchestrator for the `mo` CLI.

## Decision
Stay on **Bubble Tea v2** for the TUI. Do not integrate OpenTUI, and do not rewrite the UI in TypeScript.

A framework swap here is a **full UI rewrite, not a refactor** — it would invalidate the five deepening candidates in the architecture-review stack and the framework boundary of the ADR-008 package layout, for marginal gain against hard constraints (language, CGO, distribution).

Revisit only if a future product requirement demands a browser/SSH-renderable surface (e.g. `@opentui/ssh`) — and treat that as a **new TypeScript surface**, not a change to this Go module.

## Consequences

### Positive
- No rewrite; the existing 5-screen UI (`internal/ui/model.go`) and the 5-branch architecture stack remain valid.
- Preserves the no-CGO constraint and a single-language (Go) codebase; no Zig/Bun toolchain requirement.
- `go install` / `make install` distribution story unchanged.
- Bubble Tea is actively maintained and remains the de-facto Go TUI standard (Charm).

### Negative
- No access to OpenTUI's flexbox layout engine, Tree-sitter syntax highlighting, or default mouse support.
- A future web/SSH-accessible UI requires a separate TypeScript surface (or revisiting this ADR).
- Note for future reviews: OpenTUI is a TUI, not a windowed GUI; windowed GUIs (Wails/Fyne/Gio) remain out of scope per PRD NG-8.
