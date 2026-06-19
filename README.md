# mole-tui

A keyboard-driven TUI that orchestrates [Mole](https://github.com/tw93/mole) (`mo`) for disk cleanup.

## Prerequisites

- [Go](https://go.dev) 1.22+
- [Mole CLI](https://github.com/tw93/mole) (`mo`) on `$PATH`

```bash
brew install mo
```

## Install

**Via `go install`:**

```bash
go install github.com/tw93/mole-tui/cmd/mole-tui@latest
```

**Via `make`:**

```bash
git clone https://github.com/tw93/mole-tui.git
cd mole-tui
make install
```

## Usage

```bash
mole-tui           # launch TUI — scans and shows cleanup dashboard
mole-tui --dry-run # simulate cleanup (no files modified)
mole-tui -n        # same as --dry-run
mole-tui --version # print version
```

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

## Build

```bash
make build          # builds to ./bin/mole-tui
just build          # same, via just
make build VERSION=v0.1.0   # override version string
```

## License

MIT
