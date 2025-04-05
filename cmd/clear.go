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

func newClearCmd() *cobra.Command {
	var (
		force      bool
		keepConfig bool
		secure     bool
	)

	cmd := &cobra.Command{
		Use:   "clear [options]",
		Short: "Remove repository directories while preserving configuration",
		Long: `Remove repository directories while preserving configuration.

This command removes all repository directories managed by MCTL.
By default, it preserves the configuration, but you can also remove
the configuration with the --keep-config=false flag.

Examples:
  mctl clear
  mctl clear --force
  mctl clear --keep-config=false
  mctl clear --secure`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClear(force, keepConfig, secure)
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&force, "force", false, "Override confirmation requirement")
	cmd.Flags().BoolVar(&keepConfig, "keep-config", true, "Preserve configuration during clearing operation")
	cmd.Flags().BoolVar(&secure, "secure", false, "Use secure deletion methods")

	return cmd
}

func runClear(force, keepConfig, secure bool) error {
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

	// Get all repositories
	repositories, err := repoManager.GetAllRepositories()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to get repositories")
	}

	if len(repositories) == 0 {
		fmt.Println("No repositories to clear")
		return nil
	}

	// Confirm operation if not forced
	if !force {
		fmt.Printf("This will remove %d repository directories", len(repositories))
		if !keepConfig {
			fmt.Printf(" and the MCTL configuration")
		}
		fmt.Printf(".\nAre you sure? [y/N] ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation canceled by user")
			return errors.New(errors.ErrInvalidArgument, "Operation canceled by user")
		}
	}

	// Create logger
	logger := logging.NewLogger(currentDir)
	logger.LogOperation(logging.LogLevelInfo, "Clearing repositories")
	logger.LogAudit(logging.LogLevelInfo, fmt.Sprintf("Clearing %d repositories", len(repositories)))

	// Remove repository directories
	successCount := 0
	for _, repo := range repositories {
		repoPath := repo.FullPath()

		// Check if repository exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			fmt.Printf("✓ %s: Directory does not exist, skipping\n", repo.Config.Name)
			successCount++
			continue
		}

		// Remove repository directory
		var removeErr error
		if secure {
			// Secure deletion (simple implementation - in a real system, this would use more secure methods)
			removeErr = secureDelete(repoPath)
		} else {
			// Standard deletion
			removeErr = os.RemoveAll(repoPath)
		}

		if removeErr != nil {
			fmt.Printf("✗ %s: Failed to remove directory: %v\n", repo.Config.Name, removeErr)
		} else {
			fmt.Printf("✓ %s: Removed directory\n", repo.Config.Name)
			successCount++
		}
	}

	// Remove configuration if requested
	if !keepConfig {
		configDir := config.GetConfigDirPath(currentDir)
		var removeErr error
		if secure {
			removeErr = secureDelete(configDir)
		} else {
			removeErr = os.RemoveAll(configDir)
		}

		if removeErr != nil {
			fmt.Printf("✗ Failed to remove configuration directory: %v\n", removeErr)
		} else {
			fmt.Println("✓ Removed configuration directory")
		}
	}

	fmt.Printf("\nCleared %d/%d repositories\n", successCount, len(repositories))

	// Return error if any repository failed to clear
	if successCount < len(repositories) {
		return errors.New(errors.ErrInternalError, "Failed to clear one or more repositories")
	}

	return nil
}

// secureDelete implements a simple secure deletion
// In a real implementation, this would use more secure methods
func secureDelete(path string) error {
	// For now, just use standard deletion
	// In a real implementation, this would overwrite files with random data before deletion
	return os.RemoveAll(path)
}
