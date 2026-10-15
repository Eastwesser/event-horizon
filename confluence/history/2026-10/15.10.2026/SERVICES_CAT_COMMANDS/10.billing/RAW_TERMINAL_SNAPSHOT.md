[denismatveev@c0der event_horizon]$ find services/billing -type f -name '*.go' | sort
echo "===== main ====="
cat services/billing/cmd/main.go
echo "===== app ====="
for f in services/billing/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/billing/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/billing/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/billing/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== converter ====="
for f in services/billing/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/billing/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
services/billing/cmd/main.go
services/billing/internal/app/app.go
services/billing/internal/app/di.go
services/billing/internal/config/config.go
services/billing/internal/config/interfaces.go
services/billing/internal/converter/currency.go
services/billing/internal/converter/currency_test.go
services/billing/internal/handler/grpc_handler.go
services/billing/internal/interceptor/logger.go
services/billing/internal/interceptor/recovery.go
services/billing/internal/interceptor/validate.go
services/billing/internal/model/balance.go
services/billing/internal/model/errors.go
services/billing/internal/repository/postgres_integration_test.go
services/billing/internal/repository/postgres_repo.go
services/billing/internal/repository/redis_repo.go
services/billing/internal/service/billing_batch_test.go
services/billing/internal/service/billing_ops_test.go
services/billing/internal/service/billing_service.go
services/billing/internal/service/billing_validation_test.go
services/billing/internal/worker/outbox_worker.go
services/billing/migrations/embed.go
services/billing/proto/billing_grpc.pb.go
services/billing/proto/billing.pb.go
services/billing/proto/validate.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/billing/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init billing app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("billing stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/billing/internal/app/app.go ---
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/billing/internal/config"
)

// App is the runnable billing service process (gRPC + metrics HTTP).
type App struct {
	cfg           config.ConfigProvider
	log           *slog.Logger
	closer        *closer.Closer
	di            *diContainer
	grpcServer    *grpc.Server
	metricsServer *http.Server
	listener      net.Listener
}

// New wires dependencies and prepares servers (does not Serve yet).
func New(ctx context.Context) (*App, error) {
	cfg := config.Load()
	a := &App{
		cfg: cfg,
		di:  newDiContainer(cfg),
	}
	if err := a.init(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

// RunUntilSignal serves gRPC until SIGINT/SIGTERM, then closes resources via closer.
func (a *App) RunUntilSignal(ctx context.Context) error {
	go func() {
		a.log.Info("billing gRPC listening", "port", a.cfg.GRPCConfig().Port())
		if err := a.grpcServer.Serve(a.listener); err != nil {
			a.log.Error("grpc serve stopped", "err", err)
		}
	}()
	if err := a.closer.WaitSignal(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

// Close triggers graceful shutdown (idempotent via closer).
func (a *App) Close(ctx context.Context) error {
	return a.closer.CloseAll(ctx)
}

// GRPCAddr returns the bound address (useful for tests).
func (a *App) GRPCAddr() string {
	if a.listener == nil {
		return ""
	}
	return a.listener.Addr().String()
}
--- services/billing/internal/app/di.go ---
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
===== handler =====
--- services/billing/internal/handler/grpc_handler.go ---
package handler

import (
    "context"
    "log"
    "strings"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    "github.com/Eastwesser/event-horizon/services/billing/internal/converter"
    "github.com/Eastwesser/event-horizon/services/billing/internal/repository"
    "github.com/Eastwesser/event-horizon/services/billing/internal/service"
    pb "github.com/Eastwesser/event-horizon/services/billing/proto"
)

type BillingHandler struct {
    pb.UnimplementedBillingServiceServer
    billingService service.BillingService
}

func NewBillingHandler(svc service.BillingService) *BillingHandler {
    return &BillingHandler{
        billingService: svc,
    }
}

func (h *BillingHandler) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id is required")
    }

    currency := converter.ProtoCurrencyToRepo(req.Currency)
    balance, err := h.billingService.GetBalance(ctx, req.UserId, currency)
    if err != nil {
        log.Printf("GetBalance error: %v", err)
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.GetBalanceResponse{
        UserId:   req.UserId,
        Currency: req.Currency,
        Balance:  int32(balance),
    }, nil
}

func (h *BillingHandler) GetAllBalances(ctx context.Context, req *pb.GetAllBalancesRequest) (*pb.GetAllBalancesResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id is required")
    }

    lamps, err := h.billingService.GetBalance(ctx, req.UserId, repository.Lamps)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    tickets, err := h.billingService.GetBalance(ctx, req.UserId, repository.Tickets)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.GetAllBalancesResponse{
        UserId: req.UserId,
        Balances: []*pb.BalanceEntry{
            {Currency: pb.CurrencyType_LAMPS, Balance: int32(lamps)},
            {Currency: pb.CurrencyType_TICKETS, Balance: int32(tickets)},
        },
    }, nil
}

func (h *BillingHandler) AddCurrency(ctx context.Context, req *pb.AddCurrencyRequest) (*pb.AddCurrencyResponse, error) {
    if req.UserId == "" || req.Amount <= 0 {
        return nil, status.Error(codes.InvalidArgument, "user_id and valid amount are required")
    }

    currency := converter.ProtoCurrencyToRepo(req.Currency)
    newBalance, err := h.billingService.AddCurrency(ctx, req.UserId, currency, int(req.Amount), req.Reason, req.ReferenceId)
    if err != nil {
        log.Printf("AddCurrency error: %v", err)
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.AddCurrencyResponse{
        Success:     true,
        NewBalance:  int32(newBalance),
        Message:     "currency added successfully",
    }, nil
}

func (h *BillingHandler) SpendCurrency(ctx context.Context, req *pb.SpendCurrencyRequest) (*pb.SpendCurrencyResponse, error) {
    if req.UserId == "" || req.Amount <= 0 {
        return nil, status.Error(codes.InvalidArgument, "user_id and valid amount are required")
    }

    currency := converter.ProtoCurrencyToRepo(req.Currency)
    newBalance, err := h.billingService.SpendCurrency(
        ctx, 
        req.UserId, 
        currency, 
        int(req.Amount), 
        req.Reason, 
        req.ReferenceId, 
        req.CheckOnly,
    )
    if err != nil {
        log.Printf("SpendCurrency error: %v", err)
        if strings.Contains(err.Error(), "insufficient balance") {
            return nil, status.Error(codes.FailedPrecondition, err.Error())
        }
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.SpendCurrencyResponse{
        Success:    true,
        NewBalance: int32(newBalance),
        Message:    "currency spent successfully",
    }, nil
}

func (h *BillingHandler) GetTransactionHistory(ctx context.Context, req *pb.GetTransactionHistoryRequest) (*pb.GetTransactionHistoryResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id is required")
    }

    currency := converter.ProtoCurrencyToRepo(req.Currency)
    limit := int(req.Limit)
    if limit <= 0 || limit > 100 {
        limit = 20
    }

    transactions, total, err := h.billingService.GetTransactionHistory(ctx, req.UserId, currency, limit, int(req.Offset))
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    pbTransactions := make([]*pb.Transaction, len(transactions))
    for i, t := range transactions {
        pbTransactions[i] = &pb.Transaction{
            Id:           t.ID,
            UserId:       t.UserID,
            Currency:     converter.RepoCurrencyToProto(t.Currency),
            Amount:       int32(t.Amount),
            BalanceAfter: int32(t.BalanceAfter),
            Reason:       t.Reason,
            ReferenceId:  t.ReferenceID,
            CreatedAt:    t.CreatedAt.Unix(),
        }
    }

    return &pb.GetTransactionHistoryResponse{
        Transactions: pbTransactions,
        Total:        int32(total),
    }, nil
}
===== service =====
--- services/billing/internal/service/billing_batch_test.go ---
package service

import (
	"context"
	"sync"
	"testing"

	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
)

func TestTransactionBatch_AddAndLen(t *testing.T) {
	b := NewTransactionBatch()
	if b.Len() != 0 {
		t.Fatalf("new batch len=%d", b.Len())
	}
	b.Add(repository.Transaction{UserID: "u1", Amount: 10})
	b.Add(repository.Transaction{UserID: "u2", Amount: 20})
	if b.Len() != 2 {
		t.Fatalf("len=%d", b.Len())
	}
}

func TestTransactionBatch_EmptyFlushNoOp(t *testing.T) {
	b := NewTransactionBatch()
	b.Flush(context.Background(), nil) // must not panic on empty batch
}

func TestTransactionBatch_ConcurrentAdd(t *testing.T) {
	b := NewTransactionBatch()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			b.Add(repository.Transaction{UserID: "u", Amount: n})
		}(i)
	}
	wg.Wait()
	if b.Len() != 50 {
		t.Fatalf("len=%d", b.Len())
	}
}

func TestAddCurrency_TableNonPositive(t *testing.T) {
	s := &billingService{}
	cases := []struct {
		name   string
		amount int
	}{
		{"zero", 0},
		{"negative_one", -1},
		{"large_negative", -999},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.AddCurrency(context.Background(), "u1", repository.Tickets, tc.amount, "reason", "ref")
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestSpendCurrency_TableNonPositive(t *testing.T) {
	s := &billingService{}
	cases := []struct {
		name   string
		amount int
	}{
		{"zero", 0},
		{"negative", -5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.SpendCurrency(context.Background(), "u1", repository.Tickets, tc.amount, "reason", "ref", false)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
--- services/billing/internal/service/billing_ops_test.go ---
package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
)

type mockBillingStore struct {
	balance     int
	addBal      int
	spendBal    int
	txs         []repository.Transaction
	txTotal     int
	lastLimit   int
	addCalls    int
	spendCalls  int
	deleted     bool
	err         error
	spendErr    error
}

func (m *mockBillingStore) GetBalance(context.Context, string, repository.CurrencyType) (int, error) {
	return m.balance, m.err
}
func (m *mockBillingStore) AddBalance(context.Context, string, repository.CurrencyType, int, string, string) (int, error) {
	m.addCalls++
	if m.err != nil {
		return 0, m.err
	}
	return m.addBal, nil
}
func (m *mockBillingStore) SpendBalance(context.Context, string, repository.CurrencyType, int, string, string) (int, error) {
	m.spendCalls++
	if m.spendErr != nil {
		return 0, m.spendErr
	}
	return m.spendBal, nil
}
func (m *mockBillingStore) GetTransactionHistory(_ context.Context, _ string, _ repository.CurrencyType, limit, _ int) ([]repository.Transaction, int, error) {
	m.lastLimit = limit
	return m.txs, m.txTotal, m.err
}
func (m *mockBillingStore) BatchInsertTransactions(context.Context, []repository.Transaction) error {
	return nil
}

type mockBillingCache struct {
	cached   int
	cacheErr error
	setCalls int
	delCalls int
}

func (m *mockBillingCache) GetBalance(context.Context, string, repository.CurrencyType) (int, error) {
	return m.cached, m.cacheErr
}
func (m *mockBillingCache) SetBalance(context.Context, string, repository.CurrencyType, int, time.Duration) error {
	m.setCalls++
	return nil
}
func (m *mockBillingCache) DeleteBalance(context.Context, string, repository.CurrencyType) error {
	m.delCalls++
	return nil
}

func TestGetBalance_CacheHit(t *testing.T) {
	pg := &mockBillingStore{balance: 999}
	cache := &mockBillingCache{cached: 42}
	bal, err := NewBillingService(pg, cache).GetBalance(context.Background(), "u1", repository.Tickets)
	if err != nil || bal != 42 || cache.setCalls != 0 {
		t.Fatalf("bal=%d err=%v sets=%d", bal, err, cache.setCalls)
	}
}

func TestGetBalance_CacheMissLoadsPG(t *testing.T) {
	pg := &mockBillingStore{balance: 77}
	cache := &mockBillingCache{cached: -1}
	bal, err := NewBillingService(pg, cache).GetBalance(context.Background(), "u1", repository.Lamps)
	if err != nil || bal != 77 || cache.setCalls != 1 {
		t.Fatalf("bal=%d err=%v sets=%d", bal, err, cache.setCalls)
	}
}

func TestAddCurrency_TicketsAndLamps(t *testing.T) {
	pg := &mockBillingStore{addBal: 110}
	cache := &mockBillingCache{cached: -1}
	svc := NewBillingService(pg, cache)
	bal, err := svc.AddCurrency(context.Background(), "u1", repository.Tickets, 10, "reward", "ref-1")
	if err != nil || bal != 110 || pg.addCalls != 1 || cache.delCalls != 1 {
		t.Fatalf("bal=%d err=%v", bal, err)
	}
	_, err = svc.AddCurrency(context.Background(), "u1", repository.Lamps, 0, "x", "r")
	if err == nil || !strings.Contains(err.Error(), "positive") {
		t.Fatalf("%v", err)
	}
}

func TestSpendCurrency_CheckOnlyAndReal(t *testing.T) {
	pg := &mockBillingStore{balance: 5, spendBal: 0}
	cache := &mockBillingCache{cached: -1}
	svc := NewBillingService(pg, cache)
	_, err := svc.SpendCurrency(context.Background(), "u1", repository.Tickets, 10, "buy", "r", true)
	if err == nil || !strings.Contains(err.Error(), "insufficient") {
		t.Fatalf("%v", err)
	}
	pg.balance = 50
	bal, err := svc.SpendCurrency(context.Background(), "u1", repository.Tickets, 10, "buy", "r", true)
	if err != nil || bal != 50 || pg.spendCalls != 0 {
		t.Fatalf("checkOnly bal=%d err=%v spends=%d", bal, err, pg.spendCalls)
	}
	bal, err = svc.SpendCurrency(context.Background(), "u1", repository.Tickets, 10, "buy", "r", false)
	if err != nil || bal != 0 || pg.spendCalls != 1 || cache.delCalls != 1 {
		t.Fatalf("spend bal=%d err=%v", bal, err)
	}
}

func TestSpendCurrency_RepoError(t *testing.T) {
	pg := &mockBillingStore{spendErr: errors.New("db")}
	_, err := NewBillingService(pg, &mockBillingCache{cached: -1}).
		SpendCurrency(context.Background(), "u1", repository.Tickets, 1, "x", "r", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetTransactionHistory_ClampsLimit(t *testing.T) {
	pg := &mockBillingStore{txs: []repository.Transaction{{Amount: 1}}, txTotal: 1}
	svc := NewBillingService(pg, &mockBillingCache{cached: -1})
	_, _, err := svc.GetTransactionHistory(context.Background(), "u1", repository.Tickets, 0, 0)
	if err != nil || pg.lastLimit != 20 {
		t.Fatalf("limit=%d err=%v", pg.lastLimit, err)
	}
	_, _, _ = svc.GetTransactionHistory(context.Background(), "u1", repository.Tickets, 200, 0)
	if pg.lastLimit != 20 {
		t.Fatalf("limit=%d", pg.lastLimit)
	}
	_, _, _ = svc.GetTransactionHistory(context.Background(), "u1", repository.Tickets, 50, 0)
	if pg.lastLimit != 50 {
		t.Fatalf("limit=%d", pg.lastLimit)
	}
}
--- services/billing/internal/service/billing_service.go ---
package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
)

type BillingService interface {
	GetBalance(ctx context.Context, userID string, currency repository.CurrencyType) (int, error)
	AddCurrency(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error)
	SpendCurrency(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string, checkOnly bool) (int, error)
	GetTransactionHistory(ctx context.Context, userID string, currency repository.CurrencyType, limit, offset int) ([]repository.Transaction, int, error)
}

// BillingStore is the PostgreSQL port.
type BillingStore interface {
	GetBalance(ctx context.Context, userID string, currency repository.CurrencyType) (int, error)
	AddBalance(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error)
	SpendBalance(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error)
	GetTransactionHistory(ctx context.Context, userID string, currency repository.CurrencyType, limit, offset int) ([]repository.Transaction, int, error)
	BatchInsertTransactions(ctx context.Context, transactions []repository.Transaction) error
}

// BillingCache is the Redis port (GetBalance returns -1 on miss).
type BillingCache interface {
	GetBalance(ctx context.Context, userID string, currency repository.CurrencyType) (int, error)
	SetBalance(ctx context.Context, userID string, currency repository.CurrencyType, balance int, ttl time.Duration) error
	DeleteBalance(ctx context.Context, userID string, currency repository.CurrencyType) error
}

type billingService struct {
	pgRepo    BillingStore
	redisRepo BillingCache
}

func NewBillingService(pgRepo BillingStore, redisRepo BillingCache) BillingService {
	return &billingService{
		pgRepo:    pgRepo,
		redisRepo: redisRepo,
	}
}

func (s *billingService) GetBalance(ctx context.Context, userID string, currency repository.CurrencyType) (int, error) {
	cached, err := s.redisRepo.GetBalance(ctx, userID, currency)
	if err != nil && err != redis.Nil {
		log.Printf("Redis get error: %v", err)
	}
	if cached >= 0 {
		return cached, nil
	}

	balance, err := s.pgRepo.GetBalance(ctx, userID, currency)
	if err != nil {
		return 0, err
	}

	_ = s.redisRepo.SetBalance(ctx, userID, currency, balance, 5*time.Minute)
	return balance, nil
}

func (s *billingService) AddCurrency(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error) {
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be positive")
	}

	newBalance, err := s.pgRepo.AddBalance(ctx, userID, currency, amount, reason, referenceID)
	if err != nil {
		return 0, err
	}

	_ = s.redisRepo.DeleteBalance(ctx, userID, currency)

	return newBalance, nil
}

func (s *billingService) SpendCurrency(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string, checkOnly bool) (int, error) {
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be positive")
	}

	if checkOnly {
		balance, err := s.pgRepo.GetBalance(ctx, userID, currency)
		if err != nil {
			return 0, err
		}
		if balance < amount {
			return 0, fmt.Errorf("insufficient balance: have %d, need %d", balance, amount)
		}
		return balance, nil
	}

	newBalance, err := s.pgRepo.SpendBalance(ctx, userID, currency, amount, reason, referenceID)
	if err != nil {
		return 0, err
	}

	_ = s.redisRepo.DeleteBalance(ctx, userID, currency)

	return newBalance, nil
}

func (s *billingService) GetTransactionHistory(ctx context.Context, userID string, currency repository.CurrencyType, limit, offset int) ([]repository.Transaction, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.pgRepo.GetTransactionHistory(ctx, userID, currency, limit, offset)
}

// TransactionBatch для групповой записи
type TransactionBatch struct {
	transactions []repository.Transaction
	mu           sync.Mutex
}

func NewTransactionBatch() *TransactionBatch {
	return &TransactionBatch{
		transactions: make([]repository.Transaction, 0, 100),
	}
}

func (b *TransactionBatch) Add(tx repository.Transaction) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.transactions = append(b.transactions, tx)
}

func (b *TransactionBatch) Flush(ctx context.Context, pgRepo BillingStore) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.transactions) == 0 {
		return
	}

	if err := pgRepo.BatchInsertTransactions(ctx, b.transactions); err != nil {
		log.Printf("Failed to batch insert transactions: %v", err)
	}

	b.transactions = b.transactions[:0]
}

func (b *TransactionBatch) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.transactions)
}
--- services/billing/internal/service/billing_validation_test.go ---
package service

import (
	"context"
	"testing"

	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
)

func TestAddCurrency_RejectsNonPositive(t *testing.T) {
	s := &billingService{} // pg/redis unused until amount validated
	for _, amount := range []int{0, -1, -100} {
		_, err := s.AddCurrency(context.Background(), "u1", repository.Tickets, amount, "x", "ref")
		if err == nil {
			t.Fatalf("amount=%d: want error", amount)
		}
	}
}

func TestSpendCurrency_RejectsNonPositive(t *testing.T) {
	s := &billingService{}
	for _, amount := range []int{0, -5} {
		_, err := s.SpendCurrency(context.Background(), "u1", repository.Tickets, amount, "x", "ref", false)
		if err == nil {
			t.Fatalf("amount=%d: want error", amount)
		}
	}
}
===== repository =====
--- services/billing/internal/repository/postgres_integration_test.go ---
//go:build integration

package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
	"github.com/Eastwesser/event-horizon/services/billing/migrations"
)

func billingPool(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	if dsn := os.Getenv("BILLING_TEST_DATABASE_URL"); dsn != "" {
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			t.Fatalf("connect env dsn: %v", err)
		}
		if err := migrator.Up(stdlib.OpenDBFromPool(pool), migrations.FS); err != nil {
			pool.Close()
			t.Fatalf("migrate: %v", err)
		}
		return pool, func() { pool.Close() }
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		Env:          map[string]string{"POSTGRES_PASSWORD": "test", "POSTGRES_DB": "billing_test"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("testcontainers postgres unavailable (Docker?): %v", err)
	}
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/billing_test?sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatalf("pool: %v", err)
	}
	// wait ready
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if err := pool.Ping(ctx); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err := migrator.Up(stdlib.OpenDBFromPool(pool), migrations.FS); err != nil {
		pool.Close()
		_ = c.Terminate(ctx)
		t.Fatalf("migrate: %v", err)
	}
	cleanup := func() {
		pool.Close()
		_ = c.Terminate(ctx)
	}
	return pool, cleanup
}

func TestBilling_AddSpend_OutboxAndOptimisticVersion(t *testing.T) {
	pool, cleanup := billingPool(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresBillingRepo(pool)
	userID := uuid.NewString()

	bal, err := repo.AddBalance(ctx, userID, repository.Tickets, 100, "test_credit", "ref-add-1")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if bal != 100 {
		t.Fatalf("balance=%d want 100", bal)
	}

	var version int
	if err := pool.QueryRow(ctx,
		`SELECT version FROM user_currencies WHERE user_id=$1 AND currency_type=$2`,
		userID, repository.Tickets,
	).Scan(&version); err != nil {
		t.Fatalf("version: %v", err)
	}
	if version < 1 {
		t.Fatalf("version=%d", version)
	}

	var outbox int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM outbox WHERE processed=false`).Scan(&outbox); err != nil {
		t.Fatalf("outbox: %v", err)
	}
	if outbox < 1 {
		t.Fatal("expected outbox row after AddBalance")
	}

	bal, err = repo.SpendBalance(ctx, userID, repository.Tickets, 40, "test_spend", "ref-spend-1")
	if err != nil {
		t.Fatalf("spend: %v", err)
	}
	if bal != 60 {
		t.Fatalf("balance=%d want 60", bal)
	}

	_, err = repo.SpendBalance(ctx, userID, repository.Tickets, 1000, "over", "ref-over")
	if err == nil {
		t.Fatal("expected insufficient balance error")
	}
}

func TestBilling_ConcurrentSpend_NoNegative(t *testing.T) {
	pool, cleanup := billingPool(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresBillingRepo(pool)
	userID := uuid.NewString()
	if _, err := repo.AddBalance(ctx, userID, repository.Tickets, 50, "seed", "ref-seed"); err != nil {
		t.Fatal(err)
	}

	errCh := make(chan error, 2)
	for i := 0; i < 2; i++ {
		i := i
		go func() {
			_, err := repo.SpendBalance(ctx, userID, repository.Tickets, 40, "race", fmt.Sprintf("ref-race-%d", i))
			errCh <- err
		}()
	}
	var fails int
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			fails++
		}
	}
	if fails != 1 {
		t.Fatalf("exactly one spend should fail under race, fails=%d", fails)
	}
	bal, err := repo.GetBalance(ctx, userID, repository.Tickets)
	if err != nil && err != sql.ErrNoRows {
		// GetBalance returns 0,nil on no rows; pgx may differ — just read
	}
	_ = err
	if bal < 0 {
		t.Fatalf("negative balance %d", bal)
	}
	if bal != 10 {
		// 50-40=10 if one succeeded
		got, _ := repo.GetBalance(ctx, userID, repository.Tickets)
		if got != 10 {
			t.Fatalf("balance=%d want 10", got)
		}
	}
}
--- services/billing/internal/repository/postgres_repo.go ---
package repository

import (
    "context"
	"errors"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"
	
	"github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type CurrencyType string

const (
    Lamps   CurrencyType = "lamps"
    Tickets CurrencyType = "tickets"
)

type Balance struct {
    UserID     string
    Currency   CurrencyType
    Balance    int
    UpdatedAt  time.Time
}

type Transaction struct {
    ID           string
    UserID       string
    Currency     CurrencyType
    Amount       int
    BalanceAfter int
    Reason       string
    ReferenceID  string
    CreatedAt    time.Time
}

type BillingRepository interface {
    GetBalance(ctx context.Context, userID string, currency CurrencyType) (int, error)
    AddBalance(ctx context.Context, userID string, currency CurrencyType, amount int, reason, referenceID string) (int, error)
    SpendBalance(ctx context.Context, userID string, currency CurrencyType, amount int, reason, referenceID string) (int, error)
    GetTransactionHistory(ctx context.Context, userID string, currency CurrencyType, limit, offset int) ([]Transaction, int, error)
}

type PostgresBillingRepo struct {
    db *pgxpool.Pool
}

func NewPostgresBillingRepo(db *pgxpool.Pool) *PostgresBillingRepo {
    return &PostgresBillingRepo{db: db}
}

func (r *PostgresBillingRepo) GetBalance(ctx context.Context, userID string, currency CurrencyType) (int, error) {
    var balance int
    query := `SELECT balance FROM user_currencies WHERE user_id = $1 AND currency_type = $2`
    err := r.db.QueryRow(ctx, query, userID, currency).Scan(&balance)
    if err != nil {
        if err == sql.ErrNoRows {
            return 0, nil // новый пользователь, баланс 0
        }
        return 0, err
    }
    return balance, nil
}

// AddBalance credits currency atomically via an upsert-increment (INSERT ... ON CONFLICT
// DO UPDATE SET balance = balance + amount), which avoids the classic
// SELECT-then-UPDATE lost-update race that the old implementation had: two concurrent
// credits could both read the same stale balance and one would silently overwrite the
// other. The balance change, transaction row, and outbox event are all written in one
// transaction so an event is never lost or emitted without the balance actually changing.
func (r *PostgresBillingRepo) AddBalance(ctx context.Context, userID string, currency CurrencyType, amount int, reason, referenceID string) (int, error) {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return 0, err
    }
    defer tx.Rollback(ctx)

    var newBalance int
    upsertQuery := `
        INSERT INTO user_currencies (user_id, currency_type, balance, version)
        VALUES ($1, $2, $3, 1)
        ON CONFLICT (user_id, currency_type)
        DO UPDATE SET
            balance = user_currencies.balance + EXCLUDED.balance,
            version = user_currencies.version + 1,
            updated_at = NOW()
        RETURNING balance
    `
    if err := tx.QueryRow(ctx, upsertQuery, userID, currency, amount).Scan(&newBalance); err != nil {
        return 0, err
    }

    txQuery := `INSERT INTO transactions (user_id, currency_type, amount, balance_after, reason, reference_id) 
                VALUES ($1, $2, $3, $4, $5, $6)`
    if _, err = tx.Exec(ctx, txQuery, userID, currency, amount, newBalance, reason, referenceID); err != nil {
        return 0, err
    }

    if err := insertOutboxEvent(ctx, tx, "balance.updated", balanceUpdatedEvent{
        UserID: userID, Currency: string(currency), Delta: amount, Balance: newBalance, Reason: reason,
    }); err != nil {
        return 0, err
    }

    if err := tx.Commit(ctx); err != nil {
        return 0, err
    }

    return newBalance, nil
}

// SpendBalance debits currency atomically via a conditional UPDATE
// (`SET balance = balance - $1 WHERE balance >= $1`). Postgres re-checks the WHERE
// clause against the row's current committed value when acquiring the row lock for the
// UPDATE, so this can never drive the balance negative or lose a concurrent update —
// unlike the previous SELECT-then-UPDATE approach.
func (r *PostgresBillingRepo) SpendBalance(ctx context.Context, userID string, currency CurrencyType, amount int, reason, referenceID string) (int, error) {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return 0, err
    }
    defer tx.Rollback(ctx)

    var newBalance int
    updateQuery := `
        UPDATE user_currencies
        SET balance = balance - $1, version = version + 1, updated_at = NOW()
        WHERE user_id = $2 AND currency_type = $3 AND balance >= $1
        RETURNING balance
    `
    err = tx.QueryRow(ctx, updateQuery, amount, userID, currency).Scan(&newBalance)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            var currentBalance int
            checkErr := tx.QueryRow(ctx,
                `SELECT balance FROM user_currencies WHERE user_id = $1 AND currency_type = $2`,
                userID, currency).Scan(&currentBalance)
            if checkErr != nil {
                return 0, fmt.Errorf("user not found or no balance: %w", checkErr)
            }
            return 0, fmt.Errorf("insufficient balance: have %d, need %d", currentBalance, amount)
        }
        return 0, err
    }

    txQuery := `INSERT INTO transactions (user_id, currency_type, amount, balance_after, reason, reference_id) 
                VALUES ($1, $2, $3, $4, $5, $6)`
    if _, err = tx.Exec(ctx, txQuery, userID, currency, -amount, newBalance, reason, referenceID); err != nil {
        return 0, err
    }

    if err := insertOutboxEvent(ctx, tx, "balance.updated", balanceUpdatedEvent{
        UserID: userID, Currency: string(currency), Delta: -amount, Balance: newBalance, Reason: reason,
    }); err != nil {
        return 0, err
    }

    if err := tx.Commit(ctx); err != nil {
        return 0, err
    }

    return newBalance, nil
}

type balanceUpdatedEvent struct {
    UserID   string `json:"user_id"`
    Currency string `json:"currency"`
    Delta    int    `json:"delta"`
    Balance  int    `json:"balance"`
    Reason   string `json:"reason"`
}

func insertOutboxEvent(ctx context.Context, tx pgx.Tx, eventType string, event any) error {
    payload, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("marshal outbox payload: %w", err)
    }
    _, err = tx.Exec(ctx, `INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`, eventType, payload)
    return err
}

func (r *PostgresBillingRepo) GetTransactionHistory(ctx context.Context, userID string, currency CurrencyType, limit, offset int) ([]Transaction, int, error) {
    var total int
    countQuery := `SELECT COUNT(*) FROM transactions WHERE user_id = $1 AND currency_type = $2`
    err := r.db.QueryRow(ctx, countQuery, userID, currency).Scan(&total)
    if err != nil {
        return nil, 0, err
    }

    rows, err := r.db.Query(ctx, `
        SELECT id, user_id, currency_type, amount, balance_after, reason, reference_id, created_at
        FROM transactions 
        WHERE user_id = $1 AND currency_type = $2
        ORDER BY created_at DESC
        LIMIT $3 OFFSET $4
    `, userID, currency, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var transactions []Transaction
    for rows.Next() {
        var t Transaction
        err := rows.Scan(&t.ID, &t.UserID, &t.Currency, &t.Amount, &t.BalanceAfter, &t.Reason, &t.ReferenceID, &t.CreatedAt)
        if err != nil {
            return nil, 0, err
        }
        transactions = append(transactions, t)
    }

    return transactions, total, nil
}

// BatchInsertTransactions массовая вставка транзакций
func (r *PostgresBillingRepo) BatchInsertTransactions(ctx context.Context, transactions []Transaction) error {
    if len(transactions) == 0 {
        return nil
    }
    
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    
    // Используем COPY для массовой вставки
    copyFrom := pgx.CopyFromSlice(len(transactions), func(i int) ([]any, error) {
        t := transactions[i]
        return []any{
            t.ID, t.UserID, t.Currency, t.Amount,
            t.BalanceAfter, t.Reason, t.ReferenceID, t.CreatedAt,
        }, nil
    })
    
    _, err = tx.CopyFrom(ctx, pgx.Identifier{"transactions"},
        []string{"id", "user_id", "currency_type", "amount", "balance_after", "reason", "reference_id", "created_at"},
        copyFrom)
    if err != nil {
        return err
    }
    
    return tx.Commit(ctx)
}
--- services/billing/internal/repository/redis_repo.go ---
package repository

import (
    "context"
    "fmt"
    "strconv"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisBillingRepo struct {
    client *redis.Client
}

func NewRedisBillingRepo(addr string, db int) *RedisBillingRepo {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
        DB:   db,
    })
    return &RedisBillingRepo{client: client}
}

func (r *RedisBillingRepo) GetBalance(ctx context.Context, userID string, currency CurrencyType) (int, error) {
    key := fmt.Sprintf("billing:%s:%s", userID, currency)
    val, err := r.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return -1, nil // not found in cache
    }
    if err != nil {
        return 0, err
    }
    return strconv.Atoi(val)
}

func (r *RedisBillingRepo) SetBalance(ctx context.Context, userID string, currency CurrencyType, balance int, ttl time.Duration) error {
    key := fmt.Sprintf("billing:%s:%s", userID, currency)
    return r.client.Set(ctx, key, balance, ttl).Err()
}

func (r *RedisBillingRepo) DeleteBalance(ctx context.Context, userID string, currency CurrencyType) error {
    key := fmt.Sprintf("billing:%s:%s", userID, currency)
    return r.client.Del(ctx, key).Err()
}
===== converter =====
--- services/billing/internal/converter/currency.go ---
package converter

import (
	"github.com/Eastwesser/event-horizon/services/billing/internal/model"
	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
	pb "github.com/Eastwesser/event-horizon/services/billing/proto"
)

func ProtoCurrencyToRepo(c pb.CurrencyType) repository.CurrencyType {
	switch c {
	case pb.CurrencyType_LAMPS:
		return repository.Lamps
	case pb.CurrencyType_TICKETS:
		return repository.Tickets
	default:
		return repository.Lamps
	}
}

func RepoCurrencyToProto(c repository.CurrencyType) pb.CurrencyType {
	switch c {
	case repository.Lamps:
		return pb.CurrencyType_LAMPS
	case repository.Tickets:
		return pb.CurrencyType_TICKETS
	default:
		return pb.CurrencyType_LAMPS
	}
}

func BalanceToProto(userID string, currency repository.CurrencyType, amount int32, updatedAt int64) *pb.GetBalanceResponse {
	return &pb.GetBalanceResponse{
		UserId:    userID,
		Currency:  RepoCurrencyToProto(currency),
		Balance:   amount,
		UpdatedAt: updatedAt,
	}
}

func DomainBalanceToEntry(b model.Balance) *pb.BalanceEntry {
	var c repository.CurrencyType
	switch b.Currency {
	case string(repository.Tickets):
		c = repository.Tickets
	default:
		c = repository.Lamps
	}
	return &pb.BalanceEntry{
		Currency:  RepoCurrencyToProto(c),
		Balance:   b.Amount,
		UpdatedAt: b.UpdatedAt,
	}
}
--- services/billing/internal/converter/currency_test.go ---
package converter

import (
	"testing"

	"github.com/Eastwesser/event-horizon/services/billing/internal/model"
	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
	pb "github.com/Eastwesser/event-horizon/services/billing/proto"
)

func TestCurrencyConverters(t *testing.T) {
	if ProtoCurrencyToRepo(pb.CurrencyType_TICKETS) != repository.Tickets {
		t.Fatal("tickets")
	}
	if ProtoCurrencyToRepo(pb.CurrencyType_LAMPS) != repository.Lamps {
		t.Fatal("lamps")
	}
	if ProtoCurrencyToRepo(pb.CurrencyType(99)) != repository.Lamps {
		t.Fatal("default proto")
	}
	if RepoCurrencyToProto(repository.Lamps) != pb.CurrencyType_LAMPS {
		t.Fatal("lamps repo")
	}
	if RepoCurrencyToProto(repository.Tickets) != pb.CurrencyType_TICKETS {
		t.Fatal("tickets repo")
	}
	if RepoCurrencyToProto(repository.CurrencyType("x")) != pb.CurrencyType_LAMPS {
		t.Fatal("default repo")
	}
}

func TestBalanceToProto(t *testing.T) {
	p := BalanceToProto("u1", repository.Tickets, 42, 100)
	if p.UserId != "u1" || p.Balance != 42 || p.Currency != pb.CurrencyType_TICKETS {
		t.Fatalf("%+v", p)
	}
}

func TestDomainBalanceToEntry(t *testing.T) {
	e := DomainBalanceToEntry(model.Balance{Currency: "tickets", Amount: 7, UpdatedAt: 1})
	if e.Balance != 7 || e.Currency != pb.CurrencyType_TICKETS {
		t.Fatalf("%+v", e)
	}
	e2 := DomainBalanceToEntry(model.Balance{Currency: "lamps", Amount: 3, UpdatedAt: 2})
	if e2.Currency != pb.CurrencyType_LAMPS {
		t.Fatalf("%+v", e2)
	}
}

===== worker =====
--- services/billing/internal/worker/outbox_worker.go ---
// services/billing/internal/worker/outbox_worker.go
package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

// OutboxWorker polls the billing outbox table and republishes any events that
// haven't made it to NATS yet, guaranteeing at-least-once delivery for
// balance.updated even if NATS was briefly unavailable when the balance changed.
type OutboxWorker struct {
	db *pgxpool.Pool
	js nats.JetStreamContext
}

func NewOutboxWorker(db *pgxpool.Pool, js nats.JetStreamContext) *OutboxWorker {
	return &OutboxWorker{db: db, js: js}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	rows, err := w.db.Query(ctx, `
        SELECT id, event_type, payload FROM outbox
        WHERE processed = false
        ORDER BY created_at
        LIMIT 100
    `)
	if err != nil {
		slog.Error("billing outbox: query", "err", err)
		return
	}
	defer rows.Close()

	type pending struct {
		id, eventType string
		payload       []byte
	}
	var batch []pending

	for rows.Next() {
		var p pending
		if err := rows.Scan(&p.id, &p.eventType, &p.payload); err != nil {
			slog.Error("billing outbox: scan", "err", err)
			continue
		}
		batch = append(batch, p)
	}
	rows.Close()

	for _, p := range batch {
		if _, err := w.js.Publish(p.eventType, p.payload); err != nil {
			slog.Error("billing outbox: publish", "event_type", p.eventType, "err", err)
			continue
		}

		if _, err := w.db.Exec(ctx, `
            UPDATE outbox SET processed = true, processed_at = NOW()
            WHERE id = $1
        `, p.id); err != nil {
			slog.Error("billing outbox: mark processed", "err", err)
		} else {
			slog.Info("billing outbox: published", "event_type", p.eventType, "id", p.id)
		}
	}
}
[denismatveev@c0der event_horizon]$ 
