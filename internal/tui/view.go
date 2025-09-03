package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	// Padding constants for UI elements.
	horizontalPadding = 2
	verticalPadding   = 1
	noMargin          = 0
)

// StyleSet contains all TUI styles.
type StyleSet struct {
	Title          lipgloss.Style
	Column         lipgloss.Style
	FocusedColumn  lipgloss.Style
	LockedColumn   lipgloss.Style
	Status         lipgloss.Style
	Help           lipgloss.Style
	Passphrase     lipgloss.Style
	StrengthStyles map[string]lipgloss.Style
}

// NewStyleSet creates a new set of TUI styles.
func NewStyleSet() StyleSet {
	columnStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(verticalPadding).
		Margin(noMargin, verticalPadding)

	return StyleSet{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(noMargin, verticalPadding),
		Column: columnStyle,
		FocusedColumn: columnStyle.
			BorderForeground(lipgloss.Color("205")),
		LockedColumn: columnStyle.
			BorderForeground(lipgloss.Color("240")).
			Foreground(lipgloss.Color("240")),
		Status: lipgloss.NewStyle().
			Padding(verticalPadding, horizontalPadding).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(verticalPadding, horizontalPadding),
		Passphrase: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Padding(noMargin, horizontalPadding),
		StrengthStyles: map[string]lipgloss.Style{
			"Weak":      lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
			"Okay":      lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
			"Strong":    lipgloss.NewStyle().Foreground(lipgloss.Color("46")),
			"Excellent": lipgloss.NewStyle().Foreground(lipgloss.Color("82")),
		},
	}
}

// View implements the Bubble Tea Model interface.
func (m Model) View() string {
	if m.showHelp {
		return m.helpView()
	}

	var parts []string

	// Title
	parts = append(parts, m.styles.Title.Render("pwgen"))
	parts = append(parts, "")

	// Passphrase display
	passphrase := m.getPassphrase()

	parts = append(parts, m.styles.Passphrase.Render(passphrase))
	parts = append(parts, "")

	// Columns (slot machine view)
	columns := m.renderColumns()

	parts = append(parts, columns)
	parts = append(parts, "")

	// Status line
	status := m.renderStatus()

	parts = append(parts, status)
	parts = append(parts, "")

	// Instructions
	instructions := m.renderInstructions()

	parts = append(parts, instructions)

	return strings.Join(parts, "\n")
}

// renderColumns renders the slot machine columns.
func (m Model) renderColumns() string {
	if len(m.columns) == 0 {
		return "No columns"
	}

	columns := make([]string, 0, len(m.columns))
	for _, col := range m.columns {
		content := fmt.Sprintf("%s\n%s", col.Token.Type(), col.Value)

		var style lipgloss.Style

		switch {
		case col.Locked:
			style = m.styles.LockedColumn

			content += "\n🔒"
		case col.Focused:
			style = m.styles.FocusedColumn
		default:
			style = m.styles.Column
		}

		columns = append(columns, style.Render(content))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

// renderStatus renders the status line with entropy and strength information.
func (m Model) renderStatus() string {
	entropyStr := fmt.Sprintf("%.1f bits", m.entropy)
	lengthStr := fmt.Sprintf("%d chars", len(m.getPassphrase()))

	strengthStyle, ok := m.styles.StrengthStyles[m.strength]
	if !ok {
		strengthStyle = lipgloss.NewStyle()
	}

	strengthStr := strengthStyle.Render(m.strength)

	// Current configuration
	casingStr := m.casing.String()

	currentSep := m.separators[m.currentSep]
	switch currentSep {
	case "":
		currentSep = "none"
	case " ":
		currentSep = "space"
	}

	configStr := fmt.Sprintf("Config: %dw %dd %ds %s sep=%s",
		m.words, m.digits, m.symbols, casingStr, currentSep)

	status := fmt.Sprintf("Entropy: %s  |  Length: %s  |  Strength: %s  |  %s",
		entropyStr, lengthStr, strengthStr, configStr)

	return m.styles.Status.Render(status)
}

// renderInstructions renders the key instructions using bubbles help.
func (m Model) renderInstructions() string {
	// Use the help component's short help view
	return m.help.View(m.keys)
}

// helpView renders the help screen using bubbles help component.
func (m Model) helpView() string {
	var parts []string

	// Title
	parts = append(parts, m.styles.Title.Render("pwgen - Interactive Passphrase Generator"))
	parts = append(parts, "")

	// Key bindings help
	helpContent := m.help.View(m.keys)

	parts = append(parts, helpContent)
	parts = append(parts, "")

	// Additional instructions
	instructions := `COLUMN STATES:
  Normal      Column will regenerate when Space is pressed
  Focused     Column is currently selected (highlighted border)
  Locked 🔒   Column value is locked and won't change

CUSTOMIZATION CONTROLS:
  s           Cycle separators: - → _ → . → space → none → -
  w / W       Increase / decrease word count (1-10)
  d / D       Increase / decrease digit count (0-5)
  x / X       Increase / decrease symbol count (0-5)
  + / -       Universal increase / smart decrease
  ctrl+s      Cycle casing: mixed → lower → upper → title → mixed
  1-9         Lock/unlock specific columns

DISPLAY CONTROLS:
  v           Toggle passphrase visibility (masked/visible)
  c           Copy passphrase to clipboard
  n           Generate completely new passphrase (unlock all)

The status line shows current configuration, entropy, and strength.
Locked columns preserve their separators when cycling separators.

Press ? again to return to the main interface.`

	parts = append(parts, m.styles.Help.Render(instructions))

	return strings.Join(parts, "\n")
}
