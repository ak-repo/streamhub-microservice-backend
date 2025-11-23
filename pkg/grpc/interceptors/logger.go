package interceptors

import (
	"context"
	"time"

	"github.com/ak-repo/stream-hub/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {

		start := time.Now()

		resp, err = handler(ctx, req)

		logger.Log.Info("gRPC request",
			zap.String("method", info.FullMethod),
			zap.Duration("latency", time.Since(start)),
			zap.Error(err),
		)

		return resp, err
	}
}
