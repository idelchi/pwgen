package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI application.
func Run() error {
	// Create model
	model, err := NewModel()
	if err != nil {
		return fmt.Errorf("creating TUI model: %w", err)
	}

	// Create program
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("TUI program failed: %w", err)
	}

	return nil
}
