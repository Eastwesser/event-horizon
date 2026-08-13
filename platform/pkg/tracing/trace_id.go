package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// GetTraceID returns the hex trace id from ctx, or empty if none.
func GetTraceID(ctx context.Context) string {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}

// GetSpanID returns the hex span id from ctx, or empty if none.
func GetSpanID(ctx context.Context) string {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		return ""
	}
	return sc.SpanID().String()
}
