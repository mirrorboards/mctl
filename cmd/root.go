package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mctl",
		Short: "mctl - git repositories mesh management",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newVersionCmd(version)) // version subcommand
	cmd.AddCommand(newExampleCmd())        // example subcommand
	cmd.AddCommand(newInitCmd())           // init subcommand
	cmd.AddCommand(newAddCmd())            // add subcommand
	cmd.AddCommand(newSyncCmd())           // sync subcommand
	cmd.AddCommand(newClearCmd())          // clear subcommand
	cmd.AddCommand(newStatusCmd())         // status subcommand
	cmd.AddCommand(newSaveCmd())           // save subcommand

	return cmd
}

// Execute invokes the command.
func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
