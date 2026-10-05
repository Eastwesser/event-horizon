package interceptor

import (
	"context"
	"log/slog"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type protoValidator interface {
	Validate() error
}

func Validate() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if v, ok := req.(protoValidator); ok {
			if err := v.Validate(); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
			}
		}
		return handler(ctx, req)
	}
}

func Logger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := path.Base(info.FullMethod)
		start := time.Now()
		resp, err := handler(ctx, req)
		if err != nil {
			st, _ := status.FromError(err)
			slog.Info("gRPC request", "method", method, "code", st.Code().String(), "duration_ms", time.Since(start).Milliseconds(), "err", err.Error())
		} else {
			slog.Info("gRPC request", "method", method, "code", "OK", "duration_ms", time.Since(start).Milliseconds())
		}
		return resp, err
	}
}

func Recovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", "method", info.FullMethod, "panic", r)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}
