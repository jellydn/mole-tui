package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

// Color palette — minimal, dark/light aware, matching Mole CLI aesthetic.
var (
	normalFg  color.Color = compat.AdaptiveColor{Light: lipgloss.Color("#1a1a1a"), Dark: lipgloss.Color("#e0e0e0")}
	dimFg     color.Color = compat.AdaptiveColor{Light: lipgloss.Color("#666666"), Dark: lipgloss.Color("#888888")}
	subtleFg  color.Color = compat.AdaptiveColor{Light: lipgloss.Color("#999999"), Dark: lipgloss.Color("#666666")}
	borderFg  color.Color = compat.AdaptiveColor{Light: lipgloss.Color("#cccccc"), Dark: lipgloss.Color("#444444")}
	accentFg  color.Color = compat.AdaptiveColor{Light: lipgloss.Color("#1a6dd4"), Dark: lipgloss.Color("#6ab0f3")}
	warningFg color.Color = compat.AdaptiveColor{Light: lipgloss.Color("#c97a00"), Dark: lipgloss.Color("#f5a623")}
	successFg color.Color = compat.AdaptiveColor{Light: lipgloss.Color("#1a7f37"), Dark: lipgloss.Color("#3fb950")}
	errorFg   color.Color = compat.AdaptiveColor{Light: lipgloss.Color("#cf222e"), Dark: lipgloss.Color("#f85149")}
	titleFg   color.Color = compat.AdaptiveColor{Light: lipgloss.Color("#8250df"), Dark: lipgloss.Color("#d2a8ff")}

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
			Background(compat.AdaptiveColor{Light: lipgloss.Color("#fff8c5"), Dark: lipgloss.Color("#3a3000")}).
			Padding(0, 1).
			MarginBottom(1)

	errorBannerStyle = lipgloss.NewStyle().
				Foreground(errorFg).
				Background(compat.AdaptiveColor{Light: lipgloss.Color("#ffebe9"), Dark: lipgloss.Color("#3d1114")}).
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
