package cli

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/idelchi/pwgen/internal/tui"
)

// Options represents the root level configuration for the CLI application.
type Options struct {
	// Verbose enables verbose output.
	Verbose bool
}

// Execute runs the root command for the pwgen CLI application.
func Execute(version string) error {
	root := &cobra.Command{
		Use:   "pwgen",
		Short: "Generate memorable and secure passphrases",
		Long: heredoc.Doc(`
			pwgen is a CLI tool for generating memorable yet secure passphrases
			with interactive TUI and non-interactive modes.

			Generate passphrases using words from built-in dictionaries with
			configurable patterns, casing, digits, and symbols. Features secure
			random generation, entropy calculation, and clipboard integration.
		`),
		Example: heredoc.Doc(`
			# Start interactive TUI mode
			pwgen

			# Generate a passphrase non-interactively
			pwgen gen --words 4 --sep "-" --caps mixed

			# Generate with custom pattern
			pwgen gen --pattern "W:title SEP W:lower SEP DD{2} SEP S"

			# Check entropy of existing passphrase
			echo "correct-horse-battery-staple" | pwgen check --min-entropy 60
		`),
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			// Default to TUI mode if no subcommand specified
			return runTUI()
		},
	}

	root.SetVersionTemplate("{{ .Version }}\n")
	root.SetHelpCommand(&cobra.Command{Hidden: true})

	root.Flags().SortFlags = false
	root.PersistentFlags().SortFlags = false

	root.CompletionOptions.DisableDefaultCmd = true
	cobra.EnableCommandSorting = false

	root.AddCommand(
		Gen(),
		Check(),
		Dicts(),
		Version(),
	)

	return root.Execute()
}

// runTUI starts the interactive TUI mode.
func runTUI() error {
	return tui.Run()
}
