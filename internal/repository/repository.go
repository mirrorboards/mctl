package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mirrorboards/mctl/internal/config"
)

// Status represents the status of a repository
type Status string

const (
	// StatusClean indicates the repository has no changes
	StatusClean Status = "CLEAN"
	// StatusModified indicates the repository has uncommitted changes
	StatusModified Status = "MODIFIED"
	// StatusAhead indicates the repository is ahead of the remote
	StatusAhead Status = "AHEAD"
	// StatusBehind indicates the repository is behind the remote
	StatusBehind Status = "BEHIND"
	// StatusDiverged indicates the repository has diverged from the remote
	StatusDiverged Status = "DIVERGED"
	// StatusUnknown indicates the repository status could not be determined
	StatusUnknown Status = "UNKNOWN"
)

// Metadata represents repository metadata
type Metadata struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Basic  BasicInfo  `json:"basic"`
	Status StatusInfo `json:"status"`
	// Reserved for future extensions
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// BasicInfo contains basic repository information
type BasicInfo struct {
	CreationDate time.Time `json:"creation_date"`
	LastSync     time.Time `json:"last_sync"`
}

// StatusInfo contains repository status information
type StatusInfo struct {
	Current Status `json:"current"`
	Branch  string `json:"branch"`
}

// Repository represents a Git repository managed by MCTL
type Repository struct {
	Config   config.RepositoryConfig
	Metadata Metadata
	BaseDir  string
}

// New creates a new Repository instance
func New(cfg config.RepositoryConfig, baseDir string) *Repository {
	return &Repository{
		Config:  cfg,
		BaseDir: baseDir,
		Metadata: Metadata{
			ID:   cfg.ID,
			Name: cfg.Name,
			Basic: BasicInfo{
				CreationDate: time.Now(),
				LastSync:     time.Time{},
			},
			Status: StatusInfo{
				Current: StatusUnknown,
				Branch:  cfg.Branch,
			},
		},
	}
}

// FullPath returns the full path to the repository
func (r *Repository) FullPath() string {
	return filepath.Join(r.BaseDir, r.Config.Path)
}

// MetadataPath returns the path to the repository metadata file
func (r *Repository) MetadataPath() string {
	return filepath.Join(
		config.GetMetadataDirPath(r.BaseDir),
		fmt.Sprintf("%s.json", r.Config.ID),
	)
}

// Clone clones the repository
func (r *Repository) Clone() error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(r.FullPath())
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("error creating parent directory: %w", err)
	}

	// Build clone command
	args := []string{"clone"}
	if r.Config.Branch != "" {
		args = append(args, "--branch", r.Config.Branch)
	}
	args = append(args, r.Config.URL, r.FullPath())

	// Execute git clone
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, output)
	}

	// Update metadata
	r.Metadata.Status.Current = StatusClean
	return r.SaveMetadata()
}

// SaveMetadata saves repository metadata
func (r *Repository) SaveMetadata() error {
	// Ensure metadata directory exists
	metadataDir := config.GetMetadataDirPath(r.BaseDir)
	if err := os.MkdirAll(metadataDir, 0700); err != nil {
		return fmt.Errorf("error creating metadata directory: %w", err)
	}

	// Marshal metadata to JSON
	data, err := json.MarshalIndent(r.Metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling metadata: %w", err)
	}

	// Write metadata file
	if err := os.WriteFile(r.MetadataPath(), data, 0600); err != nil {
		return fmt.Errorf("error writing metadata file: %w", err)
	}

	return nil
}

// LoadMetadata loads repository metadata
func (r *Repository) LoadMetadata() error {
	data, err := os.ReadFile(r.MetadataPath())
	if err != nil {
		return fmt.Errorf("error reading metadata file: %w", err)
	}

	if err := json.Unmarshal(data, &r.Metadata); err != nil {
		return fmt.Errorf("error unmarshaling metadata: %w", err)
	}

	return nil
}

// UpdateStatus updates the repository status
func (r *Repository) UpdateStatus() error {
	// Check if repository exists
	if _, err := os.Stat(r.FullPath()); os.IsNotExist(err) {
		r.Metadata.Status.Current = StatusUnknown
		return nil
	}

	// Get current branch
	branch, err := r.GetCurrentBranch()
	if err != nil {
		return err
	}
	r.Metadata.Status.Branch = branch

	// Check for local changes
	hasChanges, err := r.HasLocalChanges()
	if err != nil {
		return err
	}

	// Check remote status
	if hasChanges {
		r.Metadata.Status.Current = StatusModified
	} else {
		// Fetch from remote
		if err := r.Fetch(); err != nil {
			// If fetch fails, we can still report local status
			r.Metadata.Status.Current = StatusClean
			return nil
		}

		// Check relationship with remote
		ahead, behind, err := r.GetRemoteStatus()
		if err != nil {
			return err
		}

		if ahead > 0 && behind > 0 {
			r.Metadata.Status.Current = StatusDiverged
		} else if ahead > 0 {
			r.Metadata.Status.Current = StatusAhead
		} else if behind > 0 {
			r.Metadata.Status.Current = StatusBehind
		} else {
			r.Metadata.Status.Current = StatusClean
		}
	}

	return r.SaveMetadata()
}

// GetCurrentBranch returns the current branch name
func (r *Repository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "-C", r.FullPath(), "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// HasLocalChanges checks if the repository has uncommitted changes
func (r *Repository) HasLocalChanges() (bool, error) {
	cmd := exec.Command("git", "-C", r.FullPath(), "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("error checking for local changes: %w", err)
	}
	return len(output) > 0, nil
}

// Fetch fetches updates from the remote
func (r *Repository) Fetch() error {
	cmd := exec.Command("git", "-C", r.FullPath(), "fetch")
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error fetching from remote: %w", err)
	}
	return nil
}

// GetRemoteStatus returns the number of commits ahead and behind the remote
func (r *Repository) GetRemoteStatus() (int, int, error) {
	branch, err := r.GetCurrentBranch()
	if err != nil {
		return 0, 0, err
	}

	remote := "origin"
	if r.Config.Branch != "" {
		remote = r.Config.Branch
	}

	// Get ahead count
	cmdAhead := exec.Command(
		"git", "-C", r.FullPath(),
		"rev-list", "--count", fmt.Sprintf("%s/%s..%s", remote, branch, branch),
	)
	outputAhead, err := cmdAhead.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("error checking ahead status: %w", err)
	}
	ahead := strings.TrimSpace(string(outputAhead))

	// Get behind count
	cmdBehind := exec.Command(
		"git", "-C", r.FullPath(),
		"rev-list", "--count", fmt.Sprintf("%s..%s/%s", branch, remote, branch),
	)
	outputBehind, err := cmdBehind.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("error checking behind status: %w", err)
	}
	behind := strings.TrimSpace(string(outputBehind))

	// Convert to integers
	aheadCount := 0
	if ahead != "" {
		fmt.Sscanf(ahead, "%d", &aheadCount)
	}

	behindCount := 0
	if behind != "" {
		fmt.Sscanf(behind, "%d", &behindCount)
	}

	return aheadCount, behindCount, nil
}

// Sync synchronizes the repository with the remote
func (r *Repository) Sync() error {
	// Fetch from remote
	if err := r.Fetch(); err != nil {
		return err
	}

	// Check if we have local changes
	hasChanges, err := r.HasLocalChanges()
	if err != nil {
		return err
	}
	if hasChanges {
		return fmt.Errorf("repository has uncommitted changes")
	}

	// Get current branch
	branch, err := r.GetCurrentBranch()
	if err != nil {
		return err
	}

	// Pull changes
	cmd := exec.Command(
		"git", "-C", r.FullPath(),
		"pull", "origin", branch,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error pulling from remote: %w\nOutput: %s", err, output)
	}

	// Update metadata
	r.Metadata.Basic.LastSync = time.Now()
	return r.UpdateStatus()
}

// CreateBranch creates a new branch
func (r *Repository) CreateBranch(name string, fromBranch string) error {
	args := []string{"-C", r.FullPath(), "checkout", "-b", name}
	if fromBranch != "" {
		args = append(args, fromBranch)
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creating branch: %w\nOutput: %s", err, output)
	}

	return r.UpdateStatus()
}

// CheckoutBranch checks out an existing branch
func (r *Repository) CheckoutBranch(name string) error {
	cmd := exec.Command("git", "-C", r.FullPath(), "checkout", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error checking out branch: %w\nOutput: %s", err, output)
	}

	return r.UpdateStatus()
}

// ListBranches lists all branches in the repository
func (r *Repository) ListBranches() ([]string, error) {
	cmd := exec.Command("git", "-C", r.FullPath(), "branch", "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error listing branches: %w", err)
	}

	branches := strings.Split(strings.TrimSpace(string(output)), "\n")
	return branches, nil
}

// Commit creates a new commit with the specified message
func (r *Repository) Commit(message string, all bool) error {
	args := []string{"-C", r.FullPath(), "commit", "-m", message}
	if all {
		args = append(args, "-a")
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creating commit: %w\nOutput: %s", err, output)
	}

	return r.UpdateStatus()
}

// Push pushes changes to the remote
func (r *Repository) Push() error {
	cmd := exec.Command("git", "-C", r.FullPath(), "push")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error pushing to remote: %w\nOutput: %s", err, output)
	}

	r.Metadata.Basic.LastSync = time.Now()
	return r.UpdateStatus()
}

// GenerateID generates a unique repository identifier
func GenerateID(name, url, branch, path string) string {
	// Normalize inputs
	name = strings.ToLower(strings.TrimSpace(name))
	url = strings.TrimSpace(url)
	branch = strings.TrimSpace(branch)
	path = strings.TrimSpace(path)

	// Concatenate and hash
	input := fmt.Sprintf("%s|%s|%s|%s", name, url, branch, path)
	hash := sha256.Sum256([]byte(input))

	// Return first 10 characters of hex representation
	return hex.EncodeToString(hash[:])[:10]
}

// Manager manages multiple repositories
type Manager struct {
	Config  *config.Config
	BaseDir string
}

// NewManager creates a new repository manager
func NewManager(cfg *config.Config, baseDir string) *Manager {
	return &Manager{
		Config:  cfg,
		BaseDir: baseDir,
	}
}

// GetRepository returns a repository by ID, name, or path
func (m *Manager) GetRepository(identifier string) (*Repository, error) {
	// Try to find by ID
	for _, repoCfg := range m.Config.Repositories {
		if repoCfg.ID == identifier {
			repo := New(repoCfg, m.BaseDir)
			if err := repo.LoadMetadata(); err != nil {
				return nil, err
			}
			return repo, nil
		}
	}

	// Try to find by name
	for _, repoCfg := range m.Config.Repositories {
		if repoCfg.Name == identifier {
			repo := New(repoCfg, m.BaseDir)
			if err := repo.LoadMetadata(); err != nil {
				return nil, err
			}
			return repo, nil
		}
	}

	// Try to find by path
	for _, repoCfg := range m.Config.Repositories {
		if repoCfg.Path == identifier || filepath.Join(m.BaseDir, repoCfg.Path) == identifier {
			repo := New(repoCfg, m.BaseDir)
			if err := repo.LoadMetadata(); err != nil {
				return nil, err
			}
			return repo, nil
		}
	}

	return nil, fmt.Errorf("repository not found: %s", identifier)
}

// GetAllRepositories returns all repositories
func (m *Manager) GetAllRepositories() ([]*Repository, error) {
	repos := make([]*Repository, 0, len(m.Config.Repositories))

	for _, repoCfg := range m.Config.Repositories {
		repo := New(repoCfg, m.BaseDir)
		if err := repo.LoadMetadata(); err != nil {
			// If metadata can't be loaded, initialize with defaults
			if err := repo.SaveMetadata(); err != nil {
				return nil, err
			}
		}
		repos = append(repos, repo)
	}

	return repos, nil
}

// AddRepository adds a new repository to the configuration
func (m *Manager) AddRepository(name, url, path, branch string, noClone bool) (*Repository, error) {
	// Check if path is already used
	for _, repoCfg := range m.Config.Repositories {
		if repoCfg.Path == path {
			return nil, fmt.Errorf("repository already exists at path: %s", path)
		}
	}

	// If the repository with the same name already exists, generate a unique name
	uniqueName := name
	nameCounter := 1
	for {
		nameExists := false
		for _, repoCfg := range m.Config.Repositories {
			if repoCfg.Name == uniqueName {
				nameExists = true
				break
			}
		}
		if !nameExists {
			break
		}
		uniqueName = fmt.Sprintf("%s-%d", name, nameCounter)
		nameCounter++
	}

	// Generate repository ID based on the unique name and path
	id := GenerateID(uniqueName, url, branch, path)

	// Create repository configuration
	repoCfg := config.RepositoryConfig{
		ID:     id,
		Name:   uniqueName,
		Path:   path,
		URL:    url,
		Branch: branch,
	}

	// Create repository instance
	repo := New(repoCfg, m.BaseDir)

	// Clone repository if requested
	if !noClone {
		if err := repo.Clone(); err != nil {
			return nil, err
		}
	}

	// Save metadata
	if err := repo.SaveMetadata(); err != nil {
		return nil, err
	}

	// Update configuration
	m.Config.Repositories = append(m.Config.Repositories, repoCfg)
	if err := config.SaveConfig(m.Config, m.BaseDir); err != nil {
		return nil, err
	}

	return repo, nil
}

// RemoveRepository removes a repository from the configuration
func (m *Manager) RemoveRepository(identifier string, delete bool) error {
	// Find repository
	repo, err := m.GetRepository(identifier)
	if err != nil {
		return err
	}

	// Remove from configuration
	for i, repoCfg := range m.Config.Repositories {
		if repoCfg.ID == repo.Config.ID {
			// Remove from slice
			m.Config.Repositories = append(m.Config.Repositories[:i], m.Config.Repositories[i+1:]...)
			break
		}
	}

	// Save configuration
	if err := config.SaveConfig(m.Config, m.BaseDir); err != nil {
		return err
	}

	// Delete metadata file
	if err := os.Remove(repo.MetadataPath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing metadata file: %w", err)
	}

	// Delete repository directory if requested
	if delete {
		if err := os.RemoveAll(repo.FullPath()); err != nil {
			return fmt.Errorf("error removing repository directory: %w", err)
		}
	}

	return nil
}
