package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [subcommand]",
		Short: "Manage MCTL configuration",
		Long: `Manage MCTL configuration.

This command provides subcommands for managing MCTL configuration.
If no subcommand is provided, it displays the current configuration.

Examples:
  mctl config
  mctl config get global.default_branch
  mctl config set global.default_branch main
  mctl config validate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, default to show
			if len(args) == 0 {
				return runConfigShow()
			}
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigValidateCmd())

	return cmd
}

func newConfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value",
		Long: `Get a configuration value.

This command retrieves a value from the MCTL configuration.
The key should be in the format "section.key", for example "global.default_branch".

Examples:
  mctl config get global.default_branch
  mctl config get global.parallel_operations`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			return runConfigGet(key)
		},
	}

	return cmd
}

func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Long: `Set a configuration value.

This command sets a value in the MCTL configuration.
The key should be in the format "section.key", for example "global.default_branch".

Examples:
  mctl config set global.default_branch main
  mctl config set global.parallel_operations 8`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]
			return runConfigSet(key, value)
		},
	}

	return cmd
}

func newConfigValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the configuration",
		Long: `Validate the configuration.

This command validates the MCTL configuration and reports any errors.

Examples:
  mctl config validate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigValidate()
		},
	}

	return cmd
}

func runConfigShow() error {
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

	// Display global configuration
	fmt.Println("Global Configuration:")
	fmt.Printf("  default_branch: %s\n", cfg.Global.DefaultBranch)
	fmt.Printf("  parallel_operations: %d\n", cfg.Global.ParallelOperations)
	fmt.Printf("  default_remote: %s\n", cfg.Global.DefaultRemote)

	// Display repositories
	fmt.Printf("\nRepositories (%d):\n", len(cfg.Repositories))
	for _, repo := range cfg.Repositories {
		fmt.Printf("  %s (%s):\n", repo.Name, repo.ID)
		fmt.Printf("    path: %s\n", repo.Path)
		fmt.Printf("    url: %s\n", repo.URL)
		fmt.Printf("    branch: %s\n", repo.Branch)
	}

	return nil
}

func runConfigGet(key string) error {
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

	// Parse key
	section, key, err := parseConfigKey(key)
	if err != nil {
		return err
	}

	// Get value
	value, err := getConfigValue(cfg, section, key)
	if err != nil {
		return err
	}

	// Display value
	fmt.Println(value)

	return nil
}

func runConfigSet(key, value string) error {
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

	// Parse key
	section, key, err := parseConfigKey(key)
	if err != nil {
		return err
	}

	// Set value
	if err := setConfigValue(cfg, section, key, value); err != nil {
		return err
	}

	// Save configuration
	if err := config.SaveConfig(cfg, currentDir); err != nil {
		return errors.Wrap(err, errors.ErrInvalidConfig, "Failed to save configuration")
	}

	fmt.Printf("Set %s.%s to %s\n", section, key, value)

	return nil
}

func runConfigValidate() error {
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

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return err
	}

	fmt.Println("Configuration is valid")

	return nil
}

func parseConfigKey(key string) (string, string, error) {
	// Split key into section and key
	parts := []string{}
	for _, part := range key {
		if part == '.' {
			break
		}
		parts = append(parts, string(part))
	}

	if len(parts) != 2 {
		return "", "", errors.New(errors.ErrInvalidArgument, "Invalid key format, expected 'section.key'")
	}

	return parts[0], parts[1], nil
}

func getConfigValue(cfg *config.Config, section, key string) (string, error) {
	switch section {
	case "global":
		switch key {
		case "default_branch":
			return cfg.Global.DefaultBranch, nil
		case "parallel_operations":
			return strconv.Itoa(cfg.Global.ParallelOperations), nil
		case "default_remote":
			return cfg.Global.DefaultRemote, nil
		default:
			return "", errors.New(errors.ErrInvalidArgument, fmt.Sprintf("Unknown key: %s.%s", section, key))
		}
	default:
		return "", errors.New(errors.ErrInvalidArgument, fmt.Sprintf("Unknown section: %s", section))
	}
}

func setConfigValue(cfg *config.Config, section, key, value string) error {
	switch section {
	case "global":
		switch key {
		case "default_branch":
			cfg.Global.DefaultBranch = value
			return nil
		case "parallel_operations":
			parallelOps, err := strconv.Atoi(value)
			if err != nil {
				return errors.New(errors.ErrInvalidArgument, "Invalid value for parallel_operations, expected an integer")
			}
			cfg.Global.ParallelOperations = parallelOps
			return nil
		case "default_remote":
			cfg.Global.DefaultRemote = value
			return nil
		default:
			return errors.New(errors.ErrInvalidArgument, fmt.Sprintf("Unknown key: %s.%s", section, key))
		}
	default:
		return errors.New(errors.ErrInvalidArgument, fmt.Sprintf("Unknown section: %s", section))
	}
}

func validateConfig(cfg *config.Config) error {
	// Validate global configuration
	if cfg.Global.DefaultBranch == "" {
		return errors.New(errors.ErrInvalidConfig, "Default branch is not set")
	}
	if cfg.Global.ParallelOperations <= 0 {
		return errors.New(errors.ErrInvalidConfig, "Parallel operations must be greater than 0")
	}
	if cfg.Global.DefaultRemote == "" {
		return errors.New(errors.ErrInvalidConfig, "Default remote is not set")
	}

	// Validate repositories
	for _, repo := range cfg.Repositories {
		if repo.ID == "" {
			return errors.New(errors.ErrInvalidConfig, fmt.Sprintf("Repository %s has no ID", repo.Name))
		}
		if repo.Name == "" {
			return errors.New(errors.ErrInvalidConfig, fmt.Sprintf("Repository %s has no name", repo.ID))
		}
		if repo.Path == "" {
			return errors.New(errors.ErrInvalidConfig, fmt.Sprintf("Repository %s has no path", repo.Name))
		}
		if repo.URL == "" {
			return errors.New(errors.ErrInvalidConfig, fmt.Sprintf("Repository %s has no URL", repo.Name))
		}
	}

	return nil
}
