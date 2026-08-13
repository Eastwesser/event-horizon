package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
	"github.com/Eastwesser/event-horizon/platform/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"github.com/Eastwesser/event-horizon/services/shop/internal/config"
	"github.com/Eastwesser/event-horizon/services/shop/internal/handler"
	"github.com/Eastwesser/event-horizon/services/shop/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/shop/internal/repository"
	"github.com/Eastwesser/event-horizon/services/shop/internal/service"
	"github.com/Eastwesser/event-horizon/services/shop/migrations"
	pb "github.com/Eastwesser/event-horizon/services/shop/proto"
)

type diContainer struct {
	cfg config.ConfigProvider

	db        *sql.DB
	redisRepo *repository.RedisShopRepo
	nc        *nats.Conn
	js        nats.JetStreamContext
	pgRepo    *repository.PostgresShopRepo
	svc       service.ShopService
	api       pb.ShopServiceServer
	kafkaProd kafka.Producer
}

func newDiContainer(cfg config.ConfigProvider) *diContainer {
	return &diContainer{cfg: cfg}
}

func (a *App) init(ctx context.Context) error {
	steps := []func(context.Context) error{
		a.initLogger,
		a.initCloser,
		a.initTracer,
		a.initPostgres,
		a.initRedis,
		a.initNATS,
		a.initKafka,
		a.initDomain,
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
	lc := a.cfg.LoggerConfig()
	a.log = logger.NewFrom(lc.Level(), lc.Format())
	return nil
}

func (a *App) initCloser(_ context.Context) error {
	a.closer = closer.New(a.log)
	return nil
}

func (a *App) initTracer(ctx context.Context) error {
	endpoint := getenv("JAEGER_ENDPOINT", "localhost:4317")
	shutdown, err := tracing.Init(ctx, tracing.Config{
		ServiceName: "shop",
		Endpoint:    endpoint,
		Environment: getenv("ENVIRONMENT", "development"),
		Insecure:    true,
	})
	if err != nil {
		a.log.Warn("tracer init failed; continuing without OTLP", "err", err)
		return nil
	}
	a.closer.AddNamed("otel tracer", shutdown)
	return nil
}

func (a *App) initPostgres(_ context.Context) error {
	db, err := sql.Open("postgres", a.cfg.PostgresConfig().DSN())
	if err != nil {
		return fmt.Errorf("postgres open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("postgres ping: %w", err)
	}
	if err := migrator.Up(db, migrations.FS); err != nil {
		_ = db.Close()
		return fmt.Errorf("migrations: %w", err)
	}

	a.di.db = db
	a.closer.AddNamed("postgres", func(context.Context) error {
		return db.Close()
	})
	a.log.Info("postgres ready, migrations applied")
	return nil
}

func (a *App) initRedis(_ context.Context) error {
	rc := a.cfg.RedisConfig()
	a.di.redisRepo = repository.NewRedisShopRepo(rc.Addr(), rc.DB())
	a.log.Info("redis repo ready", "addr", rc.Addr(), "db", rc.DB())
	return nil
}

func (a *App) initNATS(_ context.Context) error {
	url := a.cfg.NATSConfig().URL()
	var nc *nats.Conn
	var err error
	for i := 0; i < 30; i++ {
		nc, err = nats.Connect(url)
		if err == nil {
			break
		}
		a.log.Warn("nats connection attempt failed", "attempt", i+1, "of", 30, "err", err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("nats connect after 30 attempts: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return fmt.Errorf("jetstream: %w", err)
	}

	a.di.nc = nc
	a.di.js = js
	a.closer.AddNamed("nats", func(context.Context) error {
		return nc.Drain()
	})
	a.log.Info("nats connected", "url", url)
	return nil
}

func (a *App) initKafka(_ context.Context) error {
	cfg := kafka.LoadConfig()
	a.di.kafkaProd = kafka.NewProducer(cfg, kafka.TopicPurchasePaid, a.log)
	a.closer.AddNamed("kafka producer", func(context.Context) error {
		return a.di.kafkaProd.Close()
	})
	a.log.Info("kafka producer ready", "enabled", cfg.Enabled, "topic", kafka.TopicPurchasePaid)
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	a.di.pgRepo = repository.NewPostgresShopRepo(a.di.db)
	a.di.svc = service.NewShopService(
		a.di.pgRepo,
		a.di.redisRepo,
		a.di.js,
		a.cfg.BillingConfig().Addr(),
		a.cfg.PaymentConfig().Addr(),
	)
	if a.di.svc != nil {
		a.di.svc.SetKafkaProducer(a.di.kafkaProd)
	}
	a.di.api = handler.NewShopHandler(a.di.svc)
	return nil
}

func (a *App) initGRPC(_ context.Context) error {
	lis, err := net.Listen("tcp", ":"+a.cfg.GRPCConfig().Port())
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	a.listener = lis
	a.closer.AddNamed("tcp listener", func(context.Context) error {
		return lis.Close()
	})

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptor.Recovery(),
			interceptor.Logger(),
			interceptor.Validate(),
			metrics.UnaryServerInterceptor("shop"),
		),
	)
	pb.RegisterShopServiceServer(grpcServer, a.di.api)
	reflection.Register(grpcServer)

	a.grpcServer = grpcServer
	metrics.SetHealthy("shop")
	a.closer.AddNamed("grpc server", func(context.Context) error {
		grpcServer.GracefulStop()
		return nil
	})
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	db := a.di.db
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "shop"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded", "reason": "postgres unreachable"})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "shop"})
	})

	srv := &http.Server{Addr: ":" + a.cfg.MetricsConfig().Port(), Handler: mux}
	a.metricsServer = srv
	go func() {
		a.log.Info("metrics/health listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("metrics server error", "err", err)
		}
	}()
	a.closer.AddNamed("metrics http", func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	})
	return nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
