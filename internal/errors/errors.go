package errors

import (
	"fmt"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeAPI        ErrorType = "api"
	ErrorTypeDatabase   ErrorType = "database"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeNetwork    ErrorType = "network"
)

// AppError represents an application error with context
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
	}
}

// Wrap wraps an existing error with context
func Wrap(err error, errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// IsType checks if error is of specific type
func IsType(err error, errType ErrorType) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Type == errType
}
