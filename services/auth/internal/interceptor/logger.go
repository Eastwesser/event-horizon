package interceptor

import (
	"context"
	"log/slog"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Logger returns a unary server interceptor that logs method, status, and duration.
func Logger() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := path.Base(info.FullMethod)
		start := time.Now()

		resp, err := handler(ctx, req)
		duration := time.Since(start)

		if err != nil {
			st, _ := status.FromError(err)
			slog.Info("gRPC request",
				"method", method,
				"code", st.Code().String(),
				"duration_ms", duration.Milliseconds(),
				"err", err.Error(),
			)
		} else {
			slog.Info("gRPC request",
				"method", method,
				"code", "OK",
				"duration_ms", duration.Milliseconds(),
			)
		}
		return resp, err
	}
}
