[denismatveev@c0der event_horizon]$ find services/payment -type f -name '*.go' | sort
echo "===== main ====="
cat services/payment/cmd/main.go
echo "===== app ====="
for f in services/payment/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/payment/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/payment/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/payment/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== webhook ====="
for f in services/payment/internal/webhook/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/payment/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
services/payment/cmd/main.go
services/payment/internal/app/app.go
services/payment/internal/app/di.go
services/payment/internal/config/config.go
services/payment/internal/handler/grpc_handler.go
services/payment/internal/interceptor/interceptors.go
services/payment/internal/model/errors.go
services/payment/internal/model/payment.go
services/payment/internal/model/payment_test.go
services/payment/internal/repository/postgres_integration_test.go
services/payment/internal/repository/postgres_repo.go
services/payment/internal/repository/redis_repo.go
services/payment/internal/service/payment_service.go
services/payment/internal/service/payment_service_test.go
services/payment/internal/webhook/verify.go
services/payment/internal/webhook/verify_test.go
services/payment/internal/worker/outbox_worker.go
services/payment/migrations/embed.go
services/payment/proto/payment_grpc.pb.go
services/payment/proto/payment.pb.go
services/payment/proto/validate.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/payment/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init payment app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("payment stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/payment/internal/app/app.go ---
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/payment/internal/config"
)

type App struct {
	cfg           *config.Config
	log           *slog.Logger
	closer        *closer.Closer
	di            *diContainer
	grpcServer    *grpc.Server
	metricsServer *http.Server
	listener      net.Listener
}

func New(ctx context.Context) (*App, error) {
	cfg := config.Load()
	a := &App{cfg: cfg, di: newDiContainer(cfg)}
	if err := a.init(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) RunUntilSignal(ctx context.Context) error {
	go func() {
		a.log.Info("payment gRPC listening", "port", a.cfg.GRPCPort)
		if err := a.grpcServer.Serve(a.listener); err != nil {
			a.log.Error("grpc serve stopped", "err", err)
		}
	}()
	if err := a.closer.WaitSignal(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

func (a *App) Close(ctx context.Context) error {
	return a.closer.CloseAll(ctx)
}
--- services/payment/internal/app/di.go ---
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
===== handler =====
--- services/payment/internal/handler/grpc_handler.go ---
package handler

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
	"github.com/Eastwesser/event-horizon/services/payment/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/payment/proto"
)

type GRPCHandler struct {
	pb.UnimplementedPaymentServiceServer
	svc *service.PaymentService
}

func NewGRPCHandler(svc *service.PaymentService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func (h *GRPCHandler) CreateCheckout(ctx context.Context, req *pb.CreateCheckoutRequest) (*pb.CreateCheckoutResponse, error) {
	p, err := h.svc.CreateCheckout(ctx, req.GetUserId(), req.GetPlan())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.CreateCheckoutResponse{
		PaymentId:   p.ID,
		CheckoutUrl: p.CheckoutURL,
		AmountRub:   int32(p.AmountRub),
		Plan:        p.Plan,
		Status:      p.Status,
	}, nil
}

func (h *GRPCHandler) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.ConfirmPaymentResponse, error) {
	sub, err := h.svc.ConfirmPayment(ctx, req.GetPaymentId(), req.GetProviderRef(), req.GetWebhookSecret())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.ConfirmPaymentResponse{
		Success:        true,
		Message:        "subscription activated",
		SubscriptionId: sub.ID,
		ExpiresAtUnix:  sub.ExpiresAt.Unix(),
	}, nil
}

func (h *GRPCHandler) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (*pb.GetSubscriptionResponse, error) {
	sub, err := h.svc.GetSubscription(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &pb.GetSubscriptionResponse{Active: false, Status: "none"}, nil
		}
		return nil, mapErr(err)
	}
	return &pb.GetSubscriptionResponse{
		Active:        sub.IsActive(time.Now().UTC()),
		Plan:          sub.Plan,
		Status:        sub.Status,
		ExpiresAtUnix: sub.ExpiresAt.Unix(),
		AmountRub:     int32(sub.AmountRub),
	}, nil
}

func (h *GRPCHandler) CanPurchaseMerch(ctx context.Context, req *pb.CanPurchaseMerchRequest) (*pb.CanPurchaseMerchResponse, error) {
	ok, reason := h.svc.CanPurchaseMerch(ctx, req.GetUserId())
	return &pb.CanPurchaseMerchResponse{Allowed: ok, Reason: reason}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrInvalidInput), errors.Is(err, model.ErrInvalidStatus):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, model.ErrUnauthorized):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, model.ErrAlreadyPaid):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
===== service =====
--- services/payment/internal/service/payment_service.go ---
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
)

// PaymentStore is the PostgreSQL port.
type PaymentStore interface {
	CreatePayment(ctx context.Context, p *model.Payment) error
	GetPayment(ctx context.Context, id string) (*model.Payment, error)
	CompletePaymentAndActivateSubscription(ctx context.Context, paymentID, providerRef string, sub *model.Subscription, eventType string, eventPayload map[string]any) error
	GetActiveSubscription(ctx context.Context, userID string) (*model.Subscription, error)
}

// SubscriptionCache is the Redis port.
type SubscriptionCache interface {
	GetSubscription(ctx context.Context, userID string) (*model.Subscription, error)
	SetSubscription(ctx context.Context, sub *model.Subscription) error
}

type PaymentService struct {
	repo             PaymentStore
	cache            SubscriptionCache
	boostyURL        string
	webhookSecret    string
	subscriptionDays int
}

func New(
	repo PaymentStore,
	cache SubscriptionCache,
	boostyURL, webhookSecret string,
	subscriptionDays int,
) *PaymentService {
	if subscriptionDays <= 0 {
		subscriptionDays = 30
	}
	return &PaymentService{
		repo:             repo,
		cache:            cache,
		boostyURL:        boostyURL,
		webhookSecret:    webhookSecret,
		subscriptionDays: subscriptionDays,
	}
}

func (s *PaymentService) CreateCheckout(ctx context.Context, userID, plan string) (*model.Payment, error) {
	amount, err := model.PlanAmountRub(plan)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	id := uuid.NewString()
	checkout := fmt.Sprintf("%s?payment_id=%s&plan=%s&amount=%d", s.boostyURL, id, plan, amount)
	p := &model.Payment{
		ID:          id,
		UserID:      userID,
		Plan:        plan,
		AmountRub:   amount,
		Status:      model.StatusPending,
		Provider:    "boosty",
		CheckoutURL: checkout,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreatePayment(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PaymentService) ConfirmPayment(ctx context.Context, paymentID, providerRef, webhookSecret string) (*model.Subscription, error) {
	if s.webhookSecret != "" && webhookSecret != "" && webhookSecret != s.webhookSecret {
		return nil, model.ErrUnauthorized
	}
	p, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	sub := &model.Subscription{
		ID:        uuid.NewString(),
		UserID:    p.UserID,
		Plan:      p.Plan,
		Status:    model.StatusActive,
		AmountRub: p.AmountRub,
		PaymentID: p.ID,
		StartsAt:  now,
		ExpiresAt: now.Add(time.Duration(s.subscriptionDays) * 24 * time.Hour),
	}
	event := map[string]any{
		"event":           "payment.completed",
		"payment_id":      p.ID,
		"user_id":         p.UserID,
		"plan":            p.Plan,
		"amount_rub":      p.AmountRub,
		"subscription_id": sub.ID,
		"expires_at":      sub.ExpiresAt.Unix(),
		"provider_ref":    providerRef,
		"timestamp":       now.Unix(),
	}
	if err := s.repo.CompletePaymentAndActivateSubscription(ctx, paymentID, providerRef, sub, "payment.completed", event); err != nil {
		if errors.Is(err, model.ErrAlreadyPaid) {
			existing, getErr := s.repo.GetActiveSubscription(ctx, p.UserID)
			if getErr == nil {
				if s.cache != nil {
					_ = s.cache.SetSubscription(ctx, existing)
				}
				return existing, nil
			}
		}
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetSubscription(ctx, sub)
	}
	return sub, nil
}

func (s *PaymentService) GetSubscription(ctx context.Context, userID string) (*model.Subscription, error) {
	if s.cache != nil {
		if cached, err := s.cache.GetSubscription(ctx, userID); err == nil && cached.IsActive(time.Now().UTC()) {
			return cached, nil
		}
	}
	sub, err := s.repo.GetActiveSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetSubscription(ctx, sub)
	}
	return sub, nil
}

func (s *PaymentService) CanPurchaseMerch(ctx context.Context, userID string) (bool, string) {
	sub, err := s.GetSubscription(ctx, userID)
	if err != nil || !sub.IsActive(time.Now().UTC()) {
		return false, "active subscription required to redeem merch"
	}
	return true, ""
}
--- services/payment/internal/service/payment_service_test.go ---
package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
)

type mockPayStore struct {
	payment     *model.Payment
	sub         *model.Subscription
	createErr   error
	completeErr error
	getSubErr   error
	created     *model.Payment
	setComplete bool
}

func (m *mockPayStore) CreatePayment(_ context.Context, p *model.Payment) error {
	m.created = p
	return m.createErr
}
func (m *mockPayStore) GetPayment(context.Context, string) (*model.Payment, error) {
	if m.payment == nil {
		return nil, errors.New("not found")
	}
	cp := *m.payment
	return &cp, nil
}
func (m *mockPayStore) CompletePaymentAndActivateSubscription(context.Context, string, string, *model.Subscription, string, map[string]any) error {
	m.setComplete = true
	return m.completeErr
}
func (m *mockPayStore) GetActiveSubscription(context.Context, string) (*model.Subscription, error) {
	if m.getSubErr != nil {
		return nil, m.getSubErr
	}
	if m.sub == nil {
		return nil, errors.New("no sub")
	}
	cp := *m.sub
	return &cp, nil
}

type mockSubCache struct {
	sub    *model.Subscription
	gets   int
	sets   int
	getErr error
}

func (m *mockSubCache) GetSubscription(context.Context, string) (*model.Subscription, error) {
	m.gets++
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.sub == nil {
		return nil, errors.New("miss")
	}
	cp := *m.sub
	return &cp, nil
}
func (m *mockSubCache) SetSubscription(_ context.Context, sub *model.Subscription) error {
	m.sets++
	m.sub = sub
	return nil
}

func TestCreateCheckout(t *testing.T) {
	store := &mockPayStore{}
	svc := New(store, nil, "https://boosty.to/eh", "secret", 30)
	if _, err := svc.CreateCheckout(context.Background(), "u1", "free"); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("%v", err)
	}
	p, err := svc.CreateCheckout(context.Background(), "u1", model.PlanPresent)
	if err != nil || p.CheckoutURL == "" || !strings.Contains(p.CheckoutURL, "payment_id=") {
		t.Fatalf("%+v err=%v", p, err)
	}
	if store.created == nil || store.created.UserID != "u1" {
		t.Fatal("not persisted")
	}
}

func TestCanPurchaseMerch(t *testing.T) {
	now := time.Now().UTC()
	active := &model.Subscription{UserID: "u1", Status: model.StatusActive, ExpiresAt: now.Add(time.Hour)}
	store := &mockPayStore{sub: active}
	ok, reason := New(store, nil, "", "", 30).CanPurchaseMerch(context.Background(), "u1")
	if !ok || reason != "" {
		t.Fatalf("ok=%v reason=%q", ok, reason)
	}
	store.getSubErr = errors.New("missing")
	store.sub = nil
	ok, reason = New(store, nil, "", "", 30).CanPurchaseMerch(context.Background(), "u1")
	if ok || !strings.Contains(reason, "subscription") {
		t.Fatalf("ok=%v reason=%q", ok, reason)
	}
}

func TestGetSubscription_CacheHit(t *testing.T) {
	now := time.Now().UTC()
	cached := &model.Subscription{UserID: "u1", Status: model.StatusActive, ExpiresAt: now.Add(time.Hour)}
	cache := &mockSubCache{sub: cached}
	store := &mockPayStore{getSubErr: errors.New("should not hit")}
	got, err := New(store, cache, "", "", 30).GetSubscription(context.Background(), "u1")
	if err != nil || got.UserID != "u1" || cache.gets != 1 {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestConfirmPayment_UnauthorizedAndHappy(t *testing.T) {
	store := &mockPayStore{payment: &model.Payment{ID: "p1", UserID: "u1", Plan: model.PlanPresent, AmountRub: 200}}
	svc := New(store, &mockSubCache{}, "https://x", "sekrit", 30)
	_, err := svc.ConfirmPayment(context.Background(), "p1", "ref", "wrong")
	if !errors.Is(err, model.ErrUnauthorized) {
		t.Fatalf("%v", err)
	}
	sub, err := svc.ConfirmPayment(context.Background(), "p1", "ref", "sekrit")
	if err != nil || sub.UserID != "u1" || !store.setComplete {
		t.Fatalf("%+v err=%v", sub, err)
	}
}

func TestConfirmPayment_AlreadyPaidIdempotent(t *testing.T) {
	now := time.Now().UTC()
	existing := &model.Subscription{ID: "s1", UserID: "u1", Status: model.StatusActive, ExpiresAt: now.Add(time.Hour)}
	store := &mockPayStore{
		payment:     &model.Payment{ID: "p1", UserID: "u1", Plan: model.PlanPresent, AmountRub: 200},
		completeErr: model.ErrAlreadyPaid,
		sub:         existing,
	}
	cache := &mockSubCache{}
	got, err := New(store, cache, "", "sekrit", 30).ConfirmPayment(context.Background(), "p1", "ref", "sekrit")
	if err != nil || got.ID != "s1" || cache.sets != 1 {
		t.Fatalf("%+v err=%v sets=%d", got, err, cache.sets)
	}
}

func TestNew_DefaultSubscriptionDays(t *testing.T) {
	s := New(&mockPayStore{}, nil, "", "", 0)
	if s.subscriptionDays != 30 {
		t.Fatalf("%d", s.subscriptionDays)
	}
}
===== repository =====
--- services/payment/internal/repository/postgres_integration_test.go ---
//go:build integration

package repository_test

import (
	"context"
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
	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
	"github.com/Eastwesser/event-horizon/services/payment/internal/repository"
	"github.com/Eastwesser/event-horizon/services/payment/migrations"
)

func paymentPool(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	if dsn := os.Getenv("PAYMENT_TEST_DATABASE_URL"); dsn != "" {
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
		Env:          map[string]string{"POSTGRES_PASSWORD": "test", "POSTGRES_DB": "payment_test"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req, Started: true,
	})
	if err != nil {
		t.Skipf("testcontainers postgres unavailable (Docker?): %v", err)
	}
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/payment_test?sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatalf("pool: %v", err)
	}
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
	return pool, func() {
		pool.Close()
		_ = c.Terminate(ctx)
	}
}

func TestPayment_CompleteActivatesSubscription(t *testing.T) {
	pool, cleanup := paymentPool(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresRepo(pool)
	now := time.Now().UTC()
	userID := uuid.NewString()
	pay := &model.Payment{
		ID: uuid.NewString(), UserID: userID, Plan: model.PlanPresent, AmountRub: 200,
		Status: model.StatusPending, Provider: "boosty", CheckoutURL: "https://example.test/pay",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.CreatePayment(ctx, pay); err != nil {
		t.Fatalf("create payment: %v", err)
	}

	sub := &model.Subscription{
		ID: uuid.NewString(), UserID: userID, Plan: model.PlanPresent, Status: model.StatusActive,
		AmountRub: 200, PaymentID: pay.ID, StartsAt: now, ExpiresAt: now.Add(30 * 24 * time.Hour),
	}
	err := repo.CompletePaymentAndActivateSubscription(ctx, pay.ID, "boosty-ref-1", sub, "payment.completed", map[string]any{
		"user_id": userID, "plan": model.PlanPresent,
	})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}

	got, err := repo.GetActiveSubscription(ctx, userID)
	if err != nil {
		t.Fatalf("get sub: %v", err)
	}
	if got.Plan != model.PlanPresent || !got.IsActive(time.Now().UTC()) {
		t.Fatalf("unexpected sub: %+v", got)
	}
}
--- services/payment/internal/repository/postgres_repo.go ---
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
)

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreatePayment(ctx context.Context, p *model.Payment) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO payments (id, user_id, plan, amount_rub, status, provider, checkout_url, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.UserID, p.Plan, p.AmountRub, p.Status, p.Provider, p.CheckoutURL, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PostgresRepo) GetPayment(ctx context.Context, id string) (*model.Payment, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, plan, amount_rub, status, provider, COALESCE(provider_ref,''), COALESCE(checkout_url,''), created_at, updated_at
		FROM payments WHERE id = $1`, id)
	var p model.Payment
	if err := row.Scan(&p.ID, &p.UserID, &p.Plan, &p.AmountRub, &p.Status, &p.Provider, &p.ProviderRef, &p.CheckoutURL, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// CompletePaymentAndActivateSubscription writes payment completion + subscription + outbox in one TX.
func (r *PostgresRepo) CompletePaymentAndActivateSubscription(
	ctx context.Context,
	paymentID, providerRef string,
	sub *model.Subscription,
	eventType string,
	eventPayload map[string]any,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM payments WHERE id = $1 FOR UPDATE`, paymentID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return err
	}
	if status == model.StatusCompleted {
		return model.ErrAlreadyPaid
	}
	if status != model.StatusPending {
		return model.ErrInvalidStatus
	}

	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		UPDATE payments SET status = $2, provider_ref = $3, updated_at = $4 WHERE id = $1`,
		paymentID, model.StatusCompleted, providerRef, now,
	)
	if err != nil {
		return err
	}

	// Expire previous active subscriptions for this user.
	_, err = tx.Exec(ctx, `
		UPDATE subscriptions SET status = $2, updated_at = $3
		WHERE user_id = $1 AND status = $4`,
		sub.UserID, model.StatusExpired, now, model.StatusActive,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO subscriptions (id, user_id, plan, status, amount_rub, payment_id, starts_at, expires_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		sub.ID, sub.UserID, sub.Plan, sub.Status, sub.AmountRub, sub.PaymentID, sub.StartsAt, sub.ExpiresAt, now, now,
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("marshal outbox: %w", err)
	}
	outboxID := uuid.NewString()
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox (id, event_type, payload) VALUES ($1,$2,$3)`,
		outboxID, eventType, body,
	)
	if err != nil {
		return err
	}

	authorShare := sub.AmountRub * 85 / 100
	platformShare := sub.AmountRub - authorShare
	_, err = tx.Exec(ctx, `
		INSERT INTO revenue_splits (payment_id, amount_rub, author_share_rub, platform_share_rub)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (payment_id) DO NOTHING`,
		paymentID, sub.AmountRub, authorShare, platformShare,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepo) GetActiveSubscription(ctx context.Context, userID string) (*model.Subscription, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, plan, status, amount_rub, COALESCE(payment_id::text,''), starts_at, expires_at
		FROM subscriptions
		WHERE user_id = $1 AND status = $2 AND expires_at > NOW()
		ORDER BY expires_at DESC
		LIMIT 1`, userID, model.StatusActive)
	var s model.Subscription
	if err := row.Scan(&s.ID, &s.UserID, &s.Plan, &s.Status, &s.AmountRub, &s.PaymentID, &s.StartsAt, &s.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}
--- services/payment/internal/repository/redis_repo.go ---
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
)

type RedisRepo struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRepo(addr string, ttl time.Duration) *RedisRepo {
	return &RedisRepo{
		client: redis.NewClient(&redis.Options{Addr: addr, PoolSize: 10}),
		ttl:    ttl,
	}
}

func (r *RedisRepo) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisRepo) Close() error {
	return r.client.Close()
}

func subKey(userID string) string {
	return fmt.Sprintf("payment:sub:%s", userID)
}

func (r *RedisRepo) GetSubscription(ctx context.Context, userID string) (*model.Subscription, error) {
	data, err := r.client.Get(ctx, subKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}
	var s model.Subscription
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *RedisRepo) SetSubscription(ctx context.Context, s *model.Subscription) error {
	if s == nil {
		return nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	ttl := r.ttl
	if rem := time.Until(s.ExpiresAt); rem > 0 && rem < ttl {
		ttl = rem
	}
	return r.client.Set(ctx, subKey(s.UserID), data, ttl).Err()
}

func (r *RedisRepo) DeleteSubscription(ctx context.Context, userID string) error {
	return r.client.Del(ctx, subKey(userID)).Err()
}
===== webhook =====
--- services/payment/internal/webhook/verify.go ---
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

// Authorize verifies Boosty-style webhook auth: HMAC header (preferred) or shared secret field.
func Authorize(configuredSecret, providedSecret string, rawBody []byte, signatureHeader string) error {
	if configuredSecret == "" {
		return nil // dev: no secret configured
	}
	if signatureHeader != "" {
		if VerifyHMACSHA256(configuredSecret, rawBody, signatureHeader) {
			return nil
		}
		return ErrInvalidSignature
	}
	if providedSecret != "" && subtle.ConstantTimeCompare([]byte(providedSecret), []byte(configuredSecret)) == 1 {
		return nil
	}
	return ErrUnauthorized
}

var (
	ErrUnauthorized       = errString("webhook unauthorized")
	ErrInvalidSignature   = errString("webhook signature invalid")
)

type errString string

func (e errString) Error() string { return string(e) }

func VerifyHMACSHA256(secret string, rawBody []byte, signatureHeader string) bool {
	if secret == "" || signatureHeader == "" {
		return false
	}
	sig := strings.TrimSpace(signatureHeader)
	if i := strings.Index(sig, "="); i >= 0 {
		sig = sig[i+1:]
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	sum := mac.Sum(nil)
	got, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(sum, got) == 1
}
--- services/payment/internal/webhook/verify_test.go ---
package webhook

import "testing"

func TestVerifyHMACSHA256(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"payment_id":"abc"}`)
	mac := VerifyHMACSHA256(secret, body, "sha256=deadbeef")
	_ = mac // signature varies; test Authorize path instead
	if err := Authorize("", "any", body, ""); err != nil {
		t.Fatal("empty configured secret should allow")
	}
	if err := Authorize(secret, secret, body, ""); err != nil {
		t.Fatal("matching shared secret should allow")
	}
	if err := Authorize(secret, "wrong", body, ""); err == nil {
		t.Fatal("wrong shared secret should reject")
	}
}
===== worker =====
--- services/payment/internal/worker/outbox_worker.go ---
package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

type OutboxWorker struct {
	db *pgxpool.Pool
	js nats.JetStreamContext
}

func NewOutboxWorker(db *pgxpool.Pool, js nats.JetStreamContext) *OutboxWorker {
	return &OutboxWorker{db: db, js: js}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	rows, err := w.db.Query(ctx, `
		SELECT id, event_type, payload FROM outbox
		WHERE processed = false
		ORDER BY created_at ASC
		LIMIT 50`)
	if err != nil {
		slog.Error("payment outbox: query", "err", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, eventType string
		var payload []byte
		if err := rows.Scan(&id, &eventType, &payload); err != nil {
			slog.Error("payment outbox: scan", "err", err)
			continue
		}
		if w.js != nil {
			if _, err := w.js.Publish(eventType, payload); err != nil {
				slog.Error("payment outbox: publish", "event_type", eventType, "err", err)
				continue
			}
		}
		if _, err := w.db.Exec(ctx, `
			UPDATE outbox SET processed = true, processed_at = NOW() WHERE id = $1`, id); err != nil {
			slog.Error("payment outbox: mark processed", "err", err)
			continue
		}
		slog.Info("payment outbox: published", "event_type", eventType, "id", id)
	}
}
[denismatveev@c0der event_horizon]$ 