package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total gRPC requests handled",
		},
		[]string{"service", "method", "code"},
	)
	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)
	grpcRequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_request_errors_total",
			Help: "Total gRPC requests that returned an error",
		},
		[]string{"service", "method"},
	)
	serviceHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_health",
			Help: "1 when service process is up and serving",
		},
		[]string{"service"},
	)
)

// SetHealthy marks service as up (call once after gRPC server starts).
func SetHealthy(serviceName string) {
	serviceHealth.WithLabelValues(serviceName).Set(1)
}

// UnaryServerInterceptor records request count, latency, and errors.
func UnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		method := info.FullMethod
		code := "OK"
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code().String()
			} else {
				code = "Unknown"
			}
			grpcRequestErrors.WithLabelValues(serviceName, method).Inc()
		}
		grpcRequestsTotal.WithLabelValues(serviceName, method, code).Inc()
		grpcRequestDuration.WithLabelValues(serviceName, method).Observe(time.Since(start).Seconds())
		return resp, err
	}
}
