package app

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
	"github.com/Eastwesser/event-horizon/platform/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/config"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/handler"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/repository"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/service"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/worker"
	invMigrations "github.com/Eastwesser/event-horizon/services/inventory/migrations"
	pb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

type diContainer struct {
	cfg *config.Config

	db          *sql.DB
	mongoClient *mongo.Client
	nc          *nats.Conn
	js          nats.JetStreamContext
	baseRepo    repository.InventoryRepository
	cache       *repository.RedisCacheRepo
	repo        repository.InventoryRepository
	outbox      repository.ItemOutboxWriter
	svc         *service.InventoryService
	api         pb.InventoryServiceServer

	workerCancel context.CancelFunc
}

func newDiContainer(cfg *config.Config) *diContainer {
	return &diContainer{cfg: cfg}
}

func (a *App) init(ctx context.Context) error {
	steps := []func(context.Context) error{
		a.initLogger,
		a.initCloser,
		a.initTracer,
		a.initRepository,
		a.initNATS,
		a.initDomain,
		a.initGRPC,
		a.initMetricsHTTP,
		a.initOutboxWorker,
	}
	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	a.log = logger.New(logger.DefaultConfig())
	return nil
}

func (a *App) initCloser(_ context.Context) error {
	a.closer = closer.New(a.log)
	return nil
}

func (a *App) initTracer(ctx context.Context) error {
	endpoint := getenv("JAEGER_ENDPOINT", "localhost:4317")
	shutdown, err := tracing.Init(ctx, tracing.Config{
		ServiceName: "inventory",
		Endpoint:    endpoint,
		Environment: getenv("ENVIRONMENT", "development"),
		Insecure:    true,
	})
	if err != nil {
		a.log.Warn("tracer init failed; continuing without OTLP", "err", err)
		return nil
	}
	a.closer.AddNamed("otel tracer", shutdown)
	a.log.Info("jaeger tracer ready", "endpoint", endpoint)
	return nil
}

func (a *App) initRepository(ctx context.Context) error {
	switch a.cfg.Driver {
	case "postgres":
		a.log.Info("using PostgreSQL driver")
		repo, db, err := a.connectPostgres()
		if err != nil {
			return err
		}
		a.di.baseRepo = repo
		a.di.repo = repo
		a.di.db = db
		a.closer.AddNamed("postgres", func(context.Context) error {
			return db.Close()
		})
	case "mongo":
		a.log.Info("using MongoDB driver")
		repo, client, err := a.connectMongo(ctx)
		if err != nil {
			return err
		}
		a.di.baseRepo = repo
		a.di.repo = repo
		a.di.outbox = nil
		a.di.mongoClient = client
		a.closer.AddNamed("mongodb", func(ctx context.Context) error {
			return client.Disconnect(ctx)
		})
	default:
		return fmt.Errorf("unknown driver: %s (use 'postgres' or 'mongo')", a.cfg.Driver)
	}
	return nil
}

func (a *App) connectPostgres() (repository.InventoryRepository, *sql.DB, error) {
	db, err := sql.Open("postgres", a.cfg.PGDSN())
	if err != nil {
		return nil, nil, fmt.Errorf("postgres open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("postgres ping: %w", err)
	}
	if err := migrator.Up(db, invMigrations.FS); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("migrations: %w", err)
	}
	a.log.Info("postgres ready, migrations applied")
	return repository.NewPostgresRepo(db), db, nil
}

func (a *App) connectMongo(ctx context.Context) (repository.InventoryRepository, *mongo.Client, error) {
	mongoCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(a.cfg.MongoURI))
	if err != nil {
		return nil, nil, fmt.Errorf("mongo connect: %w", err)
	}
	if err := client.Ping(mongoCtx, nil); err != nil {
		_ = client.Disconnect(mongoCtx)
		return nil, nil, fmt.Errorf("mongo ping: %w", err)
	}
	a.log.Info("mongodb ready")
	return repository.NewMongoRepo(client.Database(a.cfg.MongoDBName)), client, nil
}

func (a *App) initNATS(_ context.Context) error {
	nc, err := nats.Connect(a.cfg.NATSURL)
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
	a.closer.AddNamed("nats", func(context.Context) error {
		return nc.Drain()
	})
	a.log.Info("nats connected", "url", a.cfg.NATSURL)
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	cache := repository.NewRedisCacheRepo(a.cfg.RedisAddr, 5*time.Minute)
	a.di.cache = cache
	a.closer.AddNamed("redis cache", func(context.Context) error {
		return cache.Close()
	})
	a.di.repo = repository.NewCachedRepository(a.di.baseRepo, cache)
	if ow, ok := a.di.repo.(repository.ItemOutboxWriter); ok {
		a.di.outbox = ow
	}
	a.di.svc = service.NewInventoryService(a.di.repo, a.di.outbox, a.di.js)
	a.di.api = handler.NewGRPCHandler(a.di.svc)
	return nil
}

func (a *App) initGRPC(_ context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	a.listener = lis
	a.closer.AddNamed("tcp listener", func(context.Context) error {
		return lis.Close()
	})

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptor.Recovery(),
			interceptor.Logger(),
			interceptor.Validate(),
			interceptor.RequireRoles(
				[]string{"author", "admin"},
				[]string{
					"/CreateItem", "/CreateItems", "/UpdateItem", "/DeleteItem",
					"/ReserveItem", "/SoftDeleteItem", "/RestoreItem", "/GetStats",
				},
			),
			metrics.UnaryServerInterceptor("inventory"),
		),
	)
	pb.RegisterInventoryServiceServer(grpcServer, a.di.api)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	reflection.Register(grpcServer)

	a.grpcServer = grpcServer
	metrics.SetHealthy("inventory")
	a.closer.AddNamed("grpc server", func(context.Context) error {
		grpcServer.GracefulStop()
		return nil
	})
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	cfg := a.cfg
	db := a.di.db
	mongoClient := a.di.mongoClient

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"inventory","driver":"%s"}`, cfg.Driver)
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := "ready"
		httpStatus := http.StatusOK

		if cfg.Driver == "postgres" && db != nil {
			if err := db.Ping(); err != nil {
				status = "degraded"
				httpStatus = http.StatusServiceUnavailable
				a.log.Warn("ready check: postgres ping failed", "err", err)
			}
		}
		if cfg.Driver == "mongo" && mongoClient != nil {
			pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := mongoClient.Ping(pingCtx, nil); err != nil {
				status = "degraded"
				httpStatus = http.StatusServiceUnavailable
				a.log.Warn("ready check: mongodb ping failed", "err", err)
			}
		}

		w.WriteHeader(httpStatus)
		fmt.Fprintf(w, `{"status":"%s","service":"inventory","driver":"%s"}`, status, cfg.Driver)
	})

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.MetricsPort), Handler: mux}
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

func (a *App) initOutboxWorker(_ context.Context) error {
	if a.cfg.Driver != "postgres" || a.di.db == nil {
		return nil
	}
	workerCtx, cancel := context.WithCancel(context.Background())
	a.di.workerCancel = cancel
	a.closer.AddNamed("outbox worker", func(context.Context) error {
		cancel()
		return nil
	})
	outboxWorker := worker.NewOutboxWorker(a.di.db, a.di.js)
	go 	outboxWorker.Start(workerCtx)
	a.log.Info("outbox worker started")
	return nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
