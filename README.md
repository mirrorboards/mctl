# mirrorboards-mctl

`mctl` is a command-line tool for Git repositories mesh management. It helps you manage multiple related Git repositories by treating them as a unified mesh.

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

Add a Git repository to the mirror.toml configuration and clone it to the specified path.

```
mctl add https://github.com/example/repo.git [path]
```

**Options:**
- `--path`, `-p` string: Path where to clone the repository (default ".")
- `--name`, `-n` string: Custom name for the repository (defaults to repo name)
- `--flat`: Clone directly into the path instead of creating a subdirectory

### `sync`

Clone all repositories defined in mirror.toml that haven't been cloned yet.

```
mctl sync
```

### `status`

Run git status on all repositories defined in mirror.toml and display the results in a colorful, elegant way.

```
mctl status
```

### `save [commit-message]`

Add, commit, and push changes to repositories. 

```
mctl save "Your commit message"
```

**Options:**
- `--all`, `-a`: Save all repositories even if they have no changes

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

## Example Workflow

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

## License

[LICENSE]
