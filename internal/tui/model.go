package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/idelchi/pwgen/internal/dictionary"
	"github.com/idelchi/pwgen/internal/generate"
	"github.com/idelchi/pwgen/internal/safety"
)

const (
	// UI constraints.
	maxWords   = 10
	maxDigits  = 5
	maxSymbols = 5
	minWords   = 1
	minDigits  = 0
	minSymbols = 0

	// Default configuration.
	defaultWords = 4
)

// Model represents the TUI application state.
//
//nolint:recvcheck // Mixed receivers required by Bubble Tea framework patterns
type Model struct {
	// Generator and dictionary
	generator  *generate.Generator
	dictionary dictionary.Dictionary

	// Current passphrase
	passphrase *safety.SecureString
	pattern    *generate.Pattern
	entropy    float64
	strength   string

	// UI state
	masked   bool
	focused  int
	columns  []Column
	showHelp bool

	// Configuration state
	words      int
	digits     int
	symbols    int
	casing     generate.CaseStyle
	separators []string
	currentSep int

	// Key bindings and help
	keys   KeyMap
	help   help.Model
	styles StyleSet

	// Window dimensions
	width  int
	height int
}

// Column represents a token column in the slot-machine interface.
type Column struct {
	Token   generate.Token
	Value   string
	Locked  bool
	Focused bool
}

// NewModel creates a new TUI model.
func NewModel() (*Model, error) {
	// Load default dictionary
	dict, err := dictionary.GetDictionary("eff")
	if err != nil {
		return nil, err
	}

	// Create generator
	generator := generate.NewGenerator(dict, "-")

	// Initialize with default pattern
	builder := generate.NewPatternBuilder(dict, "-")

	pattern, err := builder.BuildFromOptions(defaultWords, minDigits, minSymbols, "mixed", "-", false, false, false)
	if err != nil {
		return nil, err
	}

	// Create columns from pattern
	columns := make([]Column, len(pattern.Tokens))
	for i, token := range pattern.Tokens {
		columns[i] = Column{
			Token:   token,
			Locked:  false,
			Focused: i == 0, // Focus first column initially
		}
	}

	model := &Model{
		generator:  generator,
		dictionary: dict,
		pattern:    pattern,
		masked:     true,
		focused:    0,
		columns:    columns,
		showHelp:   false,

		// Default configuration
		words:      defaultWords,
		digits:     minDigits,
		symbols:    minSymbols,
		casing:     generate.CaseMixed,
		separators: []string{"-", "_", ".", " ", ""},
		currentSep: 0,

		keys:   DefaultKeyMap(),
		help:   help.New(),
		styles: NewStyleSet(),
	}

	// Generate initial passphrase
	if err := model.generatePassphrase(); err != nil {
		return nil, err
	}

	return model, nil
}

// Init implements the Bubble Tea Model interface.
func (m Model) Init() tea.Cmd {
	return nil
}

// generatePassphrase creates a new passphrase by regenerating unlocked columns.
func (m *Model) generatePassphrase() error {
	parts := make([]string, 0, len(m.columns))

	for i, column := range m.columns { //nolint:varnamelen // i is standard loop var
		if !column.Locked {
			value, err := column.Token.Generate()
			if err != nil {
				return err
			}

			m.columns[i].Value = value
		}

		parts = append(parts, m.columns[i].Value)
	}

	// Join all parts to create the passphrase - tokens already include separators
	passphraseStr := strings.Join(parts, "")

	// Store in secure string
	if m.passphrase != nil {
		m.passphrase.Wipe()
	}

	m.passphrase = safety.NewSecureString(passphraseStr)

	// Calculate entropy and strength
	m.entropy = m.pattern.EntropyBits()
	m.strength = m.calculateStrength(m.entropy)

	return nil
}

// calculateStrength returns a human-readable strength assessment.
func (m Model) calculateStrength(entropy float64) string {
	switch {
	case entropy < generate.WeakEntropyThreshold:
		return "Weak"
	case entropy < generate.OkayEntropyThreshold:
		return "Okay"
	case entropy < generate.StrongEntropyThreshold:
		return "Strong"
	default:
		return "Excellent"
	}
}

// toggleColumnLock toggles the lock state of a column.
func (m *Model) toggleColumnLock(index int) {
	if index >= 0 && index < len(m.columns) {
		m.columns[index].Locked = !m.columns[index].Locked
	}
}

// moveFocus moves the focus to a different column.
func (m *Model) moveFocus(direction int) {
	if len(m.columns) == 0 {
		return
	}

	m.columns[m.focused].Focused = false

	m.focused += direction

	if m.focused < 0 {
		m.focused = len(m.columns) - 1
	} else if m.focused >= len(m.columns) {
		m.focused = 0
	}

	m.columns[m.focused].Focused = true
}

// getPassphrase returns the current passphrase (respecting masking).
func (m Model) getPassphrase() string {
	if m.passphrase == nil {
		return ""
	}

	if m.masked {
		return safety.MaskString(m.passphrase.String(), '•')
	}

	return m.passphrase.String()
}

// cycleSeparator cycles to the next separator and regenerates the pattern.
func (m *Model) cycleSeparator() error {
	m.currentSep = (m.currentSep + 1) % len(m.separators)

	return m.regeneratePattern()
}

// regeneratePattern rebuilds the pattern based on current configuration.
func (m *Model) regeneratePattern() error {
	currentSep := m.separators[m.currentSep]
	builder := generate.NewPatternBuilder(m.dictionary, currentSep)

	var casingStr string

	switch m.casing {
	case generate.CaseLower:
		casingStr = "lower"
	case generate.CaseUpper:
		casingStr = "upper"
	case generate.CaseTitle:
		casingStr = "title"
	case generate.CaseMixed:
		casingStr = "mixed"
	}

	pattern, err := builder.BuildFromOptions(
		m.words, m.digits, m.symbols,
		casingStr, currentSep,
		false, false, false, // kebab, snake, camel
	)
	if err != nil {
		return err
	}

	m.pattern = pattern

	// Recreate columns from new pattern
	m.columns = make([]Column, len(pattern.Tokens))
	for i, token := range pattern.Tokens {
		m.columns[i] = Column{
			Token:   token,
			Locked:  false,
			Focused: i == m.focused,
		}
	}

	// Ensure focus is within bounds
	if m.focused >= len(m.columns) {
		m.focused = 0
		if len(m.columns) > 0 {
			m.columns[0].Focused = true
		}
	}

	// Generate new passphrase
	return m.generatePassphrase()
}

// adjustWords changes the word count and regenerates the pattern.
func (m *Model) adjustWords(delta int) error {
	newWords := m.words + delta
	if newWords < minWords {
		newWords = minWords
	}

	if newWords > maxWords {
		newWords = maxWords
	}

	m.words = newWords

	return m.regeneratePattern()
}

// adjustDigits changes the digit count and regenerates the pattern.
func (m *Model) adjustDigits(delta int) error {
	newDigits := m.digits + delta
	if newDigits < minDigits {
		newDigits = minDigits
	}

	if newDigits > maxDigits {
		newDigits = maxDigits
	}

	m.digits = newDigits

	return m.regeneratePattern()
}

// adjustSymbols changes the symbol count and regenerates the pattern.
func (m *Model) adjustSymbols(delta int) error {
	newSymbols := m.symbols + delta
	if newSymbols < minSymbols {
		newSymbols = minSymbols
	}

	if newSymbols > maxSymbols {
		newSymbols = maxSymbols
	}

	m.symbols = newSymbols

	return m.regeneratePattern()
}

// cycleCasing cycles to the next casing style and regenerates the pattern.
func (m *Model) cycleCasing() error {
	switch m.casing {
	case generate.CaseLower:
		m.casing = generate.CaseUpper
	case generate.CaseUpper:
		m.casing = generate.CaseTitle
	case generate.CaseTitle:
		m.casing = generate.CaseMixed
	case generate.CaseMixed:
		m.casing = generate.CaseLower
	}

	return m.regeneratePattern()
}

// cleanup performs cleanup when the model is destroyed.
func (m *Model) cleanup() {
	if m.passphrase != nil {
		m.passphrase.Wipe()
	}
}
