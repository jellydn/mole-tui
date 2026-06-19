package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette — minimal, dark/light aware, matching Mole CLI aesthetic.
var (
	normalFg  = lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#e0e0e0"}
	dimFg     = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#888888"}
	subtleFg  = lipgloss.AdaptiveColor{Light: "#999999", Dark: "#666666"}
	borderFg  = lipgloss.AdaptiveColor{Light: "#cccccc", Dark: "#444444"}
	accentFg  = lipgloss.AdaptiveColor{Light: "#1a6dd4", Dark: "#6ab0f3"}
	warningFg = lipgloss.AdaptiveColor{Light: "#c97a00", Dark: "#f5a623"}
	successFg = lipgloss.AdaptiveColor{Light: "#1a7f37", Dark: "#3fb950"}
	errorFg   = lipgloss.AdaptiveColor{Light: "#cf222e", Dark: "#f85149"}
	titleFg   = lipgloss.AdaptiveColor{Light: "#8250df", Dark: "#d2a8ff"}

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(titleFg).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Foreground(normalFg).
			MarginBottom(1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentFg).
			MarginTop(1).
			MarginBottom(1)

	actionableStyle = lipgloss.NewStyle().
			Foreground(warningFg)

	cleanStyle = lipgloss.NewStyle().
			Foreground(successFg)

	skippedStyle = lipgloss.NewStyle().
			Foreground(dimFg)

	hintStyle = lipgloss.NewStyle().
			Foreground(dimFg).
			Italic(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(normalFg)

	dimStyle = lipgloss.NewStyle().
			Foreground(dimFg)

	footerStyle = lipgloss.NewStyle().
			Foreground(subtleFg).
			MarginTop(1).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder())

	reclaimStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(successFg)

	bannerStyle = lipgloss.NewStyle().
			Foreground(warningFg).
			Background(lipgloss.AdaptiveColor{Light: "#fff8c5", Dark: "#3a3000"}).
			Padding(0, 1).
			MarginBottom(1)

	errorBannerStyle = lipgloss.NewStyle().
				Foreground(errorFg).
				Background(lipgloss.AdaptiveColor{Light: "#ffebe9", Dark: "#3d1114"}).
				Padding(0, 1).
				MarginBottom(1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(accentFg)

	helpHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(titleFg).
			MarginBottom(1)

	helpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentFg)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(normalFg)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			Width(50)

	cursorStyle = lipgloss.NewStyle().
			Foreground(accentFg).
			Bold(true)
)
