package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/mirrorboards/mctl/pkg/config"
	"github.com/spf13/cobra"
)

var (
	remoteType     string
	remoteBranch   string
	remoteAuthType string
	remoteForce    bool
	remoteMessage  string
	mergeStrategy  string
)

func newRemoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remote",
		Short: "Manage remote configuration sources",
		Long:  `Manage remote configuration sources for synchronizing mirror.toml files across repositories.`,
	}

	// Add subcommands
	cmd.AddCommand(newRemoteAddCmd())
	cmd.AddCommand(newRemoteListCmd())
	cmd.AddCommand(newRemoteRemoveCmd())
	cmd.AddCommand(newRemotePullCmd())
	cmd.AddCommand(newRemotePushCmd())

	return cmd
}

func newRemoteAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [name] [url]",
		Short: "Add a remote configuration source",
		Long:  `Add a remote configuration source for synchronizing mirror.toml files.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			url := args[1]

			if err := config.AddRemote(name, url, remoteType, remoteBranch, remoteAuthType); err != nil {
				return err
			}

			fmt.Printf("Remote %s added successfully\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&remoteType, "type", "", "Type of remote (github, gitlab, bitbucket, file)")
	cmd.Flags().StringVar(&remoteBranch, "branch", "", "Branch to use for the remote")
	cmd.Flags().StringVar(&remoteAuthType, "auth", "none", "Authentication type (ssh, token, none)")

	return cmd
}

func newRemoteListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List remote configuration sources",
		Long:  `List all remote configuration sources configured in mirror.toml.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			remotes, err := config.GetAllRemotes()
			if err != nil {
				return err
			}

			if len(remotes) == 0 {
				fmt.Println("No remote configuration sources configured")
				return nil
			}

			// Define colors
			titleColor := color.New(color.FgHiWhite, color.Bold)
			nameColor := color.New(color.FgHiCyan, color.Bold)
			urlColor := color.New(color.FgHiBlue)
			typeColor := color.New(color.FgHiYellow)
			branchColor := color.New(color.FgHiGreen)
			authColor := color.New(color.FgHiMagenta)

			// Print header
			titleColor.Println("\n✨ REMOTE CONFIGURATION SOURCES ✨")
			fmt.Println(strings.Repeat("─", 60))

			for _, remote := range remotes {
				nameColor.Printf("• %s\n", remote.Name)
				urlColor.Printf("  URL: %s\n", remote.URL)
				if remote.Type != "" {
					typeColor.Printf("  Type: %s\n", remote.Type)
				}
				if remote.Branch != "" {
					branchColor.Printf("  Branch: %s\n", remote.Branch)
				}
				authColor.Printf("  Auth: %s\n", remote.AuthType)
				fmt.Println(strings.Repeat("─", 60))
			}

			return nil
		},
	}

	return cmd
}

func newRemoteRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a remote configuration source",
		Long:  `Remove a remote configuration source from mirror.toml.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := config.RemoveRemote(name); err != nil {
				return err
			}

			fmt.Printf("Remote %s removed successfully\n", name)
			return nil
		},
	}

	return cmd
}

func newRemotePullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull [name]",
		Short: "Pull configuration from a remote source",
		Long:  `Pull and merge configuration from a remote source into the local mirror.toml.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := config.SyncWithRemote(name, mergeStrategy); err != nil {
				return err
			}

			fmt.Printf("Successfully pulled configuration from remote %s\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&mergeStrategy, "merge-strategy", "union", "Merge strategy (remote-wins, local-wins, union)")

	return cmd
}

func newRemotePushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push [name]",
		Short: "Push configuration to a remote source",
		Long:  `Push local configuration to a remote source.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if remoteMessage == "" {
				remoteMessage = "Update mirror.toml configuration"
			}

			if err := config.PushToRemote(name, remoteForce, remoteMessage); err != nil {
				return err
			}

			fmt.Printf("Successfully pushed configuration to remote %s\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&remoteForce, "force", false, "Force push (overwrite remote changes)")
	cmd.Flags().StringVar(&remoteMessage, "message", "", "Commit message for the push")

	return cmd
}
