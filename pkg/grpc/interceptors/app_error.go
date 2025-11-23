package interceptors

import (
	"context"

	"github.com/ak-repo/stream-hub/pkg/errors"
	"google.golang.org/grpc"
)

func AppErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		if errors.IsAppError(err) {
			return nil, errors.ToGRPC(err)
		}

		// Fallback: unknown => 500
		return nil, errors.ToGRPC(
			errors.New(errors.CodeInternal, "internal server error", err),
		)

	}
}
