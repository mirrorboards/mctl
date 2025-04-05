package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mctl",
		Short: "mctl - Multi-Repository Control System",
		Long: `MCTL provides secure, unified management of code repositories in high-security environments.
It implements a structured management layer over Git repositories, providing consistent 
operations across multiple codebases while maintaining comprehensive metadata and audit capabilities.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add all subcommands
	cmd.AddCommand(newVersionCmd(version))

	return cmd
}

// Execute invokes the command.
func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
