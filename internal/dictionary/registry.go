package dictionary

import (
	"errors"
	"fmt"
	"strings"
)

// BuiltinDict represents information about a built-in dictionary.
type BuiltinDict struct {
	Name        string
	Description string
	WordCount   int
	Factory     func() Dictionary
}

// builtinDictionaries contains all available built-in dictionaries.
//
//nolint:gochecknoglobals // Package-level registry for built-in dictionaries
var builtinDictionaries = map[string]BuiltinDict{
	"eff": {
		Name:        "eff",
		Description: "EFF Large Wordlist - 7776 cryptographically secure diceware words",
		WordCount:   0, // Will be set during initialization
		Factory:     ExternalEFF,
	},
}

// initializeWordCounts initializes word counts for built-in dictionaries on first access.
func initializeWordCounts() {
	for name, info := range builtinDictionaries {
		if info.WordCount == 0 {
			dict := info.Factory()

			info.WordCount = dict.Size()
			builtinDictionaries[name] = info
		}
	}
}

// GetBuiltin returns a built-in dictionary by name.
// Supported names: "eff".
//
//nolint:ireturn // Dictionary interface is the intended public API for polymorphism
func GetBuiltin(name string) (Dictionary, error) {
	initializeWordCounts()

	name = strings.ToLower(strings.TrimSpace(name))

	info, exists := builtinDictionaries[name]
	if !exists {
		return nil, fmt.Errorf("unknown built-in dictionary: %q", name)
	}

	return info.Factory(), nil
}

// ListBuiltin returns information about all built-in dictionaries.
func ListBuiltin() []BuiltinDict {
	initializeWordCounts()

	result := make([]BuiltinDict, 0, len(builtinDictionaries))

	for _, info := range builtinDictionaries {
		result = append(result, info)
	}

	return result
}

// GetDictionary returns a dictionary from a source specification.
// Source can be:
// - "eff" or "small" for built-in dictionaries
// - A file path for external dictionaries.
//
//nolint:ireturn // Dictionary interface is the intended public API for polymorphism
func GetDictionary(source string) (Dictionary, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, errors.New("dictionary source cannot be empty")
	}

	// Try built-in first.
	if dict, err := GetBuiltin(source); err == nil {
		return dict, nil
	}

	// Try as file path.
	return NewFromFile(source)
}
