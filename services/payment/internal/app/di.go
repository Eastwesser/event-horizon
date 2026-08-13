package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/payment/internal/config"
	"github.com/Eastwesser/event-horizon/services/payment/internal/handler"
	"github.com/Eastwesser/event-horizon/services/payment/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/payment/internal/repository"
	"github.com/Eastwesser/event-horizon/services/payment/internal/service"
	"github.com/Eastwesser/event-horizon/services/payment/internal/worker"
	"github.com/Eastwesser/event-horizon/services/payment/migrations"
	pb "github.com/Eastwesser/event-horizon/services/payment/proto"
)

type diContainer struct {
	cfg *config.Config

	dbpool *pgxpool.Pool
	cache  *repository.RedisRepo
	nc     *nats.Conn
	js     nats.JetStreamContext
	repo   *repository.PostgresRepo
	svc    *service.PaymentService
	api    pb.PaymentServiceServer
}

func newDiContainer(cfg *config.Config) *diContainer {
	return &diContainer{cfg: cfg}
}

func (a *App) init(ctx context.Context) error {
	steps := []func(context.Context) error{
		a.initLogger,
		a.initCloser,
		a.initPostgres,
		a.initRedis,
		a.initNATS,
		a.initDomain,
		a.initOutbox,
		a.initGRPC,
		a.initMetricsHTTP,
	}
	for _, step := range steps {
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
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	if strings.EqualFold(a.cfg.LogFormat, "json") {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	a.log = slog.New(h)
	return nil
}

func (a *App) initCloser(_ context.Context) error {
	a.closer = closer.New(a.log)
	return nil
}

func (a *App) initPostgres(ctx context.Context) error {
	poolCfg, err := pgxpool.ParseConfig(a.cfg.DSN())
	if err != nil {
		return fmt.Errorf("postgres config: %w", err)
	}
	poolCfg.MaxConns = 25
	poolCfg.MinConns = 10
	poolCfg.MaxConnLifetime = 5 * time.Minute

	dbpool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("postgres connect: %w", err)
	}
	if err := dbpool.Ping(ctx); err != nil {
		dbpool.Close()
		return fmt.Errorf("postgres ping: %w", err)
	}
	sqlDB := stdlib.OpenDBFromPool(dbpool)
	if err := migrator.Up(sqlDB, migrations.FS); err != nil {
		dbpool.Close()
		return fmt.Errorf("migrations: %w", err)
	}
	a.di.dbpool = dbpool
	a.closer.AddNamed("postgres", func(context.Context) error {
		dbpool.Close()
		return nil
	})
	a.log.Info("postgres ready, migrations applied")
	return nil
}

func (a *App) initRedis(ctx context.Context) error {
	cache := repository.NewRedisRepo(a.cfg.RedisAddr, 30*time.Minute)
	if err := cache.Ping(ctx); err != nil {
		a.log.Warn("redis unavailable; subscription cache degraded", "addr", a.cfg.RedisAddr, "err", err)
	} else {
		a.log.Info("redis connected", "addr", a.cfg.RedisAddr)
	}
	a.di.cache = cache
	a.closer.AddNamed("redis", func(context.Context) error { return cache.Close() })
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
		a.log.Warn("nats connect attempt failed", "attempt", i+1, "err", err)
		time.Sleep(time.Second)
	}
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return fmt.Errorf("jetstream: %w", err)
	}
	a.di.nc = nc
	a.di.js = js
	a.closer.AddNamed("nats", func(context.Context) error { return nc.Drain() })
	a.log.Info("nats connected", "url", a.cfg.NATSURL)
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	a.di.repo = repository.NewPostgresRepo(a.di.dbpool)
	a.di.svc = service.New(
		a.di.repo,
		a.di.cache,
		a.cfg.BoostyCheckoutURL,
		a.cfg.WebhookSecret,
		a.cfg.SubscriptionDays,
	)
	a.di.api = handler.NewGRPCHandler(a.di.svc)
	return nil
}

func (a *App) initOutbox(_ context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	w := worker.NewOutboxWorker(a.di.dbpool, a.di.js)
	go w.Start(ctx)
	a.closer.AddNamed("outbox worker", func(context.Context) error {
		cancel()
		return nil
	})
	a.log.Info("outbox worker started")
	return nil
}

func (a *App) initGRPC(_ context.Context) error {
	lis, err := net.Listen("tcp", ":"+a.cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	a.listener = lis
	a.closer.AddNamed("tcp listener", func(context.Context) error { return lis.Close() })

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.Recovery(),
			interceptor.Logger(),
			interceptor.Validate(),
		),
	)
	pb.RegisterPaymentServiceServer(grpcServer, a.di.api)
	reflection.Register(grpcServer)
	a.grpcServer = grpcServer
	a.closer.AddNamed("grpc server", func(context.Context) error {
		grpcServer.GracefulStop()
		return nil
	})
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	dbpool := a.di.dbpool
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "payment"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := dbpool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded", "reason": "postgres unreachable"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "payment"})
	})
	srv := &http.Server{Addr: ":" + a.cfg.MetricsPort, Handler: mux}
	a.metricsServer = srv
	go func() {
		a.log.Info("metrics/health listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("metrics server error", "err", err)
		}
	}()
	a.closer.AddNamed("metrics http", func(ctx context.Context) error { return srv.Shutdown(ctx) })
	return nil
}
