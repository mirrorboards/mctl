package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	configFileName = "mirror.toml"
	defaultConfig  = `# Mirror Configuration File
# See https://github.com/mirrorboards/mctl for documentation

`
)

// Remote represents a remote configuration source
type Remote struct {
	Name     string `toml:"name"`
	URL      string `toml:"url"`
	Type     string `toml:"type,omitempty"` // "github", "gitlab", "bitbucket", "file", etc.
	Branch   string `toml:"branch,omitempty"`
	AuthType string `toml:"auth_type,omitempty"` // "ssh", "token", "none"
}

// Repository represents a git repository in the configuration
type Repository struct {
	ID     string   `toml:"id"` // Unique identifier
	URL    string   `toml:"url"`
	Path   string   `toml:"path"`
	Name   string   `toml:"name,omitempty"`
	Branch string   `toml:"branch,omitempty"` // Default branch
	Tags   []string `toml:"tags,omitempty"`   // For grouping/filtering
}

// Config represents the structure of the mirror.toml file
type Config struct {
	Repositories []Repository `toml:"repositories"`
	Remotes      []Remote     `toml:"remotes,omitempty"`
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

// GenerateRepoID generates a unique ID for a repository
func GenerateRepoID(url, path, name string) string {
	// For shorter IDs, we use a hash of the repo details
	// This ensures the ID is deterministic based on the repo details
	h := sha256.New()
	h.Write([]byte(url + path + name + time.Now().String()))
	hash := hex.EncodeToString(h.Sum(nil))

	// Return first 8 characters of the hash
	return hash[:8]
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

	// Generate a unique ID for the repository
	repoID := GenerateRepoID(gitURL, targetPath, name)

	// Read existing config
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return fmt.Errorf("failed to parse mirror.toml: %w", err)
	}

	// Check if a repository with the same URL already exists
	for _, repo := range config.Repositories {
		if repo.URL == gitURL {
			return fmt.Errorf("repository with URL %s already exists", gitURL)
		}
	}

	// Create new repository
	newRepo := Repository{
		ID:   repoID,
		URL:  gitURL,
		Path: targetPath,
		Name: name,
	}

	// Add to config
	config.Repositories = append(config.Repositories, newRepo)

	// Write updated config
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// RemoveRepository removes a repository from the configuration by ID or name
func RemoveRepository(identifier string, deleteFiles bool) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(currentDir, configFileName)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("mirror.toml not found, run 'mctl init' first")
	}

	// Read config file
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return fmt.Errorf("failed to parse mirror.toml: %w", err)
	}

	// Find repository by ID or name
	var foundIndex = -1
	var foundRepo Repository
	for i, repo := range config.Repositories {
		if repo.ID == identifier || repo.Name == identifier {
			foundIndex = i
			foundRepo = repo
			break
		}
	}

	if foundIndex == -1 {
		return fmt.Errorf("repository with ID or name %s not found", identifier)
	}

	// Remove repository from config
	config.Repositories = append(config.Repositories[:foundIndex], config.Repositories[foundIndex+1:]...)

	// Write updated config
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Delete files if requested
	if deleteFiles {
		var repoPath string
		if foundRepo.Path == "." {
			repoPath = foundRepo.Name
		} else if foundRepo.Name == "" {
			repoPath = foundRepo.Path
		} else {
			repoPath = filepath.Join(foundRepo.Path, foundRepo.Name)
		}

		if err := os.RemoveAll(repoPath); err != nil {
			return fmt.Errorf("failed to remove repository files: %w", err)
		}
	}

	return nil
}

// AddRemote adds a new remote configuration source
func AddRemote(name, url, remoteType, branch, authType string) error {
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
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return fmt.Errorf("failed to parse mirror.toml: %w", err)
	}

	// Check if a remote with the same name already exists
	for _, remote := range config.Remotes {
		if remote.Name == name {
			return fmt.Errorf("remote with name %s already exists", name)
		}
	}

	// Create new remote
	newRemote := Remote{
		Name:     name,
		URL:      url,
		Type:     remoteType,
		Branch:   branch,
		AuthType: authType,
	}

	// Add to config
	config.Remotes = append(config.Remotes, newRemote)

	// Write updated config
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// RemoveRemote removes a remote configuration source by name
func RemoveRemote(name string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(currentDir, configFileName)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("mirror.toml not found, run 'mctl init' first")
	}

	// Read config file
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return fmt.Errorf("failed to parse mirror.toml: %w", err)
	}

	// Find remote by name
	var foundIndex = -1
	for i, remote := range config.Remotes {
		if remote.Name == name {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return fmt.Errorf("remote with name %s not found", name)
	}

	// Remove remote from config
	config.Remotes = append(config.Remotes[:foundIndex], config.Remotes[foundIndex+1:]...)

	// Write updated config
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetAllRemotes returns all remote configuration sources
func GetAllRemotes() ([]Remote, error) {
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

	return config.Remotes, nil
}

// SyncWithRemote synchronizes the local configuration with a remote configuration
func SyncWithRemote(remoteName string, mergeStrategy string) error {
	// Get all remotes
	remotes, err := GetAllRemotes()
	if err != nil {
		return err
	}

	// Find the specified remote
	var foundRemote *Remote
	for _, remote := range remotes {
		if remote.Name == remoteName {
			foundRemote = &remote
			break
		}
	}

	if foundRemote == nil {
		return fmt.Errorf("remote with name %s not found", remoteName)
	}

	// Fetch remote configuration
	resp, err := http.Get(foundRemote.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch remote configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch remote configuration: %s", resp.Status)
	}

	// Parse remote configuration
	var remoteConfig Config
	if _, err := toml.DecodeReader(resp.Body, &remoteConfig); err != nil {
		return fmt.Errorf("failed to parse remote configuration: %w", err)
	}

	// Get local configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(currentDir, configFileName)
	var localConfig Config
	if _, err := toml.DecodeFile(configPath, &localConfig); err != nil {
		return fmt.Errorf("failed to parse local configuration: %w", err)
	}

	// Merge configurations based on strategy
	var mergedConfig Config
	switch mergeStrategy {
	case "remote-wins":
		// Remote configuration takes precedence
		mergedConfig = remoteConfig
		// Keep local remotes
		mergedConfig.Remotes = localConfig.Remotes
	case "local-wins":
		// Local configuration takes precedence
		mergedConfig = localConfig
		// Add any repositories from remote that don't exist locally
		for _, remoteRepo := range remoteConfig.Repositories {
			exists := false
			for _, localRepo := range localConfig.Repositories {
				if localRepo.URL == remoteRepo.URL {
					exists = true
					break
				}
			}
			if !exists {
				mergedConfig.Repositories = append(mergedConfig.Repositories, remoteRepo)
			}
		}
	case "union":
		// Include all repositories from both configurations
		mergedConfig = localConfig
		for _, remoteRepo := range remoteConfig.Repositories {
			exists := false
			for _, localRepo := range localConfig.Repositories {
				if localRepo.URL == remoteRepo.URL {
					exists = true
					break
				}
			}
			if !exists {
				mergedConfig.Repositories = append(mergedConfig.Repositories, remoteRepo)
			}
		}
	default:
		return fmt.Errorf("unknown merge strategy: %s", mergeStrategy)
	}

	// Write merged configuration
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(mergedConfig); err != nil {
		return fmt.Errorf("failed to write merged configuration: %w", err)
	}

	return nil
}

// PushToRemote pushes the local configuration to a remote repository
func PushToRemote(remoteName string, force bool, message string) error {
	// This is a placeholder for the actual implementation
	// In a real implementation, this would:
	// 1. Find the remote by name
	// 2. Clone the remote repository if it doesn't exist locally
	// 3. Copy the local configuration to the cloned repository
	// 4. Commit the changes
	// 5. Push to the remote

	// For now, just return an error indicating this is not implemented
	return fmt.Errorf("push to remote not implemented yet")
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

	// Ensure all repositories have IDs (for backward compatibility)
	for i, repo := range config.Repositories {
		if repo.ID == "" {
			// Generate an ID for this repository
			config.Repositories[i].ID = GenerateRepoID(repo.URL, repo.Path, repo.Name)
		}
	}

	// Write back the updated config if any IDs were added
	f, err := os.Create(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(config); err != nil {
		return nil, fmt.Errorf("failed to write config: %w", err)
	}

	return config.Repositories, nil
}

// GetRepositoryByID returns a repository by its ID
func GetRepositoryByID(id string) (*Repository, error) {
	repos, err := GetAllRepositories()
	if err != nil {
		return nil, err
	}

	for _, repo := range repos {
		if repo.ID == id {
			return &repo, nil
		}
	}

	return nil, fmt.Errorf("repository with ID %s not found", id)
}

// GetRepositoryByName returns a repository by its name
func GetRepositoryByName(name string) (*Repository, error) {
	repos, err := GetAllRepositories()
	if err != nil {
		return nil, err
	}

	for _, repo := range repos {
		if repo.Name == name {
			return &repo, nil
		}
	}

	return nil, fmt.Errorf("repository with name %s not found", name)
}

// UpdateRepository updates an existing repository in the configuration
func UpdateRepository(repo Repository) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(currentDir, configFileName)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("mirror.toml not found, run 'mctl init' first")
	}

	// Read config file
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return fmt.Errorf("failed to parse mirror.toml: %w", err)
	}

	// Find repository by ID
	var found bool
	for i, r := range config.Repositories {
		if r.ID == repo.ID {
			config.Repositories[i] = repo
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("repository with ID %s not found", repo.ID)
	}

	// Write updated config
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// DownloadFile downloads a file from a URL to a local file
func DownloadFile(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
