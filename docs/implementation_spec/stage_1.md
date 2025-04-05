# MCTL Core Engine Implementation Specification - Stage 1

## 1. Introduction

This document outlines the implementation details for Stage 1 of the MCTL (Multi-Repository Control System) development, focusing on the core engine. The core engine provides the foundation for all MCTL operations, including repository management, state tracking, and metadata handling.

### 1.1 Purpose

The purpose of Stage 1 is to implement the essential components of the MCTL core engine, including:

1. Configuration management
2. Repository management
3. Metadata system
4. Basic Git operations
5. Logging system

These components will provide the foundation for subsequent stages of development, including the snapshot system and advanced Git operations.

### 1.2 Scope

Stage 1 includes the implementation of:

- Core data structures
- Configuration file handling
- Repository management operations
- Metadata tracking
- Basic Git command execution
- Logging and error handling
- Command-line interface for core commands

Stage 1 does not include:

- Snapshot system (will be implemented in Stage 2)
- Advanced Git operations (will be implemented in Stage 2)
- Parallel operations (will be implemented in Stage 3)
- Extension mechanisms (will be implemented in Stage 3)

## 2. Implementation Architecture

### 2.1 Package Structure

The MCTL codebase will be organized into the following package structure:

```
mctl/
├── cmd/                  # Command-line interface
│   ├── add.go            # Add repository command
│   ├── init.go           # Initialize command
│   ├── list.go           # List repositories command
│   ├── remove.go         # Remove repository command
│   ├── root.go           # Root command
│   ├── status.go         # Status command
│   └── version.go        # Version command
├── internal/             # Internal packages
│   ├── config/           # Configuration management
│   │   └── config.go     # Configuration handling
│   ├── errors/           # Error handling
│   │   └── errors.go     # Error types and handling
│   ├── git/              # Git operations
│   │   └── git.go        # Git command execution
│   ├── logging/          # Logging system
│   │   └── logging.go    # Logging functionality
│   ├── repository/       # Repository management
│   │   └── repository.go # Repository operations
│   └── utils/            # Utility functions
│       └── utils.go      # Common utilities
├── main.go               # Application entry point
└── go.mod                # Go module definition
```

### 2.2 Dependency Management

MCTL will use the following external dependencies:

1. **spf13/cobra**: Command-line interface framework
2. **spf13/viper**: Configuration management
3. **go-git/go-git**: Git operations (for operations that don't require the git CLI)
4. **rs/zerolog**: Structured logging
5. **BurntSushi/toml**: TOML parsing and serialization
6. **google/uuid**: UUID generation for repository and snapshot IDs
7. **fatih/color**: Terminal color output
8. **olekukonko/tablewriter**: Formatted table output

These dependencies will be managed using Go modules.

### 2.3 Error Handling Strategy

MCTL will use a structured error handling approach:

1. Define custom error types in the `errors` package
2. Use error wrapping to add context to errors
3. Provide detailed error messages with context
4. Log errors with appropriate context
5. Return meaningful error codes to the command-line interface

### 2.4 Logging Strategy

MCTL will use a structured logging approach:

1. Use zerolog for structured JSON logging
2. Log to both console and file
3. Use appropriate log levels (debug, info, warn, error)
4. Include context information in log entries
5. Separate operational and audit logs

## 3. Core Data Structures

### 3.1 Configuration Structure

```go
// Config represents the global MCTL configuration
type Config struct {
    Global       GlobalConfig     `toml:"global"`
    Repositories []RepositoryConfig `toml:"repositories"`
}

// GlobalConfig represents global configuration settings
type GlobalConfig struct {
    DefaultBranch      string `toml:"default_branch"`
    ParallelOperations int    `toml:"parallel_operations"`
    DefaultRemote      string `toml:"default_remote"`
}

// RepositoryConfig represents the configuration for a repository
type RepositoryConfig struct {
    ID     string `toml:"id"`
    Name   string `toml:"name"`
    Path   string `toml:"path"`
    URL    string `toml:"url"`
    Branch string `toml:"branch,omitempty"`
}
```

### 3.2 Repository Structure

```go
// Repository represents a Git repository managed by MCTL
type Repository struct {
    Config   RepositoryConfig
    Metadata Metadata
    BaseDir  string
}

// Metadata represents repository metadata
type Metadata struct {
    ID        string     `json:"id"`
    Name      string     `json:"name"`
    Basic     BasicInfo  `json:"basic"`
    Status    StatusInfo `json:"status"`
    Extensions map[string]interface{} `json:"extensions"`
}

// BasicInfo represents basic repository information
type BasicInfo struct {
    CreationDate time.Time `json:"creation_date"`
    LastSync     time.Time `json:"last_sync"`
}

// StatusInfo represents repository status information
type StatusInfo struct {
    Current Status `json:"current"`
    Branch  string `json:"branch"`
}

// Status represents the status of a repository
type Status string

const (
    StatusClean    Status = "CLEAN"
    StatusModified Status = "MODIFIED"
    StatusAhead    Status = "AHEAD"
    StatusBehind   Status = "BEHIND"
    StatusDiverged Status = "DIVERGED"
    StatusUnknown  Status = "UNKNOWN"
)
```

### 3.3 Error Structure

```go
// ErrorType represents the type of error
type ErrorType string

const (
    ErrorInternal          ErrorType = "internal_error"
    ErrorConfigNotFound    ErrorType = "config_not_found"
    ErrorRepositoryExists  ErrorType = "repository_exists"
    ErrorRepositoryNotFound ErrorType = "repository_not_found"
    ErrorGitCommandFailed  ErrorType = "git_command_failed"
    ErrorInvalidArgument   ErrorType = "invalid_argument"
)

// Error represents a structured error
type Error struct {
    Type    ErrorType
    Message string
    Details string
    Err     error
}

// Error implements the error interface
func (e *Error) Error() string {
    return fmt.Sprintf("Error: %s\nType: %s\nDetails: %s", e.Message, e.Type, e.Details)
}
```

### 3.4 Log Entry Structure

```go
// LogEntry represents a structured log entry
type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Level     string    `json:"level"`
    Message   string    `json:"message"`
    Context   map[string]interface{} `json:"context,omitempty"`
}
```

## 4. Configuration Management

### 4.1 Configuration File Format

The configuration file will be stored in TOML format at `.mirror/mirror.toml`:

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

### 4.2 Configuration Loading

```go
// LoadConfig loads the configuration from disk
func LoadConfig(baseDir string) (*Config, error) {
    configPath := filepath.Join(baseDir, ".mirror", "mirror.toml")
    
    // Check if config file exists
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        return nil, &Error{
            Type:    ErrorConfigNotFound,
            Message: "Configuration not found",
            Details: fmt.Sprintf("Configuration file not found at %s", configPath),
            Err:     err,
        }
    }
    
    // Read config file
    configData, err := os.ReadFile(configPath)
    if err != nil {
        return nil, &Error{
            Type:    ErrorInternal,
            Message: "Failed to read configuration",
            Details: fmt.Sprintf("Failed to read configuration file at %s", configPath),
            Err:     err,
        }
    }
    
    // Parse TOML
    var config Config
    if err := toml.Unmarshal(configData, &config); err != nil {
        return nil, &Error{
            Type:    ErrorInternal,
            Message: "Failed to parse configuration",
            Details: "Configuration file contains invalid TOML",
            Err:     err,
        }
    }
    
    // Validate config
    if err := validateConfig(&config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

### 4.3 Configuration Saving

```go
// SaveConfig saves the configuration to disk
func SaveConfig(config *Config, baseDir string) error {
    // Validate config
    if err := validateConfig(config); err != nil {
        return err
    }
    
    // Create config directory if it doesn't exist
    configDir := filepath.Join(baseDir, ".mirror")
    if err := os.MkdirAll(configDir, 0700); err != nil {
        return &Error{
            Type:    ErrorInternal,
            Message: "Failed to create configuration directory",
            Details: fmt.Sprintf("Failed to create directory at %s", configDir),
            Err:     err,
        }
    }
    
    // Serialize to TOML
    configData, err := toml.Marshal(config)
    if err != nil {
        return &Error{
            Type:    ErrorInternal,
            Message: "Failed to serialize configuration",
            Details: "Failed to convert configuration to TOML",
            Err:     err,
        }
    }
    
    // Write to file
    configPath := filepath.Join(configDir, "mirror.toml")
    if err := os.WriteFile(configPath, configData, 0600); err != nil {
        return &Error{
            Type:    ErrorInternal,
            Message: "Failed to write configuration",
            Details: fmt.Sprintf("Failed to write configuration to %s", configPath),
            Err:     err,
        }
    }
    
    return nil
}
```

### 4.4 Configuration Validation

```go
// validateConfig validates the configuration
func validateConfig(config *Config) error {
    // Validate global config
    if config.Global.DefaultBranch == "" {
        config.Global.DefaultBranch = "main"
    }
    if config.Global.ParallelOperations <= 0 {
        config.Global.ParallelOperations = 4
    }
    if config.Global.DefaultRemote == "" {
        config.Global.DefaultRemote = "origin"
    }
    
    // Validate repositories
    repoNames := make(map[string]bool)
    repoPaths := make(map[string]bool)
    
    for i, repo := range config.Repositories {
        // Validate ID
        if repo.ID == "" {
            return &Error{
                Type:    ErrorInvalidArgument,
                Message: "Invalid repository configuration",
                Details: fmt.Sprintf("Repository at index %d has no ID", i),
            }
        }
        
        // Validate name
        if repo.Name == "" {
            return &Error{
                Type:    ErrorInvalidArgument,
                Message: "Invalid repository configuration",
                Details: fmt.Sprintf("Repository with ID %s has no name", repo.ID),
            }
        }
        
        // Check for duplicate names
        if repoNames[repo.Name] {
            return &Error{
                Type:    ErrorInvalidArgument,
                Message: "Invalid repository configuration",
                Details: fmt.Sprintf("Duplicate repository name: %s", repo.Name),
            }
        }
        repoNames[repo.Name] = true
        
        // Validate path
        if repo.Path == "" {
            return &Error{
                Type:    ErrorInvalidArgument,
                Message: "Invalid repository configuration",
                Details: fmt.Sprintf("Repository with ID %s has no path", repo.ID),
            }
        }
        
        // Check for duplicate paths
        if repoPaths[repo.Path] {
            return &Error{
                Type:    ErrorInvalidArgument,
                Message: "Invalid repository configuration",
                Details: fmt.Sprintf("Duplicate repository path: %s", repo.Path),
            }
        }
        repoPaths[repo.Path] = true
        
        // Validate URL
        if repo.URL == "" {
            return &Error{
                Type:    ErrorInvalidArgument,
                Message: "Invalid repository configuration",
                Details: fmt.Sprintf("Repository with ID %s has no URL", repo.ID),
            }
        }
    }
    
    return nil
}
```

## 5. Repository Management

### 5.1 Repository Initialization

```go
// InitializeRepository initializes a new repository
func InitializeRepository(config *Config, name, url, path, branch string, baseDir string, noClone bool) (*Repository, error) {
    // Generate repository ID
    id := generateRepositoryID(name, url, branch, path)
    
    // Check if repository with same name already exists
    for _, repo := range config.Repositories {
        if repo.Name == name {
            return nil, &Error{
                Type:    ErrorRepositoryExists,
                Message: "Failed to add repository",
                Details: fmt.Sprintf("Repository with name %s already exists", name),
            }
        }
        
        if repo.Path == path {
            return nil, &Error{
                Type:    ErrorRepositoryExists,
                Message: "Failed to add repository",
                Details: fmt.Sprintf("Repository already exists at path: %s", path),
            }
        }
    }
    
    // Create repository config
    repoConfig := RepositoryConfig{
        ID:     id,
        Name:   name,
        Path:   path,
        URL:    url,
        Branch: branch,
    }
    
    // Clone repository if needed
    if !noClone {
        if err := cloneRepository(repoConfig, baseDir); err != nil {
            return nil, err
        }
    }
    
    // Create repository metadata
    metadata := Metadata{
        ID:   id,
        Name: name,
        Basic: BasicInfo{
            CreationDate: time.Now(),
            LastSync:     time.Now(),
        },
        Status: StatusInfo{
            Current: StatusClean,
            Branch:  branch,
        },
        Extensions: make(map[string]interface{}),
    }
    
    // Save metadata
    if err := saveMetadata(metadata, baseDir); err != nil {
        return nil, err
    }
    
    // Add repository to config
    config.Repositories = append(config.Repositories, repoConfig)
    
    // Save config
    if err := SaveConfig(config, baseDir); err != nil {
        return nil, err
    }
    
    // Create repository instance
    repo := &Repository{
        Config:   repoConfig,
        Metadata: metadata,
        BaseDir:  baseDir,
    }
    
    return repo, nil
}
```

### 5.2 Repository Cloning

```go
// cloneRepository clones a Git repository
func cloneRepository(config RepositoryConfig, baseDir string) error {
    // Construct repository path
    repoPath := filepath.Join(baseDir, config.Path)
    
    // Create parent directory if it doesn't exist
    parentDir := filepath.Dir(repoPath)
    if err := os.MkdirAll(parentDir, 0755); err != nil {
        return &Error{
            Type:    ErrorInternal,
            Message: "Failed to create directory",
            Details: fmt.Sprintf("Failed to create directory at %s", parentDir),
            Err:     err,
        }
    }
    
    // Construct git command
    args := []string{"clone", config.URL, repoPath}
    if config.Branch != "" {
        args = append(args, "--branch", config.Branch)
    }
    
    // Execute git command
    cmd := exec.Command("git", args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return &Error{
            Type:    ErrorGitCommandFailed,
            Message: "Failed to clone repository",
            Details: fmt.Sprintf("Git clone failed: %s", string(output)),
            Err:     err,
        }
    }
    
    return nil
}
```

### 5.3 Repository Removal

```go
// RemoveRepository removes a repository from MCTL management
func RemoveRepository(config *Config, name string, baseDir string, delete bool, force bool) error {
    // Find repository
    var repoConfig RepositoryConfig
    var repoIndex int
    found := false
    
    for i, repo := range config.Repositories {
        if repo.Name == name {
            repoConfig = repo
            repoIndex = i
            found = true
            break
        }
    }
    
    if !found {
        return &Error{
            Type:    ErrorRepositoryNotFound,
            Message: "Failed to remove repository",
            Details: fmt.Sprintf("Repository with name %s not found", name),
        }
    }
    
    // Check for uncommitted changes if delete is true and force is false
    if delete && !force {
        status, err := getRepositoryStatus(repoConfig, baseDir)
        if err != nil {
            return err
        }
        
        if status == StatusModified {
            return &Error{
                Type:    ErrorUncommittedChanges,
                Message: "Failed to remove repository",
                Details: fmt.Sprintf("Repository %s has uncommitted changes", name),
            }
        }
    }
    
    // Delete repository directory if requested
    if delete {
        repoPath := filepath.Join(baseDir, repoConfig.Path)
        if err := os.RemoveAll(repoPath); err != nil {
            return &Error{
                Type:    ErrorInternal,
                Message: "Failed to delete repository directory",
                Details: fmt.Sprintf("Failed to delete directory at %s", repoPath),
                Err:     err,
            }
        }
    }
    
    // Delete metadata
    metadataPath := filepath.Join(baseDir, ".mirror", "metadata", repoConfig.ID+".json")
    if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
        return &Error{
            Type:    ErrorInternal,
            Message: "Failed to delete repository metadata",
            Details: fmt.Sprintf("Failed to delete metadata at %s", metadataPath),
            Err:     err,
        }
    }
    
    // Remove repository from config
    config.Repositories = append(config.Repositories[:repoIndex], config.Repositories[repoIndex+1:]...)
    
    // Save config
    if err := SaveConfig(config, baseDir); err != nil {
        return err
    }
    
    return nil
}
```

### 5.4 Repository Status

```go
// GetRepositoryStatus gets the status of a repository
func GetRepositoryStatus(config RepositoryConfig, baseDir string) (Status, error) {
    // Construct repository path
    repoPath := filepath.Join(baseDir, config.Path)
    
    // Check if repository exists
    if _, err := os.Stat(repoPath); os.IsNotExist(err) {
        return StatusUnknown, &Error{
            Type:    ErrorRepositoryNotFound,
            Message: "Failed to get repository status",
            Details: fmt.Sprintf("Repository directory not found at %s", repoPath),
            Err:     err,
        }
    }
    
    // Check for uncommitted changes
    cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
    output, err := cmd.Output()
    if err != nil {
        return StatusUnknown, &Error{
            Type:    ErrorGitCommandFailed,
            Message: "Failed to get repository status",
            Details: fmt.Sprintf("Git status failed: %s", err.Error()),
            Err:     err,
        }
    }
    
    if len(output) > 0 {
        return StatusModified, nil
    }
    
    // Check relationship with remote
    cmd = exec.Command("git", "-C", repoPath, "rev-list", "--count", "--left-right", "@{upstream}...HEAD")
    output, err = cmd.Output()
    if err != nil {
        // If error, assume clean (might be no upstream)
        return StatusClean, nil
    }
    
    parts := strings.Split(strings.TrimSpace(string(output)), "\t")
    if len(parts) != 2 {
        return StatusClean, nil
    }
    
    behind, err := strconv.Atoi(parts[0])
    if err != nil {
        return StatusClean, nil
    }
    
    ahead, err := strconv.Atoi(parts[1])
    if err != nil {
        return StatusClean, nil
    }
    
    if ahead > 0 && behind > 0 {
        return StatusDiverged, nil
    } else if ahead > 0 {
        return StatusAhead, nil
    } else if behind > 0 {
        return StatusBehind, nil
    }
    
    return StatusClean, nil
}
```

### 5.5 Repository Metadata

```go
// LoadMetadata loads repository metadata from disk
func LoadMetadata(id string, baseDir string) (Metadata, error) {
    metadataPath := filepath.Join(baseDir, ".mirror", "metadata", id+".json")
    
    // Check if metadata file exists
    if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
        return Metadata{}, &Error{
            Type:    ErrorRepositoryNotFound,
            Message: "Failed to load repository metadata",
            Details: fmt.Sprintf("Metadata file not found at %s", metadataPath),
            Err:     err,
        }
    }
    
    // Read metadata file
    metadataData, err := os.ReadFile(metadataPath)
    if err != nil {
        return Metadata{}, &Error{
            Type:    ErrorInternal,
            Message: "Failed to read repository metadata",
            Details: fmt.Sprintf("Failed to read metadata file at %s", metadataPath),
            Err:     err,
        }
    }
    
    // Parse JSON
    var metadata Metadata
    if err := json.Unmarshal(metadataData, &metadata); err != nil {
        return Metadata{}, &Error{
            Type:    ErrorInternal,
            Message: "Failed to parse repository metadata",
            Details: "Metadata file contains invalid JSON",
            Err:     err,
        }
    }
    
    return metadata, nil
}

// SaveMetadata saves repository metadata to disk
func SaveMetadata(metadata Metadata, baseDir string) error {
    // Create metadata directory if it doesn't exist
    metadataDir := filepath.Join(baseDir, ".mirror", "metadata")
    if err := os.MkdirAll(metadataDir, 0700); err != nil {
        return &Error{
            Type:    ErrorInternal,
            Message: "Failed to create metadata directory",
            Details: fmt.Sprintf("Failed to create directory at %s", metadataDir),
            Err:     err,
        }
    }
    
    // Serialize to JSON
    metadataData, err := json.MarshalIndent(metadata, "", "  ")
    if err != nil {
        return &Error{
            Type:    ErrorInternal,
            Message: "Failed to serialize repository metadata",
            Details: "Failed to convert metadata to JSON",
            Err:     err,
        }
    }
    
    // Write to file
    metadataPath := filepath.Join(metadataDir, metadata.ID+".json")
    if err := os.WriteFile(metadataPath, metadataData, 0600); err != nil {
        return &Error{
            Type:    ErrorInternal,
            Message: "Failed to write repository metadata",
            Details: fmt.Sprintf("Failed to write metadata to %s", metadataPath),
            Err:     err,
        }
    }
    
    return nil
}
```

## 6. Git Operations

### 6.1 Git Command Execution

```go
// ExecuteGitCommand executes a Git command in a repository
func ExecuteGitCommand(repoPath string, args ...string) (string, error) {
    // Construct git command
    cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
    
    // Execute command
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", &Error{
            Type:    ErrorGitCommandFailed,
            Message: "Git command failed",
            Details: fmt.Sprintf("Command: git %s\nOutput: %s", strings.Join(args, " "), string(output)),
            Err:     err,
        }
    }
    
    return string(output), nil
}
```

### 6.2 Repository Status Determination

```go
// DetermineRepositoryStatus determines the status of a repository
func DetermineRepositoryStatus(repoPath string) (Status, error) {
    // Check for uncommitted changes
    output, err := ExecuteGitCommand(repoPath, "status", "--porcelain")
    if err != nil {
        return StatusUnknown, err
    }
    
    if len(output) > 0 {
        return StatusModified, nil
    }
    
    // Check relationship with remote
    output, err = ExecuteGitCommand(repoPath, "rev-list", "--count", "--left-right", "@{upstream}...HEAD")
    if err != nil {
        // If error, assume clean (might be no upstream)
        return StatusClean, nil
    }
    
    parts := strings.Split(strings.TrimSpace(output), "\t")
    if len(parts) != 2 {
        return StatusClean, nil
    }
    
    behind, err := strconv.Atoi(parts[0])
    if err != nil {
        return StatusClean, nil
    }
    
    ahead, err := strconv.Atoi(parts[1])
    if err != nil {
        return StatusClean, nil
    }
    
    if ahead > 0 && behind > 0 {
        return StatusDiverged, nil
    } else if ahead > 0 {
        return StatusAhead, nil
    } else if behind > 0 {
        return StatusBehind, nil
    }
    
    return StatusClean, nil
}
```

### 6.3 Branch Management

```go
// GetCurrentBranch gets the current branch of a repository
func GetCurrentBranch(repoPath string) (string, error) {
    output, err := ExecuteGitCommand(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
    if err != nil {
        return "", err
    }
    
    return strings.TrimSpace(output), nil
}

// CheckoutBranch checks out a branch in a repository
func CheckoutBranch(repoPath, branch string, create bool) error {
    args := []string{"checkout"}
    if create {
        args = append(args, "-b")
    }
    args = append(args, branch)
    
    _, err := ExecuteGitCommand(repoPath, args...)
    return err
}
```

## 7. Logging System

### 7.1 Logger Initialization

```go
// InitializeLogger initializes the logging system
func InitializeLogger(baseDir string) error {
    // Create logs directory if it doesn't exist
    logsDir := filepath.Join(baseDir, ".mirror", "logs")
    if err := os.MkdirAll(logsDir, 0700); err != nil {
        return &Error{
            Type:    ErrorInternal,
            Message: "Failed to create logs directory",
            Details: fmt.Sprintf("Failed to create directory at %s", logsDir),
            Err:     err,
        }
    }
    
    // Open operations log file
    operationsLogPath := filepath.Join(logsDir, "operations.log")
    operationsLogFile, err := os.OpenFile(operationsLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
    if err != nil {
        return &Error{
            Type:    ErrorInternal,
            Message: "Failed to open operations log file",
            Details: fmt.Sprintf("Failed to open file at %s", operationsLogPath),
            Err:     err,
        }
    }
    
    // Open audit log file
    auditLogPath := filepath.Join(logsDir, "audit.log")
    auditLogFile, err := os.OpenFile(auditLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
    if err != nil {
        operationsLogFile.Close()
        return &Error{
            Type:    ErrorInternal,
            Message: "Failed to open audit log file",
            Details: fmt.Sprintf("Failed to open file at %s", auditLogPath),
            Err:     err,
        }
    }
    
    // Initialize operations logger
    operationsLogger := zerolog.New(zerolog.MultiLevelWriter(
        zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
        operationsLogFile,
    )).With().Timestamp().Logger()
    
    // Initialize audit logger
    auditLogger := zerolog.New(auditLogFile).With().Timestamp().Logger()
    
    // Set global loggers
    SetOperationsLogger(operationsLogger)
    SetAuditLogger(auditLogger)
    
    return nil
}
```

### 7.2 Logging Operations

```go
// LogOperation logs an operation
func LogOperation(level zerolog.Level, message string, context ...map[string]interface{}) {
    event := operationsLogger.WithLevel(level)
    
    if len(context) > 0 {
        for k, v := range context[0] {
            event = event.Interface(k, v)
        }
    }
    
    event.Msg(message)
}

// LogAudit logs an audit event
func LogAudit(message string, context ...map[string]interface{}) {
    event := auditLogger.Info()
    
    if len(context) > 0 {
        for k, v := range context[0] {
            event = event.Interface(k, v)
        }
    }
    
    event.Msg(message)
}
```

## 8. Command-Line Interface

### 8.1 Root Command

```go
// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
    Use:   "mctl",
    Short: "Multi-Repository Control System",
    Long: `MCTL (Multi-Repository Control System) is a command-line tool designed to provide
secure, unified management of code repositories in high-security environments.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func init() {
    cobra.OnInitialize(initConfig)
    
    // Global flags
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.mirror/mirror.toml)")
    rootCmd.PersistentFlags().StringVar(&baseDir, "base-dir", ".", "base directory for repositories")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
    if cfgFile != "" {
        // Use config file from the flag.
        viper.SetConfigFile(cfgFile)
    } else {
        // Search for config in the current directory
        viper.AddConfigPath(filepath.Join(baseDir, ".mirror"))
        viper.SetConfigName("mirror")
        viper.SetConfigType("toml")
    }

    // If a config file is found, read it in.
    if err := viper.ReadInConfig(); err == nil {
        fmt.Println("Using config file:", viper.ConfigFileUsed())
    }
}
```

### 8.2 Init Command

```go
// initCmd represents the init command
var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize a new MCTL configuration environment",
    Long: `Initialize a new MCTL configuration environment.
This will create the .mirror directory and configuration files.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // Get flags
        defaultBranch, _ := cmd.Flags().GetString("default-branch")
        parallelOperations, _ := cmd.Flags().GetInt("parallel-operations")
        defaultRemote, _ := cmd.Flags().GetString("default-remote")
        
        // Create config
        config := &Config{
            Global: GlobalConfig{
                DefaultBranch:      defaultBranch,
                ParallelOperations: parallelOperations,
                DefaultRemote:      defaultRemote,
            },
            Repositories: []RepositoryConfig{},
        }
        
        // Create directories
        dirs := []string{
            filepath.Join(baseDir, ".mirror"),
            filepath.Join(baseDir, ".mirror", "metadata"),
            filepath.Join(baseDir, ".mirror", "logs"),
            filepath.Join(baseDir, ".mirror", "snapshots"),
            filepath.Join(baseDir, ".mirror", "cache"),
            filepath.Join(baseDir, ".mirror", "cache", "status"),
        }
        
        for _, dir := range dirs {
            if err := os.MkdirAll(dir, 0700); err != nil {
                return &Error{
                    Type:    ErrorInternal,
                    Message: "Failed to create directory",
                    Details: fmt.Sprintf("Failed to create directory at %s", dir),
                    Err:     err,
                }
            }
        }
        
        // Save config
        if err := SaveConfig(config, baseDir); err != nil {
            return err
        }
        
        // Initialize logger
        if err := InitializeLogger(baseDir); err != nil {
            return err
        }
        
        // Log initialization
        LogOperation(zerolog.InfoLevel, "Initialized MCTL configuration", map[string]interface{}{
            "base_dir": baseDir,
        })
        
        LogAudit("Initialized MCTL configuration", map[string]interface{}{
            "base_dir": baseDir,
        })
        
        fmt.Println("Initialized MCTL configuration in", filepath.Join(baseDir, ".mirror"))
        return nil
    },
}

func init() {
    rootCmd.AddCommand(initCmd)
    
    // Flags
    initCmd.Flags().String("default-branch", "main", "Default branch for new repositories")
    initCmd.Flags().Int("parallel-operations", 4, "Number of parallel operations")
    initCmd.Flags().String("default-remote", "origin", "Default remote name")
}
```

### 8.3 Add Command

```go
// addCmd represents the add command
var addCmd = &cobra.Command{
    Use:   "add [flags] repository-url [path]",
    Short: "Add a Git repository to MCTL management",
    Long: `Add a Git repository to MCTL management.
This will clone the repository and add it to the MCTL configuration.`,
    Args: cobra.RangeArgs(1, 2),
    RunE: func(cmd *cobra.Command, args []string) error {
        // Get repository URL
        url := args[0]
        
        // Get repository path
        var path string
        if len(args) > 1 {
            path = args[1]
        } else {
            pathFlag, _ := cmd.Flags().GetString("path")
            if pathFlag != "" {
                path = pathFlag
            } else {
                // Derive path from URL
                urlParts := strings.Split(url, "/")
                repoName := strings.TrimSuffix(urlParts[len(urlParts)-1], ".git")
                path = filepath.Join("repositories", repoName)
            }
        }
        
        // Get flags
        name, _ := cmd.Flags().GetString("name")
        branch, _ := cmd.Flags().GetString("branch")
        noClone, _ := cmd.Flags().GetBool("no-clone")
        flat, _ := cmd.Flags().GetBool("flat")
        
        // If name is not provided, derive from URL
        if name == "" {
            urlParts := strings.Split(url, "/")
            name = strings.TrimSuffix(urlParts[len(urlParts)-1], ".git")
        }
        
        // If branch is not provided, use default branch
        if branch == "" {
            config, err := LoadConfig(baseDir)
            if err != nil {
                return err
            }
            branch = config.Global.DefaultBranch
        }
        
        // If flat is true, adjust path
        if flat {
            path = filepath.Base(path)
        }
        
        // Load config
        config, err := LoadConfig(baseDir)
        if err != nil {
            return err
        }
        
        // Initialize repository
        repo, err := InitializeRepository(config, name, url, path, branch, baseDir, noClone)
        if err != nil {
            return err
        }
        
        // Log addition
        LogOperation(zerolog.InfoLevel, "Added repository", map[string]interface{}{
            "name": name,
            "url":  url,
            "path": path,
        })
        
        LogAudit("Added repository", map[string]interface{}{
            "name": name,
            "url":  url,
            "path": path,
        })
        
        fmt.Printf("Added repository %s (%s) at %s\n", name, url, path)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(addCmd)
    
    // Flags
    addCmd.Flags().String("path", "", "Target location for repository")
    addCmd.Flags().String("name", "", "Custom repository designation")
    addCmd.Flags().String("branch", "", "Specific branch to clone")
    addCmd.Flags().Bool("no-clone", false, "Add to configuration without cloning")
    addCmd.Flags().Bool("flat", false, "Clone directly to path without creating subdirectory")
}
```

### 8.4 List Command

```go
// listCmd represents the list command
var listCmd = &cobra.Command{
    Use:   "list [flags]",
    Short: "List repositories under MCTL management",
    Long: `List repositories under MCTL management.
This will display information about all repositories in the MCTL configuration.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // Get flags
        detailed, _ := cmd.Flags().GetBool("detailed")
        format, _ := cmd.Flags().GetString("format")
        filter, _ := cmd.Flags().GetString("filter")
        
        // Load config
        config, err := LoadConfig(baseDir)
        if err != nil {
            return err
        }
        
        // Filter repositories
        var repositories []RepositoryConfig
        if filter == "" {
            repositories = config.Repositories
        } else {
            for _, repo := range config.Repositories {
                if strings.Contains(repo.Name, filter) || strings.Contains(repo.Path, filter) {
                    repositories = append(repositories, repo)
                }
            }
        }
        
        // Display repositories
        switch format {
        case "json":
            displayRepositoriesJSON(repositories, detailed, baseDir)
        case "yaml":
            displayRepositoriesYAML(repositories, detailed, baseDir)
        default:
            displayRepositoriesTable(repositories, detailed, baseDir)
        }
        
        return nil
    },
}

func init() {
    rootCmd.AddCommand(listCmd)
    
    // Flags
    listCmd.Flags().Bool("detailed", false, "Show detailed information about each repository")
    listCmd.Flags().String("format", "table", "Output format (table, json, yaml)")
    listCmd.Flags().String("filter", "", "Filter repositories by name, path, or status")
}
```

### 8.5 Remove Command

```go
// removeCmd represents the remove command
var removeCmd = &cobra.Command{
    Use:   "remove [flags] <repository-name>",
    Short: "Remove a repository from MCTL management",
    Long: `Remove a repository from MCTL management.
This will remove the repository from the MCTL configuration.`,
    Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // Get repository name
        name := args[0]
        
        // Get flags
        delete, _ := cmd.Flags().GetBool("delete")
        force, _ := cmd.Flags().GetBool("force")
        
        // Load config
        config, err := LoadConfig(baseDir)
        if err != nil {
            return err
        }
        
        // Remove repository
        if err := RemoveRepository(config, name, baseDir, delete, force); err != nil {
            return err
        }
        
        // Log removal
        LogOperation(zerolog.InfoLevel, "Removed repository", map[string]interface{}{
            "name":   name,
            "delete": delete,
            "force":  force,
        })
        
        LogAudit("Removed repository", map[string]interface{}{
            "name":   name,
            "delete": delete,
            "force":  force,
        })
        
        fmt.Printf("Removed repository %s\n", name)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(removeCmd)
    
    // Flags
    removeCmd.Flags().Bool("delete", false, "Delete the repository directory from the filesystem")
    removeCmd.Flags().Bool("force", false, "Force removal even if there are uncommitted changes")
}
```

### 8.6 Status Command

```go
// statusCmd represents the status command
var statusCmd = &cobra.Command{
    Use:   "status [flags] [repository-names...]",
    Short: "Report status of managed repositories",
    Long: `Report status of managed repositories.
This will display the status of all repositories in the MCTL configuration.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // Get flags
        checkRemote, _ := cmd.Flags().GetBool("check-remote")
        format, _ := cmd.Flags().GetString("format")
        filter, _ := cmd.Flags().GetString("filter")
        
        // Load config
        config, err := LoadConfig(baseDir)
        if err != nil {
            return err
        }
        
        // Get repositories
        var repositories []RepositoryConfig
        if len(args) > 0 {
            // Get specified repositories
            for _, name := range args {
                found := false
                for _, repo := range config.Repositories {
                    if repo.Name == name {
                        repositories = append(repositories, repo)
                        found = true
                        break
                    }
                }
                if !found {
                    return &Error{
                        Type:    ErrorRepositoryNotFound,
                        Message: "Repository not found",
                        Details: fmt.Sprintf("Repository with name %s not found", name),
                    }
                }
            }
        } else {
            // Get all repositories
            repositories = config.Repositories
        }
        
        // Filter repositories by status
        if filter != "" {
            var filteredRepos []RepositoryConfig
            for _, repo := range repositories {
                status, err := GetRepositoryStatus(repo, baseDir)
                if err != nil {
                    status = StatusUnknown
                }
                if strings.EqualFold(string(status), filter) {
                    filteredRepos = append(filteredRepos, repo)
                }
            }
            repositories = filteredRepos
        }
        
        // Display repository status
        switch format {
        case "json":
            displayRepositoryStatusJSON(repositories, checkRemote, baseDir)
        case "yaml":
            displayRepositoryStatusYAML(repositories, checkRemote, baseDir)
        default:
            displayRepositoryStatusTable(repositories, checkRemote, baseDir)
        }
        
        return nil
    },
}

func init() {
    rootCmd.AddCommand(statusCmd)
    
    // Flags
    statusCmd.Flags().Bool("check-remote", false, "Check status relative to remote repositories")
    statusCmd.Flags().String("format", "table", "Output format (table, json, yaml)")
    statusCmd.Flags().String("filter", "", "Filter repositories by status")
}
```

## 9. Implementation Phases and Milestones

### 9.1 Phase 1: Core Infrastructure

**Milestone 1.1: Basic Project Setup**
- Set up project structure
- Configure Go modules
- Add dependencies
- Create basic command-line interface

**Milestone 1.2: Configuration Management**
- Implement configuration loading and saving
- Implement configuration validation
- Implement `init` command

**Milestone 1.3: Repository Management**
- Implement repository initialization
- Implement repository cloning
- Implement repository removal
- Implement repository status checking
- Implement repository metadata management

**Milestone 1.4: Git Operations**
- Implement Git command execution
- Implement repository status determination
- Implement branch management

**Milestone 1.5: Logging System**
- Implement logger initialization
- Implement operation logging
- Implement audit logging

### 9.2 Phase 2: Core Commands

**Milestone 2.1: Init Command**
- Implement `init` command
- Add command-line flags
- Add validation and error handling

**Milestone 2.2: Add Command**
- Implement `add` command
- Add command-line flags
- Add validation and error handling

**Milestone 2.3: List Command**
- Implement `list` command
- Add command-line flags
- Implement output formatting

**Milestone 2.4: Remove Command**
- Implement `remove` command
- Add command-line flags
- Add validation and error handling

**Milestone 2.5: Status Command**
- Implement `status` command
- Add command-line flags
- Implement output formatting

### 9.3 Phase 3: Testing and Documentation

**Milestone 3.1: Unit Testing**
- Write unit tests for core components
- Write unit tests for commands
- Achieve at least 80% code coverage

**Milestone 3.2: Integration Testing**
- Write integration tests for core workflows
- Test with real Git repositories
- Test error handling and edge cases

**Milestone 3.3: Documentation**
- Write code documentation
- Write user documentation
- Write developer documentation

**Milestone 3.4: Performance Testing**
- Test with large repositories
- Test with many repositories
- Identify and address performance bottlenecks

## 10. Conclusion

The Stage 1 implementation of the MCTL core engine will provide the foundation for all MCTL operations, including repository management, state tracking, and metadata handling. By implementing the core data structures, configuration management, repository management, Git operations, logging system, and command-line interface, Stage 1 will enable basic MCTL functionality.

The implementation will follow a structured approach, with clear separation of concerns and well-defined interfaces between components. This will enable future stages of development to build on the foundation established in Stage 1, adding more advanced features like the snapshot system, advanced Git operations, parallel operations, and extension mechanisms.

The implementation will prioritize security, reliability, and performance, with comprehensive error handling, logging, and testing. This will ensure that MCTL is a robust and secure tool for managing multiple Git repositories in high-security environments.
