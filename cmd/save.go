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
	saveAll bool
)

func newSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save [commit-message]",
		Short: "Save changes to repositories",
		Long:  `Add, commit, and push changes to repositories. Specify a commit message or use default "mirrorboards/mirrorboards".`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Define default commit message
			commitMessage := "mirrorboards/mirrorboards"
			if len(args) > 0 && args[0] != "" {
				commitMessage = args[0]
			}

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
			titleColor.Printf("\n✨ SAVING CHANGES TO REPOSITORIES ✨\n")
			titleColor.Printf("Commit message: \"%s\"\n", commitMessage)
			fmt.Println(strings.Repeat("─", 60))

			successCount := 0
			errorCount := 0
			skipCount := 0

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
					// Repository doesn't exist, skip
					continue
				}

				// Display repository name and path
				repoNameColor.Printf("• %s ", filepath.Base(repoPath))
				pathColor.Printf("(%s)\n", repoPath)

				// Check if repo has changes
				isDirty, err := hasChanges(repoPath)
				if err != nil {
					errorColor.Printf("  Error checking status: %v\n", err)
					errorCount++
					continue
				}

				// Only save if there are changes or saveAll flag is set
				if !isDirty && !saveAll {
					skipColor.Println("  No changes to save, skipping")
					skipCount++
					fmt.Println(strings.Repeat("─", 60))
					continue
				}

				// Step 1: git add --all
				fmt.Println("  Adding changes...")
				addCmd := exec.Command("git", "-C", repoPath, "add", "--all")
				if err := addCmd.Run(); err != nil {
					errorColor.Printf("  Failed to add changes: %v\n", err)
					errorCount++
					fmt.Println(strings.Repeat("─", 60))
					continue
				}

				// Step 2: git commit -m "message"
				fmt.Println("  Committing changes...")
				commitCmd := exec.Command("git", "-C", repoPath, "commit", "-m", commitMessage)
				if output, err := commitCmd.CombinedOutput(); err != nil {
					// Check if it's just "nothing to commit" error, which is fine
					if strings.Contains(string(output), "nothing to commit") {
						skipColor.Println("  Nothing to commit, working tree clean")
						skipCount++
						fmt.Println(strings.Repeat("─", 60))
						continue
					}

					errorColor.Printf("  Failed to commit changes: %v\n%s\n", err, string(output))
					errorCount++
					fmt.Println(strings.Repeat("─", 60))
					continue
				}

				// Step 3: git push
				fmt.Println("  Pushing changes...")
				pushCmd := exec.Command("git", "-C", repoPath, "push")
				if output, err := pushCmd.CombinedOutput(); err != nil {
					errorColor.Printf("  Failed to push changes: %v\n%s\n", err, string(output))
					errorCount++
					fmt.Println(strings.Repeat("─", 60))
					continue
				}

				successColor.Println("  Changes successfully saved and pushed")
				successCount++
				fmt.Println(strings.Repeat("─", 60))
			}

			// Print summary
			titleColor.Println("\n📊 SUMMARY")
			fmt.Printf("Total repositories processed: %d\n", successCount+errorCount+skipCount)
			if successCount > 0 {
				successColor.Printf("Successfully saved: %d\n", successCount)
			}
			if skipCount > 0 {
				skipColor.Printf("Skipped (no changes): %d\n", skipCount)
			}
			if errorCount > 0 {
				errorColor.Printf("Errors: %d\n", errorCount)
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&saveAll, "all", "a", false, "Save all repositories even if they have no changes")
	return cmd
}

// hasChanges checks if a repository has uncommitted changes
func hasChanges(repoPath string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}
