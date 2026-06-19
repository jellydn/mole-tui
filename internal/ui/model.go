// Package ui implements the Bubble Tea TUI with 5 screens:
// Loading, Dashboard, Confirmation, Log/Report, Help.
package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tw93/mole-tui/internal/cleanup"
	"github.com/tw93/mole-tui/internal/scanner"
)

// Screen represents the current UI screen.
type Screen int

const (
	screenLoading Screen = iota
	screenDashboard
	screenConfirm
	screenLog
	screenHelp
)

// Model is the root Bubble Tea model.
type Model struct {
	// Config
	DryRun bool

	// Screen state
	screen     Screen
	prevScreen Screen
	helpScreen Screen
	width      int
	height     int
	ready      bool

	// Loading
	spinner     spinner.Model
	loadingMsg  string
	loadingTime time.Time
	scanCancel  context.CancelFunc

	// Scan results
	scanResult  scanner.ScanResult
	scanErr     error
	hasPrevScan bool

	// Dashboard
	cursor     int
	expanded   map[int]bool
	sudoBanner bool
	errorMsg   string

	// Confirmation modal
	confirmDryRun bool

	// Log/Report
	logContent  string
	logStderr   string
	logVP       viewport.Model
	logDone     bool
	logExit     int
	logSummary  string
	quitConfirm bool // first ctrl+c during cleanup

	// Help
	helpKeys help.Model
}

// Key bindings per screen.
var (
	dashboardKeys = KeyMap{
		Up:     key.NewBinding(key.WithKeys("up", "k")),
		Down:   key.NewBinding(key.WithKeys("down", "j")),
		Tab:    key.NewBinding(key.WithKeys("tab")),
		Enter:  key.NewBinding(key.WithKeys("enter")),
		Rescan: key.NewBinding(key.WithKeys("r")),
		Sudo:   key.NewBinding(key.WithKeys("S")),
		Help:   key.NewBinding(key.WithKeys("?")),
		Esc:    key.NewBinding(key.WithKeys("esc")),
		Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c")),
	}

	confirmKeys = KeyMap{
		Confirm: key.NewBinding(key.WithKeys("y")),
		Cancel:  key.NewBinding(key.WithKeys("n", "esc")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c")),
	}

	logKeys = KeyMap{
		Up:    key.NewBinding(key.WithKeys("up", "k")),
		Down:  key.NewBinding(key.WithKeys("down", "j")),
		Enter: key.NewBinding(key.WithKeys("enter")),
		Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c")),
	}

	loadingKeys = KeyMap{
		Esc:  key.NewBinding(key.WithKeys("esc")),
		Quit: key.NewBinding(key.WithKeys("q", "ctrl+c")),
	}
)

// KeyMap defines key bindings for a screen.
type KeyMap struct {
	Up, Down, Tab, Enter, Rescan, Sudo, Confirm, Cancel, Help, Quit, Esc key.Binding
}

// Tea messages.

type scanCompleteMsg struct {
	result scanner.ScanResult
	err    error
}

type scanCancelledMsg struct{}

type cleanupCompleteMsg struct {
	result cleanup.Result
	err    error
}

// NewModel creates the root model.
func NewModel(dryRun bool) *Model {
	s := spinner.New()
	s.Style = spinnerStyle
	s.Spinner = spinner.Dot

	return &Model{
		DryRun:      dryRun,
		screen:      screenLoading,
		spinner:     s,
		loadingMsg:  "Scanning… this may take a few minutes",
		loadingTime: time.Now(),
		expanded:    make(map[int]bool),
		helpKeys:    help.New(),
	}
}

func (m *Model) Init() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.scanCancel = cancel
	return tea.Batch(m.spinner.Tick, m.scanCmd(ctx, false))
}

// scanCmd returns a tea.Cmd that runs `mo clean --dry-run` with the given ctx.
func (m *Model) scanCmd(ctx context.Context, sudo bool) tea.Cmd {
	return func() tea.Msg {
		result, err := scanner.Scan(ctx, sudo)
		if errors.Is(err, context.Canceled) {
			return scanCancelledMsg{}
		}
		return scanCompleteMsg{result: result, err: err}
	}
}

// cleanupCmd returns a tea.Cmd that runs `mo clean` and captures all output.
func cleanupCmd(dryRun bool) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		result, err := cleanup.Run(context.Background(), cleanup.Options{DryRun: dryRun}, &buf)
		if err != nil && result.Stdout == "" && result.Stderr == "" {
			// The error occurred before we got any output
			buf.WriteString(fmt.Sprintf("Error: %s\n", err))
		}
		result.Stdout = buf.String()
		if dryRun {
			result.FreedText = "Dry run complete — no files were modified"
		} else if result.FreedText == "" {
			result.FreedText = cleanup.ParseSummary(buf.String())
		}
		return cleanupCompleteMsg{result: result, err: err}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.logVP = viewport.New(max(0, msg.Width-4), max(0, msg.Height-8))
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case scanCompleteMsg:
		if msg.err != nil {
			m.scanErr = msg.err
			m.errorMsg = msg.err.Error()
			if m.hasPrevScan {
				m.screen = screenDashboard
			} else {
				m.loadingMsg = fmt.Sprintf("Scan failed: %s", msg.err)
			}
		} else {
			m.scanResult = msg.result
			m.scanErr = nil
			m.errorMsg = ""
			m.sudoBanner = msg.result.Summary.SystemCachesSkipped
			m.hasPrevScan = true
			m.screen = screenDashboard
			m.cursor = 0
			for i := range m.scanResult.Sections {
				m.expanded[i] = false
			}
		}
		m.scanCancel = nil
		return m, nil

	case scanCancelledMsg:
		m.scanCancel = nil
		if m.hasPrevScan {
			m.screen = screenDashboard
		} else {
			m.loadingMsg = "Scan cancelled"
		}
		return m, nil

	case cleanupCompleteMsg:
		m.logDone = true
		m.logExit = msg.result.ExitCode
		m.logSummary = msg.result.FreedText
		m.logStderr = msg.result.Stderr
		if msg.err != nil {
			m.logSummary = fmt.Sprintf("Error: %s", msg.err)
			m.logExit = 1
		}
		m.logContent = msg.result.Stdout
		if m.logStderr != "" {
			m.logContent += "\n" + m.logStderr
		}
		m.logVP.SetContent(m.logContent)
		m.logVP.GotoBottom()
		return m, nil
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenLoading:
		return m.handleLoadingKey(msg)
	case screenDashboard:
		return m.handleDashboardKey(msg)
	case screenConfirm:
		return m.handleConfirmKey(msg)
	case screenLog:
		return m.handleLogKey(msg)
	case screenHelp:
		return m.handleHelpKey(msg)
	}
	return m, nil
}

// -------------------------------------------------------
// Loading screen
// -------------------------------------------------------

func (m *Model) loadingView() string {
	elapsed := time.Since(m.loadingTime).Round(time.Second)
	content := fmt.Sprintf("%s %s\n\nElapsed: %s\n\n%s",
		m.spinner.View(), m.loadingMsg, elapsed,
		dimStyle.Render("esc — cancel  •  q/ctrl+c — quit"))
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center, content)
}

func (m *Model) handleLoadingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, loadingKeys.Esc):
		if m.scanCancel != nil {
			m.scanCancel()
		}
		if m.hasPrevScan {
			m.screen = screenDashboard
		} else {
			m.loadingMsg = "Cancelling…"
		}
		return m, nil
	case key.Matches(msg, loadingKeys.Quit):
		if m.scanCancel != nil {
			m.scanCancel()
		}
		return m, tea.Quit
	}
	return m, nil
}

// -------------------------------------------------------
// Dashboard screen
// -------------------------------------------------------

func (m *Model) dashboardView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("mole-tui"))
	b.WriteString("\n")

	if m.scanResult.Summary.Header != "" {
		b.WriteString(headerStyle.Render(m.scanResult.Summary.Header))
		b.WriteString("\n")
	}

	if m.sudoBanner {
		b.WriteString(bannerStyle.Render("◎ System caches skipped — press S to re-scan with sudo"))
		b.WriteString("\n")
	}

	if m.errorMsg != "" {
		b.WriteString(errorBannerStyle.Render(m.errorMsg))
		b.WriteString("\n")
	}

	globalLine := 0
	for i, section := range m.scanResult.Sections {
		mark := "  "
		if globalLine == m.cursor {
			mark = cursorStyle.Render("▸ ")
		}

		toggle := "[+]"
		if !m.expanded[i] {
			toggle = "[-]"
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n", mark,
			sectionStyle.Render(section.Name),
			dimStyle.Render(toggle)))
		globalLine++

		if !m.expanded[i] {
			for _, line := range section.Lines {
				mark := "  "
				if globalLine == m.cursor {
					mark = cursorStyle.Render("▸ ")
				}
				b.WriteString(fmt.Sprintf("  %s%s\n", mark, styleItemLine(line)))
				globalLine++
			}
		}
	}

	b.WriteString("\n")
	if m.scanResult.Summary.TotalReclaimable > 0 {
		b.WriteString(reclaimStyle.Render(
			fmt.Sprintf("Total reclaimable: %s", formatBytes(m.scanResult.Summary.TotalReclaimable))))
	} else {
		b.WriteString(reclaimStyle.Render("Total reclaimable: —"))
	}
	b.WriteString("\n\n")

	b.WriteString(footerStyle.Render(
		"↑/k ↓/j  •  tab toggle  •  enter clean  •  r rescan  •  S sudo  •  ? help  •  q quit"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func styleItemLine(line string) string {
	switch {
	case strings.HasPrefix(line, "→"):
		return actionableStyle.Render(line)
	case strings.HasPrefix(line, "✓"):
		return cleanStyle.Render(line)
	case strings.HasPrefix(line, "◎"):
		return skippedStyle.Render(line)
	case strings.HasPrefix(line, "☞"):
		return hintStyle.Render(line)
	case strings.HasPrefix(line, "↳"):
		return dimStyle.Render(line)
	default:
		return normalStyle.Render(line)
	}
}

func (m *Model) visibleLineCount() int {
	n := 0
	for i, s := range m.scanResult.Sections {
		n++ // header
		if !m.expanded[i] {
			n += len(s.Lines)
		}
	}
	return n
}

func (m *Model) handleDashboardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Guard: no sections means nothing to navigate
	if len(m.scanResult.Sections) == 0 {
		// Only allow rescan, sudo, help, and quit
		switch {
		case key.Matches(msg, dashboardKeys.Rescan):
			m.screen = screenLoading
			m.loadingMsg = "Re-scanning…"
			m.loadingTime = time.Now()
			ctx, cancel := context.WithCancel(context.Background())
			m.scanCancel = cancel
			return m, m.scanCmd(ctx, false)
		case key.Matches(msg, dashboardKeys.Sudo):
			m.screen = screenLoading
			m.loadingMsg = "Re-scanning with sudo…"
			m.loadingTime = time.Now()
			ctx, cancel := context.WithCancel(context.Background())
			m.scanCancel = cancel
			return m, m.scanCmd(ctx, true)
		case key.Matches(msg, dashboardKeys.Esc):
			return m, tea.Quit
		case key.Matches(msg, dashboardKeys.Help):
			m.prevScreen = m.screen
			m.helpScreen = screenDashboard
			m.screen = screenHelp
		case key.Matches(msg, dashboardKeys.Quit):
			return m, tea.Quit
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, dashboardKeys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, dashboardKeys.Down):
		if m.cursor < m.visibleLineCount()-1 {
			m.cursor++
		}
	case key.Matches(msg, dashboardKeys.Tab):
		line := 0
		for i := range m.scanResult.Sections {
			if line == m.cursor {
				m.expanded[i] = !m.expanded[i]
				break
			}
			line++
			if !m.expanded[i] {
				line += len(m.scanResult.Sections[i].Lines)
			}
		}
	case key.Matches(msg, dashboardKeys.Enter):
		m.screen = screenConfirm
		m.confirmDryRun = m.DryRun
	case key.Matches(msg, dashboardKeys.Rescan):
		m.screen = screenLoading
		m.loadingMsg = "Re-scanning…"
		m.loadingTime = time.Now()
		ctx, cancel := context.WithCancel(context.Background())
		m.scanCancel = cancel
		return m, m.scanCmd(ctx, false)
	case key.Matches(msg, dashboardKeys.Sudo):
		m.screen = screenLoading
		m.loadingMsg = "Re-scanning with sudo…"
		m.loadingTime = time.Now()
		ctx, cancel := context.WithCancel(context.Background())
		m.scanCancel = cancel
		return m, m.scanCmd(ctx, true)
	case key.Matches(msg, dashboardKeys.Esc):
		return m, tea.Quit
	case key.Matches(msg, dashboardKeys.Help):
		m.prevScreen = m.screen
		m.helpScreen = screenDashboard
		m.screen = screenHelp
	case key.Matches(msg, dashboardKeys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

// -------------------------------------------------------
// Confirmation screen
// -------------------------------------------------------

func (m *Model) confirmView() string {
	content := "Run cleanup command?\n\n"
	content += "  Command: mo clean\n"
	if m.confirmDryRun {
		content += "  [DRY RUN] — no files will be modified\n"
	} else {
		content += "  (may prompt for sudo password)\n"
	}
	content += fmt.Sprintf("  Reclaimable: %s\n", formatBytes(m.scanResult.Summary.TotalReclaimable))
	content += fmt.Sprintf("\n  y — confirm  •  n/esc — cancel  •  q — quit")

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modalStyle.Render(content))
}

func (m *Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, confirmKeys.Confirm):
		m.screen = screenLog
		m.logContent = ""
		m.logDone = false
		m.logExit = 0
		m.logSummary = ""
		if m.logVP.Height > 0 {
			m.logVP.SetContent("")
		}
		return m, cleanupCmd(m.DryRun)
	case key.Matches(msg, confirmKeys.Cancel):
		m.screen = screenDashboard
	case key.Matches(msg, confirmKeys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

// -------------------------------------------------------
// Log/Report screen
// -------------------------------------------------------

func (m *Model) logView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Cleanup Log"))
	b.WriteString("\n")

	if m.logDone {
		if m.logVP.View() == "" && m.logContent != "" {
			m.logVP.SetContent(m.logContent)
		}
		b.WriteString(m.logVP.View())
		b.WriteString("\n")

		if m.logSummary != "" {
			b.WriteString("\n")
			b.WriteString(reclaimStyle.Render(m.logSummary))
		}
		if m.logExit != 0 {
			b.WriteString("\n")
			b.WriteString(errorBannerStyle.Render(
				fmt.Sprintf("Cleanup exited with code %d", m.logExit)))
			// Show last 20 lines of stderr
			if m.logStderr != "" {
				lines := strings.Split(m.logStderr, "\n")
				if len(lines) > 20 {
					lines = lines[len(lines)-20:]
				}
				b.WriteString("\n")
				b.WriteString(errorBannerStyle.Render(
					fmt.Sprintf("Last %d lines of stderr:", len(lines))))
				for _, l := range lines {
					b.WriteString("\n  ")
					b.WriteString(l)
				}
			}
		}
	} else {
		b.WriteString(dimStyle.Render("Running mo clean…"))
		if m.quitConfirm {
			b.WriteString("\n\n")
			b.WriteString(bannerStyle.Render(
				"Press ctrl+c again to kill the cleanup process"))
		}
	}

	b.WriteString("\n\n")
	helpText := "↑/↓ j/k — scroll  •  q — quit"
	if m.logDone {
		if m.logExit != 0 {
			helpText = "enter — return  •  ↑/↓ j/k — scroll  •  q — quit"
		} else {
			helpText = "enter — return to dashboard  •  ↑/↓ j/k — scroll  •  q — quit"
		}
	} else if m.quitConfirm {
		helpText = "ctrl+c again to force quit  •  esc to cancel"
	}
	b.WriteString(footerStyle.Render(helpText))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m *Model) handleLogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, logKeys.Up):
		if m.logDone {
			m.logVP.LineUp(1)
		}
	case key.Matches(msg, logKeys.Down):
		if m.logDone {
			m.logVP.LineDown(1)
		}
	case key.Matches(msg, logKeys.Enter):
		if m.logDone {
			m.screen = screenDashboard
			m.cursor = 0
			m.logDone = false
			m.quitConfirm = false
		}
	case key.Matches(msg, logKeys.Quit):
		if m.logDone {
			return m, tea.Quit
		}
		if !m.quitConfirm {
			m.quitConfirm = true
			// First press — warn, second press will quit
		} else {
			return m, tea.Quit
		}
	case key.Matches(msg, loadingKeys.Esc):
		if !m.logDone && m.quitConfirm {
			m.quitConfirm = false
		}
	}
	return m, nil
}

// -------------------------------------------------------
// Help screen
// -------------------------------------------------------

func (m *Model) helpView() string {
	var b strings.Builder
	b.WriteString(helpHeaderStyle.Render("Keyboard Help"))
	b.WriteString("\n")

	entries := helpEntriesFor(m.helpScreen)
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			helpKeyStyle.Render(fmt.Sprintf("%-12s", e.key)),
			helpDescStyle.Render(e.desc)))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("? or esc — close  •  q — quit"))

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).Width(50).Render(b.String()))
}

type helpEntry struct{ key, desc string }

func helpEntriesFor(s Screen) []helpEntry {
	switch s {
	case screenDashboard:
		return []helpEntry{
			{"↑/k", "Move cursor up"},
			{"↓/j", "Move cursor down"},
			{"tab", "Collapse/expand section"},
			{"enter", "Open cleanup confirmation"},
			{"r", "Re-scan (dry-run)"},
			{"S", "Re-scan with sudo"},
			{"?", "Show this help"},
			{"q/ctrl+c", "Quit"},
		}
	case screenConfirm:
		return []helpEntry{
			{"y", "Confirm and run cleanup"},
			{"n/esc", "Cancel and return"},
			{"q", "Quit"},
		}
	case screenLog:
		return []helpEntry{
			{"↑/k", "Scroll log up"},
			{"↓/j", "Scroll log down"},
			{"enter", "Return to dashboard (after completion)"},
			{"q/ctrl+c", "Quit"},
		}
	default:
		return []helpEntry{
			{"?/esc", "Close help"},
			{"q", "Quit"},
		}
	}
}

func (m *Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, dashboardKeys.Help),
		key.Matches(msg, confirmKeys.Cancel):
		m.screen = m.prevScreen
	case key.Matches(msg, dashboardKeys.Quit):
		return m, tea.Quit
	default:
		m.screen = m.prevScreen
	}
	return m, nil
}

func (m *Model) View() string {
	if !m.ready {
		return "mole-tui — loading…"
	}
	switch m.screen {
	case screenLoading:
		return m.loadingView()
	case screenDashboard:
		return m.dashboardView()
	case screenConfirm:
		return m.confirmView()
	case screenLog:
		return m.logView()
	case screenHelp:
		return m.helpView()
	}
	return ""
}

// -------------------------------------------------------
// Helpers
// -------------------------------------------------------

func formatBytes(b float64) string {
	switch {
	case b >= 1024*1024*1024:
		return fmt.Sprintf("%.2f GB", b/(1024*1024*1024))
	case b >= 1024*1024:
		return fmt.Sprintf("%.1f MB", b/(1024*1024))
	case b >= 1024:
		return fmt.Sprintf("%.0f KB", b/1024)
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
