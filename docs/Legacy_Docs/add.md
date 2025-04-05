# MCTL Add Command

## 1. Introduction

The `mctl add` command is a core component of the MCTL system that enables users to add Git repositories to MCTL management. This command is typically one of the first commands used when setting up MCTL, as it establishes the repositories that will be managed by the system.

### 1.1 Purpose

The primary purposes of the `add` command are:

- **Repository Registration**: Register Git repositories with MCTL for unified management
- **Repository Cloning**: Clone repositories to the local filesystem (optional)
- **Metadata Creation**: Generate and store metadata about the repository
- **Configuration Update**: Update the MCTL configuration to include the new repository

### 1.2 Key Concepts

- **Repository URL**: The Git URL of the repository to be added
- **Repository Path**: The local filesystem path where the repository will be cloned
- **Repository Name**: A unique identifier for the repository within MCTL
- **Repository ID**: An automatically generated unique ID based on name, URL, branch, and path
- **Repository Branch**: The Git branch to clone (optional)

## 2. Command Syntax

### 2.1 Basic Syntax

```
mctl add [options] repository-url [path]
```

### 2.2 Arguments

- `repository-url`: (Required) The Git URL of the repository to add
- `path`: (Optional) The target location for the repository (can also be specified with `--path`)

### 2.3 Options

- `--path=<path>`: Target location for repository
- `--name=<name>`: Custom repository designation (if not provided, derived from URL)
- `--branch=<branch>`: Specific branch to clone
- `--no-clone`: Add to configuration without cloning
- `--flat`: Clone directly to path without creating subdirectory

## 3. Examples

### 3.1 Basic Usage

Add a repository with default settings:

```bash
mctl add git@github.com:example/repo.git
```

This will:
- Clone the repository to a directory named "repo" in the current directory
- Register the repository with MCTL using the name "repo"

### 3.2 Specifying Path

Add a repository to a specific path:

```bash
mctl add git@github.com:example/repo.git classified
```

or

```bash
mctl add git@github.com:example/repo.git --path=classified
```

This will:
- Clone the repository to a directory named "repo" inside the "classified" directory
- Register the repository with MCTL using the name "repo"

### 3.3 Custom Name

Add a repository with a custom name:

```bash
mctl add git@github.com:example/repo.git --name=custom-name
```

This will:
- Clone the repository to a directory named "custom-name" in the current directory
- Register the repository with MCTL using the name "custom-name"

### 3.4 Specific Branch

Add a repository and clone a specific branch:

```bash
mctl add git@github.com:example/repo.git --branch=develop
```

This will:
- Clone the "develop" branch of the repository to a directory named "repo" in the current directory
- Register the repository with MCTL using the name "repo"

### 3.5 No Clone

Add a repository to MCTL management without cloning it:

```bash
mctl add git@github.com:example/repo.git --no-clone
```

This will:
- Register the repository with MCTL using the name "repo"
- Not clone the repository to the local filesystem

### 3.6 Flat Structure

Add a repository and clone it directly to the specified path without creating a subdirectory:

```bash
mctl add git@github.com:example/repo.git --path=classified --flat
```

This will:
- Clone the repository directly to the "classified" directory (not to "classified/repo")
- Register the repository with MCTL using the name "repo"

### 3.7 Combined Options

Add a repository with multiple options:

```bash
mctl add git@github.com:example/repo.git --path=classified --name=secure-repo --branch=release-1.2
```

This will:
- Clone the "release-1.2" branch of the repository to a directory named "secure-repo" inside the "classified" directory
- Register the repository with MCTL using the name "secure-repo"

## 4. Behavior and Workflow

### 4.1 Repository Name Derivation

If a repository name is not provided with the `--name` option, MCTL derives the name from the repository URL:

1. Extract the last part of the URL after the last "/"
2. Remove the ".git" extension if present

For example:
- `git@github.com:example/repo.git` → "repo"
- `https://github.com/example/my-project.git` → "my-project"

### 4.2 Repository Path Determination

The repository path is determined based on the provided options:

1. If `path` is not provided (neither as an argument nor with `--path`):
   - Use the repository name as the path (relative to the current directory)

2. If `path` is provided and `--flat` is false (default):
   - Append the repository name to the path: `<path>/<name>`

3. If `path` is provided and `--flat` is true:
   - Use the path directly without appending the repository name

### 4.3 Repository ID Generation

MCTL generates a unique ID for each repository based on:
- Repository name (normalized to lowercase)
- Repository URL
- Repository branch
- Repository path

The ID is created by:
1. Concatenating these values with a separator
2. Computing a SHA-256 hash of the concatenated string
3. Taking the first 10 characters of the hex-encoded hash

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

### 4.4 Repository Cloning

Unless the `--no-clone` option is specified, MCTL clones the repository:

1. Ensure the parent directory exists
2. Build the Git clone command with appropriate options
3. Execute the Git clone command
4. Update repository metadata

If the `--branch` option is specified, MCTL clones the specified branch.

### 4.5 Configuration Update

After adding a repository, MCTL updates its configuration:

1. Add the repository configuration to the list of managed repositories
2. Save the updated configuration to disk

### 4.6 Logging

MCTL logs the add operation:

1. Log the operation to the operations log
2. Log an audit entry for security and compliance purposes

## 5. Implementation Details

### 5.1 Command Implementation

The `add` command is implemented in the `cmd/add.go` file:

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

### 5.2 Repository Manager Integration

The `add` command uses the Repository Manager to add repositories:

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

### 5.3 Repository Manager's AddRepository Method

The Repository Manager's `AddRepository` method handles the actual repository addition:

```go
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
```

## 6. Error Handling

The `add` command handles various error cases:

### 6.1 Configuration Errors

- **Configuration Not Found**: If MCTL is not initialized in the current directory
- **Configuration Save Failure**: If the configuration cannot be saved after adding the repository

### 6.2 Repository Errors

- **Repository Already Exists**: If a repository with the same path already exists
- **Name Conflict**: If a repository with the same name already exists (resolved by generating a unique name)
- **Clone Failure**: If the repository cannot be cloned
- **Metadata Save Failure**: If the repository metadata cannot be saved

### 6.3 Error Reporting

Errors are reported with context to help users understand and resolve issues:

```go
return errors.Wrap(err, errors.ErrRepositoryExists, "Failed to add repository")
```

## 7. Best Practices

### 7.1 Repository Naming

- Use descriptive names that reflect the repository's purpose
- Keep names concise but meaningful
- Use consistent naming conventions across repositories
- Consider using prefixes for related repositories (e.g., "api-", "ui-", "lib-")

### 7.2 Repository Organization

- Organize repositories in a logical directory structure
- Group related repositories in subdirectories
- Use the `--path` option to specify the target location
- Consider using the `--flat` option for repositories that should be at the root of a directory

### 7.3 Branch Management

- Use the `--branch` option to clone specific branches
- Consider standardizing on a default branch across repositories
- For feature development, add repositories with the feature branch

### 7.4 No-Clone Usage

The `--no-clone` option is useful in several scenarios:

- When the repository is already cloned manually
- When you want to prepare the configuration before cloning
- When you want to clone the repository later with different options
- When you're migrating an existing project to MCTL

## 8. Common Use Cases

### 8.1 Setting Up a New Project

When setting up a new project with multiple repositories:

```bash
# Initialize MCTL
mctl init

# Add repositories
mctl add git@github.com:example/api.git --path=services
mctl add git@github.com:example/ui.git --path=services
mctl add git@github.com:example/docs.git --path=documentation
mctl add git@github.com:example/lib.git --path=libraries

# Check status
mctl status
```

### 8.2 Adding Existing Repositories

When adding existing repositories to MCTL:

```bash
# Initialize MCTL
mctl init

# Add repositories without cloning
mctl add git@github.com:example/api.git --path=services/api --no-clone
mctl add git@github.com:example/ui.git --path=services/ui --no-clone

# Check status
mctl status
```

### 8.3 Feature Development

When setting up repositories for feature development:

```bash
# Initialize MCTL
mctl init

# Add repositories with feature branch
mctl add git@github.com:example/api.git --branch=feature/auth
mctl add git@github.com:example/ui.git --branch=feature/auth

# Create consistent branch across repositories
mctl branch create feature/auth

# Check status
mctl status
```

### 8.4 Microservice Architecture

When working with a microservice architecture:

```bash
# Initialize MCTL
mctl init

# Add microservices
mctl add git@github.com:example/auth-service.git --path=services
mctl add git@github.com:example/user-service.git --path=services
mctl add git@github.com:example/payment-service.git --path=services
mctl add git@github.com:example/notification-service.git --path=services

# Add shared libraries
mctl add git@github.com:example/common-lib.git --path=libraries

# Add frontend
mctl add git@github.com:example/frontend.git

# Check status
mctl status
```

## 9. Conclusion

The `mctl add` command is a fundamental component of the MCTL system, enabling users to add Git repositories to unified management. By providing flexible options for repository naming, path determination, branch selection, and cloning behavior, it accommodates a wide range of use cases and workflows.

Whether setting up a new project, adding existing repositories, or preparing for feature development, the `add` command provides the foundation for MCTL's repository management capabilities. Combined with other MCTL commands, it enables consistent operations across multiple repositories, simplifying complex development workflows and enhancing productivity.
