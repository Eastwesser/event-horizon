[denismatveev@c0der event_horizon]$ find services/analytics -type f -name '*.go' | sort
echo "===== main ====="
cat services/analytics/cmd/main.go
echo "===== app ====="
for f in services/analytics/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/analytics/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/analytics/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/analytics/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/analytics/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
services/analytics/cmd/main.go
services/analytics/internal/app/app.go
services/analytics/internal/app/di.go
services/analytics/internal/config/config.go
services/analytics/internal/handler/grpc_handler.go
services/analytics/internal/interceptor/interceptors.go
services/analytics/internal/model/model.go
services/analytics/internal/repository/clickhouse/client.go
services/analytics/internal/repository/clickhouse/client_integration_test.go
services/analytics/internal/repository/clickhouse/client_test.go
services/analytics/internal/repository/repo.go
services/analytics/internal/service/analytics_service.go
services/analytics/internal/service/analytics_service_test.go
services/analytics/internal/worker/ingest.go
services/analytics/proto/analytics_grpc.pb.go
services/analytics/proto/analytics.pb.go
services/analytics/proto/validate.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init analytics app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("analytics stopped: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/analytics/internal/app/app.go ---
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/analytics/internal/config"
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
		a.log.Info("analytics gRPC listening", "port", a.cfg.GRPCPort)
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
--- services/analytics/internal/app/di.go ---
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
===== handler =====
--- services/analytics/internal/handler/grpc_handler.go ---
package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/model"
	"github.com/Eastwesser/event-horizon/services/analytics/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/analytics/proto"
)

type GRPCHandler struct {
	pb.UnimplementedAnalyticsServiceServer
	svc *service.AnalyticsService
}

func NewGRPCHandler(svc *service.AnalyticsService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func (h *GRPCHandler) RecordEvent(ctx context.Context, req *pb.RecordEventRequest) (*pb.RecordEventResponse, error) {
	if err := h.svc.RecordEvent(ctx, req.GetUserId(), req.GetEventType(), req.GetPayloadJson()); err != nil {
		return nil, mapErr(err)
	}
	return &pb.RecordEventResponse{Ok: true}, nil
}

func (h *GRPCHandler) GetDAU(ctx context.Context, req *pb.GetDAURequest) (*pb.GetDAUResponse, error) {
	days, err := h.svc.GetDAU(ctx, int(req.GetDays()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.DayCount, 0, len(days))
	for _, d := range days {
		out = append(out, &pb.DayCount{Day: d.Day, Count: d.Count})
	}
	return &pb.GetDAUResponse{Days: out}, nil
}

func (h *GRPCHandler) GetMAU(ctx context.Context, req *pb.GetMAURequest) (*pb.GetMAUResponse, error) {
	mau, window, err := h.svc.GetMAU(ctx, int(req.GetDays()))
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.GetMAUResponse{Mau: mau, WindowDays: int32(window)}, nil
}

func (h *GRPCHandler) GetRetention(ctx context.Context, req *pb.GetRetentionRequest) (*pb.GetRetentionResponse, error) {
	ret, err := h.svc.GetRetention(ctx, int(req.GetCohortDaysAgo()), int(req.GetWindowDays()))
	if err != nil {
		return nil, mapErr(err)
	}
	points := make([]*pb.RetentionPoint, 0, len(ret.Points))
	for _, p := range ret.Points {
		points = append(points, &pb.RetentionPoint{DayN: p.DayN, Rate: p.Rate})
	}
	return &pb.GetRetentionResponse{
		CohortDay:  ret.CohortDay,
		CohortSize: ret.CohortSize,
		Points:     points,
	}, nil
}

func mapErr(err error) error {
	if errors.Is(err, model.ErrInvalidInput) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Errorf(codes.Internal, "%v", err)
}
===== service =====
--- services/analytics/internal/service/analytics_service.go ---
package service

import (
	"context"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/model"
)

// AnalyticsStore is the persistence port for AnalyticsService.
type AnalyticsStore interface {
	Record(ctx context.Context, userID, eventType, payload string) error
	DAU(ctx context.Context, days int) ([]model.DayCount, error)
	MAU(ctx context.Context, days int) (int64, error)
	Retention(ctx context.Context, cohortDaysAgo, windowDays int) (*model.Retention, error)
}

type AnalyticsService struct {
	repo AnalyticsStore
}

func New(repo AnalyticsStore) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) RecordEvent(ctx context.Context, userID, eventType, payload string) error {
	if eventType == "" {
		return model.ErrInvalidInput
	}
	if payload == "" {
		payload = "{}"
	}
	return s.repo.Record(ctx, userID, eventType, payload)
}

func (s *AnalyticsService) GetDAU(ctx context.Context, days int) ([]model.DayCount, error) {
	return s.repo.DAU(ctx, days)
}

func (s *AnalyticsService) GetMAU(ctx context.Context, days int) (int64, int, error) {
	if days <= 0 {
		days = 30
	}
	n, err := s.repo.MAU(ctx, days)
	return n, days, err
}

func (s *AnalyticsService) GetRetention(ctx context.Context, cohortDaysAgo, windowDays int) (*model.Retention, error) {
	return s.repo.Retention(ctx, cohortDaysAgo, windowDays)
}
--- services/analytics/internal/service/analytics_service_test.go ---
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/model"
)

type mockAnalytics struct {
	lastPayload string
	lastType    string
	lastDays    int
	dau         []model.DayCount
	mau         int64
	ret         *model.Retention
	err         error
}

func (m *mockAnalytics) Record(_ context.Context, _, eventType, payload string) error {
	m.lastType, m.lastPayload = eventType, payload
	return m.err
}
func (m *mockAnalytics) DAU(_ context.Context, days int) ([]model.DayCount, error) {
	m.lastDays = days
	return m.dau, m.err
}
func (m *mockAnalytics) MAU(_ context.Context, days int) (int64, error) {
	m.lastDays = days
	return m.mau, m.err
}
func (m *mockAnalytics) Retention(context.Context, int, int) (*model.Retention, error) {
	return m.ret, m.err
}

func TestRecordEvent(t *testing.T) {
	repo := &mockAnalytics{}
	svc := New(repo)
	if err := svc.RecordEvent(context.Background(), "u", "", "x"); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("%v", err)
	}
	if err := svc.RecordEvent(context.Background(), "u", "play", ""); err != nil || repo.lastPayload != "{}" {
		t.Fatalf("payload=%q err=%v", repo.lastPayload, err)
	}
}

func TestGetMAU_ClampsDays(t *testing.T) {
	repo := &mockAnalytics{mau: 42}
	n, days, err := New(repo).GetMAU(context.Background(), 0)
	if err != nil || n != 42 || days != 30 || repo.lastDays != 30 {
		t.Fatalf("n=%d days=%d last=%d err=%v", n, days, repo.lastDays, err)
	}
	n, days, err = New(repo).GetMAU(context.Background(), 14)
	if err != nil || days != 14 || repo.lastDays != 14 {
		t.Fatalf("days=%d err=%v", days, err)
	}
}

func TestGetDAUAndRetention(t *testing.T) {
	repo := &mockAnalytics{
		dau: []model.DayCount{{Day: "2026-10-01", Count: 5}},
		ret: &model.Retention{CohortSize: 10},
	}
	dau, err := New(repo).GetDAU(context.Background(), 7)
	if err != nil || len(dau) != 1 || repo.lastDays != 7 {
		t.Fatalf("%v err=%v", dau, err)
	}
	ret, err := New(repo).GetRetention(context.Background(), 7, 7)
	if err != nil || ret.CohortSize != 10 {
		t.Fatalf("%+v err=%v", ret, err)
	}
}
===== repository =====
--- services/analytics/internal/repository/repo.go ---
package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/model"
	ch "github.com/Eastwesser/event-horizon/services/analytics/internal/repository/clickhouse"
)

type AnalyticsRepo struct {
	ch *ch.Client
	db string
}

func New(client *ch.Client, db string) *AnalyticsRepo {
	return &AnalyticsRepo{ch: client, db: db}
}

func (r *AnalyticsRepo) Record(ctx context.Context, userID, eventType, payload string) error {
	return r.ch.InsertEvent(ctx, userID, eventType, payload, time.Now().UTC())
}

func (r *AnalyticsRepo) DAU(ctx context.Context, days int) ([]model.DayCount, error) {
	if days <= 0 {
		days = 30
	}
	sql := fmt.Sprintf(`
		SELECT toString(event_date), uniqExact(user_id)
		FROM %s.analytics_events
		WHERE event_date >= today() - {days:Int32} AND user_id != ''
		GROUP BY event_date
		ORDER BY event_date`, r.db)
	raw, err := r.ch.QueryTSV(ctx, sql, map[string]string{"days": ch.IntParam(days)})
	if err != nil {
		return nil, err
	}
	var out []model.DayCount
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		n, _ := strconv.ParseInt(parts[1], 10, 64)
		out = append(out, model.DayCount{Day: parts[0], Count: n})
	}
	return out, nil
}

func (r *AnalyticsRepo) MAU(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		days = 30
	}
	sql := fmt.Sprintf(`
		SELECT uniqExact(user_id)
		FROM %s.analytics_events
		WHERE event_date >= today() - {days:Int32} AND user_id != ''`, r.db)
	raw, err := r.ch.QueryTSV(ctx, sql, map[string]string{"days": ch.IntParam(days)})
	if err != nil {
		return 0, err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func (r *AnalyticsRepo) Retention(ctx context.Context, cohortDaysAgo, windowDays int) (*model.Retention, error) {
	if cohortDaysAgo <= 0 {
		cohortDaysAgo = 7
	}
	if windowDays <= 0 {
		windowDays = 7
	}
	cohortSQL := `SELECT toString(today() - {cohort_days_ago:Int32})`
	cohortDayRaw, err := r.ch.QueryTSV(ctx, cohortSQL, map[string]string{
		"cohort_days_ago": ch.IntParam(cohortDaysAgo),
	})
	if err != nil {
		return nil, err
	}
	cohortDay := strings.TrimSpace(cohortDayRaw)

	sizeSQL := fmt.Sprintf(`
		SELECT uniqExact(user_id) FROM %s.analytics_events
		WHERE event_date = toDate({cohort_day:String}) AND user_id != ''`, r.db)
	sizeRaw, err := r.ch.QueryTSV(ctx, sizeSQL, map[string]string{"cohort_day": cohortDay})
	if err != nil {
		return nil, err
	}
	cohortSize, _ := strconv.ParseInt(strings.TrimSpace(sizeRaw), 10, 64)

	pointsSQL := fmt.Sprintf(`
		SELECT dateDiff('day', toDate({cohort_day:String}), event_date) AS day_n, uniqExact(user_id)
		FROM %s.analytics_events
		WHERE user_id IN (
			SELECT user_id FROM %s.analytics_events
			WHERE event_date = toDate({cohort_day:String}) AND user_id != ''
		)
		AND event_date BETWEEN toDate({cohort_day:String}) AND toDate({cohort_day:String}) + {window_days:Int32}
		GROUP BY day_n
		ORDER BY day_n`, r.db, r.db)
	raw, err := r.ch.QueryTSV(ctx, pointsSQL, map[string]string{
		"cohort_day":  cohortDay,
		"window_days": ch.IntParam(windowDays),
	})
	if err != nil {
		return nil, err
	}

	byDay := map[int32]int64{}
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		d, _ := strconv.Atoi(parts[0])
		n, _ := strconv.ParseInt(parts[1], 10, 64)
		byDay[int32(d)] = n
	}

	points := make([]model.RetentionPoint, 0, windowDays+1)
	for i := 0; i <= windowDays; i++ {
		active := byDay[int32(i)]
		rate := 0.0
		if cohortSize > 0 {
			rate = float64(active) / float64(cohortSize)
		}
		points = append(points, model.RetentionPoint{DayN: int32(i), Rate: rate})
	}
	return &model.Retention{CohortDay: cohortDay, CohortSize: cohortSize, Points: points}, nil
}
===== worker =====
--- services/analytics/internal/worker/ingest.go ---
package worker

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/service"
)

var subjects = []string{
	"payment.completed",
	"author.upserted",
	"shop.purchased",
	"score.updated",
	"user.registered",
}

type IngestWorker struct {
	js  nats.JetStreamContext
	svc *service.AnalyticsService
}

func NewIngestWorker(js nats.JetStreamContext, svc *service.AnalyticsService) *IngestWorker {
	return &IngestWorker{js: js, svc: svc}
}

func (w *IngestWorker) Start(ctx context.Context) {
	if w.js == nil {
		return
	}
	for _, subj := range subjects {
		s := subj
		durable := "analytics-" + strings.ReplaceAll(s, ".", "-")
		_, err := w.js.Subscribe(s, func(msg *nats.Msg) {
			userID := ""
			var raw map[string]any
			if err := json.Unmarshal(msg.Data, &raw); err == nil {
				if v, ok := raw["user_id"].(string); ok {
					userID = v
				}
			}
			if err := w.svc.RecordEvent(context.Background(), userID, s, string(msg.Data)); err != nil {
				log.Printf("analytics ingest %s: %v", s, err)
				_ = msg.Nak()
				return
			}
			_ = msg.Ack()
		}, nats.Durable(durable), nats.ManualAck())
		if err != nil {
			log.Printf("analytics subscribe %s: %v", s, err)
		}
	}
	<-ctx.Done()
}
[denismatveev@c0der event_horizon]$ 
