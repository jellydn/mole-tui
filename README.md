# mole-tui 👋

[![GitHub license](https://img.shields.io/github/license/jellydn/mole-tui)](https://github.com/jellydn/mole-tui/blob/main/LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/jellydn/mole-tui/pulls)
[![Twitter: jellydn](https://img.shields.io/twitter/follow/jellydn.svg?style=social)](https://twitter.com/jellydn)

> **A keyboard-driven TUI that orchestrates [Mole](https://github.com/tw93/mole) (`mo`) for macOS disk cleanup.**

Navigate, inspect, and reclaim disk space — all from the terminal. No clicking around System Settings or guessing what's safe to delete.

## ✨ Features

- 🎛️ **Interactive dashboard** — Browse cleanup sections and item-level details
- 🔍 **Live dry-run scan** — See exactly what `mo` would reclaim before committing
- ⌨️ **Vim-style navigation** — `j/k` to move, `tab` to collapse/expand, `enter` to clean
- 🛡️ **Sudo-aware re-scan** — Press `S` to include system caches that need elevated access
- 📊 **Reclaimable size estimate** — Best-effort total of all dry-run sizes at a glance
- 📝 **Live cleanup log** — Watch `mo clean` output stream in real-time
- 🕹️ **Confirmation gate** — Review the command and impact before running

## 📹 Demo

[![IT Man Channel](https://img.shields.io/badge/YouTube-IT%20Man%20Channel-red?logo=youtube)](https://www.youtube.com/@it-man)

## 🛠️ Prerequisites

- [Go](https://go.dev) 1.26+
- [Mole CLI](https://github.com/tw93/mole) (`mo`) on `$PATH`

```bash
brew install mole
```

## 🚀 Install

**Via `go install`:**

```bash
go install github.com/jellydn/mole-tui/cmd/mole-tui@latest
```

**Via `make`:**

```bash
git clone https://github.com/jellydn/mole-tui.git
cd mole-tui
make install
```

## 🔌 Programmatic access with `mole-sidecar`

The repository also ships `mole-sidecar`, a small companion process for desktop
GUIs and other clients. It exposes the same scan and cleanup engine over
newline-delimited [JSON-RPC 2.0](https://www.jsonrpc.org/specification) on
stdin/stdout. The sidecar requires the Mole CLI (`mo`) to be available on
`$PATH` and never deletes files directly; destructive work still goes through
`mo clean`.

Install it with Go:

```bash
go install github.com/jellydn/mole-tui/cmd/mole-sidecar@latest
```

Or build it locally:

```bash
go build -o ./bin/mole-sidecar ./cmd/mole-sidecar
```

Start the sidecar as a long-lived process and send one JSON request per line:

```json
{"jsonrpc":"2.0","id":1,"method":"ping"}
```

Available methods:

| Method | Parameters | Behaviour |
|--------|------------|-----------|
| `ping` | — | Returns the sidecar version and resolved `mo` path |
| `scan.start` | `{"sudo": true\|false}` | Starts `mo clean --dry-run`; emits `scan.done`, `scan.error`, or `scan.cancelled` |
| `scan.cancel` | — | Cancels the active scan |
| `cleanup.start` | `{"dryRun": true\|false, "sudo": true\|false}` | Starts cleanup; emits `cleanup.line` chunks followed by `cleanup.done` or `cleanup.error` |
| `cleanup.cancel` | — | Cancels the active cleanup |

Start methods immediately return an `accepted` response. Clients should keep
reading stdout for asynchronous notifications such as `cleanup.line` while the
operation is running. A cancelled cleanup ends with `cleanup.done` and
`params.cancelled: true`; scans emit `scan.cancelled` when cancelled.

## 🎮 Usage

```bash
mole-tui                # launch TUI — scans and shows cleanup dashboard
mole-tui --dry-run      # simulate cleanup (no files modified)
mole-tui -n             # alias for --dry-run
mole-tui --version      # print version and exit
```

### Flags

| Flag | Description |
|------|-------------|
| `--dry-run` / `-n` | Run in dry-run mode (simulate cleanup, no files modified) |
| `--version` | Print version and exit |

### Keybindings

| Key | Action |
|-----|--------|
| `↑/k` | Move cursor up |
| `↓/j` | Move cursor down |
| `tab` | Collapse / expand section |
| `enter` | Open cleanup confirmation |
| `r` | Re-scan |
| `S` | Re-scan with sudo |
| `?` | Show help |
| `q/ctrl+c` | Quit |

## 🏗️ Build

```bash
make build                # builds to ./bin/mole-tui
just build                # same, via just
make build VERSION=v0.1.0 # override version string
```

### Scripts

| Command | Description |
|---------|-------------|
| `just build` | Build to `./bin/mole-tui` |
| `just install` | Install to `$GOBIN` |
| `just test` | Run all unit tests |
| `just fmt` | Check gofmt |
| `just vet` | Run go vet |
| `just dev` | Build and run |
| `just ci` | Full pipeline: fmt → vet → test → build |

## 📁 Project Structure

```
cmd/mole-tui/main.go       # TUI binary entrypoint
cmd/mole-sidecar/main.go   # JSON-RPC sidecar binary entrypoint
internal/
  mo/                      # Injectable boundary around the `mo` subprocess
  scanner/                 # Parse `mo clean --dry-run` → structured sections
  cleanup/                 # Shell out to `mo clean`, stream live output/events
  ui/                      # Bubble Tea models, views, keybindings (5 screens)
```

### Stack

| Layer    | Choice                                                       |
| -------- | ------------------------------------------------------------ |
| TUI      | [Bubble Tea](https://charm.land/bubbletea) v2.0.8             |
| Widgets  | [Bubbles](https://charm.land/bubbles) v2.1.1                  |
| Styling  | [Lip Gloss](https://charm.land/lipgloss) v2.0.5               |
| Sidecar  | Newline-delimited JSON-RPC 2.0 over stdio                   |
| Backend  | `mo clean` / `mo clean --dry-run` (Mole CLI)                 |

## 📄 License

MIT

## 👤 Author

**Dung Huynh Duc**

- Website: [https://productsway.com/](https://productsway.com/)
- YouTube: [IT Man Channel](https://www.youtube.com/@it-man)
- Twitter: [@jellydn](https://twitter.com/jellydn)
- GitHub: [@jellydn](https://github.com/jellydn)

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](https://github.com/jellydn/mole-tui/issues).

## Show your support

[![kofi](https://img.shields.io/badge/Ko--fi-F16061?style=for-the-badge&logo=ko-fi&logoColor=white)](https://ko-fi.com/dunghd)
[![paypal](https://img.shields.io/badge/PayPal-00457C?style=for-the-badge&logo=paypal&logoColor=white)](https://paypal.me/dunghd)
[![buymeacoffee](https://img.shields.io/badge/Buy_Me_A_Coffee-FFDD00?style=for-the-badge&logo=buy-me-a-coffee&logoColor=black)](https://www.buymeacoffee.com/dunghd)

Give a ⭐️ if this project helped you!
