package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	configFileName = "mirror.toml"
	defaultConfig  = `# Mirror Configuration File
# See https://github.com/mirrorboards/mctl for documentation

`
)

// Repository represents a git repository in the configuration
type Repository struct {
	URL  string `toml:"url"`
	Path string `toml:"path"`
	Name string `toml:"name,omitempty"`
}

// Config represents the structure of the mirror.toml file
type Config struct {
	Repositories []Repository `toml:"repositories"`
}

// GetConfigFileName returns the name of the configuration file
func GetConfigFileName() string {
	return configFileName
}

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

// ExtractRepoName extracts the repository name from a Git URL
func ExtractRepoName(gitURL string) string {
	// Remove .git extension if present
	gitURL = strings.TrimSuffix(gitURL, ".git")

	// Get the last part of the URL (after the last / or :)
	parts := regexp.MustCompile(`[/:]+`).Split(gitURL, -1)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}

// AddRepository adds a new repository to the mirror.toml configuration
func AddRepository(gitURL, targetPath, name string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(currentDir, configFileName)

	// Check if config file exists, create it if it doesn't
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := InitConfig(); err != nil {
			return err
		}
	}

	// Read existing config
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Prepare the new repository entry
	repoEntry := fmt.Sprintf("\n[[repositories]]\nurl = \"%s\"\npath = \"%s\"\n", gitURL, targetPath)
	if name != "" {
		repoEntry += fmt.Sprintf("name = \"%s\"\n", name)
	}

	// Append the new entry to the config
	newConfig := string(configData) + repoEntry

	// Write the updated config back to the file
	err = os.WriteFile(configPath, []byte(newConfig), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated config file: %w", err)
	}

	return nil
}

// GetAllRepositories returns all repositories defined in the mirror.toml file
func GetAllRepositories() ([]Repository, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(currentDir, configFileName)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("mirror.toml not found, run 'mctl init' first")
	}

	// Read config file
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse mirror.toml: %w", err)
	}

	return config.Repositories, nil
}
