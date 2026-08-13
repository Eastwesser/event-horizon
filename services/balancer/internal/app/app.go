package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/services/balancer/internal/balancer"
)

// App is the least-connections load balancer process.
type App struct {
	log        *slog.Logger
	closer     *closer.Closer
	httpServer *http.Server
	metricsSrv *http.Server
}

// New wires balancer HTTP + metrics servers.
func New(ctx context.Context) (*App, error) {
	_ = ctx
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	format := os.Getenv("LOG_FORMAT")
	if format == "" {
		format = "text"
	}
	log := logger.NewFrom(level, format)
	a := &App{log: log, closer: closer.New(log)}

	backends := []string{
		getenv("BALANCER_BACKEND_1", "http://gateway:8080"),
		getenv("BALANCER_BACKEND_2", "http://gateway-2:8080"),
		getenv("BALANCER_BACKEND_3", "http://gateway-3:8080"),
	}
	lb := balancer.NewLeastConnBalancer(backends)

	addr := getenv("BALANCER_ADDR", ":8079")
	a.httpServer = &http.Server{Addr: addr, Handler: lb}
	a.closer.AddNamed("http server", func(ctx context.Context) error {
		return a.httpServer.Shutdown(ctx)
	})

	metricsAddr := getenv("METRICS_ADDR", ":9098")
	a.metricsSrv = &http.Server{Addr: metricsAddr, Handler: metricsHandler()}
	a.closer.AddNamed("metrics http", func(ctx context.Context) error {
		return a.metricsSrv.Shutdown(ctx)
	})

	a.log.Info("balancer configured", "addr", addr, "backends", backends, "metrics", metricsAddr)
	return a, nil
}

// RunUntilSignal serves traffic until SIGINT/SIGTERM.
func (a *App) RunUntilSignal(ctx context.Context) error {
	go func() {
		a.log.Info("load balancer listening", "addr", a.httpServer.Addr)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("http serve stopped", "err", err)
		}
	}()
	go func() {
		a.log.Info("metrics listening", "addr", a.metricsSrv.Addr)
		if err := a.metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("metrics serve stopped", "err", err)
		}
	}()
	if err := a.closer.WaitSignal(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

// Close triggers graceful shutdown.
func (a *App) Close(ctx context.Context) error {
	return a.closer.CloseAll(ctx)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
