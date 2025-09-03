// Package outfmt provides output formatting functionality for different formats.
package outfmt

import (
	"io"
	"math"

	"github.com/idelchi/pwgen/internal/dictionary"
	"github.com/idelchi/pwgen/internal/generate"
)

// Formatter defines the interface for output formatting.
type Formatter interface {
	// FormatResults formats passphrase generation results.
	FormatResults(results []generate.Result) error

	// FormatAnalysis formats entropy analysis results.
	FormatAnalysis(analysis generate.AnalysisResult) error

	// FormatDictionaries formats dictionary information.
	FormatDictionaries(dicts []DictionaryInfo) error
}

// DictionaryInfo represents information about a dictionary for display.
type DictionaryInfo struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	WordCount   int     `json:"wordCount"`
	EntropyBits float64 `json:"entropyBits"`
	Path        string  `json:"path,omitempty"`
	Type        string  `json:"type"`
}

// NewFormatter creates a new formatter based on the specified type.
//
//nolint:ireturn // Formatter interface is required for polymorphism in output formatting
func NewFormatter(format string, writer io.Writer, options Options) Formatter {
	switch format {
	case "json":
		return NewJSONFormatter(writer, options.Pretty)
	case "text", "":
		return NewTextFormatter(writer, options.Verbose, options.Colors)
	default:
		// Default to text format
		return NewTextFormatter(writer, options.Verbose, options.Colors)
	}
}

// Options configures formatter behavior.
type Options struct {
	Verbose bool
	Pretty  bool
	Colors  bool
}

// DictionaryInfoFromDict creates DictionaryInfo from a Dictionary.
func DictionaryInfoFromDict(dict dictionary.Dictionary, dictType string) DictionaryInfo {
	return DictionaryInfo{
		Name:        dict.Name(),
		Description: dict.Name() + " dictionary",
		WordCount:   dict.Size(),
		EntropyBits: dict.EntropyBits(),
		Type:        dictType,
	}
}

// DictionaryInfoFromBuiltin creates DictionaryInfo from builtin dictionary info.
func DictionaryInfoFromBuiltin(info dictionary.BuiltinDict) DictionaryInfo {
	return DictionaryInfo{
		Name:        info.Name,
		Description: info.Description,
		WordCount:   info.WordCount,
		EntropyBits: calculateEntropyBits(info.WordCount),
		Type:        "builtin",
	}
}

// DictionaryInfoFromFile creates DictionaryInfo for file-based dictionaries.
func DictionaryInfoFromFile(path string, dict dictionary.Dictionary) DictionaryInfo {
	return DictionaryInfo{
		Name:        dict.Name(),
		Description: "External dictionary from " + path,
		WordCount:   dict.Size(),
		EntropyBits: dict.EntropyBits(),
		Path:        path,
		Type:        "file",
	}
}

// calculateEntropyBits calculates entropy bits from word count.
func calculateEntropyBits(wordCount int) float64 {
	if wordCount <= 0 {
		return 0
	}
	// log2(wordCount)
	return math.Log2(float64(wordCount))
}
