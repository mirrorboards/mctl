package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mirrorboards/mctl/pkg/config"
	"github.com/mirrorboards/mctl/pkg/git"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync all repositories defined in mirror.toml",
		Long:  `Clone all repositories defined in mirror.toml that haven't been cloned yet.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := config.GetAllRepositories()
			if err != nil {
				return err
			}

			if len(repos) == 0 {
				fmt.Println("No repositories configured in mirror.toml")
				return nil
			}

			syncCount := 0

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

				// Skip if the repository already exists
				gitDir := filepath.Join(fullPath, ".git")
				if _, err := os.Stat(gitDir); err == nil {
					fmt.Printf("Repository already exists at %s, skipping\n", fullPath)
					continue
				}

				fmt.Printf("Cloning %s into %s...\n", repo.URL, fullPath)

				// Determine path and name for cloning
				clonePath := repo.Path
				cloneName := repo.Name

				if err := git.Clone(repo.URL, clonePath, cloneName); err != nil {
					fmt.Printf("Failed to clone %s: %v\n", repo.URL, err)
					continue
				}

				syncCount++
			}

			fmt.Printf("Sync complete: %d repositories cloned\n", syncCount)
			return nil
		},
	}

	return cmd
}
