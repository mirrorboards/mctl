package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mirrorboards/mctl/pkg/config"
	"github.com/spf13/cobra"
)

// RepoStatus represents the status of a repository for JSON output
type RepoStatus struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	URL       string   `json:"url"`
	Branch    string   `json:"branch,omitempty"`
	Exists    bool     `json:"exists"`
	IsDirty   bool     `json:"is_dirty"`
	Staged    []string `json:"staged,omitempty"`
	Modified  []string `json:"modified,omitempty"`
	Untracked []string `json:"untracked,omitempty"`
}

// StatusSummary represents the summary of all repositories for JSON output
type StatusSummary struct {
	TotalRepos    int          `json:"total_repos"`
	ExistingRepos int          `json:"existing_repos"`
	MissingRepos  int          `json:"missing_repos"`
	DirtyRepos    int          `json:"dirty_repos"`
	Repositories  []RepoStatus `json:"repositories"`
	Timestamp     string       `json:"timestamp"`
}

var (
	statusFormat string
	statusRepos  []string
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of all repositories",
		Long:  `Run git status on all repositories defined in mirror.toml and display the results in a colorful, elegant way.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := config.GetAllRepositories()
			if err != nil {
				return err
			}

			if len(repos) == 0 {
				fmt.Println("No repositories configured in mirror.toml")
				return nil
			}

			totalRepos := len(repos)
			existingRepos := 0
			dirtyRepos := 0
			missingRepos := 0

			// For JSON output
			var statusSummary StatusSummary
			statusSummary.TotalRepos = totalRepos
			statusSummary.Repositories = make([]RepoStatus, 0, totalRepos)
			statusSummary.Timestamp = time.Now().Format(time.RFC3339)

			// For text output
			var titleColor, repoNameColor, pathColor, cleanColor, dirtyColor, branchColor, missingColor, changedFileColor, stagedFileColor, untrackedFileColor *color.Color

			if statusFormat == "text" {
				// Define colors for text output
				titleColor = color.New(color.FgHiWhite, color.Bold)
				repoNameColor = color.New(color.FgHiCyan, color.Bold)
				pathColor = color.New(color.FgHiBlue)
				cleanColor = color.New(color.FgHiGreen)
				dirtyColor = color.New(color.FgHiRed)
				branchColor = color.New(color.FgHiYellow)
				missingColor = color.New(color.FgHiMagenta)
				changedFileColor = color.New(color.FgHiRed)
				stagedFileColor = color.New(color.FgHiGreen)
				untrackedFileColor = color.New(color.FgHiCyan)

				// Print header
				titleColor.Println("\n✨ MIRROR REPOSITORIES STATUS ✨")
				fmt.Println(strings.Repeat("─", 60))
			}

			for _, repo := range repos {
				var repoPath string
				if repo.Path == "." {
					// Special case for current directory
					repoPath = repo.Name
				} else if repo.Name == "" {
					repoPath = repo.Path
				} else {
					repoPath = filepath.Join(repo.Path, repo.Name)
				}

				// Skip if not in the specified repos list
				if len(statusRepos) > 0 {
					found := false
					for _, r := range statusRepos {
						if repo.ID == r || repo.Name == r {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}

				// Create repo status for JSON output
				repoStatus := RepoStatus{
					ID:   repo.ID,
					Name: repo.Name,
					Path: repoPath,
					URL:  repo.URL,
				}

				if statusFormat == "text" {
					// Display repository name, ID, and path
					repoNameColor.Printf("• %s ", filepath.Base(repoPath))
					pathColor.Printf("(%s)\n", repoPath)
					fmt.Printf("  ID: %s\n", repo.ID)
				}

				// Check if the repository exists
				gitDir := filepath.Join(repoPath, ".git")
				isRepo := false

				// First check for standard .git directory
				if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
					isRepo = true
				} else {
					// Check if it's a git worktree (in which case .git is a file pointing to the real .git dir)
					if stat, err := os.Stat(gitDir); err == nil && !stat.IsDir() {
						isRepo = true
					} else {
						// Try running git command as a final check
						cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--is-inside-work-tree")
						if err := cmd.Run(); err == nil {
							isRepo = true
						}
					}
				}

				if !isRepo {
					// Repository doesn't exist
					missingRepos++
					repoStatus.Exists = false

					if statusFormat == "text" {
						missingColor.Println("  Status: Not cloned yet")
						fmt.Println(strings.Repeat("─", 60))
					}

					statusSummary.Repositories = append(statusSummary.Repositories, repoStatus)
					continue
				}

				repoStatus.Exists = true

				existingRepos++

				// Get current branch
				branch, err := getGitBranch(repoPath)
				if err == nil {
					repoStatus.Branch = branch
					if statusFormat == "text" {
						branchColor.Printf("  Branch: %s\n", branch)
					}
				}

				// Get git status
				status, isDirty, err := getGitStatus(repoPath)
				if err != nil {
					fmt.Printf("  Error: %v\n", err)
					continue
				}

				if isDirty {
					dirtyRepos++
					repoStatus.IsDirty = true

					// Get list of modified files
					modified, staged, untracked, err := getChangedFiles(repoPath)
					if err == nil {
						repoStatus.Modified = modified
						repoStatus.Staged = staged
						repoStatus.Untracked = untracked

						if statusFormat == "text" {
							dirtyColor.Println("  Status: Changes detected")
							fmt.Println()

							if len(staged) > 0 {
								fmt.Println("  Changes to be committed:")
								for _, file := range staged {
									stagedFileColor.Printf("    %s\n", file)
								}
								fmt.Println()
							}

							if len(modified) > 0 {
								fmt.Println("  Changes not staged for commit:")
								for _, file := range modified {
									changedFileColor.Printf("    %s\n", file)
								}
								fmt.Println()
							}

							if len(untracked) > 0 {
								fmt.Println("  Untracked files:")
								for _, file := range untracked {
									untrackedFileColor.Printf("    %s\n", file)
								}
								fmt.Println()
							}
						}
					} else if statusFormat == "text" {
						// Fallback to formatted git status output
						fmt.Println()
						fmt.Print(formatGitStatus(status))
					}
				} else {
					repoStatus.IsDirty = false
					if statusFormat == "text" {
						cleanColor.Println("  Status: Clean")
					}
				}

				statusSummary.Repositories = append(statusSummary.Repositories, repoStatus)

				if statusFormat == "text" {
					fmt.Println(strings.Repeat("─", 60))
				}
			}

			// Update summary counts
			statusSummary.ExistingRepos = existingRepos
			statusSummary.MissingRepos = missingRepos
			statusSummary.DirtyRepos = dirtyRepos

			if statusFormat == "json" {
				// Output JSON
				jsonData, err := json.MarshalIndent(statusSummary, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(jsonData))
			} else {
				// Print text summary
				titleColor.Println("\n📊 SUMMARY")
				fmt.Printf("Total repositories: %d\n", totalRepos)
				fmt.Printf("Existing repositories: %d\n", existingRepos)

				if missingRepos > 0 {
					missingColor.Printf("Not cloned repositories: %d\n", missingRepos)
				}

				if dirtyRepos > 0 {
					dirtyColor.Printf("Repositories with changes: %d\n", dirtyRepos)
				} else if existingRepos > 0 {
					cleanColor.Println("All existing repositories are clean ✓")
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&statusFormat, "format", "text", "Output format (text, json)")
	cmd.Flags().StringSliceVar(&statusRepos, "repos", []string{}, "Only show status for specified repositories (by ID or name)")

	return cmd
}

// getGitBranch returns the current branch of a git repository
func getGitBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getGitStatus runs git status and returns the output and whether the repository is dirty
func getGitStatus(repoPath string) (string, bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	porcelainOutput, err := cmd.Output()
	if err != nil {
		return "", false, err
	}

	isDirty := len(bytes.TrimSpace(porcelainOutput)) > 0

	if !isDirty {
		return "", false, nil
	}

	// If the repo is dirty, get the full git status
	cmd = exec.Command("git", "-C", repoPath, "status")
	output, err := cmd.Output()
	if err != nil {
		return "", true, err
	}

	return string(output), true, nil
}

// getChangedFiles returns lists of modified, staged, and untracked files
func getChangedFiles(repoPath string) ([]string, []string, []string, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, nil, nil, err
	}

	lines := strings.Split(string(output), "\n")
	var modified, staged, untracked []string

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		statusCode := line[0:2]
		fileName := strings.TrimSpace(line[3:])

		if fileName == "" {
			continue
		}

		switch {
		case statusCode == "??":
			untracked = append(untracked, fileName)
		case statusCode[0] == ' ' && statusCode[1] != ' ':
			modified = append(modified, fileName)
		case statusCode[0] != ' ':
			staged = append(staged, fileName)
		}
	}

	return modified, staged, untracked, nil
}

// formatGitStatus formats the git status output
func formatGitStatus(status string) string {
	// Split the status output into lines
	lines := strings.Split(status, "\n")

	// Remove the first two lines (header lines)
	if len(lines) > 2 {
		lines = lines[2:]
	}

	// Indent all lines
	for i, line := range lines {
		if len(line) > 0 {
			lines[i] = "    " + line
		}
	}

	return strings.Join(lines, "\n")
}
