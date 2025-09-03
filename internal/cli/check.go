// Package cli provides command-line interface functionality for the pwgen tool.
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/idelchi/pwgen/internal/generate"
	"github.com/idelchi/pwgen/internal/outfmt"
)

const (
	formatJSON = "json"
	formatText = "text"

	// stdinReadTimeout is the timeout for reading from stdin.
	stdinReadTimeout = 100 * time.Millisecond
)

// CheckOptions represents the configuration for the check command.
type CheckOptions struct {
	MinEntropy int
	MinLength  int
	JSON       bool
}

// Check returns the check command.
func Check() *cobra.Command {
	opts := &CheckOptions{}

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check entropy and policy compliance of a passphrase",
		Long: `Read a passphrase from stdin and analyze its entropy and policy compliance.

Calculates estimated entropy bits, checks length requirements, and provides
a security assessment. Useful for validating existing passphrases.`,
		Example: `  # Check a passphrase from stdin
  echo "correct-horse-battery-staple" | pwgen check

  # Check with minimum requirements
  echo "my-passphrase" | pwgen check --min-entropy 60 --min-length 20

  # Get results in JSON format
  echo "test123" | pwgen check --json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runCheck(opts)
		},
	}

	cmd.Flags().IntVar(&opts.MinEntropy, "min-entropy", opts.MinEntropy, "Minimum entropy requirement")
	cmd.Flags().IntVar(&opts.MinLength, "min-length", opts.MinLength, "Minimum length requirement")
	cmd.Flags().BoolVar(&opts.JSON, "json", opts.JSON, "Output in JSON format")

	cmd.Flags().SortFlags = false

	return cmd
}

// runCheck executes the passphrase analysis.
func runCheck(opts *CheckOptions) error {
	// Check if stdin has data with a short timeout
	done := make(chan bool, 1)

	var (
		passphrase string
		scanErr    error
	)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			passphrase = strings.TrimSpace(scanner.Text())
		}

		scanErr = scanner.Err()

		done <- true
	}()

	select {
	case <-done:
		if scanErr != nil {
			return fmt.Errorf("reading from stdin: %w", scanErr)
		}

		if passphrase == "" {
			return errors.New("no passphrase provided on stdin")
		}
	case <-time.After(stdinReadTimeout):
		return errors.New("no input provided on stdin")
	}

	// Analyze the passphrase
	calculator := generate.NewEntropyCalculator()
	analysis := calculator.CalculateEntropy(passphrase)

	// Check policy if requirements specified
	if opts.MinEntropy > 0 && analysis.Entropy < float64(opts.MinEntropy) {
		fmt.Fprintf(os.Stderr, "Policy violation: entropy %.1f < required %d\n",
			analysis.Entropy, opts.MinEntropy)
	}

	if opts.MinLength > 0 && analysis.Length < opts.MinLength {
		fmt.Fprintf(os.Stderr, "Policy violation: length %d < required %d\n",
			analysis.Length, opts.MinLength)
	}

	// Format output
	var format string

	if opts.JSON {
		format = formatJSON
	} else {
		format = formatText
	}

	formatter := outfmt.NewFormatter(format, os.Stdout, outfmt.Options{
		Colors:  !opts.JSON,
		Verbose: true,
	})

	return formatter.FormatAnalysis(analysis)
}
