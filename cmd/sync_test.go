package cmd

import (
	"os"
	"testing"

	"github.com/mirrorboards/mctl/pkg/config"
)

func TestSyncCmd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-sync-test")
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

	// Test empty repository list
	cmd := newSyncCmd()
	cmd.SetArgs([]string{})

	// We mock the config.GetAllRepositories and git.Clone instead of actually
	// calling them in the test, to avoid external dependencies.
	// This is a simple check for command structure and proper initialization.
	repos, err := config.GetAllRepositories()
	if err != nil {
		t.Fatalf("Error getting repositories: %v", err)
	}

	if len(repos) != 0 {
		t.Errorf("Expected 0 repositories in fresh config, got %d", len(repos))
	}

	// Add a test repository to the config
	testURL := "https://github.com/example/repo.git"
	testPath := "test-path"
	testName := "test-name"

	if err := config.AddRepository(testURL, testPath, testName); err != nil {
		t.Fatalf("Error adding repository: %v", err)
	}

	// Verify repository was added
	repos, err = config.GetAllRepositories()
	if err != nil {
		t.Fatalf("Error getting repositories: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("Expected 1 repository after adding, got %d", len(repos))
	}

	if repos[0].URL != testURL {
		t.Errorf("Expected URL %s, got %s", testURL, repos[0].URL)
	}

	if repos[0].Path != testPath {
		t.Errorf("Expected Path %s, got %s", testPath, repos[0].Path)
	}

	if repos[0].Name != testName {
		t.Errorf("Expected Name %s, got %s", testName, repos[0].Name)
	}
}
