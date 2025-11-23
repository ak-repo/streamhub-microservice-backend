package errors

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGRPC(err error) error {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case CodeInvalidInput:
			return status.Error(codes.InvalidArgument, appErr.Message)
		case CodeNotFound:
			return status.Error(codes.NotFound, appErr.Message)
		case CodeUnauthorized:
			return status.Error(codes.Unauthenticated, appErr.Message)
		case CodeConflict:
			return status.Error(codes.AlreadyExists, appErr.Message)
		default:
			return status.Error(codes.Internal, "internal server error")
		}
	}

	// Fallback unknown internal error
	return status.Error(codes.Internal, "internal server error")
}
