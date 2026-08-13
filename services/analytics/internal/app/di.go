package app

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/analytics/internal/config"
	"github.com/Eastwesser/event-horizon/services/analytics/internal/handler"
	"github.com/Eastwesser/event-horizon/services/analytics/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/analytics/internal/repository"
	ch "github.com/Eastwesser/event-horizon/services/analytics/internal/repository/clickhouse"
	"github.com/Eastwesser/event-horizon/services/analytics/internal/service"
	"github.com/Eastwesser/event-horizon/services/analytics/internal/worker"
	pb "github.com/Eastwesser/event-horizon/services/analytics/proto"
)

type diContainer struct {
	cfg    *config.Config
	ch     *ch.Client
	js     nats.JetStreamContext
	nc     *nats.Conn
	repo   *repository.AnalyticsRepo
	svc    *service.AnalyticsService
	api    pb.AnalyticsServiceServer
}

func newDiContainer(cfg *config.Config) *diContainer { return &diContainer{cfg: cfg} }

func (a *App) init(ctx context.Context) error {
	for _, step := range []func(context.Context) error{
		a.initLogger, a.initCloser, a.initClickHouse, a.initNATS, a.initDomain, a.initWorkers, a.initGRPC, a.initMetricsHTTP,
	} {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	level := slog.LevelInfo
	switch strings.ToLower(a.cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	var h slog.Handler
	if strings.EqualFold(a.cfg.LogFormat, "json") {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	a.log = slog.New(h)
	return nil
}

func (a *App) initCloser(_ context.Context) error {
	a.closer = closer.New(a.log)
	return nil
}

func (a *App) initClickHouse(ctx context.Context) error {
	client := ch.New(a.cfg.ClickHouseURL, a.cfg.ClickHouseDB)
	var err error
	for i := 0; i < 30; i++ {
		if err = client.Ping(ctx); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return err
	}
	if err := client.EnsureSchema(ctx); err != nil {
		return err
	}
	a.di.ch = client
	return nil
}

func (a *App) initNATS(_ context.Context) error {
	var nc *nats.Conn
	var err error
	for i := 0; i < 20; i++ {
		nc, err = nats.Connect(a.cfg.NATSURL)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return err
	}
	a.di.nc, a.di.js = nc, js
	a.closer.AddNamed("nats", func(context.Context) error { return nc.Drain() })
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	a.di.repo = repository.New(a.di.ch, a.cfg.ClickHouseDB)
	a.di.svc = service.New(a.di.repo)
	a.di.api = handler.NewGRPCHandler(a.di.svc)
	return nil
}

func (a *App) initWorkers(_ context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	go worker.NewIngestWorker(a.di.js, a.di.svc).Start(ctx)
	a.closer.AddNamed("workers", func(context.Context) error { cancel(); return nil })
	return nil
}

func (a *App) initGRPC(_ context.Context) error {
	lis, err := net.Listen("tcp", ":"+a.cfg.GRPCPort)
	if err != nil {
		return err
	}
	a.listener = lis
	a.closer.AddNamed("tcp", func(context.Context) error { return lis.Close() })
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor.Recovery(), interceptor.Logger(), interceptor.Validate()))
	pb.RegisterAnalyticsServiceServer(s, a.di.api)
	reflection.Register(s)
	a.grpcServer = s
	a.closer.AddNamed("grpc", func(context.Context) error { s.GracefulStop(); return nil })
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "analytics"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.di.ch.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "analytics"})
	})
	srv := &http.Server{Addr: ":" + a.cfg.MetricsPort, Handler: mux}
	a.metricsServer = srv
	go func() { _ = srv.ListenAndServe() }()
	a.closer.AddNamed("metrics", func(ctx context.Context) error { return srv.Shutdown(ctx) })
	return nil
}
