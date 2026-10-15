[denismatveev@c0der event_horizon]$ find services/leaderboard -type f -name '*.go' | sort
echo "===== main ====="
cat services/leaderboard/cmd/main.go
echo "===== app ====="
for f in services/leaderboard/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/leaderboard/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/leaderboard/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/leaderboard/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
services/leaderboard/cmd/main.go
services/leaderboard/internal/app/app.go
services/leaderboard/internal/app/di.go
services/leaderboard/internal/config/config.go
services/leaderboard/internal/converter/score.go
services/leaderboard/internal/converter/score_test.go
services/leaderboard/internal/handler/grpc_handler.go
services/leaderboard/internal/interceptor/logger.go
services/leaderboard/internal/interceptor/recovery.go
services/leaderboard/internal/interceptor/validate.go
services/leaderboard/internal/model/errors.go
services/leaderboard/internal/model/score.go
services/leaderboard/internal/repository/redis_repo.go
services/leaderboard/internal/service/leaderboard_service.go
services/leaderboard/internal/service/leaderboard_service_test.go
services/leaderboard/migrations/embed.go
services/leaderboard/proto/leaderboard_grpc.pb.go
services/leaderboard/proto/leaderboard.pb.go
services/leaderboard/proto/validate.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init leaderboard app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("leaderboard stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/leaderboard/internal/app/app.go ---
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/config"
)

// App is the runnable leaderboard service process (gRPC + metrics HTTP).
type App struct {
	cfg           *config.Config
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
		a.log.Info("leaderboard gRPC listening",
			"port", a.cfg.GRPCPort,
			"redis", a.cfg.RedisAddr,
			"nats", a.cfg.NATSUrl,
		)
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
--- services/leaderboard/internal/app/di.go ---
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
	Level     int    `json:"level"`
}

type nicknameUpdatedEvent struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}

var allGameIDs = []string{
	"hexagon", "flappy", "towers", "hanoi", "memory", "twenty48", "gears", "companion",
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

	games := []string{"hexagon", "flappy", "memory", "towers", "hanoi", "twenty48", "gears", "companion"}
	for _, gameID := range games {
		maxLevel := 1
		if gameID == "flappy" {
			maxLevel = 10
		}
		for level := 1; level <= maxLevel; level++ {
			a.log.Info("restoring leaderboard", "game", gameID, "level", level)
			if err := redisRepo.RestoreFromPostgres(ctx, gameID, level); err != nil {
				a.log.Warn("failed to restore leaderboard", "game", gameID, "level", level, "err", err)
			}
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
		level := event.Level
		if level < 1 {
			level = 1
		}
		a.log.Info("received score",
			"game", event.GameID,
			"user", event.UserID,
			"score", event.Score,
			"level", level,
		)

		msgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Notify previous #1 when a different player takes the lead.
		if top, topErr := svc.GetTopScores(msgCtx, event.GameID, 1, level); topErr == nil && len(top) > 0 {
			prev := top[0]
			if prev.UserID != "" && prev.UserID != event.UserID && event.Score > prev.Score {
				payload, _ := json.Marshal(map[string]any{
					"event":             "leaderboard.record_beaten",
					"beaten_user_id":    prev.UserID,
					"new_user_id":       event.UserID,
					"new_nickname":      event.Nickname,
					"game_id":           event.GameID,
					"level":             level,
					"new_score":         event.Score,
					"previous_score":    prev.Score,
				})
				if _, pubErr := js.Publish("event.leaderboard.record_beaten", payload); pubErr != nil {
					a.log.Error("publish record_beaten", "err", pubErr)
				}
			}
		}

		if err := svc.SaveUserInfo(msgCtx, event.GameID, event.UserID, event.UserEmail, event.Nickname, level); err != nil {
			a.log.Error("failed to save user info", "err", err)
		}
		if err := svc.UpdateScoreOnly(msgCtx, event.GameID, event.UserID, event.UserEmail, event.Score, level); err != nil {
			a.log.Error("failed to update score", "err", err)
		}
		ackChan <- msg
	}, nats.Durable("leaderboard-durable"), nats.ManualAck())
	if err != nil {
		return fmt.Errorf("nats subscribe score.updated: %w", err)
	}
	a.log.Info("subscribed to nats", "subject", "score.updated")

	_, err = js.Subscribe("event.user.nickname.updated", func(msg *nats.Msg) {
		var event nicknameUpdatedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil || event.UserID == "" || event.Nickname == "" {
			_ = msg.Ack()
			return
		}
		msgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, gameID := range allGameIDs {
			levels := []int{1}
			if gameID == "flappy" {
				levels = []int{1, 2, 3}
			}
			for _, level := range levels {
				if err := svc.SaveUserInfo(msgCtx, gameID, event.UserID, "", event.Nickname, level); err != nil {
					a.log.Error("nickname refresh", "game", gameID, "err", err)
				}
			}
		}
		_ = msg.Ack()
	}, nats.Durable("leaderboard-nickname"), nats.ManualAck())
	if err != nil {
		a.log.Error("nats subscribe nickname.updated", "err", err)
	} else {
		a.log.Info("subscribed to nats", "subject", "event.user.nickname.updated")
	}
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
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptor.Recovery(),
			interceptor.Logger(),
			interceptor.Validate(),
		),
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
===== handler =====
--- services/leaderboard/internal/handler/grpc_handler.go ---
package handler

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/leaderboard/proto"
)

type LeaderboardHandler struct {
	pb.UnimplementedLeaderboardServiceServer
	leaderboardService service.LeaderboardService
}

func NewLeaderboardHandler(svc service.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: svc,
	}
}

func (h *LeaderboardHandler) GetTopScores(ctx context.Context, req *pb.GetTopScoresRequest) (*pb.GetTopScoresResponse, error) {
	if req.GameId == "" {
		return nil, status.Error(codes.InvalidArgument, "game_id is required")
	}

	limit := int(req.Limit)
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	level := int(req.Level)

	entries, err := h.leaderboardService.GetTopScores(ctx, req.GameId, limit, level)
	if err != nil {
		log.Printf("Failed to get top scores: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbEntries := make([]*pb.ScoreEntry, 0, len(entries))
	for _, e := range entries {
		pbEntries = append(pbEntries, &pb.ScoreEntry{
			Rank:      int32(e.Rank),
			UserId:    e.UserID,
			UserEmail: e.UserEmail,
			Nickname:  e.Nickname,
			Score:     int32(e.Score),
			UpdatedAt: e.UpdatedAt,
			Level:     int32(e.Level),
		})
	}

	return &pb.GetTopScoresResponse{Entries: pbEntries}, nil
}

func (h *LeaderboardHandler) GetPlayerRank(ctx context.Context, req *pb.GetPlayerRankRequest) (*pb.GetPlayerRankResponse, error) {
	if req.GameId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "game_id and user_id are required")
	}

	rank, score, err := h.leaderboardService.GetPlayerRank(ctx, req.GameId, req.UserId, int(req.Level))
	if err != nil {
		log.Printf("Failed to get player rank: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetPlayerRankResponse{
		Rank:  int32(rank),
		Score: int32(score),
	}, nil
}

func (h *LeaderboardHandler) UpdateScore(ctx context.Context, req *pb.UpdateScoreRequest) (*pb.UpdateScoreResponse, error) {
	if req.GameId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "game_id and user_id are required")
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = req.UserEmail
		if idx := strings.Index(nickname, "@"); idx > 0 {
			nickname = nickname[:idx]
		}
	}

	newRank, err := h.leaderboardService.UpdateScore(
		ctx, req.GameId, req.UserId, req.UserEmail, nickname, int(req.Score), int(req.Level),
	)
	if err != nil {
		log.Printf("Failed to update score: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateScoreResponse{
		Success: true,
		NewRank: int32(newRank),
		Message: "score updated successfully",
	}, nil
}
===== service =====
--- services/leaderboard/internal/service/leaderboard_service.go ---
package service

import (
	"context"
	"fmt"

	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/repository"
)

type LeaderboardService interface {
	GetTopScores(ctx context.Context, gameID string, limit, level int) ([]repository.ScoreEntry, error)
	GetPlayerRank(ctx context.Context, gameID, userID string, level int) (int, int, error)
	UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score, level int) (int, error)
	UpdateScoreOnly(ctx context.Context, gameID, userID, userEmail string, score, level int) error
	SaveUserInfo(ctx context.Context, gameID, userID, userEmail, nickname string, level int) error
}

type leaderboardService struct {
	repo repository.LeaderboardRepository
}

func NewLeaderboardService(repo repository.LeaderboardRepository) LeaderboardService {
	return &leaderboardService{repo: repo}
}

func (s *leaderboardService) GetTopScores(ctx context.Context, gameID string, limit, level int) ([]repository.ScoreEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.GetTopScores(ctx, gameID, limit, level)
}

func (s *leaderboardService) GetPlayerRank(ctx context.Context, gameID, userID string, level int) (int, int, error) {
	return s.repo.GetPlayerRank(ctx, gameID, userID, level)
}

func (s *leaderboardService) UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score, level int) (int, error) {
	if gameID == "" || userID == "" {
		return 0, fmt.Errorf("game_id and user_id are required")
	}
	return s.repo.UpdateScore(ctx, gameID, userID, userEmail, nickname, score, level)
}

func (s *leaderboardService) UpdateScoreOnly(ctx context.Context, gameID, userID, userEmail string, score, level int) error {
	return s.repo.UpdateScoreOnly(ctx, gameID, userID, userEmail, score, level)
}

func (s *leaderboardService) SaveUserInfo(ctx context.Context, gameID, userID, userEmail, nickname string, level int) error {
	return s.repo.SaveUserInfo(ctx, gameID, userID, userEmail, nickname, level)
}
--- services/leaderboard/internal/service/leaderboard_service_test.go ---
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/repository"
)

type mockLB struct {
	top          []repository.ScoreEntry
	topLimit     int
	rank         int
	score        int
	updateRank   int
	lastGame     string
	lastUser     string
	updateCalled bool
	err          error
}

func (m *mockLB) GetTopScores(_ context.Context, gameID string, limit, _ int) ([]repository.ScoreEntry, error) {
	m.lastGame = gameID
	m.topLimit = limit
	return m.top, m.err
}
func (m *mockLB) GetPlayerRank(_ context.Context, gameID, userID string, _ int) (int, int, error) {
	m.lastGame, m.lastUser = gameID, userID
	return m.rank, m.score, m.err
}
func (m *mockLB) UpdateScore(_ context.Context, gameID, userID, _, _ string, score, _ int) (int, error) {
	m.updateCalled = true
	m.lastGame, m.lastUser = gameID, userID
	m.score = score
	return m.updateRank, m.err
}
func (m *mockLB) UpdateScoreOnly(context.Context, string, string, string, int, int) error {
	return m.err
}
func (m *mockLB) SaveUserInfo(context.Context, string, string, string, string, int) error {
	return m.err
}

func TestGetTopScores_ClampsLimit(t *testing.T) {
	repo := &mockLB{top: []repository.ScoreEntry{{UserID: "u1", Score: 10}}}
	svc := NewLeaderboardService(repo)
	got, err := svc.GetTopScores(context.Background(), "flappy", 0, 1)
	if err != nil || len(got) != 1 || repo.topLimit != 10 {
		t.Fatalf("got=%v limit=%d err=%v", got, repo.topLimit, err)
	}
	_, _ = svc.GetTopScores(context.Background(), "flappy", 500, 1)
	if repo.topLimit != 10 {
		t.Fatalf("want clamp to 10, got %d", repo.topLimit)
	}
	_, _ = svc.GetTopScores(context.Background(), "flappy", 25, 1)
	if repo.topLimit != 25 {
		t.Fatalf("want 25, got %d", repo.topLimit)
	}
}

func TestGetPlayerRank_Delegates(t *testing.T) {
	repo := &mockLB{rank: 3, score: 900}
	rank, score, err := NewLeaderboardService(repo).GetPlayerRank(context.Background(), "hexagon", "u1", 1)
	if err != nil || rank != 3 || score != 900 || repo.lastUser != "u1" {
		t.Fatalf("rank=%d score=%d err=%v", rank, score, err)
	}
}

func TestUpdateScore_Validation(t *testing.T) {
	svc := NewLeaderboardService(&mockLB{})
	if _, err := svc.UpdateScore(context.Background(), "", "u1", "", "", 1, 1); err == nil {
		t.Fatal("expected error")
	}
	if _, err := svc.UpdateScore(context.Background(), "g", "", "", "", 1, 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateScore_Success(t *testing.T) {
	repo := &mockLB{updateRank: 2}
	rank, err := NewLeaderboardService(repo).UpdateScore(context.Background(), "flappy", "u1", "a@b.c", "n", 50, 1)
	if err != nil || rank != 2 || !repo.updateCalled {
		t.Fatalf("rank=%d err=%v called=%v", rank, err, repo.updateCalled)
	}
}

func TestUpdateScoreOnlyAndSaveUserInfo(t *testing.T) {
	repo := &mockLB{}
	svc := NewLeaderboardService(repo)
	if err := svc.UpdateScoreOnly(context.Background(), "g", "u", "e", 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := svc.SaveUserInfo(context.Background(), "g", "u", "e", "n", 1); err != nil {
		t.Fatal(err)
	}
	repo.err = errors.New("fail")
	if err := svc.SaveUserInfo(context.Background(), "g", "u", "e", "n", 1); err == nil {
		t.Fatal("expected error")
	}
}
===== repository =====
--- services/leaderboard/internal/repository/redis_repo.go ---
package repository

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ScoreEntry struct {
	Rank      int
	UserID    string
	UserEmail string
	Nickname  string
	Score     int
	UpdatedAt int64
	Level     int
}

type LeaderboardRepository interface {
	UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score, level int) (int, error)
	UpdateScoreOnly(ctx context.Context, gameID, userID, userEmail string, score, level int) error
	GetTopScores(ctx context.Context, gameID string, limit, level int) ([]ScoreEntry, error)
	GetPlayerRank(ctx context.Context, gameID, userID string, level int) (int, int, error)
	SaveUserInfo(ctx context.Context, gameID, userID, userEmail, nickname string, level int) error
}

type RedisLeaderboardRepo struct {
	client *redis.Client
	db     *pgxpool.Pool
}

func NewRedisLeaderboardRepo(addr string, db int, dbPool *pgxpool.Pool) *RedisLeaderboardRepo {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})
	return &RedisLeaderboardRepo{
		client: client,
		db:     dbPool,
	}
}

func normalizeLevel(level int) int {
	if level < 1 {
		return 1
	}
	if level > 100 {
		return 100
	}
	return level
}

func boardKey(gameID string, level int) string {
	return fmt.Sprintf("leaderboard:%s:%d", gameID, normalizeLevel(level))
}

func infoKey(gameID string, level int) string {
	return boardKey(gameID, level) + ":info"
}

func emailKey(gameID string, level int) string {
	return boardKey(gameID, level) + ":emails"
}

func (r *RedisLeaderboardRepo) UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score, level int) (int, error) {
	level = normalizeLevel(level)
	key := boardKey(gameID, level)
	ik := infoKey(gameID, level)

	if userEmail != "" {
		r.client.HSet(ctx, ik, userID+"_email", userEmail)
	}
	if nickname != "" {
		r.client.HSet(ctx, ik, userID+"_nickname", nickname)
	} else if userEmail != "" {
		defaultNick := userEmail
		if idx := strings.Index(defaultNick, "@"); idx > 0 {
			defaultNick = defaultNick[:idx]
		}
		r.client.HSet(ctx, ik, userID+"_nickname", defaultNick)
	}

	if err := r.client.ZAddArgs(ctx, key, redis.ZAddArgs{
		GT:      true,
		Members: []redis.Z{{Score: float64(score), Member: userID}},
	}).Err(); err != nil {
		return 0, err
	}

	if r.db != nil {
		if _, err := r.db.Exec(ctx, `
            INSERT INTO leaderboard_backup (game_id, level, user_id, score, user_email, updated_at)
            VALUES ($1, $2, $3, $4, $5, NOW())
            ON CONFLICT (game_id, level, user_id) DO UPDATE
            SET score = GREATEST(leaderboard_backup.score, EXCLUDED.score),
                user_email = COALESCE(NULLIF(EXCLUDED.user_email, ''), leaderboard_backup.user_email),
                updated_at = NOW()
        `, gameID, level, userID, score, userEmail); err != nil {
			log.Printf("⚠️ Failed to persist leaderboard backup: %v", err)
		}
	}

	rank, err := r.client.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		return 0, err
	}

	return int(rank) + 1, nil
}

func (r *RedisLeaderboardRepo) UpdateScoreOnly(ctx context.Context, gameID, userID, userEmail string, score, level int) error {
	level = normalizeLevel(level)
	key := boardKey(gameID, level)
	ek := emailKey(gameID, level)

	if userEmail != "" {
		r.client.HSet(ctx, ek, userID, userEmail)
	}

	if err := r.client.ZAddArgs(ctx, key, redis.ZAddArgs{
		GT:      true,
		Members: []redis.Z{{Score: float64(score), Member: userID}},
	}).Err(); err != nil {
		return err
	}

	if r.db != nil {
		if _, err := r.db.Exec(ctx, `
            INSERT INTO leaderboard_backup (game_id, level, user_id, score, user_email, updated_at)
            VALUES ($1, $2, $3, $4, $5, NOW())
            ON CONFLICT (game_id, level, user_id) DO UPDATE
            SET score = GREATEST(leaderboard_backup.score, EXCLUDED.score),
                user_email = COALESCE(NULLIF(EXCLUDED.user_email, ''), leaderboard_backup.user_email),
                updated_at = NOW()
        `, gameID, level, userID, score, userEmail); err != nil {
			log.Printf("⚠️ Failed to persist leaderboard backup: %v", err)
		}
	}

	return nil
}

func (r *RedisLeaderboardRepo) GetTopScores(ctx context.Context, gameID string, limit, level int) ([]ScoreEntry, error) {
	level = normalizeLevel(level)
	key := boardKey(gameID, level)
	ik := infoKey(gameID, level)

	results, err := r.client.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]ScoreEntry, 0, len(results))
	for i, result := range results {
		userID := result.Member.(string)
		score := int(result.Score)

		email, _ := r.client.HGet(ctx, ik, userID+"_email").Result()
		nickname, _ := r.client.HGet(ctx, ik, userID+"_nickname").Result()

		if nickname == "" {
			if email != "" {
				nickname = email
				if idx := strings.Index(nickname, "@"); idx > 0 {
					nickname = nickname[:idx]
				}
			} else if len(userID) >= 8 {
				nickname = userID[:8]
			} else {
				nickname = userID
			}
		}

		entries = append(entries, ScoreEntry{
			Rank:      i + 1,
			UserID:    userID,
			UserEmail: email,
			Nickname:  nickname,
			Score:     score,
			UpdatedAt: time.Now().Unix(),
			Level:     level,
		})
	}

	return entries, nil
}

func (r *RedisLeaderboardRepo) GetPlayerRank(ctx context.Context, gameID, userID string, level int) (int, int, error) {
	level = normalizeLevel(level)
	key := boardKey(gameID, level)

	rank, err := r.client.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	score, err := r.client.ZScore(ctx, key, userID).Result()
	if err != nil {
		return 0, 0, err
	}

	return int(rank) + 1, int(score), nil
}

func (r *RedisLeaderboardRepo) SaveUserInfo(ctx context.Context, gameID, userID, userEmail, nickname string, level int) error {
	ik := infoKey(gameID, level)

	if userEmail != "" {
		if err := r.client.HSet(ctx, ik, userID+"_email", userEmail).Err(); err != nil {
			return err
		}
	}

	if nickname != "" {
		if err := r.client.HSet(ctx, ik, userID+"_nickname", nickname).Err(); err != nil {
			return err
		}
	} else if userEmail != "" {
		defaultNick := userEmail
		if idx := strings.Index(defaultNick, "@"); idx > 0 {
			defaultNick = defaultNick[:idx]
		}
		if err := r.client.HSet(ctx, ik, userID+"_nickname", defaultNick).Err(); err != nil {
			return err
		}
	}

	log.Printf("💾 Saved user info: user=%s level=%d email=%s nickname=%s", userID, normalizeLevel(level), userEmail, nickname)
	return nil
}

// RestoreFromPostgres loads records for one game+level into Redis.
func (r *RedisLeaderboardRepo) RestoreFromPostgres(ctx context.Context, gameID string, level int) error {
	level = normalizeLevel(level)
	rows, err := r.db.Query(ctx, `
        SELECT user_id, score
        FROM leaderboard_backup
        WHERE game_id = $1 AND level = $2
        ORDER BY score DESC
    `, gameID, level)
	if err != nil {
		return err
	}
	defer rows.Close()

	pipe := r.client.Pipeline()
	key := boardKey(gameID, level)
	for rows.Next() {
		var userID string
		var score int64
		if err := rows.Scan(&userID, &score); err != nil {
			continue
		}
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(score),
			Member: userID,
		})
	}
	_, err = pipe.Exec(ctx)
	return err
}
[denismatveev@c0der event_horizon]$ 
