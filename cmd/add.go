package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mirrorboards/mctl/pkg/config"
	"github.com/mirrorboards/mctl/pkg/git"
	"github.com/spf13/cobra"
)

var (
	addPath  string
	addName  string
	addFlat  bool
	addNoOrg bool
)

// extractOrgName extracts the organization name from a Git URL
// e.g., git@github.com:LFDT-Lockness/cggmp21.git -> LFDT-Lockness
func extractOrgName(gitURL string) string {
	// Remove .git extension if present
	gitURL = strings.TrimSuffix(gitURL, ".git")

	// Handle SSH URLs (git@github.com:org/repo)
	sshRegex := regexp.MustCompile(`git@[^:]+:([^/]+)/`)
	matches := sshRegex.FindStringSubmatch(gitURL)
	if len(matches) > 1 {
		return matches[1]
	}

	// Handle HTTPS URLs (https://github.com/org/repo)
	httpsRegex := regexp.MustCompile(`https?://[^/]+/([^/]+)/`)
	matches = httpsRegex.FindStringSubmatch(gitURL)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [git-url] [path]",
		Short: "Add a Git repository",
		Long:  `Add a Git repository to the mirror.toml configuration and clone it to the specified path.`,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			gitURL := args[0]

			// Override path if second argument is provided
			targetPath := addPath
			if len(args) > 1 {
				targetPath = args[1]
			}

			// Extract repository name from URL if not provided
			repoName := addName
			if repoName == "" {
				repoName = config.ExtractRepoName(gitURL)
			}

			// Default behavior is to use organization structure unless --no-org is specified
			if !addNoOrg {
				orgName := extractOrgName(gitURL)
				if orgName == "" {
					return fmt.Errorf("could not extract organization name from URL: %s", gitURL)
				}

				// Update the path to use the org name
				if targetPath == "." {
					targetPath = orgName
				} else {
					targetPath = filepath.Join(targetPath, orgName)
				}

				// Keep the repo name for the subdirectory
				addFlat = false
			} else {
				// For flat mode with name, we adjust the path to include the name
				if addFlat && addName != "" {
					// In flat mode with a name, we use the name as directory
					if targetPath == "." {
						// If we're in current directory, use just the name
						targetPath = addName
					} else {
						// Otherwise, join the path with the name
						targetPath = filepath.Join(targetPath, addName)
					}
					// Record the name as empty since we're treating it as part of the path
					addName = ""
					repoName = ""
				}
			}

			// Determine where to clone
			targetName := ""

			// Handle cloning modes
			if addFlat {
				// Flat mode: clone directly into the target path, no subdirectory
				targetName = ""

				// Check if the target path already contains a git repository
				gitDir := filepath.Join(targetPath, ".git")
				if _, err := os.Stat(gitDir); err == nil {
					return fmt.Errorf("destination path %s already contains a git repository", targetPath)
				}
			} else if targetPath == "." {
				// Current directory mode: clone to a subdirectory with repo name
				targetName = repoName

				// Check if the target directory exists and contains a git repository
				gitDir := filepath.Join(targetPath, targetName, ".git")
				if _, err := os.Stat(gitDir); err == nil {
					return fmt.Errorf("repository already exists at %s/%s", targetPath, targetName)
				}
			} else {
				// Normal mode: use the provided name or default to repo name
				targetName = addName
				if targetName == "" {
					targetName = repoName
				}

				// Check if the target directory exists and contains a git repository
				gitDir := filepath.Join(targetPath, targetName, ".git")
				if _, err := os.Stat(gitDir); err == nil {
					return fmt.Errorf("repository already exists at %s/%s", targetPath, targetName)
				}
			}

			// Add repository to the configuration with appropriate name
			configName := repoName
			if addFlat {
				configName = "" // Empty name means it's cloned directly into the path
			}

			if err := config.AddRepository(gitURL, targetPath, configName); err != nil {
				return err
			}

			// Clone the repository
			if err := git.Clone(gitURL, targetPath, targetName); err != nil {
				return err
			}

			if targetName == "" {
				fmt.Printf("Successfully added and cloned repository to %s\n", targetPath)
			} else {
				fmt.Printf("Successfully added and cloned repository to %s/%s\n", targetPath, targetName)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&addPath, "path", "p", ".", "Path where to clone the repository")
	cmd.Flags().StringVarP(&addName, "name", "n", "", "Custom name for the repository (defaults to repo name)")
	cmd.Flags().BoolVar(&addFlat, "flat", false, "Clone directly into the path instead of creating a subdirectory")
	cmd.Flags().BoolVar(&addNoOrg, "no-org", false, "Don't use organization name as directory structure (legacy behavior)")

	return cmd
}
