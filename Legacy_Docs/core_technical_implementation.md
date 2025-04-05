# MCTL Core Technical Implementation

## 1. Introduction

This document provides a detailed technical overview of the MCTL (Multi-Repository Control System) core implementation. MCTL is designed as a management layer that operates on top of Git repositories, providing unified control, consistent operations, and comprehensive metadata tracking across multiple repositories.

## 2. Core Architecture

### 2.1 Architectural Overview

MCTL follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                      Command Interface                       │
│  (add, remove, list, status, sync, branch, save, load, etc.) │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Repository Manager                        │
│      (repository operations, metadata management)            │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     Git Interface Layer                      │
│           (git command execution, error handling)            │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     Git Repositories                         │
│                (actual git repositories)                     │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Key Components

1. **Command Interface**: Implemented in the `cmd` package, provides the CLI commands that users interact with.
2. **Repository Manager**: Implemented in the `internal/repository` package, manages repository operations and metadata.
3. **Git Interface Layer**: Embedded within the Repository implementation, handles Git command execution.
4. **Configuration Manager**: Implemented in the `internal/config` package, manages MCTL configuration.
5. **Snapshot System**: Implemented in the `internal/snapshot` package, manages repository snapshots.
6. **Logging System**: Implemented in the `internal/logging` package, provides operation and audit logging.
7. **Error Handling**: Implemented in the `internal/errors` package, provides structured error handling.

## 3. Git Integration

### 3.1 Git Command Execution

MCTL interacts with Git repositories by executing Git commands through the Go `os/exec` package. This approach allows MCTL to leverage the full power of Git while providing a higher-level abstraction for multi-repository operations.

Example of Git command execution from `repository.go`:

```go
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
```

### 3.2 Git Operations

MCTL implements the following Git operations:

1. **Clone**: Clone a repository to the local filesystem
2. **Checkout**: Check out a specific branch
3. **Branch**: Create and manage branches
4. **Status**: Check repository status
5. **Commit**: Create commits
6. **Push**: Push changes to remote
7. **Pull**: Pull changes from remote
8. **Fetch**: Fetch updates from remote
9. **Reset**: Reset to a specific commit

Each operation is wrapped with additional functionality for metadata tracking, error handling, and logging.

### 3.3 Repository Status Tracking

MCTL tracks the status of each repository, including:

- Current branch
- Local changes
- Relationship with remote (ahead, behind, diverged)
- Last sync time

This status information is stored in the repository metadata and updated whenever repository operations are performed.

Example of status checking from `repository.go`:

```go
func (r *Repository) UpdateStatus(checkRemote ...bool) error {
    // Check if we should check remote status
    shouldCheckRemote := false
    if len(checkRemote) > 0 {
        shouldCheckRemote = checkRemote[0]
    }
    // Check if repository exists
    if _, err := os.Stat(r.FullPath()); os.IsNotExist(err) {
        r.Metadata.Status.Current = StatusUnknown
        return nil
    }

    // Check if repository is empty (no commits yet)
    cmd := exec.Command("git", "-C", r.FullPath(), "rev-parse", "--verify", "HEAD")
    if err := cmd.Run(); err != nil {
        // Repository is empty (no commits yet)
        r.Metadata.Status.Current = StatusClean
        r.Metadata.Status.Branch = "main" // Default branch for empty repos
        return r.SaveMetadata()
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

    // Set status based on local changes
    if hasChanges {
        r.Metadata.Status.Current = StatusModified
    } else {
        r.Metadata.Status.Current = StatusClean
    }

    // Check remote status if requested
    if shouldCheckRemote && !hasChanges {
        // Fetch from remote
        if err := r.Fetch(); err != nil {
            // If fetch fails, we can still report local status
            return r.SaveMetadata()
        }

        // Check relationship with remote
        ahead, behind, err := r.GetRemoteStatus()
        if err != nil {
            return r.SaveMetadata()
        }

        if ahead > 0 && behind > 0 {
            r.Metadata.Status.Current = StatusDiverged
        } else if ahead > 0 {
            r.Metadata.Status.Current = StatusAhead
        } else if behind > 0 {
            r.Metadata.Status.Current = StatusBehind
        }
    }

    return r.SaveMetadata()
}
```

## 4. Configuration and Metadata

### 4.1 Directory Structure

MCTL creates and maintains the following directory structure:

```
$BASE_DIRECTORY/
├── .mirror/                      # Configuration directory
│   ├── mirror.toml               # Primary configuration
│   ├── metadata/                 # Repository metadata
│   │   └── {id}.json             # Individual repository data
│   ├── logs/                     # Operation logs
│   │   ├── operations.log        # Command execution log
│   │   └── audit.log             # Security audit log
│   ├── snapshots/                # Repository snapshots
│   │   └── {snapshot-id}.json    # Individual snapshot data
│   └── cache/                    # Performance cache
│       └── status/               # Repository status cache
├── {repository-name}/            # Individual repository
└── {repository-name}/            # Individual repository
```

### 4.2 Configuration Format

MCTL uses TOML for its configuration file format. The configuration is stored in `.mirror/mirror.toml`:

```toml
# Global configuration parameters
[global]
default_branch = "main"
parallel_operations = 4
default_remote = "origin"

# Repository definitions
[[repositories]]
id = "a1b2c3d4e5"
name = "secure-comms"
path = "repositories/secure-comms"
url = "git@secure.gov:systems/secure-comms.git"
branch = "main"
```

### 4.3 Metadata Format

Repository metadata is stored as JSON files in the `.mirror/metadata` directory:

```json
{
  "id": "a1b2c3d4e5",
  "name": "secure-comms",
  "basic": {
    "creation_date": "2025-04-05T12:34:56Z",
    "last_sync": "2025-04-05T13:45:00Z"
  },
  "status": {
    "current": "CLEAN",
    "branch": "main"
  },
  "extensions": {}
}
```

### 4.4 Snapshot Format

Snapshots are stored as JSON files in the `.mirror/snapshots` directory:

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

### 4.5 Configuration and Metadata Management

MCTL provides a structured approach to managing configuration and metadata:

1. **Loading Configuration**: The `config.LoadConfig` function loads the configuration from the `.mirror/mirror.toml` file.
2. **Saving Configuration**: The `config.SaveConfig` function saves the configuration to the `.mirror/mirror.toml` file.
3. **Loading Metadata**: The `repository.LoadMetadata` method loads repository metadata from the `.mirror/metadata/{id}.json` file.
4. **Saving Metadata**: The `repository.SaveMetadata` method saves repository metadata to the `.mirror/metadata/{id}.json` file.

## 5. Repository Management

### 5.1 Repository Manager

The Repository Manager is the central component of MCTL that manages repositories:

```go
type Manager struct {
    Config  *config.Config
    BaseDir string
}
```

It provides methods for:

1. **Getting Repositories**: `GetRepository` and `GetAllRepositories`
2. **Adding Repositories**: `AddRepository`
3. **Removing Repositories**: `RemoveRepository`

### 5.2 Repository Representation

Each repository is represented by the `Repository` struct:

```go
type Repository struct {
    Config   config.RepositoryConfig
    Metadata Metadata
    BaseDir  string
}
```

The `Repository` struct provides methods for interacting with the repository, such as:

1. **Clone**: Clone the repository
2. **UpdateStatus**: Update the repository status
3. **Sync**: Synchronize with remote
4. **CreateBranch**: Create a new branch
5. **CheckoutBranch**: Check out a branch
6. **Commit**: Create a commit
7. **Push**: Push changes to remote

### 5.3 Repository Identification

Repositories are identified by:

1. **ID**: A unique identifier generated from the repository name, URL, branch, and path
2. **Name**: A human-readable name for the repository
3. **Path**: The path to the repository relative to the base directory

The ID is generated using a SHA-256 hash of the concatenated name, URL, branch, and path:

```go
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
```

## 6. Command Execution Flow

### 6.1 Command Structure

MCTL uses the Cobra library for command-line interface implementation. Each command follows a similar structure:

1. **Command Definition**: Define the command, its flags, and its arguments
2. **Command Execution**: Implement the command's functionality
3. **Error Handling**: Handle errors and provide meaningful error messages

Example of command definition from `add.go`:

```go
func newAddCmd() *cobra.Command {
    var (
        path    string
        name    string
        branch  string
        noClone bool
        flat    bool
    )

    cmd := &cobra.Command{
        Use:   "add [options] repository-url [path]",
        Short: "Add a Git repository to MCTL management",
        Long:  `Add a Git repository to MCTL management.

This command adds a Git repository to MCTL management. By default, it clones
the repository to the current directory, but you can specify a different
path with the --path flag or as a second argument.

You can also specify a custom name for the repository with the --name flag.
If not provided, the name will be derived from the repository URL.

Examples:
  mctl add git@secure.gov:system/comms.git
  mctl add git@secure.gov:system/comms.git classified
  mctl add git@secure.gov:system/comms.git --path=classified --name=secure-comms
  mctl add git@secure.gov:system/comms.git --branch=release-1.2
  mctl add git@secure.gov:system/comms.git --no-clone`,
        Args: cobra.MinimumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            repoURL := args[0]

            // If path is provided as second argument, use it
            if len(args) > 1 && path == "" {
                path = args[1]
            }

            return runAdd(repoURL, path, name, branch, noClone, flat)
        },
    }

    // Add flags
    cmd.Flags().StringVar(&path, "path", "", "Target location for repository")
    cmd.Flags().StringVar(&name, "name", "", "Custom repository designation")
    cmd.Flags().StringVar(&branch, "branch", "", "Specific branch to clone")
    cmd.Flags().BoolVar(&noClone, "no-clone", false, "Add to configuration without cloning")
    cmd.Flags().BoolVar(&flat, "flat", false, "Clone directly to path without creating subdirectory")

    return cmd
}
```

### 6.2 Command Execution Flow

The typical flow of command execution is:

1. **Parse Command-Line Arguments**: Parse the command-line arguments and flags
2. **Load Configuration**: Load the MCTL configuration
3. **Create Repository Manager**: Create a repository manager with the loaded configuration
4. **Execute Command**: Execute the command's functionality
5. **Log Operation**: Log the operation for audit purposes
6. **Return Result**: Return the result to the user

Example of command execution from `add.go`:

```go
func runAdd(repoURL, path, name, branch string, noClone, flat bool) error {
    // Get current directory
    currentDir, err := os.Getwd()
    if err != nil {
        return errors.Wrap(err, errors.ErrInternalError, "Failed to get current directory")
    }

    // Load configuration
    cfg, err := config.LoadConfig(currentDir)
    if err != nil {
        return errors.Wrap(err, errors.ErrConfigNotFound, "Failed to load configuration")
    }

    // Determine repository name if not provided
    if name == "" {
        name = deriveRepositoryName(repoURL)
    }

    // Determine repository path
    repoPath, err := determineRepositoryPath(currentDir, path, name, flat, repoURL)
    if err != nil {
        return err
    }

    // Create repository manager
    repoManager := repository.NewManager(cfg, currentDir)

    // Add repository
    repo, err := repoManager.AddRepository(name, repoURL, repoPath, branch, noClone)
    if err != nil {
        return errors.Wrap(err, errors.ErrRepositoryExists, "Failed to add repository")
    }

    // Log the operation
    logger := logging.NewLogger(currentDir)
    logger.LogOperation(logging.LogLevelInfo, fmt.Sprintf("Added repository %s (%s)", name, repoURL))
    logger.LogAudit(logging.LogLevelInfo, fmt.Sprintf("Repository added: %s", name))

    fmt.Printf("Added repository %s to MCTL management\n", name)
    if !noClone {
        fmt.Printf("Cloned to %s\n", repo.FullPath())
    }

    return nil
}
```

## 7. Error Handling and Logging

### 7.1 Error Handling

MCTL uses a structured approach to error handling, with error types defined in the `internal/errors` package:

```go
// Error types
const (
    ErrInternalError    = "internal_error"
    ErrConfigNotFound   = "config_not_found"
    ErrRepositoryExists = "repository_exists"
    ErrRepositoryNotFound = "repository_not_found"
    ErrGitCommandFailed = "git_command_failed"
    ErrGitCommitFailed  = "git_commit_failed"
    // ... other error types
)

// Error wraps an error with additional context
type Error struct {
    Err     error
    Type    string
    Message string
}

// Wrap wraps an error with a type and message
func Wrap(err error, errType, message string) error {
    return &Error{
        Err:     err,
        Type:    errType,
        Message: message,
    }
}
```

This allows for more meaningful error messages and better error handling.

### 7.2 Logging

MCTL provides comprehensive logging through the `internal/logging` package:

```go
// LogLevel represents the level of a log entry
type LogLevel string

const (
    LogLevelDebug LogLevel = "DEBUG"
    LogLevelInfo  LogLevel = "INFO"
    LogLevelWarn  LogLevel = "WARN"
    LogLevelError LogLevel = "ERROR"
)

// Logger provides logging functionality
type Logger struct {
    BaseDir string
}

// LogOperation logs an operation
func (l *Logger) LogOperation(level LogLevel, message string) error {
    // ... implementation
}

// LogAudit logs an audit entry
func (l *Logger) LogAudit(level LogLevel, message string) error {
    // ... implementation
}
```

This enables:

1. **Operation Logging**: Logging of command execution for debugging and troubleshooting
2. **Audit Logging**: Logging of security-relevant operations for compliance and auditing

## 8. Snapshot System

### 8.1 Snapshot Manager

The Snapshot Manager is responsible for creating, saving, loading, and applying snapshots:

```go
type Manager struct {
    BaseDir string
}
```

It provides methods for:

1. **Creating Snapshots**: `CreateSnapshot`
2. **Saving Snapshots**: `SaveSnapshot`
3. **Loading Snapshots**: `LoadSnapshot`
4. **Listing Snapshots**: `ListSnapshots`
5. **Applying Snapshots**: `ApplySnapshot`

### 8.2 Snapshot Creation

Snapshots are created by capturing the state of all repositories:

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

### 8.3 Snapshot Application

Snapshots are applied by restoring repositories to their saved state:

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

## 9. Performance Considerations

### 9.1 Parallel Operations

MCTL supports parallel operations for improved performance when working with multiple repositories. The `global.parallel_operations` configuration setting controls the number of concurrent operations.

### 9.2 Caching

MCTL uses caching to improve performance:

1. **Status Cache**: Repository status is cached to avoid repeated Git operations
2. **Metadata Cache**: Repository metadata is cached in memory during command execution

### 9.3 Efficient Git Operations

MCTL is designed to minimize the number of Git operations:

1. **Selective Updates**: Only update repositories that need to be updated
2. **Minimal Git Commands**: Use the minimum number of Git commands required
3. **Efficient Command Construction**: Construct Git commands efficiently to minimize overhead

## 10. Security Considerations

### 10.1 Secure Command Execution

MCTL executes Git commands securely:

1. **Input Validation**: Validate all user input before constructing Git commands
2. **Command Construction**: Construct Git commands using proper argument handling
3. **Output Handling**: Handle command output securely

### 10.2 Secure Configuration and Metadata

MCTL secures configuration and metadata:

1. **File Permissions**: Set appropriate file permissions for configuration and metadata files
2. **Sensitive Information**: Avoid storing sensitive information in configuration and metadata

### 10.3 Audit Logging

MCTL provides audit logging for security and compliance:

1. **Operation Logging**: Log all operations for debugging and troubleshooting
2. **Audit Logging**: Log security-relevant operations for compliance and auditing

## 11. Extensibility

### 11.1 Extension Points

MCTL is designed to be extensible:

1. **Command Extensions**: Add new commands by implementing the Cobra command interface
2. **Repository Extensions**: Extend repository functionality by adding methods to the `Repository` struct
3. **Metadata Extensions**: Extend metadata by adding fields to the `Metadata` struct

### 11.2 Extension Mechanism

The `Metadata.Extensions` field provides a mechanism for extensions to store additional metadata:

```go
type Metadata struct {
    ID     string     `json:"id"`
    Name   string     `json:"name"`
    Basic  BasicInfo  `json:"basic"`
    Status StatusInfo `json:"status"`
    // Reserved for future extensions
    Extensions map[string]interface{} `json:"extensions,omitempty"`
}
```

## 12. Conclusion

MCTL provides a powerful, structured approach to managing multiple Git repositories as a cohesive unit. By implementing a management layer over Git, it enables consistent operations, comprehensive metadata tracking, and secure audit capabilities.

The core technical implementation is designed to be:

1. **Robust**: Handle errors and edge cases gracefully
2. **Efficient**: Minimize overhead and maximize performance
3. **Secure**: Execute commands securely and provide audit logging
4. **Extensible**: Allow for future extensions and customizations

This architecture makes MCTL an ideal tool for complex, security-sensitive development environments where multiple repositories need to be managed together.
