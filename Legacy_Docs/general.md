# MCTL: Multi-Repository Control System

## Overview

MCTL (Multi-Repository Control System) is a command-line tool designed to provide secure, unified management of code repositories in high-security environments. It implements a structured management layer over Git repositories, enabling consistent operations across multiple codebases while maintaining comprehensive metadata and audit capabilities.

MCTL is particularly valuable in environments where multiple related repositories need to be managed together, such as microservice architectures, complex software systems with multiple components, or secure development environments requiring strict audit trails.

## Key Features

### Unified Repository Management

- **Centralized Control**: Manage multiple Git repositories through a single command interface
- **Consistent Operations**: Apply the same operations (branching, committing, etc.) across all repositories simultaneously
- **Synchronized State**: Ensure all repositories maintain a consistent state

### Repository Snapshot System

- **Point-in-Time Snapshots**: Create snapshots of the state of all repositories at a specific moment
- **State Restoration**: Restore repositories to a previous state using saved snapshots
- **Snapshot Management**: List, inspect, and manage repository snapshots

### Security and Compliance

- **Comprehensive Audit Logging**: Track all operations for security and compliance purposes
- **Metadata Tracking**: Maintain detailed metadata about repository status, history, and relationships
- **Secure Operations**: Implement secure, fail-safe command execution

### Workflow Optimization

- **Parallel Operations**: Execute commands across multiple repositories in parallel
- **Status Monitoring**: Quickly check the status of all managed repositories
- **Simplified Branching**: Create and manage branches across repositories with a single command

## Command Structure

MCTL provides a comprehensive set of commands for repository management:

### Core Repository Management

- `mctl init`: Initialize a new MCTL configuration environment
- `mctl add`: Add a Git repository to MCTL management
- `mctl remove`: Remove a repository from MCTL management
- `mctl list`: List repositories under MCTL management
- `mctl status`: Report status of managed repositories
- `mctl sync`: Update repositories to match remote state

### Git Operations

- `mctl branch`: Manage Git branches across repositories
- `mctl save`: Commit and push changes across repositories
- `mctl clear`: Remove repository directories while preserving configuration

### Snapshot System

- `mctl save`: Creates snapshots when committing changes (unless --no-snapshot is specified)
- `mctl load`: Restore repositories to a saved snapshot state
- `mctl snapshots`: List and inspect available snapshots

### Configuration and Utilities

- `mctl config`: Manage MCTL configuration
- `mctl logs`: Display MCTL operation logs
- `mctl version`: Display MCTL version information
- `mctl man`: Display manual pages for MCTL commands

## Repository Management

MCTL manages repositories through a structured approach:

1. **Repository Registration**: Repositories are registered with MCTL using the `add` command, which clones the repository and adds it to the configuration.

2. **Status Tracking**: MCTL tracks the status of each repository, including:
   - Current branch
   - Commit hash
   - Local changes
   - Relationship with remote (ahead, behind, diverged)

3. **Unified Operations**: Operations like branching, committing, and syncing can be performed across all repositories with a single command.

4. **Metadata Storage**: MCTL maintains metadata for each repository, including creation date, last sync time, and current status.

## Snapshot System

The Repository Snapshot System is a key feature of MCTL that enables point-in-time snapshots of repository states:

### Snapshot Creation

Snapshots are automatically created when using the `mctl save` command (unless disabled with `--no-snapshot`). Each snapshot contains:

- A unique ID (timestamp + hash)
- Creation timestamp
- Description (defaults to commit message)
- State information for each repository:
  - Repository ID and name
  - Current branch
  - Commit hash
  - Status

### Snapshot Management

Snapshots can be managed using the `mctl snapshots` command, which allows:

- Listing all available snapshots
- Viewing detailed information about specific snapshots
- Limiting the list to recent snapshots

### Snapshot Restoration

The `mctl load` command restores repositories to a saved snapshot state by:

- Checking out the correct branch for each repository
- Resetting to the saved commit hash
- Supporting options for dry-run, force loading, and limiting to specific repositories

## Configuration

MCTL uses a configuration file located at `.mirror/mirror.toml` in the base directory, containing:

### Global Configuration

```toml
[global]
default_branch = "main"
parallel_operations = 4
default_remote = "origin"
```

### Repository Definitions

```toml
[[repositories]]
id = "a1b2c3d4e5"
name = "secure-comms"
path = "repositories/secure-comms"
url = "git@secure.gov:systems/secure-comms.git"
branch = "main"
```

## Directory Structure

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

## Use Cases

MCTL is particularly valuable in the following scenarios:

### Microservice Development

When working with microservice architectures involving multiple repositories, MCTL simplifies:
- Creating consistent feature branches across all services
- Committing changes to multiple repositories with a single command
- Maintaining a snapshot of a working state across all services

### Secure Development Environments

In high-security environments, MCTL provides:
- Comprehensive audit logging of all repository operations
- Consistent application of security policies across repositories
- Ability to restore to known-good states using snapshots

### Complex System Development

For complex systems with multiple components, MCTL enables:
- Synchronized state management across components
- Simplified workflow for developers working across multiple repositories
- Point-in-time snapshots for testing and deployment

## Conclusion

MCTL provides a powerful, structured approach to managing multiple Git repositories as a cohesive unit. By implementing a management layer over Git, it enables consistent operations, comprehensive metadata tracking, and secure audit capabilities, making it an ideal tool for complex, security-sensitive development environments.
