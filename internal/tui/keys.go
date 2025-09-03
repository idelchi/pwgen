// Package tui provides terminal user interface functionality for interactive passphrase generation.
package tui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines the key bindings for the TUI.
type KeyMap struct {
	Quit        key.Binding
	Help        key.Binding
	Generate    key.Binding
	ToggleLock  key.Binding
	FocusNext   key.Binding
	FocusPrev   key.Binding
	Copy        key.Binding
	ToggleView  key.Binding
	NewAll      key.Binding
	Separators  key.Binding
	Patterns    key.Binding
	WordsUp     key.Binding
	WordsDown   key.Binding
	DigitsUp    key.Binding
	DigitsDown  key.Binding
	SymbolsUp   key.Binding
	SymbolsDown key.Binding
	CycleCasing key.Binding
	IncreaseKey key.Binding
	DecreaseKey key.Binding
	Column1     key.Binding
	Column2     key.Binding
	Column3     key.Binding
	Column4     key.Binding
	Column5     key.Binding
	Column6     key.Binding
	Column7     key.Binding
	Column8     key.Binding
	Column9     key.Binding
}

// newKeyBindingWithHelp creates a key binding with keys and help text.
func newKeyBindingWithHelp(help, desc string, keys ...string) key.Binding {
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(help, desc))
}

// newColumnKeyBinding creates a column toggle key binding.
func newColumnKeyBinding(col int) key.Binding {
	colStr := strconv.Itoa(col)
	help := fmt.Sprintf("toggle column %d", col)

	return newKeyBindingWithHelp(colStr, help, colStr)
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:        newKeyBindingWithHelp("q", "quit", "q", "ctrl+c"),
		Help:        newKeyBindingWithHelp("?", "toggle help", "?"),
		Generate:    newKeyBindingWithHelp("space", "regenerate unlocked", " "),
		ToggleLock:  newKeyBindingWithHelp("enter", "lock/unlock column", "enter"),
		FocusNext:   newKeyBindingWithHelp("tab", "next column", "tab"),
		FocusPrev:   newKeyBindingWithHelp("shift+tab", "prev column", "shift+tab"),
		Copy:        newKeyBindingWithHelp("c", "copy to clipboard", "c"),
		ToggleView:  newKeyBindingWithHelp("v", "show/hide passphrase", "v"),
		NewAll:      newKeyBindingWithHelp("n", "new passphrase", "n"),
		Separators:  newKeyBindingWithHelp("s", "toggle separators", "s"),
		Patterns:    newKeyBindingWithHelp("p", "cycle patterns", "p"),
		WordsUp:     newKeyBindingWithHelp("w", "increase words", "w"),
		WordsDown:   newKeyBindingWithHelp("W", "decrease words", "W"),
		DigitsUp:    newKeyBindingWithHelp("d", "increase digits", "d"),
		DigitsDown:  newKeyBindingWithHelp("D", "decrease digits", "D"),
		SymbolsUp:   newKeyBindingWithHelp("x", "increase symbols", "x"),
		SymbolsDown: newKeyBindingWithHelp("X", "decrease symbols", "X"),
		IncreaseKey: newKeyBindingWithHelp("+", "increase (context sensitive)", "+", "="),
		DecreaseKey: newKeyBindingWithHelp("-", "decrease (context sensitive)", "-", "_"),
		CycleCasing: newKeyBindingWithHelp("ctrl+s", "cycle casing", "ctrl+s"),
		Column1:     newColumnKeyBinding(1),
		Column2:     newColumnKeyBinding(2), //nolint:mnd // column numbers are contextually clear
		Column3:     newColumnKeyBinding(3), //nolint:mnd // column numbers are contextually clear
		Column4:     newColumnKeyBinding(4), //nolint:mnd // column numbers are contextually clear
		Column5:     newColumnKeyBinding(5), //nolint:mnd // column numbers are contextually clear
		Column6:     newColumnKeyBinding(6), //nolint:mnd // column numbers are contextually clear
		Column7:     newColumnKeyBinding(7), //nolint:mnd // column numbers are contextually clear
		Column8:     newColumnKeyBinding(8), //nolint:mnd // column numbers are contextually clear
		Column9:     newColumnKeyBinding(9), //nolint:mnd // column numbers are contextually clear
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Generate, k.Separators, k.WordsUp, k.IncreaseKey, k.DecreaseKey, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Generate, k.ToggleLock, k.FocusNext, k.FocusPrev},
		{k.Copy, k.ToggleView, k.NewAll},
		{k.Separators, k.WordsUp, k.WordsDown, k.DigitsUp, k.DigitsDown},
		{k.SymbolsUp, k.SymbolsDown, k.IncreaseKey, k.DecreaseKey},
		{k.CycleCasing, k.Patterns},
		{k.Column1, k.Column2, k.Column3, k.Column4, k.Column5},
		{k.Column6, k.Column7, k.Column8, k.Column9},
		{k.Help, k.Quit},
	}
}
