package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"

	"github.com/jellydn/mole-tui/internal/ui"
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

	// Resolve the absolute mo path for safety and display
	moPath, err := exec.LookPath("mo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[1;31mError:\033[0m mo is not on $PATH\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "mole-tui requires the \033[1mmo\033[0m CLI from Mole:\n")
		fmt.Fprintf(os.Stderr, "  brew install mo\n")
		fmt.Fprintf(os.Stderr, "  or visit https://github.com/tw93/mole\n")
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewModel(dryRunMode, moPath))

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
