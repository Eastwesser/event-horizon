[denismatveev@c0der event_horizon]$ find services/history -type f -name '*.go' | sort
echo "===== main ====="
cat services/history/cmd/main.go
echo "===== app ====="
for f in services/history/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/history/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/history/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/history/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/history/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== validate ====="
for f in services/history/internal/validate/*.go; do echo "--- $f ---"; cat "$f"; done
services/history/cmd/main.go
services/history/internal/app/app.go
services/history/internal/app/di.go
services/history/internal/config/config.go
services/history/internal/handler/grpc_handler.go
services/history/internal/interceptor/interceptors.go
services/history/internal/model/event.go
services/history/internal/repository/postgres_integration_test.go
services/history/internal/repository/repo.go
services/history/internal/service/history_service.go
services/history/internal/service/history_service_test.go
services/history/internal/validate/validate_test.go
services/history/internal/worker/workers.go
services/history/migrations/embed.go
services/history/proto/history_grpc.pb.go
services/history/proto/history.pb.go
services/history/proto/validate.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/history/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init history app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("history stopped: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/history/internal/app/app.go ---
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/history/internal/config"
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
		a.log.Info("history gRPC listening", "port", a.cfg.GRPCPort)
		if err := a.grpcServer.Serve(a.listener); err != nil {
			a.log.Error("grpc serve stopped", "err", err)
		}
	}()
	if err := a.closer.WaitSignal(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

func (a *App) Close(ctx context.Context) error { return a.closer.CloseAll(ctx) }
--- services/history/internal/app/di.go ---
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
	"github.com/Eastwesser/event-horizon/services/history/internal/config"
	"github.com/Eastwesser/event-horizon/services/history/internal/handler"
	"github.com/Eastwesser/event-horizon/services/history/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/history/internal/repository"
	"github.com/Eastwesser/event-horizon/services/history/internal/service"
	"github.com/Eastwesser/event-horizon/services/history/internal/worker"
	"github.com/Eastwesser/event-horizon/services/history/migrations"
	pb "github.com/Eastwesser/event-horizon/services/history/proto"
)

type diContainer struct {
	cfg    *config.Config
	dbpool *pgxpool.Pool
	js     nats.JetStreamContext
	nc     *nats.Conn
	repo   *repository.PostgresRepo
	svc    *service.HistoryService
	api    pb.HistoryServiceServer
}

func newDiContainer(cfg *config.Config) *diContainer { return &diContainer{cfg: cfg} }

func (a *App) init(ctx context.Context) error {
	for _, step := range []func(context.Context) error{
		a.initLogger, a.initCloser, a.initPostgres, a.initNATS, a.initDomain, a.initWorkers, a.initGRPC, a.initMetricsHTTP,
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

func (a *App) initPostgres(ctx context.Context) error {
	poolCfg, err := pgxpool.ParseConfig(a.cfg.DSN())
	if err != nil {
		return err
	}
	poolCfg.MaxConns, poolCfg.MinConns, poolCfg.MaxConnLifetime = 25, 10, 5*time.Minute
	db, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return err
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return err
	}
	if err := migrator.Up(stdlib.OpenDBFromPool(db), migrations.FS); err != nil {
		db.Close()
		return fmt.Errorf("migrations: %w", err)
	}
	a.di.dbpool = db
	a.closer.AddNamed("postgres", func(context.Context) error { db.Close(); return nil })
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
	a.di.repo = repository.NewPostgresRepo(a.di.dbpool)
	a.di.svc = service.New(a.di.repo, a.cfg.RetentionDays)
	a.di.api = handler.NewGRPCHandler(a.di.svc)
	return nil
}

func (a *App) initWorkers(_ context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	go worker.NewIngestWorker(a.di.js, a.di.svc).Start(ctx)
	go worker.NewRetentionWorker(a.di.svc).Start(ctx)
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
	pb.RegisterHistoryServiceServer(s, a.di.api)
	reflection.Register(s)
	a.grpcServer = s
	a.closer.AddNamed("grpc", func(context.Context) error { s.GracefulStop(); return nil })
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "history"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.di.dbpool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "history"})
	})
	srv := &http.Server{Addr: ":" + a.cfg.MetricsPort, Handler: mux}
	a.metricsServer = srv
	go func() { _ = srv.ListenAndServe() }()
	a.closer.AddNamed("metrics", func(ctx context.Context) error { return srv.Shutdown(ctx) })
	return nil
}
===== handler =====
--- services/history/internal/handler/grpc_handler.go ---
package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/history/internal/model"
	"github.com/Eastwesser/event-horizon/services/history/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/history/proto"
)

type GRPCHandler struct {
	pb.UnimplementedHistoryServiceServer
	svc *service.HistoryService
}

func NewGRPCHandler(svc *service.HistoryService) *GRPCHandler { return &GRPCHandler{svc: svc} }

func (h *GRPCHandler) RecordEvent(ctx context.Context, req *pb.RecordEventRequest) (*pb.RecordEventResponse, error) {
	id, err := h.svc.RecordEvent(ctx, req.GetUserId(), req.GetEventType(), req.GetPayloadJson())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.RecordEventResponse{EventId: id}, nil
}

func (h *GRPCHandler) ListEvents(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	list, total, err := h.svc.ListEvents(ctx, req.GetUserId(), req.GetEventType(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.HistoryEvent, 0, len(list))
	for _, e := range list {
		out = append(out, &pb.HistoryEvent{
			Id: e.ID, UserId: e.UserID, EventType: e.EventType,
			PayloadJson: e.Payload, CreatedAtUnix: e.CreatedAt.Unix(),
		})
	}
	return &pb.ListEventsResponse{Events: out, Total: total}, nil
}

func mapErr(err error) error {
	if errors.Is(err, model.ErrInvalidInput) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Errorf(codes.Internal, "%v", err)
}
===== service =====
--- services/history/internal/service/history_service.go ---
package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Eastwesser/event-horizon/services/history/internal/model"
)

// HistoryStore is the persistence port for HistoryService.
type HistoryStore interface {
	Insert(ctx context.Context, userID, eventType, payload string) (string, error)
	List(ctx context.Context, userID, eventType string, limit, offset int) ([]*model.Event, int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}

type HistoryService struct {
	repo          HistoryStore
	retentionDays int
}

func New(repo HistoryStore, retentionDays int) *HistoryService {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	return &HistoryService{repo: repo, retentionDays: retentionDays}
}

func (s *HistoryService) RecordEvent(ctx context.Context, userID, eventType, payload string) (string, error) {
	if eventType == "" {
		return "", model.ErrInvalidInput
	}
	if payload == "" {
		payload = "{}"
	}
	if !json.Valid([]byte(payload)) {
		b, _ := json.Marshal(map[string]string{"raw": payload})
		payload = string(b)
	}
	return s.repo.Insert(ctx, userID, eventType, payload)
}

func (s *HistoryService) ListEvents(ctx context.Context, userID, eventType string, limit, offset int) ([]*model.Event, int64, error) {
	return s.repo.List(ctx, userID, eventType, limit, offset)
}

func (s *HistoryService) PurgeExpired(ctx context.Context) (int64, error) {
	before := time.Now().UTC().AddDate(0, 0, -s.retentionDays)
	return s.repo.DeleteOlderThan(ctx, before)
}
--- services/history/internal/service/history_service_test.go ---
package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/history/internal/model"
)

type mockHistory struct {
	lastPayload string
	lastType    string
	lastUser    string
	before      time.Time
	insertID    string
	list        []*model.Event
	purged      int64
	err         error
}

func (m *mockHistory) Insert(_ context.Context, userID, eventType, payload string) (string, error) {
	m.lastUser, m.lastType, m.lastPayload = userID, eventType, payload
	if m.err != nil {
		return "", m.err
	}
	if m.insertID == "" {
		return "evt-1", nil
	}
	return m.insertID, nil
}
func (m *mockHistory) List(_ context.Context, userID, eventType string, _, _ int) ([]*model.Event, int64, error) {
	m.lastUser, m.lastType = userID, eventType
	return m.list, int64(len(m.list)), m.err
}
func (m *mockHistory) DeleteOlderThan(_ context.Context, before time.Time) (int64, error) {
	m.before = before
	return m.purged, m.err
}

func TestNew_DefaultRetention(t *testing.T) {
	s := New(&mockHistory{}, 0)
	if s.retentionDays != 30 {
		t.Fatalf("%d", s.retentionDays)
	}
}

func TestRecordEvent_ValidationAndNormalize(t *testing.T) {
	repo := &mockHistory{}
	svc := New(repo, 30)
	if _, err := svc.RecordEvent(context.Background(), "u1", "", "{}"); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("%v", err)
	}
	id, err := svc.RecordEvent(context.Background(), "u1", "login", "")
	if err != nil || id == "" || repo.lastPayload != "{}" {
		t.Fatalf("id=%s payload=%q err=%v", id, repo.lastPayload, err)
	}
	_, err = svc.RecordEvent(context.Background(), "u1", "x", "not-json")
	if err != nil {
		t.Fatal(err)
	}
	var wrapped map[string]string
	if err := json.Unmarshal([]byte(repo.lastPayload), &wrapped); err != nil || wrapped["raw"] != "not-json" {
		t.Fatalf("payload=%s", repo.lastPayload)
	}
}

func TestListEvents_Delegates(t *testing.T) {
	repo := &mockHistory{list: []*model.Event{{ID: "1", EventType: "login"}}}
	got, total, err := New(repo, 7).ListEvents(context.Background(), "u1", "login", 10, 0)
	if err != nil || total != 1 || got[0].ID != "1" {
		t.Fatalf("%v %d %v", got, total, err)
	}
}

func TestPurgeExpired_UsesRetention(t *testing.T) {
	repo := &mockHistory{purged: 3}
	n, err := New(repo, 7).PurgeExpired(context.Background())
	if err != nil || n != 3 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	want := time.Now().UTC().AddDate(0, 0, -7)
	if repo.before.Before(want.Add(-time.Minute)) || repo.before.After(want.Add(time.Minute)) {
		t.Fatalf("before=%v want~%v", repo.before, want)
	}
}
===== repository =====
--- services/history/internal/repository/postgres_integration_test.go ---
//go:build integration

package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/services/history/internal/repository"
	"github.com/Eastwesser/event-horizon/services/history/migrations"
)

func historyPool(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	if dsn := os.Getenv("HISTORY_TEST_DATABASE_URL"); dsn != "" {
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
		Env:          map[string]string{"POSTGRES_PASSWORD": "test", "POSTGRES_DB": "history_test"},
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
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/history_test?sslmode=disable", host, port.Port())
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

func TestHistory_InsertAndList(t *testing.T) {
	pool, cleanup := historyPool(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresRepo(pool)
	id, err := repo.Insert(ctx, "u1", "shop.purchased", `{"item":"hat"}`)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if id == "" {
		t.Fatal("empty id")
	}

	events, total, err := repo.List(ctx, "u1", "shop.purchased", 10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(events) != 1 || events[0].EventType != "shop.purchased" {
		t.Fatalf("total=%d events=%+v", total, events)
	}
}
--- services/history/internal/repository/repo.go ---
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Eastwesser/event-horizon/services/history/internal/model"
)

type PostgresRepo struct{ db *pgxpool.Pool }

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo { return &PostgresRepo{db: db} }

func (r *PostgresRepo) Insert(ctx context.Context, userID, eventType, payload string) (string, error) {
	id := uuid.NewString()
	_, err := r.db.Exec(ctx, `
		INSERT INTO events (id, user_id, event_type, payload, created_at)
		VALUES ($1,$2,$3,$4::jsonb,$5)`,
		id, userID, eventType, payload, time.Now().UTC())
	return id, err
}

func (r *PostgresRepo) List(ctx context.Context, userID, eventType string, limit, offset int) ([]*model.Event, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int64
	countQ := `SELECT COUNT(*) FROM events WHERE ($1 = '' OR user_id = $1) AND ($2 = '' OR event_type = $2)`
	if err := r.db.QueryRow(ctx, countQ, userID, eventType).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, event_type, payload::text, created_at FROM events
		WHERE ($1 = '' OR user_id = $1) AND ($2 = '' OR event_type = $2)
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`, userID, eventType, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []*model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.UserID, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, &e)
	}
	return out, total, nil
}

func (r *PostgresRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM events WHERE created_at < $1`, before)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
===== worker =====
--- services/history/internal/worker/workers.go ---
package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/Eastwesser/event-horizon/services/history/internal/service"
)

// NATS subjects mirrored into history trail.
// Gateway register publishes event.user.registered; store as user.registered for FE filters.
var subjects = []string{
	"payment.completed",
	"author.upserted",
	"shop.purchased",
	"score.updated",
	"user.registered",
	"event.user.registered",
}

type IngestWorker struct {
	js  nats.JetStreamContext
	svc *service.HistoryService
}

func NewIngestWorker(js nats.JetStreamContext, svc *service.HistoryService) *IngestWorker {
	return &IngestWorker{js: js, svc: svc}
}

func (w *IngestWorker) Start(ctx context.Context) {
	if w.js == nil {
		return
	}
	for _, subj := range subjects {
		s := subj
		_, err := w.js.Subscribe(s, func(msg *nats.Msg) {
			var raw map[string]any
			userID := ""
			if err := json.Unmarshal(msg.Data, &raw); err == nil {
				if v, ok := raw["user_id"].(string); ok {
					userID = v
				}
			}
			eventType := s
			if s == "event.user.registered" {
				eventType = "user.registered"
			}
			if _, err := w.svc.RecordEvent(context.Background(), userID, eventType, string(msg.Data)); err != nil {
				log.Printf("history ingest %s: %v", s, err)
				_ = msg.Nak()
				return
			}
			_ = msg.Ack()
		}, nats.Durable("history-"+s), nats.ManualAck())
		if err != nil {
			log.Printf("history subscribe %s: %v", s, err)
		}
	}
	<-ctx.Done()
}

type RetentionWorker struct {
	svc *service.HistoryService
}

func NewRetentionWorker(svc *service.HistoryService) *RetentionWorker {
	return &RetentionWorker{svc: svc}
}

func (w *RetentionWorker) Start(ctx context.Context) {
	t := time.NewTicker(1 * time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n, err := w.svc.PurgeExpired(ctx)
			if err != nil {
				log.Printf("history retention: %v", err)
			} else if n > 0 {
				log.Printf("history retention: deleted %d rows", n)
			}
		}
	}
}
===== validate =====
--- services/history/internal/validate/validate_test.go ---
package validate_test

import (
	"testing"

	historypb "github.com/Eastwesser/event-horizon/services/history/proto"
)

func TestRecordEventRequest_Validate(t *testing.T) {
	if err := (&historypb.RecordEventRequest{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
	if err := (&historypb.RecordEventRequest{EventType: "purchase.completed"}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestListEventsRequest_Validate(t *testing.T) {
	if err := (&historypb.ListEventsRequest{}).Validate(); err != nil {
		t.Fatal(err)
	}
}
[denismatveev@c0der event_horizon]$ 
