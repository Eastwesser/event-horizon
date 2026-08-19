package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/services/billing/internal/config"
	"github.com/Eastwesser/event-horizon/services/billing/internal/handler"
	"github.com/Eastwesser/event-horizon/services/billing/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
	"github.com/Eastwesser/event-horizon/services/billing/internal/service"
	"github.com/Eastwesser/event-horizon/services/billing/internal/worker"
	"github.com/Eastwesser/event-horizon/services/billing/migrations"
	pb "github.com/Eastwesser/event-horizon/services/billing/proto"
)

// ScoreEvent is the NATS payload for score.updated reward grants.
type ScoreEvent struct {
	UserID        string `json:"user_id"`
	GameID        string `json:"game_id"`
	Score         int    `json:"score"`
	IsRecord      bool   `json:"is_record"`
	Level         int    `json:"level"`
	LampsEarned   int    `json:"lamps_earned"`
	TicketsEarned int    `json:"tickets_earned"`
	Timestamp     int64  `json:"timestamp"`
}

type diContainer struct {
	cfg config.ConfigProvider

	dbpool    *pgxpool.Pool
	redisRepo *repository.RedisBillingRepo
	nc        *nats.Conn
	js        nats.JetStreamContext
	pgRepo    *repository.PostgresBillingRepo
	svc       service.BillingService
	api       pb.BillingServiceServer
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
		a.initDomain,
		a.initOutbox,
		a.initGRPC,
		a.initMetricsHTTP,
		a.initScoreSubscribe,
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
	a.log.Info("initializing tracer", "endpoint", endpoint)

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		a.log.Warn("tracer exporter failed; continuing without OTLP", "err", err)
		return nil
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("billing"),
			attribute.String("environment", "development"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	a.closer.AddNamed("otel tracer", func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	})
	a.log.Info("jaeger tracer ready")
	return nil
}

func (a *App) initPostgres(ctx context.Context) error {
	pg := a.cfg.PostgresConfig()
	poolCfg, err := pgxpool.ParseConfig(pg.DSN())
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
	a.closer.AddNamed("postgres pool", func(context.Context) error {
		dbpool.Close()
		return nil
	})
	a.log.Info("postgres ready, migrations applied")
	return nil
}

func (a *App) initRedis(_ context.Context) error {
	rc := a.cfg.RedisConfig()
	a.di.redisRepo = repository.NewRedisBillingRepo(rc.Addr(), rc.DB())
	a.log.Info("redis repo ready", "addr", rc.Addr(), "db", rc.DB())
	return nil
}

func (a *App) initNATS(_ context.Context) error {
	url := a.cfg.NATSConfig().URL()
	var nc *nats.Conn
	var lastErr error
	for i := 0; i < 10; i++ {
		nc, lastErr = nats.Connect(url)
		if lastErr == nil {
			break
		}
		a.log.Warn("nats connection attempt failed", "attempt", i+1, "err", lastErr)
		time.Sleep(2 * time.Second)
	}
	if lastErr != nil {
		return fmt.Errorf("nats connect after 10 attempts: %w", lastErr)
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

func (a *App) initDomain(_ context.Context) error {
	a.di.pgRepo = repository.NewPostgresBillingRepo(a.di.dbpool)
	a.di.svc = service.NewBillingService(a.di.pgRepo, a.di.redisRepo)
	a.di.api = handler.NewBillingHandler(a.di.svc)
	return nil
}

func (a *App) initOutbox(_ context.Context) error {
	outboxCtx, cancelOutbox := context.WithCancel(context.Background())
	outboxWorker := worker.NewOutboxWorker(a.di.dbpool, a.di.js)
	go outboxWorker.Start(outboxCtx)
	a.closer.AddNamed("outbox worker", func(context.Context) error {
		cancelOutbox()
		return nil
	})
	a.log.Info("outbox worker started")
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
		),
	)
	pb.RegisterBillingServiceServer(grpcServer, a.di.api)
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
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "billing"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		readyCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := dbpool.Ping(readyCtx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded", "reason": "postgres unreachable"})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "billing"})
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

func (a *App) initScoreSubscribe(_ context.Context) error {
	billingService := a.di.svc
	log := a.log
	_, err := a.di.js.Subscribe("score.updated", func(msg *nats.Msg) {
		var event ScoreEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Error("failed to unmarshal score event", "err", err)
			return
		}
		log.Info("received score event",
			"user_id", event.UserID,
			"lamps", event.LampsEarned,
			"tickets", event.TicketsEarned,
		)

		lampsRefID := fmt.Sprintf("%s-lamps-%d", event.UserID, time.Now().UnixNano())
		ticketsRefID := fmt.Sprintf("%s-tickets-%d", event.UserID, time.Now().UnixNano())

		if event.LampsEarned > 0 {
			_, err := billingService.AddCurrency(context.Background(), event.UserID,
				repository.Lamps, event.LampsEarned, "game_reward", lampsRefID)
			if err != nil {
				log.Error("failed to add lamps", "err", err)
			} else {
				log.Info("added lamps", "amount", event.LampsEarned, "user_id", event.UserID)
			}
		}

		if event.TicketsEarned > 0 {
			_, err := billingService.AddCurrency(context.Background(), event.UserID,
				repository.Tickets, event.TicketsEarned, "game_reward", ticketsRefID)
			if err != nil {
				log.Error("failed to add tickets", "err", err)
			} else {
				log.Info("added tickets", "amount", event.TicketsEarned, "user_id", event.UserID)
			}
		}

		_ = msg.Ack()
	}, nats.Durable("billing-durable"), nats.ManualAck())

	if err != nil {
		a.log.Warn("failed to subscribe to score.updated", "err", err)
	} else {
		a.log.Info("subscribed to nats", "subject", "score.updated")
	}
	return nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
