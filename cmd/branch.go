package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/mirrorboards/mctl/pkg/config"
	"github.com/spf13/cobra"
)

var (
	branchCreate bool
	branchPull   bool
	branchRepos  []string
)

func newBranchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch [branch-name]",
		Short: "Switch all repositories to a specific branch",
		Long:  `Switch all repositories to a specific branch. Optionally create the branch if it doesn't exist and pull latest changes.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]

			repos, err := config.GetAllRepositories()
			if err != nil {
				return err
			}

			if len(repos) == 0 {
				fmt.Println("No repositories configured in mirror.toml")
				return nil
			}

			// Define colors
			titleColor := color.New(color.FgHiWhite, color.Bold)
			repoNameColor := color.New(color.FgHiCyan, color.Bold)
			pathColor := color.New(color.FgHiBlue)
			successColor := color.New(color.FgHiGreen)
			errorColor := color.New(color.FgHiRed)
			skipColor := color.New(color.FgHiYellow)

			// Print header
			titleColor.Printf("\n✨ SWITCHING REPOSITORIES TO BRANCH: %s ✨\n", branchName)
			fmt.Println(strings.Repeat("─", 60))

			successCount := 0
			errorCount := 0
			skipCount := 0

			for _, repo := range repos {
				// Skip if not in the specified repos list
				if len(branchRepos) > 0 {
					found := false
					for _, r := range branchRepos {
						if repo.ID == r || repo.Name == r {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}

				var repoPath string
				if repo.Path == "." {
					// Special case for current directory
					repoPath = repo.Name
				} else if repo.Name == "" {
					repoPath = repo.Path
				} else {
					repoPath = filepath.Join(repo.Path, repo.Name)
				}

				// Display repository name and path
				repoNameColor.Printf("• %s ", filepath.Base(repoPath))
				pathColor.Printf("(%s)\n", repoPath)

				// Check if the repository exists
				gitDir := filepath.Join(repoPath, ".git")
				isRepo := false

				// First check for standard .git directory
				if _, err := os.Stat(gitDir); err == nil {
					isRepo = true
				} else {
					// Try running git command as a final check
					cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--is-inside-work-tree")
					if err := cmd.Run(); err == nil {
						isRepo = true
					}
				}

				if !isRepo {
					// Repository doesn't exist, skip
					skipColor.Println("  Repository not found, skipping")
					skipCount++
					fmt.Println(strings.Repeat("─", 60))
					continue
				}

				// Get current branch
				currentBranch, err := getGitBranch(repoPath)
				if err != nil {
					errorColor.Printf("  Error getting current branch: %v\n", err)
					errorCount++
					fmt.Println(strings.Repeat("─", 60))
					continue
				}

				// Check if already on the target branch
				if currentBranch == branchName {
					skipColor.Printf("  Already on branch %s, skipping\n", branchName)
					skipCount++

					// Pull if requested
					if branchPull {
						fmt.Println("  Pulling latest changes...")
						pullCmd := exec.Command("git", "-C", repoPath, "pull")
						if err := pullCmd.Run(); err != nil {
							errorColor.Printf("  Error pulling changes: %v\n", err)
							errorCount++
						} else {
							successColor.Println("  Successfully pulled latest changes")
						}
					}

					fmt.Println(strings.Repeat("─", 60))
					continue
				}

				// Check if the branch exists
				branchExists, err := checkBranchExists(repoPath, branchName)
				if err != nil {
					errorColor.Printf("  Error checking if branch exists: %v\n", err)
					errorCount++
					fmt.Println(strings.Repeat("─", 60))
					continue
				}

				if !branchExists {
					if branchCreate {
						// Create the branch
						fmt.Printf("  Creating branch %s...\n", branchName)
						createCmd := exec.Command("git", "-C", repoPath, "checkout", "-b", branchName)
						if err := createCmd.Run(); err != nil {
							errorColor.Printf("  Error creating branch: %v\n", err)
							errorCount++
							fmt.Println(strings.Repeat("─", 60))
							continue
						}
						successColor.Printf("  Successfully created and switched to branch %s\n", branchName)
						successCount++
					} else {
						// Branch doesn't exist and --create not specified
						skipColor.Printf("  Branch %s doesn't exist (use --create to create it), skipping\n", branchName)
						skipCount++
						fmt.Println(strings.Repeat("─", 60))
						continue
					}
				} else {
					// Switch to the branch
					fmt.Printf("  Switching to branch %s...\n", branchName)
					switchCmd := exec.Command("git", "-C", repoPath, "checkout", branchName)
					if err := switchCmd.Run(); err != nil {
						errorColor.Printf("  Error switching branch: %v\n", err)
						errorCount++
						fmt.Println(strings.Repeat("─", 60))
						continue
					}
					successColor.Printf("  Successfully switched to branch %s\n", branchName)
					successCount++
				}

				// Pull if requested
				if branchPull {
					fmt.Println("  Pulling latest changes...")
					pullCmd := exec.Command("git", "-C", repoPath, "pull")
					if err := pullCmd.Run(); err != nil {
						errorColor.Printf("  Error pulling changes: %v\n", err)
						errorCount++
					} else {
						successColor.Println("  Successfully pulled latest changes")
					}
				}

				fmt.Println(strings.Repeat("─", 60))
			}

			// Print summary
			titleColor.Println("\n📊 SUMMARY")
			fmt.Printf("Total repositories processed: %d\n", successCount+errorCount+skipCount)
			if successCount > 0 {
				successColor.Printf("Successfully switched: %d\n", successCount)
			}
			if skipCount > 0 {
				skipColor.Printf("Skipped: %d\n", skipCount)
			}
			if errorCount > 0 {
				errorColor.Printf("Errors: %d\n", errorCount)
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&branchCreate, "create", false, "Create the branch if it doesn't exist")
	cmd.Flags().BoolVar(&branchPull, "pull", false, "Pull latest changes after switching branch")
	cmd.Flags().StringSliceVar(&branchRepos, "repos", []string{}, "Only switch specified repositories (by ID or name)")

	return cmd
}

// checkBranchExists checks if a branch exists in a repository
func checkBranchExists(repoPath, branchName string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--list", branchName)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(output)) != "", nil
}
