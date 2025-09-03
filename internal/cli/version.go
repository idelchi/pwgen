package cli

import (
	"github.com/spf13/cobra"
)

// Version returns the version command.
func Version() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "version",
		Short:  "Show version information",
		Long:   `Display the current version of pwgen.`,
		Hidden: true, // Hide since version is available via --version flag
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Version is handled by root command
			return cmd.Root().RunE(cmd, []string{"--version"})
		},
	}

	return cmd
}
