package errors

import (
	"fmt"
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	// ErrorCategoryConfiguration represents configuration errors
	ErrorCategoryConfiguration ErrorCategory = "E1"
	// ErrorCategoryRepository represents repository errors
	ErrorCategoryRepository ErrorCategory = "E2"
	// ErrorCategoryGitOperation represents Git operation errors
	ErrorCategoryGitOperation ErrorCategory = "E3"
	// ErrorCategoryFilesystem represents filesystem errors
	ErrorCategoryFilesystem ErrorCategory = "E4"
	// ErrorCategoryUserInput represents user input errors
	ErrorCategoryUserInput ErrorCategory = "E5"
	// ErrorCategoryInternal represents internal errors
	ErrorCategoryInternal ErrorCategory = "E9"
)

// ErrorCode represents a specific error code
type ErrorCode struct {
	Category ErrorCategory
	Code     int
	Message  string
}

// String returns the string representation of the error code
func (e ErrorCode) String() string {
	return fmt.Sprintf("%s%03d", e.Category, e.Code)
}

// Error represents an MCTL error
type Error struct {
	Code    ErrorCode
	Message string
	Details []string
	Err     error
}

// Error returns the error message
func (e *Error) Error() string {
	return fmt.Sprintf("ERROR [%s] %s: %s", e.Code, e.Code.Message, e.Message)
}

// WithDetails adds details to the error
func (e *Error) WithDetails(details ...string) *Error {
	e.Details = append(e.Details, details...)
	return e
}

// WithError adds an underlying error to the error
func (e *Error) WithError(err error) *Error {
	e.Err = err
	return e
}

// Format returns a formatted error message
func (e *Error) Format() string {
	result := fmt.Sprintf("ERROR [%s] %s:\n- %s", e.Code, e.Code.Message, e.Message)

	for _, detail := range e.Details {
		result += fmt.Sprintf("\n- %s", detail)
	}

	return result
}

// New creates a new error
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error
func Wrap(err error, code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Common error codes
var (
	// Configuration errors (E1xxx)
	ErrConfigNotFound = ErrorCode{ErrorCategoryConfiguration, 1, "Configuration not found"}
	ErrInvalidConfig  = ErrorCode{ErrorCategoryConfiguration, 2, "Invalid configuration"}

	// Repository errors (E2xxx)
	ErrRepositoryNotFound = ErrorCode{ErrorCategoryRepository, 1, "Repository not found"}
	ErrRepositoryExists   = ErrorCode{ErrorCategoryRepository, 2, "Repository already exists"}
	ErrCloneFailed        = ErrorCode{ErrorCategoryRepository, 3, "Repository clone failed"}

	// Git operation errors (E3xxx)
	ErrGitPushFailed   = ErrorCode{ErrorCategoryGitOperation, 1, "Git push operation failed"}
	ErrGitPullFailed   = ErrorCode{ErrorCategoryGitOperation, 2, "Git pull operation failed"}
	ErrGitFetchFailed  = ErrorCode{ErrorCategoryGitOperation, 3, "Git fetch operation failed"}
	ErrGitCommitFailed = ErrorCode{ErrorCategoryGitOperation, 4, "Git commit operation failed"}
	ErrGitBranchFailed = ErrorCode{ErrorCategoryGitOperation, 5, "Git branch operation failed"}

	// Filesystem errors (E4xxx)
	ErrPermissionDenied = ErrorCode{ErrorCategoryFilesystem, 1, "Permission denied"}
	ErrDiskFull         = ErrorCode{ErrorCategoryFilesystem, 2, "Disk full"}
	ErrFileNotFound     = ErrorCode{ErrorCategoryFilesystem, 3, "File not found"}

	// User input errors (E5xxx)
	ErrInvalidArgument     = ErrorCode{ErrorCategoryUserInput, 1, "Invalid argument"}
	ErrInvalidRepositoryID = ErrorCode{ErrorCategoryUserInput, 2, "Invalid repository identifier"}
	ErrMissingArgument     = ErrorCode{ErrorCategoryUserInput, 3, "Missing required argument"}

	// Internal errors (E9xxx)
	ErrInternalError = ErrorCode{ErrorCategoryInternal, 1, "Internal error"}
	ErrUnexpected    = ErrorCode{ErrorCategoryInternal, 2, "Unexpected error"}
)
