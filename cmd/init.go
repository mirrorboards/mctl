package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/logging"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var (
		directory string
		template  string
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new MCTL configuration environment",
		Long: `Initialize a new MCTL configuration environment.

This command creates the necessary directory structure and configuration files
for MCTL to manage Git repositories. By default, it initializes in the current
directory, but you can specify a different directory with the --directory flag.

You can also specify a template for the initial configuration with the --template
flag. Available templates are:
- standard: Default configuration with common settings
- secure: Configuration with enhanced security settings
- minimal: Minimal configuration with basic settings

Examples:
  mctl init
  mctl init --directory=/secure/projects
  mctl init --template=secure
  mctl init --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(directory, template, force)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&directory, "directory", ".", "Target directory for initialization")
	cmd.Flags().StringVar(&template, "template", "standard", "Configuration template (standard, secure, minimal)")
	cmd.Flags().BoolVar(&force, "force", false, "Override existing configuration")

	return cmd
}

func runInit(directory string, template string, force bool) error {
	// Resolve absolute path
	absDir, err := filepath.Abs(directory)
	if err != nil {
		return errors.Wrap(err, errors.ErrInvalidArgument, fmt.Sprintf("Invalid directory: %s", directory))
	}

	// Check if already initialized
	if config.IsInitialized(absDir) && !force {
		err := errors.New(errors.ErrInvalidConfig, "Directory already contains MCTL configuration")
		err = err.WithDetails("Use --force to override existing configuration")
		return err
	}

	// Validate template
	if template != "standard" && template != "secure" && template != "minimal" {
		err := errors.New(errors.ErrInvalidArgument, "Invalid template specification")
		err = err.WithDetails("Available templates: standard, secure, minimal")
		return err
	}

	// Create configuration
	cfg := createConfigFromTemplate(template)

	// Create directory structure
	if err := createDirectoryStructure(absDir); err != nil {
		return err
	}

	// Save configuration
	if err := config.SaveConfig(cfg, absDir); err != nil {
		return errors.Wrap(err, errors.ErrInvalidConfig, "Failed to save configuration")
	}

	// Initialize logging
	logger := logging.NewLogger(absDir)
	if err := logger.LogOperation(logging.LogLevelInfo, "Initialized MCTL configuration"); err != nil {
		// Non-fatal error, just print a warning
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize logging: %v\n", err)
	}

	fmt.Printf("Initialized MCTL configuration in %s\n", absDir)
	return nil
}

func createConfigFromTemplate(template string) *config.Config {
	cfg := config.DefaultConfig()

	switch template {
	case "secure":
		// Enhanced security settings
		cfg.Global.DefaultBranch = "main"
		cfg.Global.ParallelOperations = 2
	case "minimal":
		// Minimal settings
		cfg.Global.ParallelOperations = 1
	}

	return cfg
}

func createDirectoryStructure(baseDir string) error {
	// Create directories with appropriate permissions
	dirs := []string{
		config.GetConfigDirPath(baseDir),
		config.GetMetadataDirPath(baseDir),
		config.GetLogsDirPath(baseDir),
		config.GetCacheDirPath(baseDir),
		config.GetStatusCacheDirPath(baseDir),
	}

	for _, dir := range dirs {
		// Set secure permissions for .mirror directory and subdirectories
		var perm os.FileMode = 0755
		if filepath.Base(dir) == config.DefaultConfigDir ||
			filepath.Dir(dir) == config.GetConfigDirPath(baseDir) {
			perm = 0700
		}

		if err := os.MkdirAll(dir, perm); err != nil {
			return errors.Wrap(err, errors.ErrPermissionDenied, fmt.Sprintf("Failed to create directory: %s", dir))
		}
	}

	return nil
}
