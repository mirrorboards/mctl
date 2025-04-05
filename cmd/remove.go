package cmd

import (
	"fmt"
	"os"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/logging"
	"github.com/mirrorboards/mctl/internal/repository"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	var (
		delete          bool
		force           bool
		preserveHistory bool
	)

	cmd := &cobra.Command{
		Use:     "remove [options] repository-identifier",
		Aliases: []string{"rm"},
		Short:   "Remove a repository from MCTL management",
		Long: `Remove a repository from MCTL management.

This command removes a repository from MCTL management. By default, it removes
the repository from both the configuration and the file system. You can use
the --delete=false flag to keep the repository files.

The repository can be identified by its unique ID, name, or path.

Examples:
  mctl remove a1b2c3d4e5
  mctl remove secure-comms
  mctl remove repositories/secure-comms
  mctl remove secure-comms --delete=false
  mctl remove secure-comms --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			identifier := args[0]
			return runRemove(identifier, delete, force, preserveHistory)
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&delete, "delete", true, "Remove repository files after configuration update")
	cmd.Flags().BoolVar(&force, "force", false, "Bypass confirmation prompts")
	cmd.Flags().BoolVar(&preserveHistory, "preserve-history", false, "Retain operational history after removal")

	return cmd
}

func runRemove(identifier string, delete, force, preserveHistory bool) error {
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

	// Find repository
	repo, err := repoManager.GetRepository(identifier)
	if err != nil {
		return errors.Wrap(err, errors.ErrRepositoryNotFound, fmt.Sprintf("Repository not found: %s", identifier))
	}

	// Confirm removal if not forced
	if !force {
		fmt.Printf("Are you sure you want to remove repository '%s'", repo.Config.Name)
		if delete {
			fmt.Printf(" and delete its files")
		}
		fmt.Printf("? [y/N] ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation canceled by user")
			return errors.New(errors.ErrInvalidArgument, "Operation canceled by user")
		}
	}

	// Log the operation before removal
	logger := logging.NewLogger(currentDir)
	logger.LogOperation(logging.LogLevelInfo, fmt.Sprintf("Removing repository %s", repo.Config.Name))
	logger.LogAudit(logging.LogLevelInfo, fmt.Sprintf("Repository removed: %s", repo.Config.Name))

	// Remove repository
	if err := repoManager.RemoveRepository(identifier, delete); err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to remove repository")
	}

	fmt.Printf("Removed repository '%s' from MCTL management\n", repo.Config.Name)
	if delete {
		fmt.Printf("Deleted repository files at %s\n", repo.FullPath())
	}

	return nil
}
