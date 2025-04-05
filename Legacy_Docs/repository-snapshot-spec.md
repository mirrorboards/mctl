# Repository Snapshot System Technical Specification

## 1. Overview

### 1.1 Purpose

The Repository Snapshot System enhances the `mctl` tool to provide point-in-time snapshots of repository states, enabling users to save and restore the state of all managed repositories. This feature allows for easier tracking of changes across repositories and provides a mechanism to return to previous states.

### 1.2 Requirements

1. Enhance `mctl save` to create snapshots with unique IDs
2. Each snapshot must contain:
   - Information about all repositories
   - Current branch for each repository
   - Commit hash for each repository
   - Timestamp and description
3. Create `mctl load` to restore repositories to a saved snapshot state
4. Provide listing and management of snapshots

### 1.3 Success Criteria

1. Users can create snapshots of the current state of all repositories
2. Users can list available snapshots
3. Users can restore repositories to a previous snapshot state
4. The system handles edge cases (conflicts, missing repositories, etc.)

## 2. Architecture

### 2.1 System Components

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│  Save Command   │────▶│ Snapshot System │◀────│  Load Command   │
│                 │     │                 │     │                 │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │                 │
                        │ Repository Mgr  │
                        │                 │
                        └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │                 │
                        │  Git Repos      │
                        │                 │
                        └─────────────────┘
```

### 2.2 Data Structures

#### 2.2.1 Snapshot

```go
// Snapshot represents a point-in-time state of all repositories
type Snapshot struct {
    ID           string             `json:"id"`
    CreatedAt    time.Time          `json:"created_at"`
    Description  string             `json:"description"`
    Repositories []RepositoryState  `json:"repositories"`
}
```

#### 2.2.2 Repository State

```go
// RepositoryState represents the state of a repository at snapshot time
type RepositoryState struct {
    ID           string  `json:"id"`
    Name         string  `json:"name"`
    Path         string  `json:"path"`
    Branch       string  `json:"branch"`
    CommitHash   string  `json:"commit_hash"`
    Status       string  `json:"status"`
}
```

### 2.3 Storage

Snapshots will be stored as JSON files in a dedicated directory:

```
.mirror/snapshots/{snapshot-id}.json
```

The snapshot ID will be a combination of a timestamp and a hash to ensure uniqueness.

## 3. Command Specifications

### 3.1 Save Command

#### 3.1.1 Current Behavior

The current `mctl save` command:
- Commits and pushes changes across repositories
- Takes a commit message as an argument
- Has options for specific repositories, no-push, amend, all, and sign

#### 3.1.2 Enhanced Behavior

The enhanced `mctl save` command will:
- Create a snapshot of all repositories with their current branches and commit hashes
- Generate a unique ID for each snapshot
- Store this snapshot in the snapshots directory
- Continue to support all existing functionality

#### 3.1.3 Command Syntax

```
mctl save [options] "commit-message"
```

Options:
- `--repos=<repos>`: Limit to specific repositories (comma-separated)
- `--no-push`: Create commit without pushing to remote
- `--amend`: Modify previous commit instead of creating new one
- `--all`: Include all changes including untracked files
- `--sign`: Cryptographically sign the commit
- `--no-snapshot`: Skip creating a snapshot (only commit/push)
- `--description="<text>"`: Add a description to the snapshot

### 3.2 Load Command

#### 3.2.1 Behavior

The new `mctl load` command will:
- Take a snapshot ID as an argument
- Find the snapshot with that ID
- For each repository in the snapshot:
  - Check out the correct branch
  - Reset to the saved commit hash

#### 3.2.2 Command Syntax

```
mctl load [options] <snapshot-id>
```

Options:
- `--repos=<repos>`: Limit to specific repositories (comma-separated)
- `--dry-run`: Show what would be done without making changes
- `--force`: Force load even if there are uncommitted changes

### 3.3 List Snapshots Command

#### 3.3.1 Behavior

The new `mctl snapshots` command will:
- List all available snapshots
- Show snapshot ID, creation time, and description
- Optionally show detailed information about repositories in a snapshot

#### 3.3.2 Command Syntax

```
mctl snapshots [options]
```

Options:
- `--detailed`: Show detailed information about repositories in each snapshot
- `--limit=<n>`: Limit to the most recent n snapshots
- `--id=<id>`: Show details for a specific snapshot ID

## 4. Implementation Details

### 4.1 File Structure Changes

1. Create a new file `internal/snapshot/snapshot.go` for the snapshot system
2. Create a new file `cmd/load.go` for the load command
3. Create a new file `cmd/snapshots.go` for the snapshots command
4. Modify `cmd/save.go` to create snapshots
5. Update `cmd/root.go` to register the new commands

### 4.2 Snapshot System Implementation

The snapshot system will provide the following functionality:

1. Create a snapshot of the current state
2. Save a snapshot to disk
3. Load a snapshot from disk
4. List available snapshots
5. Get a specific snapshot by ID
6. Delete a snapshot

### 4.3 Repository Manager Extensions

The repository manager will be extended with methods to:

1. Get the current commit hash for a repository
2. Check out a specific branch
3. Reset to a specific commit hash
4. Verify if a repository has uncommitted changes

### 4.4 Error Handling

The system will handle the following error cases:

1. Repository not found
2. Branch not found
3. Commit hash not found
4. Uncommitted changes when loading a snapshot
5. Snapshot not found
6. Invalid snapshot format

## 5. API Specifications

### 5.1 Snapshot System API

```go
// CreateSnapshot creates a new snapshot of the current state
func CreateSnapshot(description string) (*Snapshot, error)

// SaveSnapshot saves a snapshot to disk
func SaveSnapshot(snapshot *Snapshot) error

// LoadSnapshot loads a snapshot from disk
func LoadSnapshot(id string) (*Snapshot, error)

// ListSnapshots lists all available snapshots
func ListSnapshots() ([]*Snapshot, error)

// GetSnapshot gets a specific snapshot by ID
func GetSnapshot(id string) (*Snapshot, error)

// DeleteSnapshot deletes a snapshot
func DeleteSnapshot(id string) error

// ApplySnapshot applies a snapshot to the repositories
func ApplySnapshot(snapshot *Snapshot, options ApplyOptions) error
```

### 5.2 Repository Manager API Extensions

```go
// GetCommitHash gets the current commit hash for a repository
func (r *Repository) GetCommitHash() (string, error)

// CheckoutBranch checks out a specific branch
func (r *Repository) CheckoutBranch(branch string) error

// ResetToCommit resets to a specific commit hash
func (r *Repository) ResetToCommit(commitHash string) error

// HasUncommittedChanges checks if a repository has uncommitted changes
func (r *Repository) HasUncommittedChanges() (bool, error)
```

## 6. Testing Strategy

### 6.1 Unit Tests

1. Test snapshot creation and serialization
2. Test snapshot loading and deserialization
3. Test repository state capture
4. Test ID generation and uniqueness

### 6.2 Integration Tests

1. Test save command with snapshot creation
2. Test load command with snapshot restoration
3. Test snapshots command for listing
4. Test error handling and edge cases

### 6.3 Manual Testing

1. Test with multiple repositories in different states
2. Test with repositories that have uncommitted changes
3. Test with repositories that have diverged from remote
4. Test with repositories that have conflicts

## 7. Future Considerations

### 7.1 Potential Enhancements

1. Add support for partial snapshots (subset of repositories)
2. Add support for snapshot tags/labels for easier identification
3. Add support for snapshot expiration/cleanup
4. Add support for snapshot comparison
5. Add support for snapshot export/import
6. Add support for snapshot sharing

### 7.2 Performance Considerations

1. For large repositories, capturing the state may be slow
2. For many snapshots, listing may become slow
3. Consider pagination or filtering for snapshot listing
4. Consider compression for snapshot storage

## 8. Implementation Timeline

1. Phase 1: Core snapshot system implementation (2 days)
2. Phase 2: Save command enhancement (1 day)
3. Phase 3: Load command implementation (1 day)
4. Phase 4: Snapshots command implementation (1 day)
5. Phase 5: Testing and bug fixing (2 days)

Total estimated time: 7 days
