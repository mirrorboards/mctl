# MCTL: Multi-Repository Control System

MCTL provides secure, unified management of code repositories in high-security environments. It implements a structured management layer over Git repositories, providing consistent operations across multiple codebases while maintaining comprehensive metadata and audit capabilities.

## Features

- **Unified Management**: Manage multiple Git repositories with a single command
- **Consistent Operations**: Apply the same operations across all repositories
- **Comprehensive Metadata**: Track repository status, history, and relationships
- **Audit Capabilities**: Log all operations for security and compliance
- **Secure Operations**: Implement secure, fail-safe command execution

## Installation

### From Source

```bash
git clone https://github.com/mirrorboards/mctl.git
cd mctl
go build
```

### Using Go Install

```bash
go install github.com/mirrorboards/mctl@latest
```

## Quick Start

1. Initialize MCTL in your project directory:

```bash
mctl init
```

2. Add repositories to manage:

```bash
mctl add git@github.com:example/repo1.git
mctl add git@github.com:example/repo2.git
```

3. Check the status of all repositories:

```bash
mctl status
```

4. Synchronize all repositories with their remotes:

```bash
mctl sync
```

5. Create a branch across all repositories:

```bash
mctl branch create feature-branch
```

6. Commit and push changes across all repositories:

```bash
mctl save "Implement new feature"
```

## Command Reference

### Core Commands

- `mctl init`: Initialize a new MCTL configuration environment
- `mctl add`: Add a Git repository to MCTL management
- `mctl remove`: Remove a repository from MCTL management and delete its files
- `mctl list`: List repositories under MCTL management
- `mctl status`: Report status of managed repositories
- `mctl sync`: Update repositories to match remote state
- `mctl branch`: Manage Git branches across repositories
- `mctl save`: Commit and push changes across repositories
- `mctl clear`: Remove repository directories while preserving configuration
- `mctl config`: Manage MCTL configuration
- `mctl logs`: Display MCTL logs

### Additional Commands

- `mctl version`: Display MCTL version information
- `mctl help`: Display help information

## Configuration

MCTL uses a TOML configuration file located at `.mirror/mirror.toml` in the base directory. The configuration file contains global settings and repository definitions.

Example configuration:

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

## Directory Structure

```
$BASE_DIRECTORY/
├── .mirror/                      # Configuration directory
│   ├── mirror.toml               # Primary configuration
│   ├── metadata/                 # Repository metadata
│   │   └── {id}.json             # Individual repository data
│   ├── logs/                     # Operation logs
│   │   ├── operations.log        # Command execution log
│   │   └── audit.log             # Security audit log
│   └── cache/                    # Performance cache
│       └── status/               # Repository status cache
├── {repository-name}/            # Individual repository
└── {repository-name}/            # Individual repository
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
