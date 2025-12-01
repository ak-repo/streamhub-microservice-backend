package errors

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


// ============================================================================
// 1) AppError (Core Error Type)
// ============================================================================

type AppError struct {
	Code    string // machine-readable error code
	Message string // human-readable message
	Err     error  // optional underlying error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
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


// ============================================================================
// 2) Error Codes (ALL codes supported)
// ============================================================================

const (
	CodeInvalidInput  = "invalid_input"
	CodeNotFound      = "not_found"
	CodeConflict      = "conflict"
	CodeUnauthorized  = "unauthorized"
	CodeForbidden     = "forbidden"
	CodeInternal      = "internal_error"
	CodeBadRequest    = "bad_request"
	CodeAlreadyExists = "already_exists"
)


// ============================================================================
// 3) Error Type Checkers
// ============================================================================

func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

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


// ============================================================================
// 4) Convert AppError → gRPC Error
// ============================================================================

func ToGRPC(err error) error {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {

		case CodeInvalidInput, CodeBadRequest:
			return status.Error(codes.InvalidArgument, appErr.Message)

		case CodeNotFound:
			return status.Error(codes.NotFound, appErr.Message)

		case CodeUnauthorized:
			return status.Error(codes.Unauthenticated, appErr.Message)

		case CodeForbidden:
			return status.Error(codes.PermissionDenied, appErr.Message)

		case CodeConflict, CodeAlreadyExists:
			return status.Error(codes.AlreadyExists, appErr.Message)

		case CodeInternal:
			return status.Error(codes.Internal, appErr.Message)

		default:
			return status.Error(codes.Internal, "internal server error")
		}
	}

	// Non-AppError fallback
	return status.Error(codes.Internal, "internal server error")
}


// ============================================================================
// 5) Convert gRPC Error → Fiber HTTP Response
// ============================================================================

func GRPCToFiber(err error) (int, fiber.Map) {
	if err == nil {
		return fiber.StatusOK, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return fiber.StatusInternalServerError, fiber.Map{
			"error": "internal server error",
		}
	}

	switch st.Code() {

	case codes.InvalidArgument:
		return fiber.StatusBadRequest, fiber.Map{"error": st.Message()}

	case codes.NotFound:
		return fiber.StatusNotFound, fiber.Map{"error": st.Message()}

	case codes.AlreadyExists:
		return fiber.StatusConflict, fiber.Map{"error": st.Message()}

	case codes.Unauthenticated:
		return fiber.StatusUnauthorized, fiber.Map{"error": st.Message()}

	case codes.PermissionDenied:
		return fiber.StatusForbidden, fiber.Map{"error": st.Message()}

	case codes.Internal:
		return fiber.StatusInternalServerError, fiber.Map{"error": st.Message()}

	case codes.Unavailable:
		return fiber.StatusServiceUnavailable, fiber.Map{"error": "service unavailable"}

	default:
		return fiber.StatusInternalServerError, fiber.Map{"error": st.Message()}
	}
}
