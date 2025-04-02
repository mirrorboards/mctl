package cmd

import (
	"fmt"

	"github.com/mirrorboards/mctl/pkg/config"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize an empty mirror.toml file",
		Long:  `Initialize an empty mirror.toml file in the current directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.InitConfig(); err != nil {
				return err
			}
			fmt.Println("Initialized empty mirror.toml file")
			return nil
		},
	}

	return cmd
}
