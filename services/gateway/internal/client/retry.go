package client

import (
	"context"
	"math/rand"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryConfig controls unary client retries (exponential backoff + full jitter).
type RetryConfig struct {
	MaxAttempts int           // total attempts including the first call
	BaseDelay   time.Duration // backoff base; wait = rand(0, base*2^i)
	MaxDelay    time.Duration // cap on exponential term before jitter
}

// DefaultRetryConfig is tuned for short gateway→service hops.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   40 * time.Millisecond,
		MaxDelay:    400 * time.Millisecond,
	}
}

// UnaryRetryInterceptor retries transient gRPC failures with full jitter.
// Does not retry permanent codes (InvalidArgument, NotFound, PermissionDenied, …).
func UnaryRetryInterceptor(cfg RetryConfig) grpc.UnaryClientInterceptor {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 40 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 400 * time.Millisecond
	}

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var last error
		for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
			last = invoker(ctx, method, req, reply, cc, opts...)
			if last == nil {
				return nil
			}
			if !isRetryable(last) || attempt == cfg.MaxAttempts-1 {
				return last
			}
			if err := waitBackoff(ctx, cfg, attempt); err != nil {
				return last
			}
		}
		return last
	}
}

func isRetryable(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.Unavailable, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

func waitBackoff(ctx context.Context, cfg RetryConfig, attempt int) error {
	exp := cfg.BaseDelay * time.Duration(1<<attempt)
	if exp > cfg.MaxDelay {
		exp = cfg.MaxDelay
	}
	wait := time.Duration(rand.Int63n(int64(exp) + 1)) // full jitter
	if wait <= 0 {
		return nil
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
