package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/idelchi/pwgen/internal/clipboard"
	"github.com/idelchi/pwgen/internal/dictionary"
	"github.com/idelchi/pwgen/internal/generate"
	"github.com/idelchi/pwgen/internal/outfmt"
)

// GenOptions represents the configuration for the generate command.
type GenOptions struct {
	Words      int
	Sep        string
	Caps       string
	Digits     int
	Symbols    int
	Pattern    string
	Dict       string
	Kebab      bool
	Snake      bool
	Camel      bool
	Count      int
	JSON       bool
	Copy       bool
	MinEntropy int
	MinLength  int
}

const (
	// defaultWordCount is the default number of words to generate.
	defaultWordCount = 4
)

// Gen returns the generate command.
func Gen() *cobra.Command {
	opts := &GenOptions{
		Words:   defaultWordCount,
		Sep:     "-",
		Caps:    "mixed",
		Digits:  0,
		Symbols: 0,
		Dict:    "eff",
		Count:   1,
	}

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate passphrases non-interactively",
		Long: `Generate passphrases using configurable options.

Supports word-based generation with customizable separators, casing,
digits, symbols, and patterns. Output can be plain text or JSON format.`,
		Example: `  # Generate default passphrase (4 words, mixed case, hyphen-separated)
  pwgen gen

  # Generate with specific options
  pwgen gen --words 5 --sep "." --caps title --digits 2 --symbols 1

  # Generate using custom pattern
  pwgen gen --pattern "W:title SEP W:lower SEP DD{2} SEP S"

  # Generate multiple passphrases in JSON format
  pwgen gen --count 3 --json

  # Generate and copy to clipboard
  pwgen gen --copy`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runGenerate(opts)
		},
	}

	cmd.Flags().IntVar(&opts.Words, "words", opts.Words, "Number of words to generate")
	cmd.Flags().StringVar(&opts.Sep, "sep", opts.Sep, "Separator between tokens")
	cmd.Flags().StringVar(&opts.Caps, "caps", opts.Caps, "Casing style: mixed|lower|upper|title")
	cmd.Flags().IntVar(&opts.Digits, "digits", opts.Digits, "Number of digit tokens")
	cmd.Flags().IntVar(&opts.Symbols, "symbols", opts.Symbols, "Number of symbol tokens")
	cmd.Flags().StringVar(&opts.Pattern, "pattern", opts.Pattern, "Custom pattern (overrides other options)")
	cmd.Flags().StringVar(&opts.Dict, "dict", opts.Dict, "Dictionary to use: eff|small|path")
	cmd.Flags().BoolVar(&opts.Kebab, "kebab", opts.Kebab, "Use kebab-case separators")
	cmd.Flags().BoolVar(&opts.Snake, "snake", opts.Snake, "Use snake_case separators")
	cmd.Flags().BoolVar(&opts.Camel, "camel", opts.Camel, "Use camelCase (no separators)")
	cmd.Flags().IntVar(&opts.Count, "count", opts.Count, "Number of passphrases to generate")
	cmd.Flags().BoolVar(&opts.JSON, "json", opts.JSON, "Output in JSON format")
	cmd.Flags().BoolVar(&opts.Copy, "copy", opts.Copy, "Copy result to clipboard")
	cmd.Flags().IntVar(&opts.MinEntropy, "min-entropy", opts.MinEntropy, "Minimum entropy requirement")
	cmd.Flags().IntVar(&opts.MinLength, "min-length", opts.MinLength, "Minimum length requirement")

	cmd.Flags().SortFlags = false

	return cmd
}

// runGenerate executes the passphrase generation.
func runGenerate(opts *GenOptions) error {
	// Get dictionary
	dict, err := dictionary.GetDictionary(opts.Dict)
	if err != nil {
		return fmt.Errorf("loading dictionary: %w", err)
	}

	// Create generator
	generator := generate.NewGenerator(dict, opts.Sep)

	// Generate passphrases
	results, err := generator.Generate(generate.Options{
		Words:      opts.Words,
		Digits:     opts.Digits,
		Symbols:    opts.Symbols,
		Separator:  opts.Sep,
		Casing:     opts.Caps,
		Pattern:    opts.Pattern,
		Kebab:      opts.Kebab,
		Snake:      opts.Snake,
		Camel:      opts.Camel,
		Count:      opts.Count,
		MinEntropy: opts.MinEntropy,
		MinLength:  opts.MinLength,
	})
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Handle clipboard copy
	if opts.Copy && len(results) > 0 {
		if err := clipboard.Copy(results[0].Passphrase); err != nil {
			// Don't fail the command, just warn
			fmt.Fprintf(os.Stderr, "Warning: failed to copy to clipboard: %v\n", err)
		}
	}

	// Format output
	var format string

	if opts.JSON {
		format = "json"
	} else {
		format = "text"
	}

	formatter := outfmt.NewFormatter(format, os.Stdout, outfmt.Options{
		Colors: !opts.JSON,
	})

	return formatter.FormatResults(results)
}
