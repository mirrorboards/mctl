package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCmd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-test")
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

	// Execute the init command
	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Error executing init command: %v", err)
	}

	// Check that the file was created
	configPath := filepath.Join(tempDir, "mirror.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}
}
