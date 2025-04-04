package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mirrorboards/mctl/pkg/config"
	"github.com/mirrorboards/mctl/pkg/git"
	"github.com/spf13/cobra"
)

var (
	syncRemote        string
	syncMergeStrategy string
	syncRepos         []string
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync all repositories defined in mirror.toml",
		Long:  `Clone all repositories defined in mirror.toml that haven't been cloned yet. Optionally sync with a remote configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If a remote is specified, sync with it first
			if syncRemote != "" {
				fmt.Printf("Syncing with remote configuration %s...\n", syncRemote)
				if err := config.SyncWithRemote(syncRemote, syncMergeStrategy); err != nil {
					return fmt.Errorf("failed to sync with remote: %w", err)
				}
				fmt.Printf("Successfully synced with remote %s\n", syncRemote)
			}

			// Get repositories (after potential remote sync)
			repos, err := config.GetAllRepositories()
			if err != nil {
				return err
			}

			if len(repos) == 0 {
				fmt.Println("No repositories configured in mirror.toml")
				return nil
			}

			syncCount := 0
			fmt.Println("Syncing repositories...")

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

				// Skip if not in the specified repos list
				if len(syncRepos) > 0 {
					found := false
					for _, r := range syncRepos {
						if repo.ID == r || repo.Name == r {
							found = true
							break
						}
					}
					if !found {
						continue
					}
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

	cmd.Flags().StringVar(&syncRemote, "remote", "", "Remote configuration to sync with")
	cmd.Flags().StringVar(&syncMergeStrategy, "merge-strategy", "union", "Merge strategy for remote sync (remote-wins, local-wins, union)")
	cmd.Flags().StringSliceVar(&syncRepos, "repos", []string{}, "Only sync specified repositories (by ID or name)")

	return cmd
}
