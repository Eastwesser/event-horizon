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
	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/config"
	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/handler"
	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/repository"
	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/service"
	"github.com/Eastwesser/event-horizon/services/leaderboard/migrations"
	pb "github.com/Eastwesser/event-horizon/services/leaderboard/proto"
)

// ScoreEvent is the NATS payload for score.updated.
type ScoreEvent struct {
	UserID    string `json:"user_id"`
	GameID    string `json:"game_id"`
	UserEmail string `json:"user_email"`
	Nickname  string `json:"nickname"`
	Score     int    `json:"score"`
}

type diContainer struct {
	cfg *config.Config

	dbpool *pgxpool.Pool
	nc     *nats.Conn
	js     nats.JetStreamContext
	repo   *repository.RedisLeaderboardRepo
	svc    service.LeaderboardService
	api    pb.LeaderboardServiceServer

	ackCancel context.CancelFunc
}

func newDiContainer(cfg *config.Config) *diContainer {
	return &diContainer{cfg: cfg}
}

func (a *App) init(ctx context.Context) error {
	steps := []func(context.Context) error{
		a.initLogger,
		a.initCloser,
		a.initTracer,
		a.initPostgres,
		a.initRedis,
		a.initDomain,
		a.initNATS,
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
	a.log = logger.New(logger.DefaultConfig())
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
			semconv.ServiceNameKey.String("leaderboard"),
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
	dbURL := "postgres://" + a.cfg.DBUser + ":" + a.cfg.DBPassword + "@" + a.cfg.DBHost + ":" + a.cfg.DBPort + "/" + a.cfg.DBName
	poolCfg, err := pgxpool.ParseConfig(dbURL)
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
	if err := migrator.Up(stdlib.OpenDBFromPool(dbpool), migrations.FS); err != nil {
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

func (a *App) initRedis(ctx context.Context) error {
	redisRepo := repository.NewRedisLeaderboardRepo(a.cfg.RedisAddr, a.cfg.RedisDB, a.di.dbpool)
	a.di.repo = redisRepo

	games := []string{"hexagon", "flappy", "memory", "towers"}
	for _, gameID := range games {
		a.log.Info("restoring leaderboard", "game", gameID)
		if err := redisRepo.RestoreFromPostgres(ctx, gameID); err != nil {
			a.log.Warn("failed to restore leaderboard", "game", gameID, "err", err)
		} else {
			a.log.Info("restored leaderboard", "game", gameID)
		}
	}
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	a.di.svc = service.NewLeaderboardService(a.di.repo)
	a.di.api = handler.NewLeaderboardHandler(a.di.svc)
	return nil
}

func (a *App) initNATS(_ context.Context) error {
	var nc *nats.Conn
	var lastErr error
	for i := 0; i < 30; i++ {
		nc, lastErr = nats.Connect(a.cfg.NATSUrl)
		if lastErr == nil {
			break
		}
		a.log.Warn("nats connect failed", "attempt", i+1, "err", lastErr)
		time.Sleep(1 * time.Second)
	}
	if lastErr != nil {
		return fmt.Errorf("nats connect after 30 attempts: %w", lastErr)
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

	ackCtx, ackCancel := context.WithCancel(context.Background())
	a.di.ackCancel = ackCancel
	a.closer.AddNamed("ack batcher", func(context.Context) error {
		ackCancel()
		return nil
	})

	ackChan := make(chan *nats.Msg, 1000)
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		defer ticker.Stop()
		batch := make([]*nats.Msg, 0, 100)
		flush := func() {
			for _, m := range batch {
				_ = m.Ack()
			}
			batch = batch[:0]
		}
		for {
			select {
			case <-ackCtx.Done():
				flush()
				return
			case msg := <-ackChan:
				batch = append(batch, msg)
				if len(batch) >= 100 {
					flush()
				}
			case <-ticker.C:
				if len(batch) > 0 {
					flush()
				}
			}
		}
	}()

	svc := a.di.svc
	_, err = js.Subscribe("score.updated", func(msg *nats.Msg) {
		var event ScoreEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			a.log.Error("failed to unmarshal score event", "err", err)
			return
		}
		a.log.Info("received score",
			"game", event.GameID,
			"user", event.UserID,
			"score", event.Score,
		)

		msgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := svc.SaveUserInfo(msgCtx, event.GameID, event.UserID, event.UserEmail, event.Nickname); err != nil {
			a.log.Error("failed to save user info", "err", err)
		}
		if err := svc.UpdateScoreOnly(msgCtx, event.GameID, event.UserID, event.UserEmail, event.Score); err != nil {
			a.log.Error("failed to update score", "err", err)
		}
		ackChan <- msg
	}, nats.Durable("leaderboard-durable"), nats.ManualAck())
	if err != nil {
		return fmt.Errorf("nats subscribe score.updated: %w", err)
	}
	a.log.Info("subscribed to nats", "subject", "score.updated")
	return nil
}

func (a *App) initGRPC(_ context.Context) error {
	lis, err := net.Listen("tcp", ":"+a.cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	a.listener = lis
	a.closer.AddNamed("tcp listener", func(context.Context) error {
		return lis.Close()
	})

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.Recovery(),
			interceptor.Logger(),
			interceptor.Validate(),
			otelgrpc.UnaryServerInterceptor(),
		),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	pb.RegisterLeaderboardServiceServer(grpcServer, a.di.api)
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
		_, _ = w.Write([]byte(`{"status":"ok","service":"leaderboard"}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		readyCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		w.Header().Set("Content-Type", "application/json")
		if err := dbpool.Ping(readyCtx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"degraded","reason":"postgres unreachable"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","service":"leaderboard"}`))
	})

	srv := &http.Server{Addr: ":" + a.cfg.MetricsPort, Handler: mux}
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
