package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/repository"
)

const (
	// DefaultSnapshotsDir is the default snapshots directory name
	DefaultSnapshotsDir = "snapshots"
)

// RepositoryState represents the state of a repository at snapshot time
type RepositoryState struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Branch     string `json:"branch"`
	CommitHash string `json:"commit_hash"`
	Status     string `json:"status"`
}

// Snapshot represents a point-in-time state of all repositories
type Snapshot struct {
	ID           string            `json:"id"`
	CreatedAt    time.Time         `json:"created_at"`
	Description  string            `json:"description"`
	Repositories []RepositoryState `json:"repositories"`
}

// ApplyOptions represents options for applying a snapshot
type ApplyOptions struct {
	DryRun       bool
	Force        bool
	Repositories []string
}

// Manager manages snapshots
type Manager struct {
	BaseDir string
}

// NewManager creates a new snapshot manager
func NewManager(baseDir string) *Manager {
	return &Manager{
		BaseDir: baseDir,
	}
}

// GetSnapshotsDirPath returns the path to the snapshots directory
func GetSnapshotsDirPath(baseDir string) string {
	return filepath.Join(baseDir, config.DefaultConfigDir, DefaultSnapshotsDir)
}

// GetSnapshotPath returns the path to a snapshot file
func GetSnapshotPath(baseDir, id string) string {
	return filepath.Join(GetSnapshotsDirPath(baseDir), fmt.Sprintf("%s.json", id))
}

// CreateSnapshot creates a new snapshot of the current state
func (m *Manager) CreateSnapshot(repoManager *repository.Manager, description string) (*Snapshot, error) {
	// Get all repositories
	repos, err := repoManager.GetAllRepositories()
	if err != nil {
		return nil, fmt.Errorf("error getting repositories: %w", err)
	}

	// Create repository states
	repoStates := make([]RepositoryState, 0, len(repos))
	for _, repo := range repos {
		// Update repository status
		if err := repo.UpdateStatus(); err != nil {
			return nil, fmt.Errorf("error updating repository status: %w", err)
		}

		// Get commit hash
		commitHash, err := repo.GetCommitHash()
		if err != nil {
			return nil, fmt.Errorf("error getting commit hash for %s: %w", repo.Config.Name, err)
		}

		// Create repository state
		repoState := RepositoryState{
			ID:         repo.Config.ID,
			Name:       repo.Config.Name,
			Path:       repo.Config.Path,
			Branch:     repo.Metadata.Status.Branch,
			CommitHash: commitHash,
			Status:     string(repo.Metadata.Status.Current),
		}

		repoStates = append(repoStates, repoState)
	}

	// Create snapshot
	now := time.Now()
	id := generateSnapshotID(now, repoStates)
	snapshot := &Snapshot{
		ID:           id,
		CreatedAt:    now,
		Description:  description,
		Repositories: repoStates,
	}

	return snapshot, nil
}

// SaveSnapshot saves a snapshot to disk
func (m *Manager) SaveSnapshot(snapshot *Snapshot) error {
	// Ensure snapshots directory exists
	snapshotsDir := GetSnapshotsDirPath(m.BaseDir)
	if err := os.MkdirAll(snapshotsDir, 0700); err != nil {
		return fmt.Errorf("error creating snapshots directory: %w", err)
	}

	// Marshal snapshot to JSON
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling snapshot: %w", err)
	}

	// Write snapshot file
	snapshotPath := GetSnapshotPath(m.BaseDir, snapshot.ID)
	if err := os.WriteFile(snapshotPath, data, 0600); err != nil {
		return fmt.Errorf("error writing snapshot file: %w", err)
	}

	return nil
}

// LoadSnapshot loads a snapshot from disk
func (m *Manager) LoadSnapshot(id string) (*Snapshot, error) {
	snapshotPath := GetSnapshotPath(m.BaseDir, id)
	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("snapshot not found: %s", id)
		}
		return nil, fmt.Errorf("error reading snapshot file: %w", err)
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("error unmarshaling snapshot: %w", err)
	}

	return &snapshot, nil
}

// ListSnapshots lists all available snapshots
func (m *Manager) ListSnapshots() ([]*Snapshot, error) {
	snapshotsDir := GetSnapshotsDirPath(m.BaseDir)

	// Check if snapshots directory exists
	if _, err := os.Stat(snapshotsDir); os.IsNotExist(err) {
		return []*Snapshot{}, nil
	}

	// Read snapshots directory
	files, err := os.ReadDir(snapshotsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading snapshots directory: %w", err)
	}

	// Load each snapshot
	snapshots := make([]*Snapshot, 0, len(files))
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		id := file.Name()[:len(file.Name())-5] // Remove .json extension
		snapshot, err := m.LoadSnapshot(id)
		if err != nil {
			// Skip invalid snapshots
			continue
		}

		snapshots = append(snapshots, snapshot)
	}

	// Sort snapshots by creation time (newest first)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].CreatedAt.After(snapshots[j].CreatedAt)
	})

	return snapshots, nil
}

// GetSnapshot gets a specific snapshot by ID
func (m *Manager) GetSnapshot(id string) (*Snapshot, error) {
	return m.LoadSnapshot(id)
}

// DeleteSnapshot deletes a snapshot
func (m *Manager) DeleteSnapshot(id string) error {
	snapshotPath := GetSnapshotPath(m.BaseDir, id)
	if err := os.Remove(snapshotPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("snapshot not found: %s", id)
		}
		return fmt.Errorf("error deleting snapshot: %w", err)
	}

	return nil
}

// ApplySnapshot applies a snapshot to the repositories
func (m *Manager) ApplySnapshot(snapshot *Snapshot, repoManager *repository.Manager, options ApplyOptions) error {
	// Get repositories to apply
	var repoStatesToApply []RepositoryState
	if len(options.Repositories) > 0 {
		// Filter repositories by name
		repoMap := make(map[string]bool)
		for _, name := range options.Repositories {
			repoMap[name] = true
		}

		for _, repoState := range snapshot.Repositories {
			if repoMap[repoState.Name] {
				repoStatesToApply = append(repoStatesToApply, repoState)
			}
		}
	} else {
		// Apply all repositories
		repoStatesToApply = snapshot.Repositories
	}

	// Check for uncommitted changes if not force
	if !options.Force {
		for _, repoState := range repoStatesToApply {
			repo, err := repoManager.GetRepository(repoState.Name)
			if err != nil {
				return fmt.Errorf("error getting repository %s: %w", repoState.Name, err)
			}

			hasChanges, err := repo.HasLocalChanges()
			if err != nil {
				return fmt.Errorf("error checking for changes in %s: %w", repoState.Name, err)
			}

			if hasChanges {
				return fmt.Errorf("repository %s has uncommitted changes, use --force to override", repoState.Name)
			}
		}
	}

	// Apply snapshot
	for _, repoState := range repoStatesToApply {
		repo, err := repoManager.GetRepository(repoState.Name)
		if err != nil {
			return fmt.Errorf("error getting repository %s: %w", repoState.Name, err)
		}

		// Print what would be done in dry run mode
		if options.DryRun {
			fmt.Printf("Would checkout branch %s and reset to commit %s in repository %s\n",
				repoState.Branch, repoState.CommitHash, repoState.Name)
			continue
		}

		// Checkout branch
		if err := repo.CheckoutBranch(repoState.Branch); err != nil {
			return fmt.Errorf("error checking out branch %s in repository %s: %w",
				repoState.Branch, repoState.Name, err)
		}

		// Reset to commit
		if err := repo.ResetToCommit(repoState.CommitHash); err != nil {
			return fmt.Errorf("error resetting to commit %s in repository %s: %w",
				repoState.CommitHash, repoState.Name, err)
		}

		// Update repository status
		if err := repo.UpdateStatus(); err != nil {
			return fmt.Errorf("error updating repository status: %w", err)
		}

		fmt.Printf("✓ %s: Restored to branch %s at commit %s\n",
			repoState.Name, repoState.Branch, repoState.CommitHash[:8])
	}

	return nil
}

// generateSnapshotID generates a unique snapshot ID
func generateSnapshotID(timestamp time.Time, repoStates []RepositoryState) string {
	// Format timestamp as YYYYMMDD-HHMMSS
	timeStr := timestamp.Format("20060102-150405")

	// Create a hash of repository states
	h := sha256.New()
	for _, repoState := range repoStates {
		h.Write([]byte(repoState.ID))
		h.Write([]byte(repoState.Branch))
		h.Write([]byte(repoState.CommitHash))
	}
	hash := hex.EncodeToString(h.Sum(nil))[:8]

	// Combine timestamp and hash
	return fmt.Sprintf("%s-%s", timeStr, hash)
}
