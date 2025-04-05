package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/repository"
	"github.com/mirrorboards/mctl/internal/snapshot"
	"github.com/spf13/cobra"
)

func newLoadCmd() *cobra.Command {
	var (
		repos  string
		dryRun bool
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "load [options] <snapshot-id>",
		Short: "Load a repository snapshot",
		Long: `Load a repository snapshot.

This command restores repositories to the state saved in a snapshot. It will
checkout the correct branch and reset to the saved commit hash for each repository.

Examples:
  mctl load 20250405-123456-abcdef12
  mctl load --repos=secure-comms,authentication 20250405-123456-abcdef12
  mctl load --dry-run 20250405-123456-abcdef12
  mctl load --force 20250405-123456-abcdef12`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshotID := args[0]
			return runLoad(snapshotID, repos, dryRun, force)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&repos, "repos", "", "Limit to specific repositories (comma-separated)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	cmd.Flags().BoolVar(&force, "force", false, "Force load even if there are uncommitted changes")

	return cmd
}

func runLoad(snapshotID, repos string, dryRun, force bool) error {
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

	// Create snapshot manager
	snapshotManager := snapshot.NewManager(currentDir)

	// Load snapshot
	snap, err := snapshotManager.LoadSnapshot(snapshotID)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, fmt.Sprintf("Failed to load snapshot: %s", snapshotID))
	}

	// Parse repositories
	var repoNames []string
	if repos != "" {
		repoNames = strings.Split(repos, ",")
		for i, name := range repoNames {
			repoNames[i] = strings.TrimSpace(name)
		}
	}

	// Create apply options
	options := snapshot.ApplyOptions{
		DryRun:       dryRun,
		Force:        force,
		Repositories: repoNames,
	}

	// Apply snapshot
	if err := snapshotManager.ApplySnapshot(snap, repoManager, options); err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to apply snapshot")
	}

	if dryRun {
		fmt.Println("\nDry run completed. No changes were made.")
	} else {
		fmt.Printf("\nSuccessfully loaded snapshot %s\n", snapshotID)
	}

	return nil
}
