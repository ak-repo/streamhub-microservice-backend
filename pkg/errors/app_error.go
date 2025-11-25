package errors

import (
	"errors"
	"fmt"
)

// AppError represents an application-level error with a code and optional underlying error.
type AppError struct {
	Code    string // machine-readable error code
	Message string // human-readable message
	Err     error  // optional underlying error
}

// Error returns a formatted string of the AppError.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap allows errors.Unwrap to retrieve the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError wrapping an optional underlying error.
func New(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ===== AppError Codes =====
const (
	CodeInvalidInput  = "invalid_input"
	CodeNotFound      = "not_found"
	CodeConflict      = "conflict"
	CodeUnauthorized  = "unauthorized"
	CodeInternal      = "internal_error"
	CodeBadRequest    = "bad_request"
	CodeAlreadyExists = "already_exists"
	CodeForbidden     = "forbidden"
)

// ===== Helpers to check error codes =====

// IsAppError returns true if the error is an AppError.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// Specific code checks

func IsNotFound(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == CodeNotFound
}

func IsForbidden(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == CodeForbidden
}

func IsConflict(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == CodeConflict
}

func IsUnauthorized(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == CodeUnauthorized
}

func IsInvalidInput(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == CodeInvalidInput
}

func IsInternal(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == CodeInternal
}
