package client

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryRetryInterceptor_retriesUnavailable(t *testing.T) {
	var calls atomic.Int32
	invoker := func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		n := calls.Add(1)
		if n < 3 {
			return status.Error(codes.Unavailable, "down")
		}
		return nil
	}

	interceptor := UnaryRetryInterceptor(RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   time.Millisecond,
		MaxDelay:    5 * time.Millisecond,
	})

	err := interceptor(context.Background(), "/svc/Method", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if calls.Load() != 3 {
		t.Fatalf("calls=%d want 3", calls.Load())
	}
}

func TestUnaryRetryInterceptor_noRetryOnInvalidArgument(t *testing.T) {
	var calls atomic.Int32
	invoker := func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		calls.Add(1)
		return status.Error(codes.InvalidArgument, "bad")
	}

	interceptor := UnaryRetryInterceptor(RetryConfig{
		MaxAttempts: 4,
		BaseDelay:   time.Millisecond,
		MaxDelay:    5 * time.Millisecond,
	})

	err := interceptor(context.Background(), "/svc/Method", nil, nil, nil, invoker)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d want 1", calls.Load())
	}
}

func TestUnaryRetryInterceptor_respectsContextCancel(t *testing.T) {
	var calls atomic.Int32
	invoker := func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		calls.Add(1)
		return status.Error(codes.Unavailable, "down")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	interceptor := UnaryRetryInterceptor(RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   50 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
	})

	err := interceptor(ctx, "/svc/Method", nil, nil, nil, invoker)
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("want Unavailable (last RPC err), got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d want 1 (no further attempts after cancel)", calls.Load())
	}
}

func TestIsRetryable(t *testing.T) {
	if !isRetryable(status.Error(codes.Unavailable, "x")) {
		t.Fatal("Unavailable should retry")
	}
	if !isRetryable(status.Error(codes.ResourceExhausted, "x")) {
		t.Fatal("ResourceExhausted should retry")
	}
	if isRetryable(status.Error(codes.PermissionDenied, "x")) {
		t.Fatal("PermissionDenied must not retry")
	}
}
