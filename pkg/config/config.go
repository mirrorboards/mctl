package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	configFileName = "mirror.toml"
	defaultConfig  = `# Mirror Configuration File
# See https://github.com/mirrorboards/mctl for documentation

`
)

// InitConfig creates an empty config file in the current directory
func InitConfig() error {
	// Check if file already exists
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(currentDir, configFileName)
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	// Create the file
	err = os.WriteFile(configPath, []byte(defaultConfig), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
