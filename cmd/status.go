package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/repository"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	var (
		repos         string
		format        string
		showUntracked bool
		verify        bool
	)

	cmd := &cobra.Command{
		Use:   "status [options]",
		Short: "Report status of managed repositories",
		Long: `Report status of managed repositories.

This command displays the status of repositories managed by MCTL.
It shows information about the current branch, working directory state,
relationship to remote, and more.

Examples:
  mctl status
  mctl status --repos=secure-comms,authentication
  mctl status --format=json
  mctl status --show-untracked
  mctl status --verify`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(repos, format, showUntracked, verify)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&repos, "repos", "", "Limit to specific repositories (comma-separated)")
	cmd.Flags().StringVar(&format, "format", "detailed", "Output format (detailed, summary, json)")
	cmd.Flags().BoolVar(&showUntracked, "show-untracked", false, "Include information about untracked files")
	cmd.Flags().BoolVar(&verify, "verify", false, "Perform integrity verification during status check")

	return cmd
}

func runStatus(repos, format string, showUntracked, verify bool) error {
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

	// Update status for each repository
	for _, repo := range repositories {
		if err := repo.UpdateStatus(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to update status for %s: %v\n", repo.Config.Name, err)
		}
	}

	// Display status in the specified format
	switch format {
	case "detailed":
		displayDetailedStatus(repositories, showUntracked, verify)
	case "summary":
		displaySummaryStatus(repositories, showUntracked, verify)
	case "json":
		displayJSONStatus(repositories, showUntracked, verify)
	default:
		return errors.New(errors.ErrInvalidArgument, "Invalid format specification")
	}

	return nil
}

func displayDetailedStatus(repos []*repository.Repository, showUntracked, verify bool) {
	for _, repo := range repos {
		fmt.Printf("Repository: %s (%s)\n", repo.Config.Name, repo.Config.ID)
		fmt.Printf("  Path: %s\n", repo.FullPath())
		fmt.Printf("  URL: %s\n", repo.Config.URL)
		fmt.Printf("  Branch: %s\n", repo.Metadata.Status.Branch)
		fmt.Printf("  Status: %s\n", repo.Metadata.Status.Current)

		// Get additional status information
		if repo.Metadata.Status.Current != repository.StatusUnknown {
			// Check for local changes
			hasChanges, err := repo.HasLocalChanges()
			if err == nil {
				if hasChanges {
					fmt.Println("  Local Changes: Yes")
				} else {
					fmt.Println("  Local Changes: No")
				}
			}

			// Check relationship with remote
			ahead, behind, err := repo.GetRemoteStatus()
			if err == nil {
				fmt.Printf("  Ahead: %d commits\n", ahead)
				fmt.Printf("  Behind: %d commits\n", behind)
			}

			// Show last sync time
			if !repo.Metadata.Basic.LastSync.IsZero() {
				fmt.Printf("  Last Sync: %s\n", repo.Metadata.Basic.LastSync.Format("2006-01-02 15:04:05"))
			} else {
				fmt.Printf("  Last Sync: Never\n")
			}
		}

		fmt.Println()
	}
}

func displaySummaryStatus(repos []*repository.Repository, showUntracked, verify bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "NAME\tBRANCH\tSTATUS\tAHEAD\tBEHIND\tLAST SYNC")

	// Print repositories
	for _, repo := range repos {
		ahead, behind := 0, 0
		if repo.Metadata.Status.Current != repository.StatusUnknown {
			var err error
			ahead, behind, err = repo.GetRemoteStatus()
			if err != nil {
				ahead, behind = 0, 0
			}
		}

		lastSync := "Never"
		if !repo.Metadata.Basic.LastSync.IsZero() {
			lastSync = repo.Metadata.Basic.LastSync.Format("2006-01-02")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n",
			repo.Config.Name,
			repo.Metadata.Status.Branch,
			repo.Metadata.Status.Current,
			ahead,
			behind,
			lastSync,
		)
	}
}

func displayJSONStatus(repos []*repository.Repository, showUntracked, verify bool) {
	type jsonStatus struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Path       string `json:"path"`
		URL        string `json:"url"`
		Branch     string `json:"branch"`
		Status     string `json:"status"`
		Ahead      int    `json:"ahead"`
		Behind     int    `json:"behind"`
		LastSync   string `json:"last_sync"`
		HasChanges bool   `json:"has_changes"`
	}

	var result []jsonStatus
	for _, repo := range repos {
		status := jsonStatus{
			ID:     repo.Config.ID,
			Name:   repo.Config.Name,
			Path:   repo.FullPath(),
			URL:    repo.Config.URL,
			Branch: repo.Metadata.Status.Branch,
			Status: string(repo.Metadata.Status.Current),
		}

		if repo.Metadata.Status.Current != repository.StatusUnknown {
			// Check for local changes
			hasChanges, err := repo.HasLocalChanges()
			if err == nil {
				status.HasChanges = hasChanges
			}

			// Check relationship with remote
			ahead, behind, err := repo.GetRemoteStatus()
			if err == nil {
				status.Ahead = ahead
				status.Behind = behind
			}

			// Show last sync time
			if !repo.Metadata.Basic.LastSync.IsZero() {
				status.LastSync = repo.Metadata.Basic.LastSync.Format("2006-01-02 15:04:05")
			} else {
				status.LastSync = "Never"
			}
		}

		result = append(result, status)
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling to JSON: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}
