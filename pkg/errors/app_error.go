package errors

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ===== AppError Codes =====

const (
	CodeInvalidInput = "invalid_input"
	CodeNotFound     = "not_found"
	CodeConflict     = "conflict"
	CodeUnauthorized = "unauthorized"
	CodeInternal     = "internal_error"
	CodeBadRequest   = "not allowed"
)

// ===== Check if error is AppError =====

func IsAppError(err error) bool {
	if err == nil {
		return false
	}
	var appErr *AppError
	return errors.As(err, &appErr)
}
