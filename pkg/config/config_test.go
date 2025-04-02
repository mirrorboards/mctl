package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-config-test")
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

	// Test successful config creation
	err = InitConfig()
	if err != nil {
		t.Fatalf("Error initializing config: %v", err)
	}

	// Check that the file exists
	configPath := filepath.Join(tempDir, configFileName)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}

	// Test that calling it a second time returns an error
	err = InitConfig()
	if err == nil {
		t.Errorf("InitConfig should return an error when file already exists")
	}
}
