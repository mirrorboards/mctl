package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mirrorboards/mctl/pkg/config"
	"github.com/mirrorboards/mctl/pkg/git"
	"github.com/spf13/cobra"
)

func newClearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all repositories defined in mirror.toml",
		Long:  `Remove all directories created by mctl based on mirror.toml, but keep the configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := config.GetAllRepositories()
			if err != nil {
				return err
			}

			if len(repos) == 0 {
				fmt.Println("No repositories configured in mirror.toml")
				return nil
			}

			clearCount := 0
			dirsToCleanUp := make(map[string]bool)

			for _, repo := range repos {
				var fullPath string
				if repo.Path == "." {
					// Special case for current directory
					fullPath = repo.Name
				} else if repo.Name == "" {
					fullPath = repo.Path
				} else {
					fullPath = filepath.Join(repo.Path, repo.Name)
				}

				fmt.Printf("Removing %s...\n", fullPath)

				if err := git.RemoveDirectory(fullPath); err != nil {
					fmt.Printf("Failed to remove %s: %v\n", fullPath, err)
					continue
				}

				// Add directory to cleanup list
				dirsToCleanUp[fullPath] = true
				clearCount++
			}

			// Clean up empty parent directories
			fmt.Println("Cleaning up empty parent directories...")
			for dir := range dirsToCleanUp {
				if err := git.RemoveEmptyParentDirectories(dir); err != nil {
					fmt.Printf("Failed to clean up parent directories for %s: %v\n", dir, err)
				}
			}

			fmt.Printf("Clear complete: %d repositories removed\n", clearCount)
			return nil
		},
	}

	return cmd
}
