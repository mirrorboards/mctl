package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Clone clones a Git repository to the specified path.
// If targetName is provided, it will be used as the directory name within the path.
// If targetName is empty, the repository will be cloned directly into the path.
func Clone(gitURL, targetPath, targetName string) error {
	// Ensure the target directory exists
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Build the clone destination
	clonePath := targetPath
	if targetName != "" {
		clonePath = filepath.Join(targetPath, targetName)
	}

	// Execute git clone
	cmd := exec.Command("git", "clone", gitURL, clonePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

// RemoveDirectory removes a directory and all its contents.
// This is used to remove a Git repository directory.
func RemoveDirectory(path string) error {
	// Check if the directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Directory doesn't exist, nothing to do
		return nil
	}

	// Remove the directory and all its contents
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
	}

	return nil
}

// RemoveEmptyParentDirectories removes empty parent directories up to the current working directory.
// This is used to clean up empty directories after removing a Git repository.
func RemoveEmptyParentDirectories(path string) error {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Convert path to absolute path if it's relative
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get the parent directory
	dir := filepath.Dir(absPath)

	// Walk up the directory tree until we reach the current directory
	for dir != currentDir && dir != "/" && dir != "." {
		// Check if the directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// Directory doesn't exist, move up
			dir = filepath.Dir(dir)
			continue
		}

		// Check if the directory is empty
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", dir, err)
		}

		if len(entries) == 0 {
			// Directory is empty, remove it
			if err := os.Remove(dir); err != nil {
				return fmt.Errorf("failed to remove directory %s: %w", dir, err)
			}
		} else {
			// Directory is not empty, stop here
			break
		}

		// Move up to the next parent directory
		dir = filepath.Dir(dir)
	}

	return nil
}
