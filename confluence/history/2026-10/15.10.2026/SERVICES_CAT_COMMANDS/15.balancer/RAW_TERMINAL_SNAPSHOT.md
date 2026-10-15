[denismatveev@c0der event_horizon]$ # Структура
echo "===== tree ====="
find services/balancer -type f -name '*.go' | sort

# Main
echo "===== main ====="
cat services/balancer/cmd/main.go

# App
echo "===== app ====="
for f in services/balancer/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done

# Backend
echo "===== backend ====="
for f in services/balancer/internal/backend/*.go; do echo "--- $f ---"; cat "$f"; done

# Balancer
echo "===== balancer ====="
for f in services/balancer/internal/balancer/*.go; do echo "--- $f ---"; cat "$f"; done

# Metrics
echo "===== metrics ====="
for f in services/balancer/internal/metrics/*.go; do echo "--- $f ---"; cat "$f"; done

# Proxy
echo "===== proxy ====="
for f in services/balancer/internal/proxy/*.go; do echo "--- $f ---"; cat "$f"; done

# Config (если есть)
echo "===== config ====="
for f in services/balancer/internal/config/*.go; do echo "--- $f ---"; cat "$f"; done
===== tree =====
services/balancer/cmd/main.go
services/balancer/internal/app/app.go
services/balancer/internal/app/metrics.go
services/balancer/internal/backend/backend.go
services/balancer/internal/balancer/least_conn.go
services/balancer/internal/balancer/least_conn_test.go
services/balancer/internal/metrics/metrics.go
services/balancer/internal/proxy/proxy.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/balancer/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init balancer app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("balancer stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/balancer/internal/app/app.go ---
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
--- services/balancer/internal/app/metrics.go ---
package app

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func metricsHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"balancer"}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","service":"balancer"}`))
	})
	return mux
}
===== backend =====
--- services/balancer/internal/backend/backend.go ---
===== balancer =====
--- services/balancer/internal/balancer/least_conn.go ---
package balancer

import (
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "sync"
    "sync/atomic"

    "github.com/Eastwesser/event-horizon/services/balancer/internal/metrics"
)

type Backend struct {
    URL         *url.URL
    ActiveConns int32
    Mutex       sync.Mutex
    Proxy       *httputil.ReverseProxy
}

type LeastConnBalancer struct {
    backends []*Backend
    mu       sync.RWMutex
}

func NewLeastConnBalancer(urls []string) *LeastConnBalancer {
    backends := make([]*Backend, 0, len(urls))

    for _, rawURL := range urls {
        u, err := url.Parse(rawURL)
        if err != nil {
            log.Printf("Failed to parse URL %s: %v", rawURL, err)
            continue
        }

        backend := &Backend{
            URL:   u,
            Proxy: httputil.NewSingleHostReverseProxy(u),
        }

        // custom proxy for logging
        backend.Proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
            log.Printf("Proxy error: %v", err)
            http.Error(w, "Backend error", http.StatusBadGateway)
        }

        backends = append(backends, backend)
    }

    return &LeastConnBalancer{
        backends: backends,
    }
}

func (lb *LeastConnBalancer) getLeastConnBackend() *Backend {
    lb.mu.RLock()
    defer lb.mu.RUnlock()

    if len(lb.backends) == 0 {
        return nil
    }

    var selected *Backend
    var minConns int32 = 2147483647 // MaxInt32

    for _, b := range lb.backends {
        conns := atomic.LoadInt32(&b.ActiveConns)
        if conns < minConns {
            minConns = conns
            selected = b
        }
    }

    return selected
}

func (lb *LeastConnBalancer) getTotalActiveConns() int32 {
    lb.mu.RLock()
    defer lb.mu.RUnlock()
    
    var total int32
    for _, b := range lb.backends {
        total += atomic.LoadInt32(&b.ActiveConns)
    }
    return total
}

func (lb *LeastConnBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    metrics.RequestsTotal.Inc()
    metrics.ActiveConnections.Set(float64(lb.getTotalActiveConns()))
    backend := lb.getLeastConnBackend()

    if backend == nil {
        http.Error(w, "No available backends", http.StatusServiceUnavailable)
        return
    }

    atomic.AddInt32(&backend.ActiveConns, 1)
    defer atomic.AddInt32(&backend.ActiveConns, -1)

    log.Printf("🔀 Proxying to %s (active: %d)", backend.URL.Host, backend.ActiveConns)

    // Добавляем заголовок с именем бэкенда
    r.Header.Set("X-Balancer-Backend", backend.URL.Host)

    backend.Proxy.ServeHTTP(w, r)
}
--- services/balancer/internal/balancer/least_conn_test.go ---
package balancer

import (
	"sync/atomic"
	"testing"
)

func TestGetLeastConnBackend_PicksLowest(t *testing.T) {
	lb := NewLeastConnBalancer([]string{
		"http://gw-busy:8080",
		"http://gw-idle:8080",
		"http://gw-mid:8080",
	})
	if got := len(lb.backends); got != 3 {
		t.Fatalf("backends=%d want 3", got)
	}

	atomic.StoreInt32(&lb.backends[0].ActiveConns, 10)
	atomic.StoreInt32(&lb.backends[1].ActiveConns, 1)
	atomic.StoreInt32(&lb.backends[2].ActiveConns, 4)

	got := lb.getLeastConnBackend()
	if got == nil {
		t.Fatal("expected a backend")
	}
	if got.URL.Host != "gw-idle:8080" {
		t.Fatalf("host=%s want gw-idle:8080", got.URL.Host)
	}
}

func TestGetLeastConnBackend_Empty(t *testing.T) {
	lb := NewLeastConnBalancer(nil)
	if b := lb.getLeastConnBackend(); b != nil {
		t.Fatalf("expected nil, got %v", b.URL)
	}
}
===== metrics =====
--- services/balancer/internal/metrics/metrics.go ---
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    ActiveConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "balancer_active_connections",
            Help: "Total active connections across all backends",
        },
    )

    RequestsTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "balancer_requests_total",
            Help: "Total number of requests processed by balancer",
        },
    )
)
===== proxy =====
--- services/balancer/internal/proxy/proxy.go ---
===== config =====
--- services/balancer/internal/config/*.go ---
cat: 'services/balancer/internal/config/*.go': Нет такого файла или каталога
[denismatveev@c0der event_horizon]$ 

