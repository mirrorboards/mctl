package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/repository"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	var (
		repos         string
		showUntracked bool
	)

	cmd := &cobra.Command{
		Use:   "status [options]",
		Short: "Report status of managed repositories",
		Long: `Report status of managed repositories.

This command displays the status of repositories managed by MCTL.
It shows information about the current branch, working directory state,
and local changes.

Examples:
  mctl status
  mctl status --repos=repo1,repo2
  mctl status --show-untracked`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(repos, showUntracked)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&repos, "repos", "", "Limit to specific repositories (comma-separated)")
	cmd.Flags().BoolVar(&showUntracked, "show-untracked", false, "Include information about untracked files")

	return cmd
}

func runStatus(repos string, showUntracked bool) error {
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

	// Get repositories to check
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

	// Sort repositories by name for consistent output
	sort.Slice(repositories, func(i, j int) bool {
		return repositories[i].Config.Name < repositories[j].Config.Name
	})

	// Collect all changes across repositories
	var allModifiedFiles, allAddedFiles, allDeletedFiles, allUntrackedFiles []string
	var reposWithChanges int

	// Process each repository
	for _, repo := range repositories {
		// Check if repository exists
		if _, err := os.Stat(repo.FullPath()); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: Repository not found: %s at %s\n", repo.Config.Name, repo.FullPath())
			continue
		}

		// Check for local changes
		hasChanges, err := repo.HasLocalChanges()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to check for local changes in %s: %v\n", repo.Config.Name, err)
			continue
		}

		// Skip repositories with no changes
		if !hasChanges {
			continue
		}

		reposWithChanges++

		// Get changed files
		modifiedFiles, addedFiles, deletedFiles, untrackedFiles, err := getGitChangedFiles(repo, showUntracked)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to get git status for %s: %v\n", repo.Config.Name, err)
			continue
		}

		// Add files to consolidated lists
		allModifiedFiles = append(allModifiedFiles, modifiedFiles...)
		allAddedFiles = append(allAddedFiles, addedFiles...)
		allDeletedFiles = append(allDeletedFiles, deletedFiles...)
		allUntrackedFiles = append(allUntrackedFiles, untrackedFiles...)
	}

	// Print consolidated output
	fmt.Printf("Found changes in %d repositories\n\n", reposWithChanges)

	// Print changes not staged for commit
	if len(allModifiedFiles) > 0 || len(allDeletedFiles) > 0 {
		fmt.Println("Changes not staged for commit:")
		fmt.Println("  (use \"git add <file>...\" to update what will be committed)")
		fmt.Println("  (use \"git restore <file>...\" to discard changes in working directory)")
		fmt.Println()

		// Print modified files
		for _, file := range allModifiedFiles {
			fmt.Printf("\tmodified:  %s\n", file)
		}

		// Print deleted files
		for _, file := range allDeletedFiles {
			fmt.Printf("\tdeleted:    %s\n", file)
		}

		fmt.Println()
	}

	// Print changes staged for commit
	if len(allAddedFiles) > 0 {
		fmt.Println("Changes to be committed:")
		fmt.Println("  (use \"git restore --staged <file>...\" to unstage)")
		fmt.Println()

		// Print added files
		for _, file := range allAddedFiles {
			fmt.Printf("\tnew file:       %s\n", file)
		}

		fmt.Println()
	}

	// Print untracked files
	if showUntracked && len(allUntrackedFiles) > 0 {
		fmt.Println("Untracked files:")
		fmt.Println("  (use \"git add <file>...\" to include in what will be committed)")
		fmt.Println()

		// Print untracked files
		for _, file := range allUntrackedFiles {
			fmt.Printf("\t%s\n", file)
		}

		fmt.Println()
	}

	return nil
}

// getGitBranchInfo gets the current branch and remote status for a repository
func getGitBranchInfo(repo *repository.Repository) (string, string, error) {
	// Get current branch
	branchCmd := exec.Command("git", "-C", repo.FullPath(), "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("error getting current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get remote status
	remoteCmd := exec.Command("git", "-C", repo.FullPath(), "status", "-sb")
	remoteOutput, err := remoteCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("error getting remote status: %w", err)
	}
	remoteStatus := strings.Split(strings.TrimSpace(string(remoteOutput)), "\n")[0]
	remoteStatus = strings.TrimPrefix(remoteStatus, "## "+branch+" ")

	return branch, remoteStatus, nil
}

// getGitChangedFiles gets the modified, added, deleted, and untracked files for a repository
func getGitChangedFiles(repo *repository.Repository, showUntracked bool) ([]string, []string, []string, []string, error) {
	// Get changed files
	args := []string{"-C", repo.FullPath(), "status", "--porcelain"}
	if !showUntracked {
		args = append(args, "--untracked-files=no")
	}

	cmd := exec.Command("git", args...)
	statusOutput, err := cmd.Output()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error getting git status: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(statusOutput)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil, nil, nil, nil
	}

	// Collect modified, added, deleted, and untracked files
	var modifiedFiles, addedFiles, deletedFiles, untrackedFiles []string

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		statusCode := line[0:2]
		filename := line[3:]

		// Prepend repository name to filename for clarity
		// Remove redundant path prefixes for better readability
		cleanFilename := filename
		if strings.HasPrefix(filename, "scanboards-cluster/") {
			cleanFilename = strings.TrimPrefix(filename, "scanboards-cluster/")
		} else if strings.HasPrefix(filename, "canboards-cluster/") {
			cleanFilename = strings.TrimPrefix(filename, "canboards-cluster/")
		}
		fullPath := fmt.Sprintf("%s: %s", repo.Config.Name, cleanFilename)

		switch {
		case statusCode == "M " || statusCode == " M":
			modifiedFiles = append(modifiedFiles, fullPath)
		case statusCode == "A " || statusCode == "AM":
			addedFiles = append(addedFiles, fullPath)
		case statusCode == "D " || statusCode == " D":
			deletedFiles = append(deletedFiles, fullPath)
		case statusCode == "??":
			untrackedFiles = append(untrackedFiles, fullPath)
		}
	}

	return modifiedFiles, addedFiles, deletedFiles, untrackedFiles, nil
}
