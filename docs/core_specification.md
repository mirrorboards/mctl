# MCTL Core Specification

## 1. Introduction

This document provides a detailed specification of the core logic and data structures that power the MCTL (Multi-Repository Control System). It describes the state management, data flow, and design decisions that enable MCTL to provide unified control over multiple Git repositories.

### 1.1 Purpose

The core of MCTL is designed to:

1. Maintain a consistent state across multiple repositories
2. Track metadata about repositories and their relationships
3. Provide a foundation for higher-level operations like snapshots and synchronization
4. Ensure security and auditability of all operations
5. Enable extensibility for future enhancements

### 1.2 Core Components

The MCTL core consists of several key components:

1. **Configuration Manager**: Manages the global and repository-specific configuration
2. **Repository Manager**: Coordinates operations across repositories
3. **Metadata System**: Tracks and persists repository state and metadata
4. **Snapshot System**: Captures and restores point-in-time states
5. **Logging System**: Records operations and audit information

## 2. State Management

### 2.1 Global State

The global state of MCTL is maintained in the configuration file and includes:

| Property | Type | Purpose | Persistence |
|----------|------|---------|-------------|
| `default_branch` | String | Default branch for new repositories | Configuration file |
| `parallel_operations` | Integer | Number of operations to run in parallel | Configuration file |
| `default_remote` | String | Default remote name for repositories | Configuration file |

**Rationale**: These global settings provide consistent defaults across all repositories, reducing the need for repetitive configuration and ensuring uniformity in operations.

### 2.2 Repository Registry

The repository registry maintains a list of all repositories under MCTL management:

| Property | Type | Purpose | Persistence |
|----------|------|---------|-------------|
| `repositories` | Array of RepositoryConfig | List of all managed repositories | Configuration file |

**Rationale**: A centralized registry allows MCTL to operate on multiple repositories as a cohesive unit, which is essential for maintaining consistency across codebases.

### 2.3 Repository State

Each repository has its own state, which includes:

| Property | Type | Purpose | Persistence |
|----------|------|---------|-------------|
| `id` | String | Unique identifier for the repository | Configuration file, Metadata file |
| `name` | String | Human-readable name for the repository | Configuration file, Metadata file |
| `path` | String | Path to the repository relative to base directory | Configuration file |
| `url` | String | Git URL of the repository | Configuration file |
| `branch` | String | Current branch (or default branch if not specified) | Configuration file, Metadata file |
| `status` | Enum | Current status (CLEAN, MODIFIED, AHEAD, BEHIND, DIVERGED, UNKNOWN) | Metadata file |
| `creation_date` | Timestamp | When the repository was added to MCTL | Metadata file |
| `last_sync` | Timestamp | When the repository was last synchronized with remote | Metadata file |
| `extensions` | Map | Reserved for future extensions | Metadata file |

**Rationale**: Tracking detailed state for each repository enables MCTL to make informed decisions about operations, provide meaningful status information, and maintain audit trails.

### 2.4 Snapshot State

Snapshots capture the state of all repositories at a specific point in time:

| Property | Type | Purpose | Persistence |
|----------|------|---------|-------------|
| `id` | String | Unique identifier for the snapshot | Snapshot file |
| `created_at` | Timestamp | When the snapshot was created | Snapshot file |
| `description` | String | User-provided or auto-generated description | Snapshot file |
| `repositories` | Array of RepositoryState | State of each repository at snapshot time | Snapshot file |

**Rationale**: Snapshots enable point-in-time recovery and reproducibility, which is essential for complex development workflows and disaster recovery.

### 2.5 Repository State in Snapshots

Each repository's state within a snapshot includes:

| Property | Type | Purpose | Persistence |
|----------|------|---------|-------------|
| `id` | String | Repository ID | Snapshot file |
| `name` | String | Repository name | Snapshot file |
| `path` | String | Repository path | Snapshot file |
| `branch` | String | Branch at snapshot time | Snapshot file |
| `commit_hash` | String | Commit hash at snapshot time | Snapshot file |
| `status` | String | Repository status at snapshot time | Snapshot file |

**Rationale**: Capturing the exact state of each repository enables precise restoration to a previous state, which is critical for reproducibility and recovery.

## 3. Data Flow

### 3.1 Configuration Flow

1. **Loading Configuration**:
   - Read configuration file from `.mirror/mirror.toml`
   - Parse TOML into configuration structure
   - Validate configuration
   - Make configuration available to all components

2. **Saving Configuration**:
   - Validate configuration
   - Serialize configuration to TOML
   - Write to `.mirror/mirror.toml` with appropriate permissions
   - Log configuration change

**Rationale**: A structured configuration flow ensures that configuration changes are validated, persisted, and logged, which is essential for maintaining system integrity.

### 3.2 Repository Operation Flow

1. **Repository Identification**:
   - Identify repository by name or path
   - Load repository configuration
   - Load repository metadata
   - Create repository instance

2. **Operation Execution**:
   - Validate operation parameters
   - Execute Git commands
   - Update repository status
   - Update metadata
   - Log operation

3. **Error Handling**:
   - Capture Git command errors
   - Provide context-specific error messages
   - Log errors
   - Roll back changes if necessary

**Rationale**: A structured operation flow ensures that repository operations are executed consistently, with proper validation, error handling, and logging.

### 3.3 Snapshot Flow

1. **Snapshot Creation**:
   - Capture state of all repositories
   - Generate unique snapshot ID
   - Create snapshot structure
   - Serialize to JSON
   - Write to `.mirror/snapshots/{snapshot-id}.json`
   - Log snapshot creation

2. **Snapshot Application**:
   - Load snapshot from disk
   - Validate snapshot
   - For each repository:
     - Check for uncommitted changes
     - Check out branch
     - Reset to commit hash
     - Update status
   - Log snapshot application

**Rationale**: A structured snapshot flow ensures that snapshots are created and applied consistently, with proper validation, error handling, and logging.

## 4. Core Data Structures

### 4.1 Configuration Structure

```
Config
├── Global
│   ├── DefaultBranch: String
│   ├── ParallelOperations: Integer
│   └── DefaultRemote: String
└── Repositories: Array of RepositoryConfig
    ├── ID: String
    ├── Name: String
    ├── Path: String
    ├── URL: String
    └── Branch: String (optional)
```

**Rationale**: This structure provides a clear separation between global settings and repository-specific configuration, making it easy to manage and extend.

### 4.2 Repository Structure

```
Repository
├── Config: RepositoryConfig
│   ├── ID: String
│   ├── Name: String
│   ├── Path: String
│   ├── URL: String
│   └── Branch: String (optional)
├── Metadata: Metadata
│   ├── ID: String
│   ├── Name: String
│   ├── Basic: BasicInfo
│   │   ├── CreationDate: Timestamp
│   │   └── LastSync: Timestamp
│   ├── Status: StatusInfo
│   │   ├── Current: Status (enum)
│   │   └── Branch: String
│   └── Extensions: Map
└── BaseDir: String
```

**Rationale**: This structure encapsulates all the information needed to manage a repository, including its configuration, metadata, and base directory.

### 4.3 Snapshot Structure

```
Snapshot
├── ID: String
├── CreatedAt: Timestamp
├── Description: String
└── Repositories: Array of RepositoryState
    ├── ID: String
    ├── Name: String
    ├── Path: String
    ├── Branch: String
    ├── CommitHash: String
    └── Status: String
```

**Rationale**: This structure captures the state of all repositories at a specific point in time, enabling precise restoration to a previous state.

### 4.4 Repository Manager Structure

```
Manager
├── Config: Config
└── BaseDir: String
```

**Rationale**: The Repository Manager is a lightweight structure that coordinates operations across repositories, using the configuration and base directory as its primary state.

### 4.5 Snapshot Manager Structure

```
Manager
└── BaseDir: String
```

**Rationale**: The Snapshot Manager is a lightweight structure that coordinates snapshot operations, using the base directory as its primary state.

## 5. State Persistence

### 5.1 Configuration Persistence

The configuration is persisted in the `.mirror/mirror.toml` file using the TOML format:

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

**Rationale**: TOML provides a human-readable and editable format for configuration, with clear structure and support for arrays of tables, which is ideal for repository definitions.

### 5.2 Metadata Persistence

Repository metadata is persisted in the `.mirror/metadata/{id}.json` file using the JSON format:

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

**Rationale**: JSON provides a structured format for metadata, with support for nested objects and arrays, which is ideal for repository metadata.

### 5.3 Snapshot Persistence

Snapshots are persisted in the `.mirror/snapshots/{snapshot-id}.json` file using the JSON format:

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

**Rationale**: JSON provides a structured format for snapshots, with support for nested objects and arrays, which is ideal for repository state snapshots.

### 5.4 Log Persistence

Logs are persisted in the `.mirror/logs` directory:

- `operations.log`: Records all operations for debugging and troubleshooting
- `audit.log`: Records security-relevant operations for compliance and auditing

Logs are stored as line-delimited JSON:

```json
{"timestamp":"2025-04-05T12:34:56Z","level":"INFO","message":"Added repository secure-comms (git@secure.gov:systems/secure-comms.git)"}
{"timestamp":"2025-04-05T12:35:00Z","level":"INFO","message":"Created branch feature/auth in repository secure-comms"}
```

**Rationale**: Line-delimited JSON provides a structured format for logs, with each log entry as a separate JSON object, which is ideal for log processing and analysis.

## 6. Core Logic

### 6.1 Repository Identification

Repositories are identified by:

1. **ID**: A unique identifier generated from the repository name, URL, branch, and path
2. **Name**: A human-readable name for the repository
3. **Path**: The path to the repository relative to the base directory

The ID is generated using a SHA-256 hash of the concatenated name, URL, branch, and path:

```
ID = SHA-256(name|url|branch|path)[:10]
```

**Rationale**: This approach ensures that each repository has a unique identifier that is stable across operations, even if the repository name or path changes.

### 6.2 Repository Status Determination

Repository status is determined by:

1. **Local Changes**: Whether the repository has uncommitted changes
2. **Remote Relationship**: Whether the repository is ahead, behind, or diverged from remote

The status is represented by an enum:

- `CLEAN`: No local changes, in sync with remote
- `MODIFIED`: Local changes
- `AHEAD`: No local changes, ahead of remote
- `BEHIND`: No local changes, behind remote
- `DIVERGED`: No local changes, diverged from remote
- `UNKNOWN`: Status cannot be determined

**Rationale**: This approach provides a clear and consistent representation of repository status, which is essential for making informed decisions about operations.

### 6.3 Snapshot ID Generation

Snapshot IDs are generated using a combination of:

1. A timestamp in the format `YYYYMMDD-HHMMSS`
2. A hash of repository states to ensure uniqueness

```
ID = YYYYMMDD-HHMMSS-SHA-256(repo1.id|repo1.branch|repo1.commit_hash|...)[:8]
```

**Rationale**: This approach ensures that each snapshot has a unique identifier that is both human-readable (with the timestamp) and guaranteed to be unique (with the hash).

### 6.4 Error Handling

Errors are handled using a structured approach:

1. **Error Types**: Errors are categorized by type (e.g., `repository_exists`, `git_command_failed`)
2. **Error Context**: Errors include context information (e.g., repository name, command output)
3. **Error Wrapping**: Errors are wrapped with additional context as they propagate up the call stack

**Rationale**: This approach provides clear and actionable error messages, which is essential for troubleshooting and resolving issues.

### 6.5 Logging

Operations are logged using a structured approach:

1. **Operation Logging**: All operations are logged for debugging and troubleshooting
2. **Audit Logging**: Security-relevant operations are logged for compliance and auditing
3. **Log Levels**: Logs are categorized by level (DEBUG, INFO, WARN, ERROR)

**Rationale**: This approach provides comprehensive logging for both operational and security purposes, which is essential for troubleshooting and compliance.

## 7. Security Considerations

### 7.1 File Permissions

MCTL sets appropriate file permissions for security:

- Configuration directory: `0700` (owner read/write/execute)
- Configuration file: `0600` (owner read/write)
- Metadata directory: `0700` (owner read/write/execute)
- Metadata files: `0600` (owner read/write)
- Snapshots directory: `0700` (owner read/write/execute)
- Snapshot files: `0600` (owner read/write)
- Logs directory: `0700` (owner read/write/execute)
- Log files: `0600` (owner read/write)

**Rationale**: These permissions ensure that sensitive information is only accessible to the owner, which is essential for security.

### 7.2 Sensitive Information

MCTL avoids storing sensitive information in configuration and metadata:

- No passwords or tokens in configuration
- No sensitive repository content in metadata
- No sensitive repository content in snapshots

**Rationale**: This approach minimizes the risk of sensitive information exposure, which is essential for security.

### 7.3 Command Execution

MCTL executes Git commands securely:

1. **Input Validation**: Validate all user input before constructing Git commands
2. **Command Construction**: Construct Git commands using proper argument handling
3. **Output Handling**: Handle command output securely

**Rationale**: This approach minimizes the risk of command injection and other security vulnerabilities, which is essential for security.

## 8. Performance Considerations

### 8.1 Parallel Operations

MCTL supports parallel operations for improved performance:

1. **Parallel Execution**: Execute operations on multiple repositories in parallel
2. **Concurrency Control**: Limit the number of concurrent operations to avoid resource exhaustion
3. **Progress Tracking**: Track progress of parallel operations for user feedback

**Rationale**: This approach improves performance when working with multiple repositories, which is essential for large-scale development.

### 8.2 Caching

MCTL uses caching to improve performance:

1. **Status Cache**: Cache repository status to avoid repeated Git operations
2. **Metadata Cache**: Cache repository metadata in memory during command execution

**Rationale**: This approach reduces the number of Git operations and file system accesses, which improves performance.

### 8.3 Efficient Git Operations

MCTL is designed to minimize the number of Git operations:

1. **Selective Updates**: Only update repositories that need to be updated
2. **Minimal Git Commands**: Use the minimum number of Git commands required
3. **Efficient Command Construction**: Construct Git commands efficiently to minimize overhead

**Rationale**: This approach reduces the overhead of Git operations, which improves performance.

## 9. Extensibility

### 9.1 Extension Points

MCTL is designed to be extensible:

1. **Command Extensions**: Add new commands by implementing the Cobra command interface
2. **Repository Extensions**: Extend repository functionality by adding methods to the Repository struct
3. **Metadata Extensions**: Extend metadata by adding fields to the Metadata struct

**Rationale**: This approach enables future enhancements without modifying the core codebase, which is essential for maintainability.

### 9.2 Extension Mechanism

The `Metadata.Extensions` field provides a mechanism for extensions to store additional metadata:

```json
{
  "extensions": {
    "custom_extension": {
      "property1": "value1",
      "property2": "value2"
    }
  }
}
```

**Rationale**: This approach enables extensions to store their own metadata without modifying the core metadata structure, which is essential for extensibility.

## 10. Conclusion

The MCTL core provides a solid foundation for managing multiple Git repositories as a cohesive unit. By maintaining detailed state, providing structured operations, and ensuring security and auditability, it enables consistent operations across repositories, comprehensive metadata tracking, and secure audit capabilities.

The core specification outlined in this document provides a comprehensive reference for the MCTL core logic and data structures, including state management, data flow, and design decisions. It is designed to guide the implementation of the MCTL system and provide a reference for future enhancements.
