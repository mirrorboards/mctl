package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mirrorboards/mctl/pkg/config"
)

func TestStatusCmd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-status-test")
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

	// Initialize an empty mirror.toml file
	if err := config.InitConfig(); err != nil {
		t.Fatalf("Error initializing config: %v", err)
	}

	// Create a test repository structure
	testRepos := []struct {
		url  string
		path string
		name string
	}{
		{"https://github.com/test1/repo1.git", "./path1", "name1"},
		{"https://github.com/test2/repo2.git", "./path2", "name2"},
		{"https://github.com/test3/repo3.git", "./path3", ""},
	}

	// Add the repositories to the config and create dummy directories
	for _, repo := range testRepos {
		if err := config.AddRepository(repo.url, repo.path, repo.name); err != nil {
			t.Fatalf("Error adding repository %s: %v", repo.url, err)
		}

		// Create the directory structure
		var dirPath string
		if repo.name == "" {
			dirPath = repo.path
		} else {
			dirPath = filepath.Join(repo.path, repo.name)
		}

		// Create the repository directory with a .git subdirectory
		gitDir := filepath.Join(dirPath, ".git")
		if err := os.MkdirAll(gitDir, 0755); err != nil {
			t.Fatalf("Error creating test directory %s: %v", gitDir, err)
		}
	}

	// Create the status command
	cmd := newStatusCmd()
	cmd.SetArgs([]string{})

	// We're not going to actually execute the command as it would try to run git commands,
	// but we'll verify that the basic structure is in place
	// This is primarily to ensure that the command doesn't panic
	// The core functions are individually testable

	// Test getGitBranch function
	if branch, err := getGitBranch("invalid-path"); err == nil {
		t.Errorf("getGitBranch on invalid path should return an error, got branch: %s", branch)
	}

	// Test getGitStatus function
	if status, isDirty, err := getGitStatus("invalid-path"); err == nil {
		t.Errorf("getGitStatus on invalid path should return an error, got status: %s, isDirty: %v", status, isDirty)
	}

	// Test getChangedFiles function
	if modified, staged, untracked, err := getChangedFiles("invalid-path"); err == nil {
		t.Errorf("getChangedFiles on invalid path should return an error, got modified: %v, staged: %v, untracked: %v",
			modified, staged, untracked)
	}

	// Test formatGitStatus function
	testStatus := "On branch main\nChanges not staged for commit:\n  modified: file.txt\n"
	formattedStatus := formatGitStatus(testStatus)

	// Ensure the formatGitStatus function properly indents and removes headers
	if formattedStatus != "    modified: file.txt\n" {
		t.Errorf("formatGitStatus did not format correctly, got: %s", formattedStatus)
	}
}
