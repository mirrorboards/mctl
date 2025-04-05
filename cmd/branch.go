package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/logging"
	"github.com/mirrorboards/mctl/internal/repository"
	"github.com/spf13/cobra"
)

func newBranchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch [subcommand]",
		Short: "Manage Git branches across repositories",
		Long: `Manage Git branches across repositories.

This command provides subcommands for managing Git branches across repositories.
If no subcommand is provided, it lists the available branches.

Examples:
  mctl branch
  mctl branch list
  mctl branch create feature-branch
  mctl branch checkout release-branch`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, default to list
			if len(args) == 0 {
				return runBranchList("", false)
			}
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newBranchListCmd())
	cmd.AddCommand(newBranchCreateCmd())
	cmd.AddCommand(newBranchCheckoutCmd())

	return cmd
}

func newBranchListCmd() *cobra.Command {
	var (
		repos string
		all   bool
	)

	cmd := &cobra.Command{
		Use:   "list [options]",
		Short: "List branches in repositories",
		Long: `List branches in repositories.

This command lists the branches in the specified repositories.
If no repositories are specified, it lists branches for all repositories.

Examples:
  mctl branch list
  mctl branch list --repos=secure-comms,authentication
  mctl branch list --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBranchList(repos, all)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&repos, "repos", "", "Limit to specific repositories (comma-separated)")
	cmd.Flags().BoolVar(&all, "all", false, "Show all branches, including remote branches")

	return cmd
}

func newBranchCreateCmd() *cobra.Command {
	var (
		repos string
		from  string
		push  bool
		track bool
	)

	cmd := &cobra.Command{
		Use:   "create [options] branch-name",
		Short: "Create a new branch in repositories",
		Long: `Create a new branch in repositories.

This command creates a new branch in the specified repositories.
If no repositories are specified, it creates the branch in all repositories.

Examples:
  mctl branch create feature-branch
  mctl branch create --repos=secure-comms,authentication feature-branch
  mctl branch create --from=main feature-branch
  mctl branch create --push feature-branch`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			return runBranchCreate(repos, branchName, from, push, track)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&repos, "repos", "", "Limit to specific repositories (comma-separated)")
	cmd.Flags().StringVar(&from, "from", "", "Base branch for creation (default: current branch)")
	cmd.Flags().BoolVar(&push, "push", false, "Push new branch to remote after creation")
	cmd.Flags().BoolVar(&track, "track", false, "Configure tracking relationship with remote")

	return cmd
}

func newBranchCheckoutCmd() *cobra.Command {
	var (
		repos string
		force bool
	)

	cmd := &cobra.Command{
		Use:   "checkout [options] branch-name",
		Short: "Switch to an existing branch in repositories",
		Long: `Switch to an existing branch in repositories.

This command switches to an existing branch in the specified repositories.
If no repositories are specified, it switches the branch in all repositories.

Examples:
  mctl branch checkout main
  mctl branch checkout --repos=secure-comms,authentication release-branch
  mctl branch checkout --force feature-branch`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			return runBranchCheckout(repos, branchName, force)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&repos, "repos", "", "Limit to specific repositories (comma-separated)")
	cmd.Flags().BoolVar(&force, "force", false, "Force checkout even with uncommitted changes")

	return cmd
}

func runBranchList(repos string, all bool) error {
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

	// List branches for each repository
	for _, repo := range repositories {
		fmt.Printf("Repository: %s\n", repo.Config.Name)

		// Get current branch
		currentBranch, err := repo.GetCurrentBranch()
		if err != nil {
			fmt.Printf("  Error getting current branch: %v\n", err)
			continue
		}

		// List branches
		branches, err := repo.ListBranches()
		if err != nil {
			fmt.Printf("  Error listing branches: %v\n", err)
			continue
		}

		// Display branches
		for _, branch := range branches {
			if branch == currentBranch {
				fmt.Printf("* %s (current)\n", branch)
			} else {
				fmt.Printf("  %s\n", branch)
			}
		}

		fmt.Println()
	}

	return nil
}

func runBranchCreate(repos, branchName, fromBranch string, push, track bool) error {
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

	// Create branch in each repository
	successCount := 0
	for _, repo := range repositories {
		// Log operation
		logger.LogOperation(logging.LogLevelInfo, fmt.Sprintf("Creating branch %s in repository %s", branchName, repo.Config.Name))

		// Create branch
		if err := repo.CreateBranch(branchName, fromBranch); err != nil {
			fmt.Printf("✗ %s: Failed to create branch: %v\n", repo.Config.Name, err)
			continue
		}

		// Push branch if requested
		if push {
			// TODO: Implement push branch
			fmt.Printf("✓ %s: Created branch %s (push not implemented yet)\n", repo.Config.Name, branchName)
		} else {
			fmt.Printf("✓ %s: Created branch %s\n", repo.Config.Name, branchName)
		}

		successCount++
	}

	fmt.Printf("\nCreated branch in %d/%d repositories\n", successCount, len(repositories))

	// Return error if any repository failed
	if successCount < len(repositories) {
		return errors.New(errors.ErrGitBranchFailed, "Failed to create branch in one or more repositories")
	}

	return nil
}

func runBranchCheckout(repos, branchName string, force bool) error {
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

	// Checkout branch in each repository
	successCount := 0
	for _, repo := range repositories {
		// Check if already on the branch
		currentBranch, err := repo.GetCurrentBranch()
		if err == nil && currentBranch == branchName {
			fmt.Printf("✓ %s: Already on branch %s\n", repo.Config.Name, branchName)
			successCount++
			continue
		}

		// Check for uncommitted changes if not forcing
		if !force {
			hasChanges, err := repo.HasLocalChanges()
			if err != nil {
				fmt.Printf("✗ %s: Failed to check for local changes: %v\n", repo.Config.Name, err)
				continue
			}
			if hasChanges {
				fmt.Printf("✗ %s: Has uncommitted changes (use --force to override)\n", repo.Config.Name)
				continue
			}
		}

		// Log operation
		logger.LogOperation(logging.LogLevelInfo, fmt.Sprintf("Checking out branch %s in repository %s", branchName, repo.Config.Name))

		// Checkout branch
		if err := repo.CheckoutBranch(branchName); err != nil {
			fmt.Printf("✗ %s: Failed to checkout branch: %v\n", repo.Config.Name, err)
			continue
		}

		fmt.Printf("✓ %s: Checked out branch %s\n", repo.Config.Name, branchName)
		successCount++
	}

	fmt.Printf("\nChecked out branch in %d/%d repositories\n", successCount, len(repositories))

	// Return error if any repository failed
	if successCount < len(repositories) {
		return errors.New(errors.ErrGitBranchFailed, "Failed to checkout branch in one or more repositories")
	}

	return nil
}
