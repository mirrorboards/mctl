package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/logging"
	"github.com/mirrorboards/mctl/internal/repository"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	var (
		repos      string
		parallel   int
		force      bool
		dryRun     bool
		fetchOnly  bool
		autoRemove bool
	)

	cmd := &cobra.Command{
		Use:   "sync [options]",
		Short: "Update repositories to match remote state",
		Long: `Update repositories to match remote state.

This command synchronizes repositories with their remote state.
It fetches updates from the remote and merges them into the local repository.

If a repository directory does not exist, it will be cloned from the remote.
You can use the --auto-remove flag to automatically remove repositories that
cannot be cloned (e.g., if the remote repository no longer exists).

Examples:
  mctl sync
  mctl sync --repos=secure-comms,authentication
  mctl sync --parallel=8
  mctl sync --force
  mctl sync --dry-run
  mctl sync --fetch-only
  mctl sync --auto-remove`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(repos, parallel, force, dryRun, fetchOnly, autoRemove)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&repos, "repos", "", "Limit to specific repositories (comma-separated)")
	cmd.Flags().IntVar(&parallel, "parallel", 4, "Number of concurrent operations")
	cmd.Flags().BoolVar(&force, "force", false, "Override local changes warning")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Report actions without execution")
	cmd.Flags().BoolVar(&fetchOnly, "fetch-only", false, "Update remote references without merging")
	cmd.Flags().BoolVar(&autoRemove, "auto-remove", false, "Automatically remove repositories that cannot be cloned from configuration")

	return cmd
}

func runSync(repos string, parallel int, force, dryRun, fetchOnly, autoRemove bool) error {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to get current directory")
	}

	// Load configuration
	cfg, err := config.LoadConfig(currentDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigNotFound, "Failed to load configuration")
	}

	// Create repository manager
	repoManager := repository.NewManager(cfg, currentDir)

	// Get repositories to sync
	var repositories []*repository.Repository
	if repos == "" {
		// Get all repositories
		repositories, err = repoManager.GetAllRepositories()
		if err != nil {
			return errors.Wrap(err, errors.ErrInternalError, "Failed to get repositories")
		}
	} else {
		// Get specified repositories
		repoNames := strings.Split(repos, ",")
		for _, name := range repoNames {
			repo, err := repoManager.GetRepository(strings.TrimSpace(name))
			if err != nil {
				return errors.Wrap(err, errors.ErrRepositoryNotFound, fmt.Sprintf("Repository not found: %s", name))
			}
			repositories = append(repositories, repo)
		}
	}

	// Create logger
	logger := logging.NewLogger(currentDir)

	// Limit parallel operations
	if parallel <= 0 {
		parallel = 1
	}
	semaphore := make(chan struct{}, parallel)
	var wg sync.WaitGroup

	// Track results
	type syncResult struct {
		Name       string
		Success    bool
		Error      error
		NotExist   bool // Repository directory does not exist
		Repository *repository.Repository
	}
	results := make([]syncResult, len(repositories))

	// Sync repositories
	for i, repo := range repositories {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(i int, repo *repository.Repository) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			// Check if repository directory exists
			if _, err := os.Stat(repo.FullPath()); os.IsNotExist(err) {
				// Repository doesn't exist, clone it
				logger.LogOperation(logging.LogLevelInfo, fmt.Sprintf("Cloning repository %s (directory does not exist)", repo.Config.Name))

				if dryRun {
					fmt.Printf("Would clone repository %s\n", repo.Config.Name)
					results[i] = syncResult{
						Name:    repo.Config.Name,
						Success: true,
					}
					return
				}

				// Clone the repository
				if err := repo.Clone(); err != nil {
					results[i] = syncResult{
						Name:       repo.Config.Name,
						Success:    false,
						Error:      fmt.Errorf("failed to clone repository: %w", err),
						NotExist:   true,
						Repository: repo,
					}
					return
				}

				// Successfully cloned
				results[i] = syncResult{
					Name:    repo.Config.Name,
					Success: true,
					Error:   nil,
				}
				return
			}

			// Update status
			if err := repo.UpdateStatus(); err != nil {
				results[i] = syncResult{
					Name:    repo.Config.Name,
					Success: false,
					Error:   fmt.Errorf("failed to update status: %w", err),
				}
				return
			}

			// Check for local changes
			if !force && repo.Metadata.Status.Current == repository.StatusModified {
				results[i] = syncResult{
					Name:    repo.Config.Name,
					Success: false,
					Error:   fmt.Errorf("repository has uncommitted changes (use --force to override)"),
				}
				return
			}

			// Log operation
			logger.LogOperation(logging.LogLevelInfo, fmt.Sprintf("Syncing repository %s", repo.Config.Name))

			if dryRun {
				fmt.Printf("Would sync repository %s\n", repo.Config.Name)
				results[i] = syncResult{
					Name:    repo.Config.Name,
					Success: true,
				}
				return
			}

			// Fetch only
			if fetchOnly {
				if err := repo.Fetch(); err != nil {
					results[i] = syncResult{
						Name:    repo.Config.Name,
						Success: false,
						Error:   fmt.Errorf("failed to fetch: %w", err),
					}
					return
				}

				results[i] = syncResult{
					Name:    repo.Config.Name,
					Success: true,
				}
				return
			}

			// Sync repository
			if err := repo.Sync(); err != nil {
				results[i] = syncResult{
					Name:    repo.Config.Name,
					Success: false,
					Error:   fmt.Errorf("failed to sync: %w", err),
				}
				return
			}

			results[i] = syncResult{
				Name:    repo.Config.Name,
				Success: true,
			}
		}(i, repo)
	}

	// Wait for all operations to complete
	wg.Wait()

	// Handle auto-remove for repositories that don't exist
	if autoRemove {
		var reposToRemove []*repository.Repository
		for _, result := range results {
			if result.NotExist {
				reposToRemove = append(reposToRemove, result.Repository)
			}
		}

		// Remove repositories from configuration
		for _, repo := range reposToRemove {
			if dryRun {
				fmt.Printf("Would remove repository %s from configuration\n", repo.Config.Name)
				continue
			}

			// Log the operation
			logger.LogOperation(logging.LogLevelInfo, fmt.Sprintf("Removing repository %s (failed to clone)", repo.Config.Name))
			logger.LogAudit(logging.LogLevelInfo, fmt.Sprintf("Repository removed: %s", repo.Config.Name))

			// Remove repository from configuration
			if err := repoManager.RemoveRepository(repo.Config.ID, false); err != nil {
				fmt.Printf("✗ Failed to remove %s from configuration: %v\n", repo.Config.Name, err)
			} else {
				fmt.Printf("✓ Removed %s from configuration (failed to clone)\n", repo.Config.Name)
			}
		}
	}

	// Report results
	successCount := 0
	for _, result := range results {
		if result.Success {
			fmt.Printf("✓ %s: Synchronized successfully\n", result.Name)
			successCount++
		} else if result.NotExist && autoRemove {
			// Already reported above
		} else if result.NotExist {
			fmt.Printf("✗ %s: %v (use --auto-remove to remove failed repositories from configuration)\n", result.Name, result.Error)
		} else {
			fmt.Printf("✗ %s: %v\n", result.Name, result.Error)
		}
	}

	fmt.Printf("\nSynchronized %d/%d repositories\n", successCount, len(repositories))

	// Return error if any repository failed to sync and wasn't auto-removed
	if successCount < len(repositories) && (!autoRemove || successCount == 0) {
		return errors.New(errors.ErrGitPullFailed, "One or more repositories failed to synchronize")
	}

	return nil
}
