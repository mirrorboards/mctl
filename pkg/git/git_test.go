package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClone(t *testing.T) {
	// Skip actual clone tests when not in full test environment
	// This is a placeholder test since we don't want to perform actual git operations in unit tests
	if os.Getenv("MCTL_RUN_GIT_TESTS") != "true" {
		t.Skip("Skipping git clone test. Set MCTL_RUN_GIT_TESTS=true to enable")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-git-test")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases are skipped by default to avoid network calls in tests
	// This is here mainly for documentation and manual testing
	testCases := []struct {
		name       string
		gitURL     string
		targetPath string
		targetName string
	}{
		{
			name:       "Clone to subdirectory",
			gitURL:     "https://github.com/spf13/cobra",
			targetPath: tempDir,
			targetName: "cobra-test",
		},
		{
			name:       "Clone flat",
			gitURL:     "https://github.com/spf13/cobra",
			targetPath: filepath.Join(tempDir, "flat-test"),
			targetName: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Clone(tc.gitURL, tc.targetPath, tc.targetName)
			if err != nil {
				t.Fatalf("Error cloning repository: %v", err)
			}

			// Check if the repository was cloned
			var repoPath string
			if tc.targetName == "" {
				repoPath = tc.targetPath
			} else {
				repoPath = filepath.Join(tc.targetPath, tc.targetName)
			}

			gitDir := filepath.Join(repoPath, ".git")
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				t.Errorf("Git directory not found at %s", gitDir)
			}
		})
	}
}

func TestRemoveDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-git-remove-test")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test directory structure
	testDir := filepath.Join(tempDir, "test-dir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Error creating test directory: %v", err)
	}

	// Create a file in the test directory
	testFile := filepath.Join(testDir, "test-file.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Error creating test file: %v", err)
	}

	// Test removing the directory
	if err := RemoveDirectory(testDir); err != nil {
		t.Fatalf("Error removing directory: %v", err)
	}

	// Verify the directory is removed
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Errorf("Directory should have been removed: %s", testDir)
	}

	// Test removing a non-existent directory
	if err := RemoveDirectory(filepath.Join(tempDir, "non-existent")); err != nil {
		t.Errorf("Removing non-existent directory should not error: %v", err)
	}
}

func TestRemoveEmptyParentDirectories(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-git-remove-parent-test")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Error changing to temporary directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create a nested directory structure
	nestedDir := filepath.Join("parent", "child", "grandchild")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Error creating nested directory: %v", err)
	}

	// Test removing empty parent directories
	if err := RemoveEmptyParentDirectories(nestedDir); err != nil {
		t.Fatalf("Error removing empty parent directories: %v", err)
	}

	// Verify parent directories are removed
	if _, err := os.Stat("parent"); !os.IsNotExist(err) {
		t.Errorf("Parent directory should have been removed")
	}

	// Test with non-empty directory
	parentDir := filepath.Join("parent2", "child2")
	childDir := filepath.Join(parentDir, "grandchild2")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("Error creating nested directory: %v", err)
	}

	// Create a file in the parent directory to make it non-empty
	parentFile := filepath.Join("parent2", "file.txt")
	if err := os.WriteFile(parentFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Error creating test file: %v", err)
	}

	// Remove the grandchild directory
	if err := os.Remove(childDir); err != nil {
		t.Fatalf("Error removing grandchild directory: %v", err)
	}

	// Test removing empty parent directories
	if err := RemoveEmptyParentDirectories(childDir); err != nil {
		t.Fatalf("Error removing empty parent directories: %v", err)
	}

	// Verify parent directory is not removed because it has a file
	if _, err := os.Stat("parent2"); os.IsNotExist(err) {
		t.Errorf("Non-empty parent directory should not have been removed")
	}
}
