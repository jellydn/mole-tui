package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tw93/mole-tui/internal/ui"
)

// version is injected at build time via -ldflags.
var version = "dev"

func main() {
	dryRun := flag.Bool("dry-run", false, "Run in dry-run mode (simulate cleanup, no files modified)")
	dryRunShort := flag.Bool("n", false, "Alias for --dry-run")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("mole-tui version", version)
		os.Exit(0)
	}

	// Combine --dry-run and -n
	dryRunMode := *dryRun || *dryRunShort

	// Check for mo binary before starting the TUI
	if _, err := exec.LookPath("mo"); err != nil {
		fmt.Fprintf(os.Stderr, "\033[1;31mError:\033[0m mo is not on $PATH\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "mole-tui requires the \033[1mmo\033[0m CLI from Mole:\n")
		fmt.Fprintf(os.Stderr, "  brew install mo\n")
		fmt.Fprintf(os.Stderr, "  or visit https://github.com/tw93/mole\n")
		os.Exit(1)
	}

	p := tea.NewProgram(
		ui.NewModel(dryRunMode),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
