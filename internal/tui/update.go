package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/idelchi/pwgen/internal/clipboard"
)

// Update implements the Bubble Tea Model interface.
//
//nolint:ireturn // tea.Model interface is required by Bubble Tea framework
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		return m.handleKeyMessage(msg)
	default:
		return m, nil
	}
}

// handleWindowSize handles window size messages.
//
//nolint:ireturn // tea.Model interface is required by Bubble Tea framework
func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.help.Width = msg.Width

	return m, nil
}

// handleKeyMessage handles key press messages.
//
//nolint:ireturn // tea.Model interface is required by Bubble Tea framework
func (m Model) handleKeyMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		(&m).cleanup()

		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp

		return m, nil
	case key.Matches(msg, m.keys.ToggleView):
		m.masked = !m.masked

		return m, nil
	default:
		return m.handleActionKeys(msg)
	}
}

// handleActionKeys handles action-based key presses.
//
//nolint:ireturn // tea.Model interface is required by Bubble Tea framework
func (m Model) handleActionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Generate):
		if err := (&m).generatePassphrase(); err != nil {
			return m, nil
		}

		return m, nil
	case key.Matches(msg, m.keys.ToggleLock):
		(&m).toggleColumnLock(m.focused)

		return m, nil
	case key.Matches(msg, m.keys.FocusNext):
		(&m).moveFocus(1)

		return m, nil
	case key.Matches(msg, m.keys.FocusPrev):
		(&m).moveFocus(-1)

		return m, nil
	case key.Matches(msg, m.keys.Copy):
		return m.handleCopy()
	case key.Matches(msg, m.keys.NewAll):
		return m.handleNewAll()
	default:
		return m.handleConfigKeys(msg)
	}
}

// handleConfigKeys handles configuration change keys.
//
//nolint:ireturn // tea.Model interface is required by Bubble Tea framework
func (m Model) handleConfigKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Separators):
		if err := (&m).cycleSeparator(); err != nil {
			return m, nil
		}

		return m, nil
	case key.Matches(msg, m.keys.Patterns):
		return m, nil
	case key.Matches(msg, m.keys.CycleCasing):
		if err := (&m).cycleCasing(); err != nil {
			return m, nil
		}

		return m, nil
	default:
		return m.handleAdjustmentKeys(msg)
	}
}

// handleAdjustmentKeys handles increment/decrement keys.
//
//nolint:ireturn // tea.Model interface is required by Bubble Tea framework
func (m Model) handleAdjustmentKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.WordsUp):
		if err := (&m).adjustWords(1); err != nil {
			return m, nil
		}

		return m, nil
	case key.Matches(msg, m.keys.WordsDown):
		if err := (&m).adjustWords(-1); err != nil {
			return m, nil
		}

		return m, nil
	case key.Matches(msg, m.keys.DigitsUp):
		if err := (&m).adjustDigits(1); err != nil {
			return m, nil
		}

		return m, nil
	case key.Matches(msg, m.keys.DigitsDown):
		if err := (&m).adjustDigits(-1); err != nil {
			return m, nil
		}

		return m, nil
	case key.Matches(msg, m.keys.SymbolsUp):
		if err := (&m).adjustSymbols(1); err != nil {
			return m, nil
		}

		return m, nil
	case key.Matches(msg, m.keys.SymbolsDown):
		if err := (&m).adjustSymbols(-1); err != nil {
			return m, nil
		}

		return m, nil
	case key.Matches(msg, m.keys.IncreaseKey):
		if err := (&m).adjustWords(1); err != nil {
			return m, nil
		}

		return m, nil
	case key.Matches(msg, m.keys.DecreaseKey):
		return m.handleUniversalDecrease()
	default:
		return m.handleColumnKeys(msg)
	}
}

// handleCopy handles copy to clipboard.
//
//nolint:ireturn // tea.Model interface is required by Bubble Tea framework
func (m Model) handleCopy() (tea.Model, tea.Cmd) {
	if m.passphrase != nil {
		passphrase := m.passphrase.String()
		// Copy to clipboard - errors are silently ignored for now
		// In the future, this could show status messages in the UI
		_ = clipboard.Copy(passphrase)
	}

	return m, nil
}

// handleNewAll handles new passphrase generation.
//
//nolint:ireturn // tea.Model interface is required by Bubble Tea framework
func (m Model) handleNewAll() (tea.Model, tea.Cmd) {
	(&m).unlockAll()

	if err := (&m).generatePassphrase(); err != nil {
		return m, nil
	}

	return m, nil
}

// handleUniversalDecrease handles universal decrease key.
//
//nolint:ireturn // tea.Model interface is required by Bubble Tea framework
func (m Model) handleUniversalDecrease() (tea.Model, tea.Cmd) {
	switch {
	case m.words > 1:
		if err := (&m).adjustWords(-1); err != nil {
			return m, nil
		}
	case m.digits > 0:
		if err := (&m).adjustDigits(-1); err != nil {
			return m, nil
		}
	case m.symbols > 0:
		if err := (&m).adjustSymbols(-1); err != nil {
			return m, nil
		}
	}

	return m, nil
}

// handleColumnKeys handles column selection keys.
//
//nolint:ireturn // tea.Model interface is required by Bubble Tea framework
func (m Model) handleColumnKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	columnKeys := []key.Binding{
		m.keys.Column1, m.keys.Column2, m.keys.Column3,
		m.keys.Column4, m.keys.Column5, m.keys.Column6,
		m.keys.Column7, m.keys.Column8, m.keys.Column9,
	}

	for colIndex, keyBinding := range columnKeys {
		if key.Matches(msg, keyBinding) && len(m.columns) > colIndex {
			(&m).toggleColumnLock(colIndex)

			return m, nil
		}
	}

	return m, nil
}

// unlockAll unlocks all columns.
func (m *Model) unlockAll() {
	for i := range m.columns {
		m.columns[i].Locked = false
	}
}
