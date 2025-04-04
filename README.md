# mirrorboards-mctl

`mctl` is a command-line tool for Git repositories mesh management. It helps you manage multiple related Git repositories by treating them as a unified mesh. Each repository is assigned a unique ID for easy reference and management.

## Installation

```
go install github.com/mirrorboards/mctl@latest
```

## Usage

```
mctl [command]
```

## Available Commands

### `init`

Initialize an empty mirror.toml file in the current directory.

```
mctl init
```

### `add [git-url] [path]`

Add a Git repository to the mirror.toml configuration and clone it to the specified path. Each repository is assigned a unique ID.

```
mctl add https://github.com/example/repo.git [path]
```

**Options:**
- `--path`, `-p` string: Path where to clone the repository (default ".")
- `--name`, `-n` string: Custom name for the repository (defaults to repo name)
- `--flat`: Clone directly into the path instead of creating a subdirectory

### `remove [id|name]`

Remove a repository from the mirror.toml configuration by ID or name.

```
mctl remove repo-id
```

**Options:**
- `--delete`: Delete repository files in addition to removing from configuration

### `sync`

Clone all repositories defined in mirror.toml that haven't been cloned yet. Optionally sync with a remote configuration.

```
mctl sync
```

**Options:**
- `--remote` string: Remote configuration to sync with
- `--merge-strategy` string: Merge strategy for remote sync (remote-wins, local-wins, union) (default "union")
- `--repos` strings: Only sync specified repositories (by ID or name)

### `status`

Run git status on all repositories defined in mirror.toml and display the results in a colorful, elegant way. Shows repository IDs for easy reference.

```
mctl status
```

**Options:**
- `--format` string: Output format (text, json) (default "text")
- `--repos` strings: Only show status for specified repositories (by ID or name)

### `branch [branch-name]`

Switch all repositories to a specific branch. Optionally create the branch if it doesn't exist and pull latest changes.

```
mctl branch main
```

**Options:**
- `--create`: Create the branch if it doesn't exist
- `--pull`: Pull latest changes after switching branch
- `--repos` strings: Only switch specified repositories (by ID or name)

### `save [commit-message]`

Add, commit, and push changes to repositories. 

```
mctl save "Your commit message"
```

**Options:**
- `--all`, `-a`: Save all repositories even if they have no changes

### `remote`

Manage remote configuration sources for synchronizing mirror.toml files across repositories.

```
mctl remote [command]
```

**Subcommands:**
- `add [name] [url]`: Add a remote configuration source
- `list`: List remote configuration sources
- `remove [name]`: Remove a remote configuration source
- `pull [name]`: Pull and merge configuration from a remote source
- `push [name]`: Push local configuration to a remote source

### `clear`

Remove all directories created by mctl based on mirror.toml, but keep the configuration.

```
mctl clear
```

### `version`

Display the current version of mctl.

```
mctl version
```

## Configuration

The `mirror.toml` file contains the configuration for your repository mesh. It is created automatically when you run `mctl init`.

Each repository in the configuration has:
- A unique ID for easy reference
- URL, path, and optional name
- Optional branch and tags for organization

Remote configuration sources can be defined to synchronize configurations across different machines or teams.

## Example Workflows

### Basic Workflow

1. Initialize a new configuration:
   ```
   mctl init
   ```

2. Add repositories to your mesh:
   ```
   mctl add https://github.com/org/repo1.git ./packages/repo1
   mctl add https://github.com/org/repo2.git ./packages/repo2
   ```

3. Check the status of all repositories:
   ```
   mctl status
   ```

4. Make changes across repositories and save them with a single command:
   ```
   mctl save "Update all repositories with new feature"
   ```

5. Later, sync all repositories to ensure they're up to date:
   ```
   mctl sync
   ```

6. If needed, clear all repositories while keeping the configuration:
   ```
   mctl clear
   ```

### Remote Configuration Workflow

1. Add a remote configuration source:
   ```
   mctl remote add github https://raw.githubusercontent.com/mirrorboards/mirrorboards/refs/heads/main/mirror.toml
   ```

2. Sync with the remote configuration:
   ```
   mctl sync --remote github
   ```

3. Pull updates from the remote configuration:
   ```
   mctl remote pull github
   ```

4. Push your local configuration to a remote repository:
   ```
   mctl remote push github --message "Update configuration"
   ```

### Branch Management Workflow

1. Switch all repositories to a specific branch:
   ```
   mctl branch feature-branch
   ```

2. Create a new branch in all repositories:
   ```
   mctl branch new-feature --create
   ```

3. Switch to main branch and pull latest changes:
   ```
   mctl branch main --pull
   ```

4. Switch branch only for specific repositories:
   ```
   mctl branch hotfix --repos repo1-id repo2-id
   ```

## License

[LICENSE]
