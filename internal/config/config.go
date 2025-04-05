package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	// DefaultConfigDir is the default configuration directory name
	DefaultConfigDir = ".mirror"
	// DefaultConfigFile is the default configuration file name
	DefaultConfigFile = "mirror.toml"
	// DefaultReposDir is the default repositories directory name
	DefaultReposDir = "repositories"
	// DefaultMetadataDir is the default metadata directory name
	DefaultMetadataDir = "metadata"
	// DefaultLogsDir is the default logs directory name
	DefaultLogsDir = "logs"
	// DefaultCacheDir is the default cache directory name
	DefaultCacheDir = "cache"
	// DefaultStatusCacheDir is the default status cache directory name
	DefaultStatusCacheDir = "status"
	// DefaultOperationsLogFile is the default operations log file name
	DefaultOperationsLogFile = "operations.log"
	// DefaultAuditLogFile is the default audit log file name
	DefaultAuditLogFile = "audit.log"
)

// Config represents the main configuration structure
type Config struct {
	Global       GlobalConfig       `toml:"global"`
	Repositories []RepositoryConfig `toml:"repositories"`
}

// GlobalConfig represents global configuration settings
type GlobalConfig struct {
	DefaultBranch      string `toml:"default_branch"`
	ParallelOperations int    `toml:"parallel_operations"`
	DefaultRemote      string `toml:"default_remote"`
}

// RepositoryConfig represents a repository configuration
type RepositoryConfig struct {
	ID     string `toml:"id"`
	Name   string `toml:"name"`
	Path   string `toml:"path"`
	URL    string `toml:"url"`
	Branch string `toml:"branch"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Global: GlobalConfig{
			DefaultBranch:      "main",
			ParallelOperations: 4,
			DefaultRemote:      "origin",
		},
		Repositories: []RepositoryConfig{},
	}
}

// LoadConfig loads configuration from the specified directory
func LoadConfig(baseDir string) (*Config, error) {
	configPath := filepath.Join(baseDir, DefaultConfigDir, DefaultConfigFile)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found at %s", configPath)
	}

	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("error decoding configuration file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to the specified directory
func SaveConfig(config *Config, baseDir string) error {
	configDir := filepath.Join(baseDir, DefaultConfigDir)
	configPath := filepath.Join(configDir, DefaultConfigFile)

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("error creating configuration directory: %w", err)
	}

	// Create or truncate the file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("error creating configuration file: %w", err)
	}
	defer file.Close()

	// Set secure permissions
	if err := os.Chmod(configPath, 0600); err != nil {
		return fmt.Errorf("error setting configuration file permissions: %w", err)
	}

	// Encode configuration to TOML
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("error encoding configuration: %w", err)
	}

	return nil
}

// IsInitialized checks if MCTL is initialized in the specified directory
func IsInitialized(baseDir string) bool {
	configPath := filepath.Join(baseDir, DefaultConfigDir, DefaultConfigFile)
	_, err := os.Stat(configPath)
	return err == nil
}

// GetConfigDirPath returns the path to the configuration directory
func GetConfigDirPath(baseDir string) string {
	return filepath.Join(baseDir, DefaultConfigDir)
}

// GetMetadataDirPath returns the path to the metadata directory
func GetMetadataDirPath(baseDir string) string {
	return filepath.Join(baseDir, DefaultConfigDir, DefaultMetadataDir)
}

// GetLogsDirPath returns the path to the logs directory
func GetLogsDirPath(baseDir string) string {
	return filepath.Join(baseDir, DefaultConfigDir, DefaultLogsDir)
}

// GetCacheDirPath returns the path to the cache directory
func GetCacheDirPath(baseDir string) string {
	return filepath.Join(baseDir, DefaultConfigDir, DefaultCacheDir)
}

// GetStatusCacheDirPath returns the path to the status cache directory
func GetStatusCacheDirPath(baseDir string) string {
	return filepath.Join(baseDir, DefaultConfigDir, DefaultCacheDir, DefaultStatusCacheDir)
}

// GetRepositoriesDirPath returns the path to the repositories directory
func GetRepositoriesDirPath(baseDir string) string {
	return filepath.Join(baseDir, DefaultReposDir)
}

// GetOperationsLogFilePath returns the path to the operations log file
func GetOperationsLogFilePath(baseDir string) string {
	return filepath.Join(baseDir, DefaultConfigDir, DefaultLogsDir, DefaultOperationsLogFile)
}

// GetAuditLogFilePath returns the path to the audit log file
func GetAuditLogFilePath(baseDir string) string {
	return filepath.Join(baseDir, DefaultConfigDir, DefaultLogsDir, DefaultAuditLogFile)
}
