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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
	"github.com/Eastwesser/event-horizon/platform/pkg/tracing"
	"github.com/Eastwesser/event-horizon/services/auth/internal/config"
	"github.com/Eastwesser/event-horizon/services/auth/internal/handler"
	"github.com/Eastwesser/event-horizon/services/auth/internal/interceptor"
	jwtauth "github.com/Eastwesser/event-horizon/services/auth/internal/jwt"
	"github.com/Eastwesser/event-horizon/services/auth/internal/repository"
	"github.com/Eastwesser/event-horizon/services/auth/internal/service"
	"github.com/Eastwesser/event-horizon/services/auth/migrations"
	pb "github.com/Eastwesser/event-horizon/services/auth/proto"
)

type diContainer struct {
	cfg config.ConfigProvider

	dbpool    *pgxpool.Pool
	redisRepo *repository.RedisAuthRepo
	userRepo  repository.UserRepository
	svc       service.AuthService
	api       pb.AuthServiceServer
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
		ServiceName: "auth",
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

func (a *App) initRedis(ctx context.Context) error {
	rc := a.cfg.RedisConfig()
	redisRepo := repository.NewRedisAuthRepo(rc.Addr(), time.Duration(rc.UserCacheTTLMinutes())*time.Minute)
	if err := redisRepo.Ping(ctx); err != nil {
		a.log.Warn("redis unavailable; sessions/cache degraded", "addr", rc.Addr(), "err", err)
	} else {
		a.log.Info("redis connected", "addr", rc.Addr())
	}
	a.di.redisRepo = redisRepo
	a.closer.AddNamed("redis", func(context.Context) error {
		return redisRepo.Close()
	})
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	jwtCfg := a.cfg.JWTConfig()
	accessTTL := time.Duration(jwtCfg.AccessMinutes()) * time.Minute
	refreshTTL := time.Duration(jwtCfg.RefreshDays()) * 24 * time.Hour
	tokens := jwtauth.NewManager(jwtCfg.Secret(), accessTTL, refreshTTL)

	a.di.userRepo = repository.NewPostgresUserRepo(a.di.dbpool)
	a.di.svc = service.NewAuthService(a.di.userRepo, a.di.redisRepo, tokens)
	a.di.api = handler.NewAuthHandler(a.di.svc)
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
			metrics.UnaryServerInterceptor("auth"),
		),
	)
	pb.RegisterAuthServiceServer(grpcServer, a.di.api)
	reflection.Register(grpcServer)

	a.grpcServer = grpcServer
	metrics.SetHealthy("auth")
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
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "auth"})
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
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "auth"})
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
