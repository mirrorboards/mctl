package cmd

import (
	"fmt"

	"github.com/mirrorboards/mctl/pkg/config"
	"github.com/spf13/cobra"
)

var (
	removeDelete bool
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [id|name]",
		Short: "Remove a repository from the configuration",
		Long:  `Remove a repository from the mirror.toml configuration by ID or name. Optionally delete the repository files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			identifier := args[0]

			// Remove the repository
			if err := config.RemoveRepository(identifier, removeDelete); err != nil {
				return err
			}

			if removeDelete {
				fmt.Printf("Repository %s removed from configuration and files deleted\n", identifier)
			} else {
				fmt.Printf("Repository %s removed from configuration (files preserved)\n", identifier)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&removeDelete, "delete", false, "Delete repository files in addition to removing from configuration")

	return cmd
}
