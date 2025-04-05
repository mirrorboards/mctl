package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/logging"
	"github.com/mirrorboards/mctl/internal/repository"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var (
		path    string
		name    string
		branch  string
		noClone bool
		flat    bool
	)

	cmd := &cobra.Command{
		Use:   "add [options] repository-url [path]",
		Short: "Add a Git repository to MCTL management",
		Long: `Add a Git repository to MCTL management.

This command adds a Git repository to MCTL management. By default, it clones
the repository to the current directory, but you can specify a different
path with the --path flag or as a second argument.

You can also specify a custom name for the repository with the --name flag.
If not provided, the name will be derived from the repository URL.

Examples:
  mctl add git@secure.gov:system/comms.git
  mctl add git@secure.gov:system/comms.git classified
  mctl add git@secure.gov:system/comms.git --path=classified --name=secure-comms
  mctl add git@secure.gov:system/comms.git --branch=release-1.2
  mctl add git@secure.gov:system/comms.git --no-clone`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoURL := args[0]

			// If path is provided as second argument, use it
			if len(args) > 1 && path == "" {
				path = args[1]
			}

			return runAdd(repoURL, path, name, branch, noClone, flat)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&path, "path", "", "Target location for repository")
	cmd.Flags().StringVar(&name, "name", "", "Custom repository designation")
	cmd.Flags().StringVar(&branch, "branch", "", "Specific branch to clone")
	cmd.Flags().BoolVar(&noClone, "no-clone", false, "Add to configuration without cloning")
	cmd.Flags().BoolVar(&flat, "flat", false, "Clone directly to path without creating subdirectory")

	return cmd
}

func runAdd(repoURL, path, name, branch string, noClone, flat bool) error {
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

	// Determine repository name if not provided
	if name == "" {
		name = deriveRepositoryName(repoURL)
	}

	// Determine repository path
	repoPath, err := determineRepositoryPath(currentDir, path, name, flat, repoURL)
	if err != nil {
		return err
	}

	// Create repository manager
	repoManager := repository.NewManager(cfg, currentDir)

	// Add repository
	repo, err := repoManager.AddRepository(name, repoURL, repoPath, branch, noClone)
	if err != nil {
		return errors.Wrap(err, errors.ErrRepositoryExists, "Failed to add repository")
	}

	// Log the operation
	logger := logging.NewLogger(currentDir)
	logger.LogOperation(logging.LogLevelInfo, fmt.Sprintf("Added repository %s (%s)", name, repoURL))
	logger.LogAudit(logging.LogLevelInfo, fmt.Sprintf("Repository added: %s", name))

	fmt.Printf("Added repository %s to MCTL management\n", name)
	if !noClone {
		fmt.Printf("Cloned to %s\n", repo.FullPath())
	}

	return nil
}

func deriveRepositoryName(repoURL string) string {
	// Extract name from URL
	parts := strings.Split(repoURL, "/")
	lastPart := parts[len(parts)-1]

	// Remove .git extension if present
	name := strings.TrimSuffix(lastPart, ".git")

	return name
}

func determineRepositoryPath(baseDir, path, name string, flat bool, repoURL string) (string, error) {
	// If path is not provided, use the current directory
	if path == "" {
		return name, nil
	}

	// If flat is true, use the path directly
	if flat {
		return path, nil
	}

	// Otherwise, append repository name to path
	return filepath.Join(path, name), nil
}
