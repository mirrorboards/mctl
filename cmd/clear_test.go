package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mirrorboards/mctl/pkg/config"
)

func TestClearCmd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-clear-test")
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

	// Create the clear command
	cmd := newClearCmd()
	cmd.SetArgs([]string{})

	// Execute the command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Error executing clear command: %v", err)
	}

	// Verify that the directories were removed
	for _, repo := range testRepos {
		var dirPath string
		if repo.name == "" {
			dirPath = repo.path
		} else {
			dirPath = filepath.Join(repo.path, repo.name)
		}

		if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
			t.Errorf("Directory %s should have been removed", dirPath)
		}
	}

	// Verify that the config file still exists
	configPath := filepath.Join(tempDir, config.GetConfigFileName())
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file should still exist")
	}
}
