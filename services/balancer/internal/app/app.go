package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/services/balancer/internal/balancer"
	"github.com/Eastwesser/event-horizon/services/balancer/internal/metrics"
)

const (
	healthCheckInterval = 5 * time.Second
	healthCheckTimeout  = 2 * time.Second
	healthCheckPath     = "/ready" // /ready pings Redis in gateway; /health only process alive
)

// App is the least-connections load balancer process.
type App struct {
	log        *slog.Logger
	closer     *closer.Closer
	lb         *balancer.LeastConnBalancer
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
	a.lb = balancer.NewLeastConnBalancer(backends, log)

	addr := getenv("BALANCER_ADDR", ":8079")
	a.httpServer = &http.Server{Addr: addr, Handler: a.lb}
	a.closer.AddNamed("http server", func(ctx context.Context) error {
		return a.httpServer.Shutdown(ctx)
	})

	metricsAddr := getenv("METRICS_ADDR", ":9098")
	a.metricsSrv = &http.Server{Addr: metricsAddr, Handler: metricsHandler(a.lb)}
	a.closer.AddNamed("metrics http", func(ctx context.Context) error {
		return a.metricsSrv.Shutdown(ctx)
	})

	a.log.Info("balancer configured", "addr", addr, "backends", backends, "metrics", metricsAddr)
	return a, nil
}

// RunUntilSignal serves traffic and runs the health-check loop until SIGINT/SIGTERM.
func (a *App) RunUntilSignal(ctx context.Context) error {
	go a.healthCheckLoop(ctx)

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

// healthCheckLoop pings each backend's /ready every interval and flips Healthy.
func (a *App) healthCheckLoop(ctx context.Context) {
	client := &http.Client{Timeout: healthCheckTimeout}
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	check := func() {
		for _, b := range a.lb.Backends() {
			url := b.URL.Scheme + "://" + b.URL.Host + healthCheckPath
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				a.log.Warn("health check: build request", "backend", b.URL.Host, "err", err)
				b.Healthy.Store(false)
				metrics.BackendHealthy.WithLabelValues(b.URL.Host).Set(0)
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				if b.Healthy.Swap(false) {
					a.log.Warn("backend marked unhealthy", "backend", b.URL.Host, "err", err)
				}
				metrics.BackendHealthy.WithLabelValues(b.URL.Host).Set(0)
				continue
			}
			_ = resp.Body.Close()
			ok := resp.StatusCode >= 200 && resp.StatusCode < 300
			if b.Healthy.Swap(ok) != ok {
				if ok {
					a.log.Info("backend marked healthy", "backend", b.URL.Host, "status", resp.StatusCode)
				} else {
					a.log.Warn("backend marked unhealthy", "backend", b.URL.Host, "status", resp.StatusCode)
				}
			}
			if ok {
				metrics.BackendHealthy.WithLabelValues(b.URL.Host).Set(1)
			} else {
				metrics.BackendHealthy.WithLabelValues(b.URL.Host).Set(0)
			}
		}
	}

	check() // run immediately on start
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			check()
		}
	}
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
