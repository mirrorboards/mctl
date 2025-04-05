# MCTL Repository Snapshot System

## 1. Introduction

The Repository Snapshot System is a core feature of MCTL that provides point-in-time snapshots of repository states. This system enables users to save and restore the state of all managed repositories, facilitating easier tracking of changes across repositories and providing a mechanism to return to previous states.

### 1.1 Purpose and Benefits

The snapshot system serves several critical purposes:

- **State Preservation**: Captures the exact state of all repositories at a specific point in time
- **Reproducibility**: Enables returning to a known state for testing, debugging, or deployment
- **Change Tracking**: Provides a history of repository states over time
- **Coordination**: Ensures consistent states across multiple repositories
- **Disaster Recovery**: Allows recovery from undesired changes or failed experiments

### 1.2 Key Concepts

- **Snapshot**: A point-in-time record of the state of all repositories
- **Repository State**: The state of an individual repository within a snapshot, including branch, commit hash, and status
- **Snapshot ID**: A unique identifier for each snapshot, composed of a timestamp and a hash
- **Snapshot Description**: A user-provided or automatically generated description of the snapshot

## 2. Architecture

### 2.1 System Components

The snapshot system consists of several interconnected components:

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

- **Snapshot Manager**: Responsible for creating, saving, loading, and applying snapshots
- **Repository Manager**: Provides access to repository information and operations
- **Git Repositories**: The actual Git repositories being managed
- **Command Interface**: The `save`, `load`, and `snapshots` commands that interact with the snapshot system

### 2.2 Data Flow

1. **Snapshot Creation**:
   - User executes `mctl save` with a commit message
   - Command commits changes to repositories
   - Snapshot system captures the state of all repositories
   - Snapshot is saved to disk

2. **Snapshot Listing**:
   - User executes `mctl snapshots`
   - Snapshot system reads all snapshot files
   - Command displays snapshot information

3. **Snapshot Loading**:
   - User executes `mctl load` with a snapshot ID
   - Snapshot system loads the specified snapshot
   - Repository manager applies the snapshot state to repositories
   - Repositories are restored to the snapshot state

## 3. Data Structures

### 3.1 Snapshot

The `Snapshot` struct represents a point-in-time state of all repositories:

```go
type Snapshot struct {
    ID           string             `json:"id"`
    CreatedAt    time.Time          `json:"created_at"`
    Description  string             `json:"description"`
    Repositories []RepositoryState  `json:"repositories"`
}
```

- **ID**: A unique identifier for the snapshot, generated from timestamp and repository state hash
- **CreatedAt**: The timestamp when the snapshot was created
- **Description**: A user-provided or automatically generated description of the snapshot
- **Repositories**: An array of repository states included in the snapshot

### 3.2 Repository State

The `RepositoryState` struct represents the state of a repository at snapshot time:

```go
type RepositoryState struct {
    ID         string `json:"id"`
    Name       string `json:"name"`
    Path       string `json:"path"`
    Branch     string `json:"branch"`
    CommitHash string `json:"commit_hash"`
    Status     string `json:"status"`
}
```

- **ID**: The unique identifier of the repository
- **Name**: The name of the repository
- **Path**: The path to the repository relative to the base directory
- **Branch**: The current branch at snapshot time
- **CommitHash**: The commit hash at snapshot time
- **Status**: The repository status at snapshot time (CLEAN, MODIFIED, AHEAD, BEHIND, DIVERGED, UNKNOWN)

### 3.3 Apply Options

The `ApplyOptions` struct represents options for applying a snapshot:

```go
type ApplyOptions struct {
    DryRun       bool
    Force        bool
    Repositories []string
}
```

- **DryRun**: If true, show what would be done without making changes
- **Force**: If true, apply snapshot even if there are uncommitted changes
- **Repositories**: If specified, limit snapshot application to these repositories

## 4. Storage Format

### 4.1 File Structure

Snapshots are stored as JSON files in the `.mirror/snapshots` directory within the MCTL base directory:

```
.mirror/snapshots/{snapshot-id}.json
```

This location within the `.mirror` directory is intentional, as it allows the snapshot metadata to be committed to version control and shared across a team. By committing the `.mirror` directory to a shared repository, teams can share snapshot references, enabling consistent state management across development environments.

The snapshot ID is a combination of a timestamp and a hash to ensure uniqueness, following the format:

```
YYYYMMDD-HHMMSS-{hash}
```

For example: `20250405-123456-abcdef12`

### 4.2 JSON Format

Each snapshot file contains a JSON representation of the `Snapshot` struct:

```json
{
  "id": "20250405-123456-abcdef12",
  "created_at": "2025-04-05T12:34:56Z",
  "description": "Implement authentication feature",
  "repositories": [
    {
      "id": "a1b2c3d4e5",
      "name": "secure-comms",
      "path": "repositories/secure-comms",
      "branch": "feature/auth",
      "commit_hash": "7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b",
      "status": "CLEAN"
    },
    {
      "id": "f6g7h8i9j0",
      "name": "authentication",
      "path": "repositories/authentication",
      "branch": "feature/auth",
      "commit_hash": "3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f",
      "status": "CLEAN"
    }
  ]
}
```

### 4.3 ID Generation

Snapshot IDs are generated using a combination of:

1. A timestamp in the format `YYYYMMDD-HHMMSS`
2. A hash of repository states to ensure uniqueness

The hash is created by:
- Concatenating the repository ID, branch, and commit hash for each repository
- Computing a SHA-256 hash of this concatenated string
- Taking the first 8 characters of the hex-encoded hash

```go
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
```

## 5. Command Interface

### 5.1 Save Command

The `mctl save` command commits changes across repositories and creates a snapshot of the current state.

#### 5.1.1 Syntax

```
mctl save [options] "commit-message"
```

#### 5.1.2 Options

- `--repos=<repos>`: Limit to specific repositories (comma-separated)
- `--no-push`: Create commit without pushing to remote
- `--amend`: Modify previous commit instead of creating new one
- `--all`: Include all changes including untracked files
- `--sign`: Cryptographically sign the commit
- `--no-snapshot`: Skip creating a snapshot (only commit/push)
- `--description="<text>"`: Add a description to the snapshot

#### 5.1.3 Examples

```bash
# Commit changes with a message and create a snapshot
mctl save "Fix authentication bug"

# Commit changes to specific repositories
mctl save --repos=secure-comms,authentication "Update dependencies"

# Commit without pushing to remote
mctl save --no-push "Work in progress"

# Commit and amend previous commit
mctl save --amend "Fix typo in previous commit"

# Commit all changes including untracked files
mctl save --all "Add new feature"

# Commit with a signed commit
mctl save --sign "Security patch"

# Commit with a custom snapshot description
mctl save --description="Stable version for testing" "Prepare for testing"

# Commit without creating a snapshot
mctl save --no-snapshot "Minor changes"
```

#### 5.1.4 Behavior

1. Commits changes in the specified repositories (or all repositories with changes if none specified)
2. Optionally pushes changes to the remote (unless `--no-push` is specified)
3. Creates a snapshot of the current state (unless `--no-snapshot` is specified)
4. Displays the snapshot ID and instructions for restoring it

### 5.2 Load Command

The `mctl load` command restores repositories to a saved snapshot state.

#### 5.2.1 Syntax

```
mctl load [options] <snapshot-id>
```

#### 5.2.2 Options

- `--repos=<repos>`: Limit to specific repositories (comma-separated)
- `--dry-run`: Show what would be done without making changes
- `--force`: Force load even if there are uncommitted changes

#### 5.2.3 Examples

```bash
# Load a specific snapshot
mctl load 20250405-123456-abcdef12

# Load a snapshot for specific repositories
mctl load --repos=secure-comms,authentication 20250405-123456-abcdef12

# Show what would be done without making changes
mctl load --dry-run 20250405-123456-abcdef12

# Force load even if there are uncommitted changes
mctl load --force 20250405-123456-abcdef12
```

#### 5.2.4 Behavior

1. Loads the specified snapshot from disk
2. For each repository in the snapshot (or specified repositories if `--repos` is used):
   - Checks for uncommitted changes (unless `--force` is specified)
   - Checks out the correct branch
   - Resets to the saved commit hash
3. Displays the result of the operation

### 5.3 Snapshots Command

The `mctl snapshots` command lists available snapshots and shows details about them.

#### 5.3.1 Syntax

```
mctl snapshots [options]
```

#### 5.3.2 Options

- `--detailed`: Show detailed information about repositories in each snapshot
- `--limit=<n>`: Limit to the most recent n snapshots
- `--id=<id>`: Show details for a specific snapshot ID

#### 5.3.3 Examples

```bash
# List all snapshots
mctl snapshots

# Show detailed information about repositories in each snapshot
mctl snapshots --detailed

# Limit to the 5 most recent snapshots
mctl snapshots --limit=5

# Show details for a specific snapshot
mctl snapshots --id=20250405-123456-abcdef12
```

#### 5.3.4 Behavior

1. Lists all available snapshots, showing:
   - Snapshot ID
   - Creation time (relative if less than 24 hours ago, otherwise date and time)
   - Number of repositories
   - Description
2. If `--detailed` is specified, also shows for each repository:
   - Repository name
   - Branch
   - Commit hash (first 8 characters)
   - Status
3. If `--id` is specified, shows detailed information about the specified snapshot

## 6. Implementation Details

### 6.1 Snapshot Manager

The `Manager` struct in the `snapshot` package is responsible for managing snapshots:

```go
type Manager struct {
    BaseDir string
}
```

- **BaseDir**: The base directory for the MCTL configuration

#### 6.1.1 Creating a Snapshot

The `CreateSnapshot` method creates a new snapshot of the current state:

```go
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
```

#### 6.1.2 Saving a Snapshot

The `SaveSnapshot` method saves a snapshot to disk:

```go
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
```

#### 6.1.3 Loading a Snapshot

The `LoadSnapshot` method loads a snapshot from disk:

```go
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
```

#### 6.1.4 Listing Snapshots

The `ListSnapshots` method lists all available snapshots:

```go
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
```

#### 6.1.5 Applying a Snapshot

The `ApplySnapshot` method applies a snapshot to the repositories:

```go
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
```

### 6.2 Repository Extensions

The `Repository` struct in the `repository` package is extended with methods to support snapshots:

#### 6.2.1 Getting Commit Hash

The `GetCommitHash` method gets the current commit hash for a repository:

```go
func (r *Repository) GetCommitHash() (string, error) {
    cmd := exec.Command("git", "-C", r.FullPath(), "rev-parse", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("error getting commit hash: %w", err)
    }
    return strings.TrimSpace(string(output)), nil
}
```

#### 6.2.2 Checking Out Branch

The `CheckoutBranch` method checks out a specific branch:

```go
func (r *Repository) CheckoutBranch(name string) error {
    cmd := exec.Command("git", "-C", r.FullPath(), "checkout", name)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("error checking out branch: %w\nOutput: %s", err, output)
    }

    return r.UpdateStatus()
}
```

#### 6.2.3 Resetting to Commit

The `ResetToCommit` method resets to a specific commit hash:

```go
func (r *Repository) ResetToCommit(commitHash string) error {
    cmd := exec.Command("git", "-C", r.FullPath(), "reset", "--hard", commitHash)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("error resetting to commit: %w\nOutput: %s", err, output)
    }
    return nil
}
```

## 7. Error Handling

The snapshot system handles various error cases to ensure robustness:

### 7.1 Snapshot Creation Errors

- **Repository Not Found**: If a repository specified in `--repos` is not found
- **Status Update Failure**: If updating the status of a repository fails
- **Commit Hash Retrieval Failure**: If getting the commit hash for a repository fails
- **Snapshot Directory Creation Failure**: If creating the snapshots directory fails
- **JSON Marshaling Failure**: If marshaling the snapshot to JSON fails
- **File Write Failure**: If writing the snapshot file fails

### 7.2 Snapshot Loading Errors

- **Snapshot Not Found**: If the specified snapshot ID does not exist
- **File Read Failure**: If reading the snapshot file fails
- **JSON Unmarshaling Failure**: If unmarshaling the snapshot from JSON fails

### 7.3 Snapshot Application Errors

- **Repository Not Found**: If a repository in the snapshot is not found
- **Uncommitted Changes**: If a repository has uncommitted changes (unless `--force` is specified)
- **Branch Checkout Failure**: If checking out a branch fails
- **Commit Reset Failure**: If resetting to a commit hash fails
- **Status Update Failure**: If updating the status of a repository after applying a snapshot fails

### 7.4 Error Reporting

Errors are reported with context to help users understand and resolve issues:

```go
return errors.Wrap(err, errors.ErrInternalError, "Failed to create snapshot")
```

## 8. Best Practices and Workflows

### 8.1 When to Create Snapshots

Snapshots are automatically created when using `mctl save` unless disabled with `--no-snapshot`. Consider creating snapshots:

- Before making significant changes to repositories
- After completing a feature or bug fix
- Before merging branches
- Before deploying to production
- At regular intervals during development

### 8.2 Snapshot Naming and Description

While snapshot IDs are automatically generated, providing meaningful descriptions helps identify snapshots later:

```bash
mctl save --description="Stable version before refactoring" "Refactor authentication module"
```

### 8.3 Snapshot Management

Regularly review and manage snapshots to avoid accumulating unnecessary ones:

```bash
# List all snapshots
mctl snapshots

# Show details for a specific snapshot
mctl snapshots --id=20250405-123456-abcdef12

# Delete a snapshot (not yet implemented)
# mctl snapshots --delete=20250405-123456-abcdef12
```

### 8.4 Partial Snapshot Loading

When working with multiple repositories, you can load a snapshot for specific repositories:

```bash
mctl load --repos=secure-comms,authentication 20250405-123456-abcdef12
```

### 8.5 Dry Run

Before applying a snapshot, you can see what would be done without making changes:

```bash
mctl load --dry-run 20250405-123456-abcdef12
```

### 8.6 Common Workflows

#### 8.6.1 Feature Development

1. Create a feature branch across repositories:
   ```bash
   mctl branch create feature/auth
   ```

2. Make changes and commit regularly with snapshots:
   ```bash
   mctl save "Implement login form"
   ```

3. If you need to revert to a previous state:
   ```bash
   mctl load 20250405-123456-abcdef12
   ```

4. When the feature is complete, create a final snapshot:
   ```bash
   mctl save --description="Completed authentication feature" "Finalize authentication"
   ```

#### 8.6.2 Experimentation

1. Create a snapshot before experimenting:
   ```bash
   mctl save --description="Before experimental changes" "Prepare for experiment"
   ```

2. Make experimental changes and commit:
   ```bash
   mctl save "Experimental changes"
   ```

3. If the experiment fails, revert to the previous state:
   ```bash
   mctl load 20250405-123456-abcdef12
   ```

4. If the experiment succeeds, continue development.

#### 8.6.3 Deployment

1. Create a snapshot before deployment:
   ```bash
   mctl save --description="Pre-deployment state" "Prepare for deployment"
   ```

2. Deploy the application.

3. If deployment fails, revert to the pre-deployment state:
   ```bash
   mctl load 20250405-123456-abcdef12
   ```

## 9. Technical Specifications

### 9.1 Performance Considerations

- **Snapshot Creation**: O(n) where n is the number of repositories
- **Snapshot Loading**: O(1) for loading a snapshot, O(n) for applying it to repositories
- **Snapshot Listing**: O(m) where m is the number of snapshots

### 9.2 Storage Requirements

- Each snapshot file is typically a few kilobytes, depending on the number of repositories
- Storage grows linearly with the number of snapshots
- All snapshot files are stored in the `.mirror/snapshots` directory, making them easy to commit to version control

### 9.3 Sharing Snapshots

- The `.mirror` directory containing snapshots can be committed to a shared repository
- Team members can pull the shared repository to access the same snapshots
- This enables consistent state management across development environments
- Snapshots can be referenced in documentation, issue trackers, or pull requests

### 9.4 Compatibility

- Snapshots are compatible across different versions of MCTL as long as the snapshot format remains unchanged
- Snapshots are tied to the specific repositories they were created from and may not be applicable to different repositories

### 9.4 Security

- Snapshot files are created with 0600 permissions (read/write for owner only)
- The snapshots directory is created with 0700 permissions (read/write/execute for owner only)
- Snapshots contain repository state information but not the actual code

### 9.5 Limitations

- Snapshots do not capture the working directory state, only the committed state
- Snapshots do not capture stashed changes
- Snapshots do not capture Git configuration
- Snapshots are specific to the repositories they were created from and cannot be applied to different repositories

## 10. Future Enhancements

The snapshot system could be enhanced in several ways:

### 10.1 Snapshot Management

- **Deletion**: Add support for deleting snapshots
- **Expiration**: Add support for automatic snapshot expiration
- **Tagging**: Add support for tagging snapshots for easier identification
- **Comparison**: Add support for comparing snapshots to see what changed

### 10.2 Snapshot Content

- **Working Directory**: Capture uncommitted changes in the working directory
- **Stashed Changes**: Capture stashed changes
- **Git Configuration**: Capture Git configuration

### 10.3 Snapshot Application

- **Partial Application**: Apply only specific aspects of a snapshot (e.g., only branch or only commit)
- **Merge Strategy**: Add options for handling conflicts when applying snapshots
- **Post-Apply Hooks**: Add support for running commands after applying a snapshot

### 10.4 Integration

- **CI/CD Integration**: Integrate with CI/CD systems to create and apply snapshots automatically
- **Version Control Integration**: Commit the `.mirror` directory to version control to share snapshots
- **External Storage**: Optionally store snapshots in additional external storage (e.g., S3)
- **Team Collaboration**: Use shared snapshots for consistent development environments

## 11. Conclusion

The Repository Snapshot System is a powerful feature of MCTL that enables point-in-time snapshots of repository states. By capturing the exact state of all repositories, it provides reproducibility, change tracking, and disaster recovery capabilities. The system is designed to be robust, efficient, and easy to use, making it an essential tool for managing multiple repositories in complex development environments.
