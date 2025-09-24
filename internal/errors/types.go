package errors

import (
	"fmt"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeNetwork    ErrorType = "network"
	ErrorTypeAuth       ErrorType = "auth"
	ErrorTypeConfig     ErrorType = "config"
	ErrorTypeInternal   ErrorType = "internal"
	ErrorTypeOpenAPI    ErrorType = "openapi"
	ErrorTypeMCP        ErrorType = "mcp"
)

// QUrlError represents a structured error with context
type QUrlError struct {
	Type    ErrorType
	Message string
	Context map[string]interface{}
	Cause   error
}

// Error implements the error interface
func (e *QUrlError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Cause.Error())
	}
	return e.Message
}

// Unwrap returns the underlying error for error unwrapping
func (e *QUrlError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches a specific type
func (e *QUrlError) Is(target error) bool {
	if targetErr, ok := target.(*QUrlError); ok {
		return e.Type == targetErr.Type
	}
	return false
}

// WithContext adds context information to the error
func (e *QUrlError) WithContext(key string, value interface{}) *QUrlError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// New creates a new QUrlError
func New(errType ErrorType, message string) *QUrlError {
	return &QUrlError{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errType ErrorType, message string) *QUrlError {
	return &QUrlError{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
		Cause:   err,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, errType ErrorType, format string, args ...interface{}) *QUrlError {
	return Wrap(err, errType, fmt.Sprintf(format, args...))
}

// Newf creates a new QUrlError with formatted message
func Newf(errType ErrorType, format string, args ...interface{}) *QUrlError {
	return New(errType, fmt.Sprintf(format, args...))
}

// IsType checks if an error is of a specific type
func IsType(err error, errType ErrorType) bool {
	if qErr, ok := err.(*QUrlError); ok {
		return qErr.Type == errType
	}
	return false
}

// GetType returns the error type, or ErrorTypeInternal if not a QUrlError
func GetType(err error) ErrorType {
	if qErr, ok := err.(*QUrlError); ok {
		return qErr.Type
	}
	return ErrorTypeInternal
}

// GetContext returns context information from the error
func GetContext(err error) map[string]interface{} {
	if qErr, ok := err.(*QUrlError); ok {
		return qErr.Context
	}
	return nil
}