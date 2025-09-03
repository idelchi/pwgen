package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/idelchi/pwgen/internal/dictionary"
	"github.com/idelchi/pwgen/internal/outfmt"
)

// DictsOptions represents the configuration for the dicts command.
type DictsOptions struct {
	JSON bool
}

// Dicts returns the dicts command.
func Dicts() *cobra.Command {
	opts := &DictsOptions{}

	cmd := &cobra.Command{
		Use:   "dicts",
		Short: "List available dictionaries",
		Long: `List all available dictionaries with their metadata.

Shows built-in dictionaries (eff, small) and any external dictionaries
that can be loaded from file paths. Includes word count and entropy information.`,
		Example: `  # List all dictionaries
  pwgen dicts

  # Get dictionary info in JSON format
  pwgen dicts --json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runDicts(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", opts.JSON, "Output in JSON format")

	cmd.Flags().SortFlags = false

	return cmd
}

// runDicts executes the dictionary listing.
func runDicts(opts *DictsOptions) error {
	// Get built-in dictionaries
	builtins := dictionary.ListBuiltin()

	dictInfos := make([]outfmt.DictionaryInfo, 0, len(builtins))

	for _, builtin := range builtins {
		info := outfmt.DictionaryInfoFromBuiltin(builtin)

		dictInfos = append(dictInfos, info)
	}

	// Format output
	var format string

	if opts.JSON {
		format = "json"
	} else {
		format = "text"
	}

	formatter := outfmt.NewFormatter(format, os.Stdout, outfmt.Options{
		Colors:  !opts.JSON,
		Verbose: true,
	})

	return formatter.FormatDictionaries(dictInfos)
}
