package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/logging"
	"github.com/mirrorboards/mctl/internal/repository"
	"github.com/mirrorboards/mctl/internal/snapshot"
	"github.com/spf13/cobra"
)

func newSaveCmd() *cobra.Command {
	var (
		repos       string
		noPush      bool
		amend       bool
		all         bool
		sign        bool
		noSnapshot  bool
		description string
	)

	cmd := &cobra.Command{
		Use:   "save [options] \"commit-message\"",
		Short: "Commit and push changes across repositories",
		Long: `Commit and push changes across repositories.

This command commits changes in the specified repositories and optionally
pushes them to the remote. If no repositories are specified, it commits
changes in all repositories that have modifications.

A snapshot of the repositories' state is automatically created after committing
changes, unless --no-snapshot is specified.

Examples:
  mctl save "Fix bug in authentication module"
  mctl save --repos=secure-comms,authentication "Update dependencies"
  mctl save --no-push "Work in progress"
  mctl save --amend "Fix typo in previous commit"
  mctl save --all "Add new feature"
  mctl save --sign "Security patch"
  mctl save --description="Stable version for testing" "Prepare for testing"
  mctl save --no-snapshot "Minor changes"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			message := args[0]
			return runSave(repos, message, noPush, amend, all, sign, noSnapshot, description)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&repos, "repos", "", "Limit to specific repositories (comma-separated)")
	cmd.Flags().BoolVar(&noPush, "no-push", false, "Create commit without pushing to remote")
	cmd.Flags().BoolVar(&amend, "amend", false, "Modify previous commit instead of creating new one")
	cmd.Flags().BoolVar(&all, "all", false, "Include all changes including untracked files")
	cmd.Flags().BoolVar(&sign, "sign", false, "Cryptographically sign the commit")
	cmd.Flags().BoolVar(&noSnapshot, "no-snapshot", false, "Skip creating a snapshot")
	cmd.Flags().StringVar(&description, "description", "", "Add a description to the snapshot")

	return cmd
}

func runSave(repos, message string, noPush, amend, all, sign, noSnapshot bool, description string) error {
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

	// Get repositories
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

	// Filter repositories with changes
	var reposWithChanges []*repository.Repository
	for _, repo := range repositories {
		// Update status
		if err := repo.UpdateStatus(); err != nil {
			fmt.Printf("Warning: Failed to update status for %s: %v\n", repo.Config.Name, err)
			continue
		}

		// Check for changes
		hasChanges, err := repo.HasLocalChanges()
		if err != nil {
			fmt.Printf("Warning: Failed to check for changes in %s: %v\n", repo.Config.Name, err)
			continue
		}

		if hasChanges {
			reposWithChanges = append(reposWithChanges, repo)
		}
	}

	if len(reposWithChanges) == 0 {
		fmt.Println("No changes to commit in any repository")
		return nil
	}

	// Commit changes in each repository
	successCount := 0
	for _, repo := range reposWithChanges {
		// Log operation
		logger.LogOperation(logging.LogLevelInfo, fmt.Sprintf("Committing changes in repository %s", repo.Config.Name))
		logger.LogAudit(logging.LogLevelInfo, fmt.Sprintf("Commit in %s: %s", repo.Config.Name, message))

		// Create commit
		if err := repo.Commit(message, all); err != nil {
			fmt.Printf("✗ %s: Failed to commit: %v\n", repo.Config.Name, err)
			continue
		}

		// Push changes if requested
		if !noPush {
			if err := repo.Push(); err != nil {
				fmt.Printf("✗ %s: Committed but failed to push: %v\n", repo.Config.Name, err)
				continue
			}
			fmt.Printf("✓ %s: Committed and pushed\n", repo.Config.Name)
		} else {
			fmt.Printf("✓ %s: Committed (not pushed)\n", repo.Config.Name)
		}

		successCount++
	}

	fmt.Printf("\nSaved changes in %d/%d repositories\n", successCount, len(reposWithChanges))

	// Return error if any repository failed
	if successCount < len(reposWithChanges) {
		return errors.New(errors.ErrGitCommitFailed, "Failed to save changes in one or more repositories")
	}

	// Create snapshot if requested
	if !noSnapshot {
		// Create snapshot manager
		snapshotManager := snapshot.NewManager(currentDir)

		// If no description is provided, use commit message
		snapshotDesc := description
		if snapshotDesc == "" {
			snapshotDesc = message
		}

		// Create snapshot
		snap, err := snapshotManager.CreateSnapshot(repoManager, snapshotDesc)
		if err != nil {
			return errors.Wrap(err, errors.ErrInternalError, "Failed to create snapshot")
		}

		// Save snapshot
		if err := snapshotManager.SaveSnapshot(snap); err != nil {
			return errors.Wrap(err, errors.ErrInternalError, "Failed to save snapshot")
		}

		fmt.Printf("\nCreated snapshot: %s\n", snap.ID)
		fmt.Printf("Use 'mctl load %s' to restore this state\n", snap.ID)
	}

	return nil
}
