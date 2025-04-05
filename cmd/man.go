package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/spf13/cobra"
)

func newManCmd() *cobra.Command {
	var (
		directory string
	)

	cmd := &cobra.Command{
		Use:   "man",
		Short: "Generate man pages for MCTL",
		Long: `Generate man pages for MCTL.

This command generates man pages for MCTL commands.
By default, it generates man pages in the current directory,
but you can specify a different directory with the --directory flag.

Examples:
  mctl man
  mctl man --directory=/usr/share/man/man1`,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMan(cmd, directory)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&directory, "directory", ".", "Directory to write man pages to")

	return cmd
}

func runMan(cmd *cobra.Command, directory string) error {
	// Get root command
	rootCmd := cmd.Root()

	// Create manuals directory
	if err := os.MkdirAll(directory, 0755); err != nil {
		return errors.Wrap(err, errors.ErrPermissionDenied, fmt.Sprintf("Failed to create directory: %s", directory))
	}

	// Generate man pages
	fmt.Println("Generating man pages...")

	// Create a simple man page for the root command
	manFilePath := filepath.Join(directory, "mctl.1")
	f, err := os.Create(manFilePath)
	if err != nil {
		return errors.Wrap(err, errors.ErrPermissionDenied, fmt.Sprintf("Failed to create file: %s", manFilePath))
	}
	defer f.Close()

	// Write a simple man page
	_, err = f.WriteString(fmt.Sprintf(".TH MCTL 1 \"%s\" \"MCTL Manual\"\n", rootCmd.Version))
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to write man page")
	}
	_, err = f.WriteString(".SH NAME\nmctl \\- Multi-Repository Control System\n")
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to write man page")
	}
	_, err = f.WriteString(".SH SYNOPSIS\n.B mctl\n[command] [options]\n")
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to write man page")
	}
	_, err = f.WriteString(".SH DESCRIPTION\n")
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to write man page")
	}
	_, err = f.WriteString(rootCmd.Long)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to write man page")
	}
	_, err = f.WriteString("\n.SH COMMANDS\n")
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to write man page")
	}

	// Add commands
	for _, subCmd := range rootCmd.Commands() {
		if subCmd.Hidden {
			continue
		}
		_, err = f.WriteString(fmt.Sprintf(".TP\n.B %s\n%s\n", subCmd.Name(), subCmd.Short))
		if err != nil {
			return errors.Wrap(err, errors.ErrInternalError, "Failed to write man page")
		}
	}

	fmt.Printf("Generated man page at %s\n", manFilePath)

	// Generate man pages for subcommands
	for _, subCmd := range rootCmd.Commands() {
		if subCmd.Hidden {
			continue
		}

		// Create a simple man page for the subcommand
		manFilePath := filepath.Join(directory, fmt.Sprintf("mctl-%s.1", subCmd.Name()))
		f, err := os.Create(manFilePath)
		if err != nil {
			fmt.Printf("Warning: Failed to create file for %s: %v\n", subCmd.Name(), err)
			continue
		}
		defer f.Close()

		// Write a simple man page
		_, err = f.WriteString(fmt.Sprintf(".TH \"MCTL-%s\" 1 \"%s\" \"MCTL Manual\"\n", subCmd.Name(), rootCmd.Version))
		if err != nil {
			fmt.Printf("Warning: Failed to write man page for %s: %v\n", subCmd.Name(), err)
			continue
		}
		_, err = f.WriteString(fmt.Sprintf(".SH NAME\nmctl-%s \\- %s\n", subCmd.Name(), subCmd.Short))
		if err != nil {
			fmt.Printf("Warning: Failed to write man page for %s: %v\n", subCmd.Name(), err)
			continue
		}
		_, err = f.WriteString(fmt.Sprintf(".SH SYNOPSIS\n.B mctl %s\n[options]\n", subCmd.Name()))
		if err != nil {
			fmt.Printf("Warning: Failed to write man page for %s: %v\n", subCmd.Name(), err)
			continue
		}
		_, err = f.WriteString(".SH DESCRIPTION\n")
		if err != nil {
			fmt.Printf("Warning: Failed to write man page for %s: %v\n", subCmd.Name(), err)
			continue
		}
		_, err = f.WriteString(subCmd.Long)
		if err != nil {
			fmt.Printf("Warning: Failed to write man page for %s: %v\n", subCmd.Name(), err)
			continue
		}

		fmt.Printf("Generated man page at %s\n", manFilePath)
	}

	return nil
}
