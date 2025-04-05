# MCTL API Specification

## 1. Introduction

### 1.1 Overview

MCTL (Multi-Repository Control System) is a command-line tool designed to provide secure, unified management of code repositories in high-security environments. It implements a structured management layer over Git repositories, enabling consistent operations across multiple codebases while maintaining comprehensive metadata and audit capabilities.

This API specification document provides a comprehensive reference for the MCTL command-line interface, including all commands, flags, parameters, and the underlying logic. It is designed to be used as a prompt for code generation.

### 1.2 Key Features

- **Unified Repository Management**: Centralized control of multiple Git repositories through a single interface
- **Repository Snapshot System**: Point-in-time snapshots of repository states for reproducibility and recovery
- **Security and Compliance**: Comprehensive audit logging and metadata tracking
- **Workflow Optimization**: Parallel operations and simplified branching across repositories

### 1.3 Architecture Overview

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

## 2. Command Interface

### 2.1 Core Repository Management Commands

#### 2.1.1 `mctl init`

**Purpose**: Initialize a new MCTL configuration environment.

**Syntax**:
```
mctl init [options]
```

**Options**:
- `--default-branch=<branch>`: Set the default branch for new repositories (default: "main")
- `--parallel-operations=<n>`: Set the number of parallel operations (default: 4)
- `--default-remote=<remote>`: Set the default remote name (default: "origin")

**Behavior**:
1. Create the `.mirror` directory in the current directory
2. Create the configuration file `.mirror/mirror.toml` with default settings
3. Create subdirectories for metadata, logs, snapshots, and cache
4. Set appropriate file permissions for security

**Error Handling**:
- If the `.mirror` directory already exists, return an error
- If the current directory is not writable, return an error

#### 2.1.2 `mctl add`

**Purpose**: Add a Git repository to MCTL management.

**Syntax**:
```
mctl add [options] repository-url [path]
```

**Arguments**:
- `repository-url`: (Required) The Git URL of the repository to add
- `path`: (Optional) The target location for the repository (can also be specified with `--path`)

**Options**:
- `--path=<path>`: Target location for repository
- `--name=<name>`: Custom repository designation (if not provided, derived from URL)
- `--branch=<branch>`: Specific branch to clone
- `--no-clone`: Add to configuration without cloning
- `--flat`: Clone directly to path without creating subdirectory

**Behavior**:
1. Determine repository name (from `--name` or derived from URL)
2. Determine repository path (from positional argument, `--path`, or current directory)
3. Generate a unique repository ID based on name, URL, branch, and path
4. Clone the repository (unless `--no-clone` is specified)
5. Create and save repository metadata
6. Update MCTL configuration to include the new repository
7. Log the operation for audit purposes

**Error Handling**:
- If a repository with the same path already exists, return an error
- If a repository with the same name already exists, generate a unique name
- If the repository cannot be cloned, return an error
- If the configuration cannot be updated, return an error

#### 2.1.3 `mctl remove`

**Purpose**: Remove a repository from MCTL management.

**Syntax**:
```
mctl remove [options] <repository-name>
```

**Arguments**:
- `repository-name`: (Required) The name of the repository to remove

**Options**:
- `--delete`: Delete the repository directory from the filesystem
- `--force`: Force removal even if there are uncommitted changes

**Behavior**:
1. Find the repository in the configuration
2. If `--delete` is specified, delete the repository directory
3. Remove the repository from the configuration
4. Delete repository metadata
5. Log the operation for audit purposes

**Error Handling**:
- If the repository is not found, return an error
- If the repository has uncommitted changes and `--force` is not specified, return an error
- If the repository directory cannot be deleted, return an error

#### 2.1.4 `mctl list`

**Purpose**: List repositories under MCTL management.

**Syntax**:
```
mctl list [options]
```

**Options**:
- `--detailed`: Show detailed information about each repository
- `--format=<format>`: Output format (table, json, yaml)
- `--filter=<filter>`: Filter repositories by name, path, or status

**Behavior**:
1. Load the MCTL configuration
2. Get all repositories from the configuration
3. For each repository, display:
   - Repository name
   - Repository path
   - Current branch
   - Status (if `--detailed` is specified)
   - Last sync time (if `--detailed` is specified)
   - Remote URL (if `--detailed` is specified)

**Error Handling**:
- If the configuration cannot be loaded, return an error
- If a repository's status cannot be determined, display "UNKNOWN"

#### 2.1.5 `mctl status`

**Purpose**: Report status of managed repositories.

**Syntax**:
```
mctl status [options] [repository-names...]
```

**Arguments**:
- `repository-names`: (Optional) Names of repositories to check status for

**Options**:
- `--check-remote`: Check status relative to remote repositories
- `--format=<format>`: Output format (table, json, yaml)
- `--filter=<filter>`: Filter repositories by status

**Behavior**:
1. Load the MCTL configuration
2. Get specified repositories (or all if none specified)
3. For each repository, check:
   - Current branch
   - Local changes
   - Relationship with remote (if `--check-remote` is specified)
4. Display status information for each repository

**Error Handling**:
- If a repository is not found, return an error
- If a repository's status cannot be determined, display "UNKNOWN"
- If remote status check fails, display local status only

#### 2.1.6 `mctl sync`

**Purpose**: Update repositories to match remote state.

**Syntax**:
```
mctl sync [options] [repository-names...]
```

**Arguments**:
- `repository-names`: (Optional) Names of repositories to synchronize

**Options**:
- `--fetch-only`: Only fetch updates without merging
- `--rebase`: Use rebase instead of merge
- `--force`: Force sync even if there are uncommitted changes
- `--dry-run`: Show what would be done without making changes

**Behavior**:
1. Load the MCTL configuration
2. Get specified repositories (or all if none specified)
3. For each repository:
   - Check for uncommitted changes (unless `--force` is specified)
   - Fetch updates from remote
   - If `--fetch-only` is not specified:
     - If `--rebase` is specified, rebase on remote
     - Otherwise, merge remote changes
   - Update repository status
   - Update last sync time in metadata

**Error Handling**:
- If a repository is not found, return an error
- If a repository has uncommitted changes and `--force` is not specified, return an error
- If fetch, merge, or rebase fails, return an error

### 2.2 Git Operations Commands

#### 2.2.1 `mctl branch`

**Purpose**: Manage Git branches across repositories.

**Syntax**:
```
mctl branch <subcommand> [options] [repository-names...]
```

**Subcommands**:
- `create`: Create a new branch
- `checkout`: Check out an existing branch
- `list`: List branches
- `delete`: Delete a branch

**Arguments**:
- `repository-names`: (Optional) Names of repositories to operate on

**Options for `create`**:
- `--from=<branch>`: Create branch from specified branch
- `--no-checkout`: Create branch without checking it out

**Options for `checkout`**:
- `--force`: Force checkout even if there are uncommitted changes
- `--create`: Create branch if it doesn't exist

**Options for `list`**:
- `--remote`: List remote branches
- `--all`: List both local and remote branches

**Options for `delete`**:
- `--remote`: Delete remote branch
- `--force`: Force deletion even if branch is not fully merged

**Behavior**:
1. Load the MCTL configuration
2. Get specified repositories (or all if none specified)
3. Execute the specified subcommand for each repository
4. Update repository status
5. Log the operation for audit purposes

**Error Handling**:
- If a repository is not found, return an error
- If a branch operation fails, return an error
- If a repository has uncommitted changes and `--force` is not specified, return an error

#### 2.2.2 `mctl save`

**Purpose**: Commit and push changes across repositories.

**Syntax**:
```
mctl save [options] "commit-message"
```

**Arguments**:
- `commit-message`: (Required) The commit message to use

**Options**:
- `--repos=<repos>`: Limit to specific repositories (comma-separated)
- `--no-push`: Create commit without pushing to remote
- `--amend`: Modify previous commit instead of creating new one
- `--all`: Include all changes including untracked files
- `--sign`: Cryptographically sign the commit
- `--no-snapshot`: Skip creating a snapshot (only commit/push)
- `--description="<text>"`: Add a description to the snapshot

**Behavior**:
1. Load the MCTL configuration
2. Get specified repositories (or all with changes if none specified)
3. For each repository with changes:
   - Create a commit with the specified message
   - If `--no-push` is not specified, push to remote
4. If `--no-snapshot` is not specified:
   - Create a snapshot of the current state
   - Save the snapshot to disk
5. Log the operation for audit purposes

**Error Handling**:
- If a repository is not found, return an error
- If a commit or push fails, return an error
- If snapshot creation fails, return an error

#### 2.2.3 `mctl clear`

**Purpose**: Remove repository directories while preserving configuration.

**Syntax**:
```
mctl clear [options] [repository-names...]
```

**Arguments**:
- `repository-names`: (Optional) Names of repositories to clear

**Options**:
- `--force`: Force clearing even if there are uncommitted changes
- `--all`: Clear all repositories

**Behavior**:
1. Load the MCTL configuration
2. Get specified repositories (or all if none specified)
3. For each repository:
   - Check for uncommitted changes (unless `--force` is specified)
   - Delete the repository directory
   - Update repository status
4. Log the operation for audit purposes

**Error Handling**:
- If a repository is not found, return an error
- If a repository has uncommitted changes and `--force` is not specified, return an error
- If a repository directory cannot be deleted, return an error

### 2.3 Snapshot System Commands

#### 2.3.1 `mctl load`

**Purpose**: Restore repositories to a saved snapshot state.

**Syntax**:
```
mctl load [options] <snapshot-id>
```

**Arguments**:
- `snapshot-id`: (Required) The ID of the snapshot to load

**Options**:
- `--repos=<repos>`: Limit to specific repositories (comma-separated)
- `--dry-run`: Show what would be done without making changes
- `--force`: Force load even if there are uncommitted changes

**Behavior**:
1. Load the MCTL configuration
2. Load the specified snapshot
3. Get repositories to apply the snapshot to
4. For each repository:
   - Check for uncommitted changes (unless `--force` is specified)
   - If `--dry-run` is specified, show what would be done
   - Otherwise:
     - Check out the branch from the snapshot
     - Reset to the commit hash from the snapshot
     - Update repository status
5. Log the operation for audit purposes

**Error Handling**:
- If the snapshot is not found, return an error
- If a repository is not found, return an error
- If a repository has uncommitted changes and `--force` is not specified, return an error
- If branch checkout or commit reset fails, return an error

#### 2.3.2 `mctl snapshots`

**Purpose**: List and inspect available snapshots.

**Syntax**:
```
mctl snapshots [options]
```

**Options**:
- `--detailed`: Show detailed information about repositories in each snapshot
- `--limit=<n>`: Limit to the most recent n snapshots
- `--id=<id>`: Show details for a specific snapshot ID
- `--format=<format>`: Output format (table, json, yaml)

**Behavior**:
1. Load the MCTL configuration
2. List all available snapshots
3. For each snapshot, display:
   - Snapshot ID
   - Creation time
   - Description
   - Number of repositories
4. If `--detailed` is specified, also show for each repository:
   - Repository name
   - Branch
   - Commit hash
   - Status
5. If `--id` is specified, show detailed information for the specified snapshot

**Error Handling**:
- If the snapshots directory does not exist, return an empty list
- If a snapshot file is invalid, skip it and continue
- If a specific snapshot ID is requested but not found, return an error

### 2.4 Configuration and Utilities Commands

#### 2.4.1 `mctl config`

**Purpose**: Manage MCTL configuration.

**Syntax**:
```
mctl config <subcommand> [options] [key] [value]
```

**Subcommands**:
- `get`: Get a configuration value
- `set`: Set a configuration value
- `list`: List all configuration values
- `reset`: Reset configuration to defaults

**Arguments**:
- `key`: (Required for `get` and `set`) The configuration key
- `value`: (Required for `set`) The configuration value

**Options**:
- `--global`: Operate on global configuration
- `--format=<format>`: Output format for `list` (table, json, yaml)

**Behavior**:
1. Load the MCTL configuration
2. Execute the specified subcommand
3. If `set` or `reset`, save the updated configuration
4. Log the operation for audit purposes

**Error Handling**:
- If the configuration cannot be loaded, return an error
- If a key is not found for `get`, return an error
- If the configuration cannot be saved, return an error

#### 2.4.2 `mctl logs`

**Purpose**: Display MCTL operation logs.

**Syntax**:
```
mctl logs [options]
```

**Options**:
- `--audit`: Show audit logs instead of operation logs
- `--limit=<n>`: Limit to the most recent n log entries
- `--level=<level>`: Filter by log level (debug, info, warn, error)
- `--format=<format>`: Output format (table, json, yaml)
- `--since=<time>`: Show logs since specified time

**Behavior**:
1. Load the MCTL configuration
2. Determine which log file to read (operation or audit)
3. Read and parse the log file
4. Apply filters (limit, level, since)
5. Display the log entries

**Error Handling**:
- If the log file does not exist, return an empty list
- If the log file cannot be read, return an error
- If the log file is invalid, return an error

#### 2.4.3 `mctl version`

**Purpose**: Display MCTL version information.

**Syntax**:
```
mctl version [options]
```

**Options**:
- `--short`: Display version number only
- `--json`: Output in JSON format

**Behavior**:
1. Display version information:
   - Version number
   - Build date
   - Git commit hash
   - Go version
   - Platform

**Error Handling**:
- No specific error handling

#### 2.4.4 `mctl man`

**Purpose**: Display manual pages for MCTL commands.

**Syntax**:
```
mctl man [command]
```

**Arguments**:
- `command`: (Optional) The command to show manual page for

**Behavior**:
1. If no command is specified, display a list of available commands
2. If a command is specified, display the manual page for that command

**Error Handling**:
- If the specified command does not exist, return an error

## 3. Data Structures

### 3.1 Configuration Format

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

### 3.2 Metadata Format

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

### 3.3 Snapshot Format

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

### 3.4 Log Format

Logs are stored as line-delimited JSON in the `.mirror/logs` directory:

```json
{"timestamp":"2025-04-05T12:34:56Z","level":"INFO","message":"Added repository secure-comms (git@secure.gov:systems/secure-comms.git)"}
{"timestamp":"2025-04-05T12:35:00Z","level":"INFO","message":"Created branch feature/auth in repository secure-comms"}
```

## 4. Directory Structure

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

## 5. Error Handling

### 5.1 Error Types

MCTL defines the following error types:

- `internal_error`: An internal error occurred
- `config_not_found`: Configuration not found
- `repository_exists`: Repository already exists
- `repository_not_found`: Repository not found
- `git_command_failed`: Git command failed
- `git_commit_failed`: Git commit failed
- `snapshot_not_found`: Snapshot not found
- `uncommitted_changes`: Repository has uncommitted changes
- `branch_not_found`: Branch not found
- `invalid_argument`: Invalid argument provided

### 5.2 Error Format

Errors are reported with context to help users understand and resolve issues:

```
Error: Failed to add repository
Type: repository_exists
Details: Repository already exists at path: repositories/secure-comms
```

## 6. Best Practices and Workflows

### 6.1 Repository Management

#### 6.1.1 Repository Naming

- Use descriptive names that reflect the repository's purpose
- Keep names concise but meaningful
- Use consistent naming conventions across repositories
- Consider using prefixes for related repositories (e.g., "api-", "ui-", "lib-")

#### 6.1.2 Repository Organization

- Organize repositories in a logical directory structure
- Group related repositories in subdirectories
- Use the `--path` option to specify the target location
- Consider using the `--flat` option for repositories that should be at the root of a directory

### 6.2 Snapshot Management

#### 6.2.1 When to Create Snapshots

Snapshots are automatically created when using `mctl save` unless disabled with `--no-snapshot`. Consider creating snapshots:

- Before making significant changes to repositories
- After completing a feature or bug fix
- Before merging branches
- Before deploying to production
- At regular intervals during development

#### 6.2.2 Snapshot Naming and Description

While snapshot IDs are automatically generated, providing meaningful descriptions helps identify snapshots later:

```bash
mctl save --description="Stable version before refactoring" "Refactor authentication module"
```

#### 6.2.3 Snapshot Management

Regularly review and manage snapshots to avoid accumulating unnecessary ones:

```bash
# List all snapshots
mctl snapshots

# Show details for a specific snapshot
mctl snapshots --id=20250405-123456-abcdef12
```

### 6.3 Common Workflows

#### 6.3.1 Feature Development

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

#### 6.3.2 Experimentation

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

#### 6.3.3 Deployment

1. Create a snapshot before deployment:
   ```bash
   mctl save --description="Pre-deployment state" "Prepare for deployment"
   ```

2. Deploy the application.

3. If deployment fails, revert to the pre-deployment state:
   ```bash
   mctl load 20250405-123456-abcdef12
   ```

## 7. Security Considerations

### 7.1 Secure Command Execution

MCTL executes Git commands securely:

1. **Input Validation**: Validate all user input before constructing Git commands
2. **Command Construction**: Construct Git commands using proper argument handling
3. **Output Handling**: Handle command output securely

### 7.2 Secure Configuration and Metadata

MCTL secures configuration and metadata:

1. **File Permissions**: Set appropriate file permissions for configuration and metadata files
2. **Sensitive Information**: Avoid storing sensitive information in configuration and metadata

### 7.3 Audit Logging

MCTL provides audit logging for security and compliance:

1. **Operation Logging**: Log all operations for debugging and troubleshooting
2. **Audit Logging**: Log security-relevant operations for compliance and auditing

## 8. Performance Considerations

### 8.1 Parallel Operations

MCTL supports parallel operations for improved performance when working with multiple repositories. The `global.parallel_operations` configuration setting controls the number of concurrent operations.

### 8.2 Caching

MCTL uses caching to improve performance:

1. **Status Cache**: Repository status is cached to avoid repeated Git operations
2. **Metadata Cache**: Repository metadata is cached in memory during command execution

### 8.3 Efficient Git Operations

MCTL is designed to minimize the number of Git operations:

1. **Selective Updates**: Only update repositories that need to be updated
2. **Minimal Git Commands**: Use the minimum number of Git commands required
3. **Efficient Command Construction**: Construct Git commands efficiently to minimize overhead

## 9. Extensibility

### 9.1 Extension Points

MCTL is designed to be extensible:

1. **Command Extensions**: Add new commands by implementing the Cobra command interface
2. **Repository Extensions**: Extend repository functionality by adding methods to the Repository struct
3. **Metadata Extensions**: Extend metadata by adding fields to the Metadata struct

### 9.2 Extension Mechanism

The `Metadata.Extensions` field provides a mechanism for extensions to store additional metadata.

## 10. Conclusion

MCTL provides a powerful, structured approach to managing multiple Git repositories as a cohesive unit. By implementing a management layer over Git, it enables consistent operations, comprehensive metadata tracking, and secure audit capabilities, making it an ideal tool for complex, security-sensitive development environments.

The API specification outlined in this document provides a comprehensive reference for the MCTL command-line interface, including all commands, flags, parameters, and the underlying logic. It is designed to be used as a prompt for code generation, enabling developers to implement the MCTL system according to this specification.
