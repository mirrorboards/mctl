package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/logging"
	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	var (
		logType string
		limit   int
	)

	cmd := &cobra.Command{
		Use:   "logs [options]",
		Short: "Display MCTL logs",
		Long: `Display MCTL logs.

This command displays logs from MCTL operations. By default, it shows
operation logs, but you can specify the type of logs to display with
the --type flag.

Examples:
  mctl logs
  mctl logs --type=operations
  mctl logs --type=audit
  mctl logs --limit=50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogs(logType, limit)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&logType, "type", "operations", "Type of logs to display (operations, audit)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of log entries to display")

	return cmd
}

func runLogs(logType string, limit int) error {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to get current directory")
	}

	// Validate log type
	var logTypeEnum logging.LogType
	switch strings.ToLower(logType) {
	case "operations", "operation", "op":
		logTypeEnum = logging.LogTypeOperation
	case "audit", "security":
		logTypeEnum = logging.LogTypeAudit
	default:
		return errors.New(errors.ErrInvalidArgument, fmt.Sprintf("Invalid log type: %s", logType))
	}

	// Create logger
	logger := logging.NewLogger(currentDir)

	// Get logs
	logs, err := logger.GetLogs(logTypeEnum, limit)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to get logs")
	}

	// Display logs
	if len(logs) == 0 {
		fmt.Printf("No %s logs found\n", logType)
		return nil
	}

	fmt.Printf("%s Logs (showing %d entries):\n\n", strings.Title(logType), len(logs))
	for _, log := range logs {
		fmt.Println(log)
	}

	return nil
}
