package cmd

import (
	"os"
	"testing"
)

func TestAddCmd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-add-test")
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
	initCmd := newInitCmd()
	initCmd.SetArgs([]string{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("Error executing init command: %v", err)
	}

	// We're not going to actually execute git clone in the test
	// Just test the command structure and flags

	// Create the add command with test flags
	cmd := newAddCmd()

	// Test with default flags
	cmd.SetArgs([]string{"https://github.com/example/repo.git"})

	// Instead of executing, just parse the args and verify they're set correctly
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Error parsing flags: %v", err)
	}

	if addPath != "." {
		t.Errorf("Default path should be '.', got '%s'", addPath)
	}

	if addName != "" {
		t.Errorf("Default name should be empty, got '%s'", addName)
	}

	if addFlat {
		t.Errorf("Default flat should be false")
	}

	// Test with custom flags
	if err := cmd.ParseFlags([]string{"--path", "custom/path", "--name", "custom-name", "--flat"}); err != nil {
		t.Fatalf("Error parsing flags: %v", err)
	}

	if addPath != "custom/path" {
		t.Errorf("Path should be 'custom/path', got '%s'", addPath)
	}

	if addName != "custom-name" {
		t.Errorf("Name should be 'custom-name', got '%s'", addName)
	}

	if !addFlat {
		t.Errorf("Flat should be true")
	}
}
