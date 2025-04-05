package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mirrorboards/mctl/internal/config"
)

// LogType represents the type of log
type LogType string

const (
	// LogTypeOperation represents an operation log
	LogTypeOperation LogType = "OPERATION"
	// LogTypeAudit represents an audit log
	LogTypeAudit LogType = "AUDIT"
)

// LogLevel represents the level of a log entry
type LogLevel string

const (
	// LogLevelInfo represents an informational log entry
	LogLevelInfo LogLevel = "INFO"
	// LogLevelWarning represents a warning log entry
	LogLevelWarning LogLevel = "WARNING"
	// LogLevelError represents an error log entry
	LogLevelError LogLevel = "ERROR"
)

// Logger handles logging operations
type Logger struct {
	BaseDir string
}

// NewLogger creates a new logger
func NewLogger(baseDir string) *Logger {
	return &Logger{
		BaseDir: baseDir,
	}
}

// ensureLogDirectoryExists ensures the log directory exists
func (l *Logger) ensureLogDirectoryExists() error {
	logDir := config.GetLogsDirPath(l.BaseDir)
	return os.MkdirAll(logDir, 0700)
}

// getLogFilePath returns the path to the specified log file
func (l *Logger) getLogFilePath(logType LogType) string {
	var filename string
	switch logType {
	case LogTypeOperation:
		filename = config.DefaultOperationsLogFile
	case LogTypeAudit:
		filename = config.DefaultAuditLogFile
	default:
		filename = "unknown.log"
	}
	return filepath.Join(config.GetLogsDirPath(l.BaseDir), filename)
}

// Log logs a message to the specified log file
func (l *Logger) Log(logType LogType, level LogLevel, message string) error {
	// Ensure log directory exists
	if err := l.ensureLogDirectoryExists(); err != nil {
		return fmt.Errorf("error ensuring log directory exists: %w", err)
	}

	// Format log entry
	timestamp := time.Now().Format(time.RFC3339)
	logEntry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	// Open log file in append mode
	logFile, err := os.OpenFile(
		l.getLogFilePath(logType),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0600,
	)
	if err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}
	defer logFile.Close()

	// Write log entry
	if _, err := logFile.WriteString(logEntry); err != nil {
		return fmt.Errorf("error writing to log file: %w", err)
	}

	return nil
}

// LogOperation logs an operation
func (l *Logger) LogOperation(level LogLevel, message string) error {
	return l.Log(LogTypeOperation, level, message)
}

// LogAudit logs an audit event
func (l *Logger) LogAudit(level LogLevel, message string) error {
	return l.Log(LogTypeAudit, level, message)
}

// GetLogs retrieves logs from the specified log file
func (l *Logger) GetLogs(logType LogType, limit int) ([]string, error) {
	// Check if log file exists
	logPath := l.getLogFilePath(logType)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Read log file
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	// Split into lines
	lines := []string{}
	if len(data) > 0 {
		lines = filepath.SplitList(string(data))
	}

	// Apply limit if specified
	if limit > 0 && len(lines) > limit {
		// Return the most recent logs (from the end of the slice)
		return lines[len(lines)-limit:], nil
	}

	return lines, nil
}
