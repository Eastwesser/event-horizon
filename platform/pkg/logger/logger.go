package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/Eastwesser/event-horizon/platform/pkg/tracing"
)

// Config for structured slog logger.
type Config interface {
	Level() string
	Format() string // "json" | "text"
}

type staticConfig struct {
	level  string
	format string
}

func (c staticConfig) Level() string  { return c.level }
func (c staticConfig) Format() string { return c.format }

// DefaultConfig reads LOG_LEVEL / LOG_FORMAT from env (info / text by default).
func DefaultConfig() Config {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	format := os.Getenv("LOG_FORMAT")
	if format == "" {
		format = "text"
	}
	return staticConfig{level: level, format: format}
}

// New builds a slog.Logger from config.
func New(cfg Config) *slog.Logger {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return NewFrom(cfg.Level(), cfg.Format())
}

// NewFrom builds a slog.Logger from level/format strings (avoids cross-package interface friction).
func NewFrom(level, format string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var handler slog.Handler
	if strings.EqualFold(format, "json") {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

// Nop returns a discard logger for tests.
func Nop() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// WithContext enriches the logger with trace_id/span_id when present in ctx.
func WithContext(ctx context.Context, log *slog.Logger) *slog.Logger {
	if log == nil {
		log = slog.Default()
	}
	traceID := tracing.GetTraceID(ctx)
	if traceID == "" {
		return log
	}
	return log.With(
		slog.String("trace_id", traceID),
		slog.String("span_id", tracing.GetSpanID(ctx)),
	)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
