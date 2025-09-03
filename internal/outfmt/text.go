package outfmt

import (
	"fmt"
	"io"
	"strings"

	"github.com/idelchi/pwgen/internal/generate"
)

// TextFormatter formats output as plain text.
type TextFormatter struct {
	writer  io.Writer
	verbose bool
	colors  bool
}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter(writer io.Writer, verbose, colors bool) *TextFormatter {
	return &TextFormatter{
		writer:  writer,
		verbose: verbose,
		colors:  colors,
	}
}

// FormatResults formats generation results as plain text.
func (f *TextFormatter) FormatResults(results []generate.Result) error {
	for i, result := range results {
		if i > 0 {
			if _, err := fmt.Fprintln(f.writer); err != nil {
				return err
			}
		}

		if f.verbose {
			if err := f.formatResultVerbose(result); err != nil {
				return err
			}
		} else {
			if err := f.formatResultSimple(result); err != nil {
				return err
			}
		}
	}

	return nil
}

// FormatAnalysis formats entropy analysis as plain text.
func (f *TextFormatter) FormatAnalysis(analysis generate.AnalysisResult) error {
	fmt.Fprintf(f.writer, "Passphrase Analysis\n")
	fmt.Fprintf(f.writer, "==================\n\n")

	fmt.Fprintf(f.writer, "Input: %s\n", analysis.Passphrase)
	fmt.Fprintf(f.writer, "Length: %d characters\n", analysis.Length)
	fmt.Fprintf(f.writer, "Character sets: %s\n", strings.Join(analysis.Charsets, ", "))
	fmt.Fprintf(f.writer, "Charset size: %d\n", analysis.CharsetSize)
	fmt.Fprintf(f.writer, "Entropy: %.1f bits\n", analysis.Entropy)
	fmt.Fprintf(f.writer, "Strength: %s\n", f.colorizeStrength(analysis.Strength))
	fmt.Fprintf(f.writer, "Estimated crack time: %s\n", analysis.CrackTime)

	if analysis.WordBased {
		fmt.Fprintf(f.writer, "Word-based structure detected: ~%d words\n", analysis.EstimatedWords)
	}

	if len(analysis.Patterns) > 0 {
		fmt.Fprintf(f.writer, "\nDetected patterns:\n")

		for _, pattern := range analysis.Patterns {
			fmt.Fprintf(f.writer, "  - %s: %s (-%0.1f bits)\n",
				pattern.Type, pattern.Description, pattern.Penalty)
		}
	}

	return nil
}

// FormatDictionaries formats dictionary information as plain text.
func (f *TextFormatter) FormatDictionaries(dicts []DictionaryInfo) error {
	fmt.Fprintf(f.writer, "Available Dictionaries\n")
	fmt.Fprintf(f.writer, "=====================\n\n")

	for _, dict := range dicts {
		fmt.Fprintf(f.writer, "Name: %s\n", dict.Name)
		fmt.Fprintf(f.writer, "Description: %s\n", dict.Description)
		fmt.Fprintf(f.writer, "Word count: %d\n", dict.WordCount)
		fmt.Fprintf(f.writer, "Entropy per word: %.1f bits\n", dict.EntropyBits)

		if dict.Path != "" {
			fmt.Fprintf(f.writer, "Path: %s\n", dict.Path)
		}

		fmt.Fprintf(f.writer, "Type: %s\n", dict.Type)
		fmt.Fprintf(f.writer, "\n")
	}

	return nil
}

// formatResultSimple outputs just the passphrase.
func (f *TextFormatter) formatResultSimple(result generate.Result) error {
	_, err := fmt.Fprintln(f.writer, result.Passphrase)

	return err
}

// formatResultVerbose outputs detailed information about the passphrase.
func (f *TextFormatter) formatResultVerbose(result generate.Result) error {
	if _, err := fmt.Fprintf(f.writer, "Passphrase: %s\n", result.Passphrase); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(f.writer, "Length: %d characters\n", result.Length); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(f.writer, "Entropy: %.1f bits\n", result.Entropy); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(f.writer, "Strength: %s\n", f.colorizeStrength(result.Strength)); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(f.writer, "Pattern: %s\n", result.Pattern); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(f.writer, "Estimated crack time: %s\n", result.CrackTime); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(f.writer, "Policy compliance: %s\n", f.colorizePolicyStatus(result.PolicyPass)); err != nil {
		return err
	}

	return nil
}

// colorizeStrength adds color codes to strength indicators if colors are enabled.
func (f *TextFormatter) colorizeStrength(strength string) string {
	if !f.colors {
		return strength
	}

	switch strength {
	case "Weak":
		return fmt.Sprintf("\033[31m%s\033[0m", strength) // Red
	case "Okay":
		return fmt.Sprintf("\033[33m%s\033[0m", strength) // Yellow
	case "Strong":
		return fmt.Sprintf("\033[32m%s\033[0m", strength) // Green
	case "Excellent":
		return fmt.Sprintf("\033[92m%s\033[0m", strength) // Bright green
	default:
		return strength
	}
}

// colorizePolicyStatus adds color codes to policy status if colors are enabled.
func (f *TextFormatter) colorizePolicyStatus(pass bool) string {
	status := "FAIL"

	if pass {
		status = "PASS"
	}

	if !f.colors {
		return status
	}

	if pass {
		return fmt.Sprintf("\033[32m%s\033[0m", status) // Green
	}

	return fmt.Sprintf("\033[31m%s\033[0m", status) // Red
}
