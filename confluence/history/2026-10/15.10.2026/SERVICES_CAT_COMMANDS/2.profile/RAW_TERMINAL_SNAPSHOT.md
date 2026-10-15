[denismatveev@c0der event_horizon]$ find services/profile -type f -name '*.go' | sort
echo "===== main ====="
cat services/profile/cmd/main.go
echo "===== app ====="
for f in services/profile/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/profile/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/profile/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/profile/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== converter ====="
for f in services/profile/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done
services/profile/cmd/main.go
services/profile/internal/app/app.go
services/profile/internal/app/di.go
services/profile/internal/config/config.go
services/profile/internal/converter/profile.go
services/profile/internal/converter/profile_test.go
services/profile/internal/handler/grpc_handler.go
services/profile/internal/interceptor/logger.go
services/profile/internal/interceptor/recovery.go
services/profile/internal/interceptor/validate.go
services/profile/internal/model/errors.go
services/profile/internal/model/profile.go
services/profile/internal/repository/achievement_repo.go
services/profile/internal/repository/profile_repo.go
services/profile/internal/repository/redis_repo.go
services/profile/internal/service/achievements.go
services/profile/internal/service/achievements_test.go
services/profile/internal/service/profile_service.go
services/profile/internal/service/profile_service_test.go
services/profile/migrations/embed.go
services/profile/proto/profile_grpc.pb.go
services/profile/proto/profile.pb.go
services/profile/proto/validate.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/profile/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init profile app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("profile stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/profile/internal/app/app.go ---
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/profile/internal/config"
)

// App is the runnable profile service process (gRPC + metrics HTTP).
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
		a.log.Info("profile gRPC listening", "port", a.cfg.GRPCPort)
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
--- services/profile/internal/app/di.go ---
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/contracts/events"
	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/services/profile/internal/config"
	"github.com/Eastwesser/event-horizon/services/profile/internal/handler"
	"github.com/Eastwesser/event-horizon/services/profile/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/profile/internal/repository"
	"github.com/Eastwesser/event-horizon/services/profile/internal/service"
	"github.com/Eastwesser/event-horizon/services/profile/migrations"
	pb "github.com/Eastwesser/event-horizon/services/profile/proto"
)

type ScoreEvent struct {
	UserID        string `json:"user_id"`
	GameID        string `json:"game_id"`
	Score         int    `json:"score"`
	IsRecord      bool   `json:"is_record"`
	Level         int    `json:"level"`
	LampsEarned   int    `json:"lamps_earned"`
	TicketsEarned int    `json:"tickets_earned"`
}

type UserEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

type diContainer struct {
	cfg *config.Config

	dbpool    *pgxpool.Pool
	nc        *nats.Conn
	js        nats.JetStreamContext
	redisRepo *repository.RedisProfileRepo
	repo      *repository.PostgresProfileRepo
	svc       service.ProfileService
	api       pb.ProfileServiceServer
}

func newDiContainer(cfg *config.Config) *diContainer {
	return &diContainer{cfg: cfg}
}

func (a *App) init(ctx context.Context) error {
	steps := []func(context.Context) error{
		a.initLogger,
		a.initCloser,
		a.initPostgres,
		a.initNATS,
		a.initRedis,
		a.initDomain,
		a.initGRPC,
		a.initMetricsHTTP,
		a.initSubscriptions,
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

func (a *App) initPostgres(ctx context.Context) error {
	dbURL := "postgres://" + a.cfg.DBUser + ":" + a.cfg.DBPassword + "@" + a.cfg.DBHost + ":" + a.cfg.DBPort + "/" + a.cfg.DBName
	poolCfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return fmt.Errorf("postgres config: %w", err)
	}
	poolCfg.MaxConns = 25
	poolCfg.MinConns = 10
	poolCfg.MaxConnLifetime = 5 * time.Minute

	var dbpool *pgxpool.Pool
	var dbErr error
	for i := 0; i < 30; i++ {
		dbpool, dbErr = pgxpool.NewWithConfig(ctx, poolCfg)
		if dbErr == nil {
			if pingErr := dbpool.Ping(ctx); pingErr == nil {
				break
			} else {
				dbpool.Close()
				dbErr = pingErr
			}
		}
		a.log.Warn("postgres connect failed", "attempt", i+1, "err", dbErr)
		time.Sleep(2 * time.Second)
	}
	if dbErr != nil {
		return fmt.Errorf("postgres connect after 30 attempts: %w", dbErr)
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

func (a *App) initNATS(_ context.Context) error {
	var nc *nats.Conn
	var natsErr error
	for i := 0; i < 30; i++ {
		nc, natsErr = nats.Connect(a.cfg.NATSUrl)
		if natsErr == nil {
			break
		}
		a.log.Warn("nats connect failed", "attempt", i+1, "err", natsErr)
		time.Sleep(1 * time.Second)
	}
	if natsErr != nil {
		return fmt.Errorf("nats connect after 30 attempts: %w", natsErr)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return fmt.Errorf("jetstream: %w", err)
	}

	stream, err := js.StreamInfo("EVENTS")
	if err != nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     "EVENTS",
			Subjects: []string{"event.>", "score.updated", "user.registered"},
			Storage:  nats.FileStorage,
		})
		if err != nil {
			a.log.Warn("failed to create EVENTS stream", "err", err)
		}
	} else if !contains(stream.Config.Subjects, "user.registered") {
		newSubjects := append(stream.Config.Subjects, "user.registered")
		_, err = js.UpdateStream(&nats.StreamConfig{
			Name:     "EVENTS",
			Subjects: newSubjects,
			Storage:  nats.FileStorage,
		})
		if err != nil {
			a.log.Warn("failed to update EVENTS stream", "err", err)
		} else {
			a.log.Info("updated EVENTS stream with user.registered")
		}
	}

	a.di.nc = nc
	a.di.js = js
	a.closer.AddNamed("nats", func(context.Context) error {
		return nc.Drain()
	})
	a.log.Info("nats connected", "url", a.cfg.NATSUrl)
	return nil
}

func (a *App) initRedis(ctx context.Context) error {
	redisRepo := repository.NewRedisProfileRepo(a.cfg.RedisAddr)
	if err := redisRepo.Ping(ctx); err != nil {
		a.log.Warn("redis unavailable; profile will run without cache", "err", err)
	} else {
		a.log.Info("redis connected", "addr", a.cfg.RedisAddr)
	}
	a.di.redisRepo = redisRepo
	a.closer.AddNamed("redis", func(context.Context) error {
		return redisRepo.Close()
	})
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	a.di.repo = repository.NewPostgresProfileRepo(a.di.dbpool)
	achRepo := repository.NewPostgresAchievementRepo(a.di.dbpool)
	cacheTTL := time.Duration(a.cfg.ProfileCacheTTLMin) * time.Minute
	a.di.svc = service.NewProfileService(a.di.repo, achRepo, a.di.redisRepo, cacheTTL)
	a.di.api = handler.NewProfileHandler(a.di.svc)
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
		),
	)
	pb.RegisterProfileServiceServer(grpcServer, a.di.api)
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
		_, _ = w.Write([]byte(`{"status":"ok","service":"profile"}`))
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
		_, _ = w.Write([]byte(`{"status":"ready","service":"profile"}`))
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

func (a *App) initSubscriptions(_ context.Context) error {
	js := a.di.js
	svc := a.di.svc
	repo := a.di.repo

	_, err := js.Subscribe("event.user.registered", func(msg *nats.Msg) {
		var event UserEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			a.log.Error("failed to unmarshal event.user.registered", "err", err)
			return
		}

		nickname := event.Email
		if len(nickname) > 8 {
			nickname = nickname[:8]
		}

		profile := &repository.UserProfile{
			UserID:     event.UserID,
			Email:      event.Email,
			Nickname:   nickname,
			TotalScore: 0,
			BestScores: make(map[string]int32),
			Lamps:      0,
			Tickets:    0,
		}

		if err := svc.UpdateProfile(context.Background(), profile); err != nil {
			a.log.Error("failed to upsert profile", "user", event.UserID, "err", err)
		} else {
			a.log.Info("profile created", "user", event.UserID)
		}
		_ = msg.Ack()
	}, nats.Durable("profile-user-registered"), nats.ManualAck())
	if err != nil {
		a.log.Warn("failed to subscribe to event.user.registered", "err", err)
	} else {
		a.log.Info("subscribed to nats", "subject", "event.user.registered")
	}

	_, err = js.Subscribe("score.updated", func(msg *nats.Msg) {
		var event ScoreEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			a.log.Error("failed to unmarshal score.updated", "err", err)
			return
		}

		a.log.Info("updating profile",
			"user", event.UserID,
			"game", event.GameID,
			"score", event.Score,
		)

		ctx := context.Background()
		profile, err := repo.GetProfile(ctx, event.UserID)
		if err != nil || profile == nil {
			profile = &repository.UserProfile{
				UserID:     event.UserID,
				BestScores: make(map[string]int32),
			}
		}

		if event.IsRecord {
			if profile.BestScores == nil {
				profile.BestScores = make(map[string]int32)
			}
			if current, ok := profile.BestScores[event.GameID]; !ok || int32(event.Score) > current {
				profile.BestScores[event.GameID] = int32(event.Score)
			}
		}

		var total int32
		for _, s := range profile.BestScores {
			total += s
		}
		profile.TotalScore = total
		profile.Lamps += int32(event.LampsEarned)
		profile.Tickets += int32(event.TicketsEarned)

		if err := svc.UpdateProfile(ctx, profile); err != nil {
			a.log.Error("failed to update profile", "user", event.UserID, "err", err)
		} else {
			a.log.Info("profile updated", "user", event.UserID, "total", profile.TotalScore)
			// total_score = sum(best_scores); unlock after upsert (boosted runs never publish)
			if err := svc.EvaluateAndUnlock(ctx, event.UserID, profile.BestScores, profile.TotalScore, event.GameID, event.Level); err != nil {
				a.log.Error("failed to evaluate achievements", "user", event.UserID, "err", err)
			}
		}
		_ = msg.Ack()
	}, nats.Durable("profile-score-updated"), nats.ManualAck())
	if err != nil {
		a.log.Warn("failed to subscribe to score.updated", "err", err)
	} else {
		a.log.Info("subscribed to nats", "subject", "score.updated")
	}

	// Wave 2 #5b — first shop purchase.
	_, err = js.Subscribe(kafka.TopicPurchasePaid, func(msg *nats.Msg) {
		paid, err := events.UnmarshalPurchasePaid(msg.Data)
		if err != nil {
			a.log.Error("failed to unmarshal purchase.paid", "err", err)
			_ = msg.Ack()
			return
		}
		userID := paid.UserUUID
		if userID == "" {
			_ = msg.Ack()
			return
		}
		ctx := context.Background()
		if err := svc.UnlockCodes(ctx, userID, []string{"first_purchase"}); err != nil {
			a.log.Error("first_purchase unlock failed", "user", userID, "err", err)
		}
		_ = msg.Ack()
	}, nats.Durable("profile-purchase-paid"), nats.ManualAck())
	if err != nil {
		a.log.Warn("failed to subscribe to purchase.paid", "err", err)
	} else {
		a.log.Info("subscribed to nats", "subject", kafka.TopicPurchasePaid)
	}

	// Wave 2 #5b — first boost spend.
	_, err = js.Subscribe(kafka.TopicBoostStarted, func(msg *nats.Msg) {
		started, err := events.UnmarshalBoostStarted(msg.Data)
		if err != nil {
			a.log.Error("failed to unmarshal boost.started", "err", err)
			_ = msg.Ack()
			return
		}
		userID := started.UserUUID
		if userID == "" {
			_ = msg.Ack()
			return
		}
		ctx := context.Background()
		if err := svc.UnlockCodes(ctx, userID, []string{"first_boost"}); err != nil {
			a.log.Error("first_boost unlock failed", "user", userID, "err", err)
		}
		_ = msg.Ack()
	}, nats.Durable("profile-boost-started"), nats.ManualAck())
	if err != nil {
		a.log.Warn("failed to subscribe to boost.started", "err", err)
	} else {
		a.log.Info("subscribed to nats", "subject", kafka.TopicBoostStarted)
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
===== handler =====
--- services/profile/internal/handler/grpc_handler.go ---
package handler

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/profile/internal/repository"
	"github.com/Eastwesser/event-horizon/services/profile/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/profile/proto"
)

type ProfileHandler struct {
	pb.UnimplementedProfileServiceServer
	profileService service.ProfileService
}

func NewProfileHandler(profileService service.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

func (h *ProfileHandler) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}

	profile, err := h.profileService.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if profile == nil {
		return nil, status.Error(codes.NotFound, "profile not found")
	}

	// Backfill score-based achievements from stored best_scores / total_score
	// (sum of best_scores). Level achievements need a live score.updated event.
	_ = h.profileService.EvaluateAndUnlock(ctx, profile.UserID, profile.BestScores, profile.TotalScore, "", 0)

	achievements, err := h.profileService.ListAchievements(ctx, profile.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbAch := make([]*pb.Achievement, 0, len(achievements))
	for _, a := range achievements {
		pbAch = append(pbAch, &pb.Achievement{
			Code:        a.Code,
			Title:       a.Title,
			Description: a.Description,
			Icon:        a.Icon,
			UnlockedAt:  a.UnlockedAt.UTC().Format(time.RFC3339),
		})
	}

	return &pb.GetProfileResponse{
		UserId:       profile.UserID,
		Email:        profile.Email,
		Nickname:     profile.Nickname,
		TotalScore:   profile.TotalScore,
		BestScores:   profile.BestScores,
		Lamps:        profile.Lamps,
		Tickets:      profile.Tickets,
		Achievements: pbAch,
	}, nil
}

func (h *ProfileHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	profile := &repository.UserProfile{
		UserID:     req.UserId,
		Nickname:   req.Nickname,
		BestScores: req.BestScores,
	}

	if req.TotalScore != nil {
		profile.TotalScore = *req.TotalScore
	}

	if err := h.profileService.UpdateProfile(ctx, profile); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateProfileResponse{Success: true, Message: "profile updated"}, nil
}
===== service =====
--- services/profile/internal/service/achievements.go ---
package service

import (
	"context"
	"fmt"
	"log"

	"github.com/Eastwesser/event-horizon/services/profile/internal/repository"
)

// mainGames — first_play_* catalog (Wave 2 #5).
var mainGames = []string{
	"flappy", "hexagon", "memory", "towers", "hanoi", "twenty48", "gears", "companion",
}

// Per-game score tiers: любитель / профессионал / герой (новичок = first_play_*).
var gameScoreTiers = map[string][3]int32{
	"flappy":    {25, 100, 500},
	"hexagon":   {50, 200, 1000},
	"memory":    {200, 500, 900},
	"towers":    {10, 30, 80},
	"hanoi":     {100, 500, 2000},
	"twenty48":  {512, 2048, 8192},
	"gears":     {50, 200, 500},
	"companion": {1000, 5000, 15000},
}

// EvaluateAchievements derives unlock codes from best_scores / total_score / optional event level.
// totalScore should be the sum of best_scores (already maintained by the score.updated consumer).
func EvaluateAchievements(bestScores map[string]int32, totalScore int32, gameID string, level int) []string {
	codes := make([]string, 0, 40)
	if bestScores == nil {
		bestScores = map[string]int32{}
	}

	for _, g := range mainGames {
		if _, ok := bestScores[g]; ok {
			codes = append(codes, "first_play_"+g)
		}
	}
	// First play from this event even if not yet a stored best (edge: non-record).
	if gameID != "" {
		codes = append(codes, "first_play_"+gameID)
	}

	for game, tiers := range gameScoreTiers {
		score := bestScores[game]
		if score >= tiers[0] {
			codes = append(codes, game+"_amateur")
		}
		if score >= tiers[1] {
			codes = append(codes, game+"_pro")
		}
		if score >= tiers[2] {
			codes = append(codes, game+"_hero")
		}
	}

	// Legacy flappy thresholds (kept for already-unlocked badges).
	if bestScores["flappy"] >= 100 {
		codes = append(codes, "flappy_score_100")
	}
	if bestScores["flappy"] >= 500 {
		codes = append(codes, "flappy_score_500")
	}

	if gameID == "flappy" {
		if level >= 5 {
			codes = append(codes, "flappy_level_5")
		}
		if level >= 10 {
			codes = append(codes, "flappy_level_10")
		}
	}

	if totalScore >= 1000 {
		codes = append(codes, "total_score_1k")
	}
	if totalScore >= 5000 {
		codes = append(codes, "total_score_5k")
	}
	if totalScore >= 10000 {
		codes = append(codes, "total_score_10k")
	}

	return uniqueStrings(codes)
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func (s *profileService) EvaluateAndUnlock(ctx context.Context, userID string, bestScores map[string]int32, totalScore int32, gameID string, level int) error {
	if s.achievements == nil || userID == "" {
		return nil
	}
	codes := EvaluateAchievements(bestScores, totalScore, gameID, level)
	return s.UnlockCodes(ctx, userID, codes)
}

// UnlockCodes inserts catalog codes for the user (idempotent via ON CONFLICT).
func (s *profileService) UnlockCodes(ctx context.Context, userID string, codes []string) error {
	if s.achievements == nil || userID == "" || len(codes) == 0 {
		return nil
	}
	n, err := s.achievements.Unlock(ctx, userID, codes)
	if err != nil {
		return fmt.Errorf("unlock achievements: %w", err)
	}
	if n > 0 {
		log.Printf("achievements unlocked user=%s count=%d codes=%v", userID, n, codes)
	}
	return nil
}

func (s *profileService) ListAchievements(ctx context.Context, userID string) ([]repository.Achievement, error) {
	if s.achievements == nil {
		return nil, nil
	}
	return s.achievements.ListUnlocked(ctx, userID)
}
--- services/profile/internal/service/achievements_test.go ---
package service

import (
	"testing"
)

func TestEvaluateAchievements(t *testing.T) {
	codes := EvaluateAchievements(map[string]int32{
		"flappy": 120,
		"hexagon": 10,
	}, 130, "flappy", 5)

	want := map[string]bool{
		"first_play_flappy":  true,
		"first_play_hexagon": true,
		"flappy_amateur":     true,
		"flappy_pro":         true,
		"flappy_score_100":   true,
		"flappy_level_5":     true,
	}
	got := map[string]bool{}
	for _, c := range codes {
		got[c] = true
	}
	for k := range want {
		if !got[k] {
			t.Fatalf("missing %s in %v", k, codes)
		}
	}
	if got["flappy_score_500"] || got["flappy_level_10"] || got["total_score_1k"] {
		t.Fatalf("unexpected high-tier unlocks: %v", codes)
	}

	codes = EvaluateAchievements(map[string]int32{"flappy": 600}, 6000, "flappy", 10)
	got = map[string]bool{}
	for _, c := range codes {
		got[c] = true
	}
	for _, k := range []string{"flappy_score_500", "flappy_level_10", "total_score_5k", "total_score_1k"} {
		if !got[k] {
			t.Fatalf("missing %s in %v", k, codes)
		}
	}
}
--- services/profile/internal/service/profile_service.go ---
package service

import (
	"context"
	"log"
	"time"

	"github.com/Eastwesser/event-horizon/services/profile/internal/repository"
)

type ProfileService interface {
	GetProfile(ctx context.Context, userID string) (*repository.UserProfile, error)
	UpdateProfile(ctx context.Context, profile *repository.UserProfile) error
	EvaluateAndUnlock(ctx context.Context, userID string, bestScores map[string]int32, totalScore int32, gameID string, level int) error
	UnlockCodes(ctx context.Context, userID string, codes []string) error
	ListAchievements(ctx context.Context, userID string) ([]repository.Achievement, error)
}

type profileService struct {
	repo         repository.ProfileRepository
	achievements repository.AchievementRepository
	cache        *repository.RedisProfileRepo
	cacheTTL     time.Duration
}

func NewProfileService(
	repo repository.ProfileRepository,
	achievements repository.AchievementRepository,
	cache *repository.RedisProfileRepo,
	cacheTTL time.Duration,
) ProfileService {
	return &profileService{repo: repo, achievements: achievements, cache: cache, cacheTTL: cacheTTL}
}

func (s *profileService) GetProfile(ctx context.Context, userID string) (*repository.UserProfile, error) {
	if s.cache != nil {
		cached, err := s.cache.GetProfile(ctx, userID)
		if err != nil {
			log.Printf("profile cache read error for %s: %v", userID, err)
		} else if cached != nil {
			return cached, nil
		}
	}

	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil || profile == nil {
		return profile, err
	}

	if s.cache != nil {
		if err := s.cache.SetProfile(ctx, profile, s.cacheTTL); err != nil {
			log.Printf("profile cache write error for %s: %v", userID, err)
		}
	}

	return profile, nil
}

func (s *profileService) UpdateProfile(ctx context.Context, profile *repository.UserProfile) error {
	if err := s.repo.UpsertProfile(ctx, profile); err != nil {
		return err
	}

	if s.cache != nil {
		if err := s.cache.InvalidateProfile(ctx, profile.UserID); err != nil {
			log.Printf("profile cache invalidate error for %s: %v", profile.UserID, err)
		}
	}

	return nil
}
--- services/profile/internal/service/profile_service_test.go ---
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/profile/internal/repository"
)

type mockProfileRepo struct {
	profile *repository.UserProfile
	err     error
	upserts int
}

func (m *mockProfileRepo) GetProfile(context.Context, string) (*repository.UserProfile, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.profile == nil {
		return nil, nil
	}
	cp := *m.profile
	return &cp, nil
}
func (m *mockProfileRepo) UpsertProfile(_ context.Context, p *repository.UserProfile) error {
	m.upserts++
	m.profile = p
	return m.err
}

type mockAchievements struct {
	list    []repository.Achievement
	unlocks []string
	err     error
}

func (m *mockAchievements) ListUnlocked(context.Context, string) ([]repository.Achievement, error) {
	return m.list, m.err
}
func (m *mockAchievements) Unlock(_ context.Context, _ string, codes []string) (int64, error) {
	m.unlocks = append(m.unlocks, codes...)
	return int64(len(codes)), m.err
}

func TestGetProfile_FromRepo(t *testing.T) {
	repo := &mockProfileRepo{profile: &repository.UserProfile{UserID: "u1", Nickname: "neo"}}
	got, err := NewProfileService(repo, nil, nil, time.Minute).GetProfile(context.Background(), "u1")
	if err != nil || got.Nickname != "neo" {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestUpdateProfile(t *testing.T) {
	repo := &mockProfileRepo{}
	svc := NewProfileService(repo, nil, nil, time.Minute)
	if err := svc.UpdateProfile(context.Background(), &repository.UserProfile{UserID: "u1", Nickname: "x"}); err != nil {
		t.Fatal(err)
	}
	if repo.upserts != 1 {
		t.Fatal("upsert not called")
	}
	repo.err = errors.New("db")
	if err := svc.UpdateProfile(context.Background(), &repository.UserProfile{UserID: "u1"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestListAchievements(t *testing.T) {
	svc := NewProfileService(&mockProfileRepo{}, nil, nil, time.Minute)
	got, err := svc.ListAchievements(context.Background(), "u1")
	if err != nil || got != nil {
		t.Fatalf("%v err=%v", got, err)
	}
	ach := &mockAchievements{list: []repository.Achievement{{Code: "first_play_flappy"}}}
	got, err = NewProfileService(&mockProfileRepo{}, ach, nil, time.Minute).ListAchievements(context.Background(), "u1")
	if err != nil || len(got) != 1 {
		t.Fatalf("%v err=%v", got, err)
	}
}

func TestEvaluateAndUnlock(t *testing.T) {
	ach := &mockAchievements{}
	svc := NewProfileService(&mockProfileRepo{}, ach, nil, time.Minute)
	if err := svc.EvaluateAndUnlock(context.Background(), "", map[string]int32{"flappy": 1}, 1, "flappy", 1); err != nil {
		t.Fatal(err)
	}
	if len(ach.unlocks) != 0 {
		t.Fatal("empty user should no-op")
	}
	if err := svc.EvaluateAndUnlock(context.Background(), "u1", map[string]int32{"flappy": 100}, 1000, "flappy", 5); err != nil {
		t.Fatal(err)
	}
	if len(ach.unlocks) == 0 {
		t.Fatal("expected unlocks")
	}
}
===== repository =====
--- services/profile/internal/repository/achievement_repo.go ---
package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Achievement struct {
	Code        string
	Title       string
	Description string
	Icon        string
	UnlockedAt  time.Time
}

type AchievementRepository interface {
	ListUnlocked(ctx context.Context, userID string) ([]Achievement, error)
	Unlock(ctx context.Context, userID string, codes []string) (int64, error)
}

type PostgresAchievementRepo struct {
	db *pgxpool.Pool
}

func NewPostgresAchievementRepo(db *pgxpool.Pool) *PostgresAchievementRepo {
	return &PostgresAchievementRepo{db: db}
}

func (r *PostgresAchievementRepo) ListUnlocked(ctx context.Context, userID string) ([]Achievement, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.code, a.title, a.description, a.icon, ua.unlocked_at
		FROM user_achievements ua
		JOIN achievements a ON a.code = ua.code
		WHERE ua.user_id = $1
		ORDER BY a.sort_order ASC, ua.unlocked_at ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Achievement, 0)
	for rows.Next() {
		var a Achievement
		if err := rows.Scan(&a.Code, &a.Title, &a.Description, &a.Icon, &a.UnlockedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *PostgresAchievementRepo) Unlock(ctx context.Context, userID string, codes []string) (int64, error) {
	if len(codes) == 0 {
		return 0, nil
	}
	// Ensure profile row exists for FK (no-op if already present).
	_, _ = r.db.Exec(ctx, `
		INSERT INTO user_profiles (user_id, email, nickname, total_score, best_scores, lamps, tickets, updated_at)
		VALUES ($1, '', '', 0, '{}', 0, 0, NOW())
		ON CONFLICT (user_id) DO NOTHING`, userID)

	tag, err := r.db.Exec(ctx, `
		INSERT INTO user_achievements (user_id, code, unlocked_at)
		SELECT $1, c.code, NOW()
		FROM UNNEST($2::text[]) AS c(code)
		INNER JOIN achievements a ON a.code = c.code
		ON CONFLICT (user_id, code) DO NOTHING`, userID, codes)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
--- services/profile/internal/repository/profile_repo.go ---
package repository

import (
    "context"
    "encoding/json"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type UserProfile struct {
    UserID     string
    Email      string
    Nickname   string
    TotalScore int32
    BestScores map[string]int32
    Lamps      int32
    Tickets    int32
    UpdatedAt  time.Time
}

type ProfileRepository interface {
    GetProfile(ctx context.Context, userID string) (*UserProfile, error)
    UpsertProfile(ctx context.Context, profile *UserProfile) error
}

type PostgresProfileRepo struct {
    db *pgxpool.Pool
}

func NewPostgresProfileRepo(db *pgxpool.Pool) *PostgresProfileRepo {
    return &PostgresProfileRepo{db: db}
}

func (r *PostgresProfileRepo) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
    var profile UserProfile
    var bestScoresJSON []byte

    query := `SELECT user_id, email, nickname, total_score, best_scores, lamps, tickets, updated_at
              FROM user_profiles WHERE user_id = $1`

    err := r.db.QueryRow(ctx, query, userID).Scan(
        &profile.UserID,
        &profile.Email,
        &profile.Nickname,
        &profile.TotalScore,
        &bestScoresJSON,
        &profile.Lamps,
        &profile.Tickets,
        &profile.UpdatedAt,
    )

    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }

    if err := json.Unmarshal(bestScoresJSON, &profile.BestScores); err != nil {
        profile.BestScores = make(map[string]int32)
    }

    return &profile, nil
}

func (r *PostgresProfileRepo) UpsertProfile(ctx context.Context, profile *UserProfile) error {
    bestScoresJSON, err := json.Marshal(profile.BestScores)
    if err != nil {
        return err
    }

    query := `INSERT INTO user_profiles (user_id, email, nickname, total_score, best_scores, lamps, tickets, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
              ON CONFLICT (user_id) DO UPDATE
              SET email = EXCLUDED.email,
                  nickname = EXCLUDED.nickname,
                  total_score = EXCLUDED.total_score,
                  best_scores = EXCLUDED.best_scores,
                  lamps = EXCLUDED.lamps,
                  tickets = EXCLUDED.tickets,
                  updated_at = NOW()`

    _, err = r.db.Exec(ctx, query,
        profile.UserID,
        profile.Email,
        profile.Nickname,
        profile.TotalScore,
        bestScoresJSON,
        profile.Lamps,
        profile.Tickets,
    )

    return err
}
--- services/profile/internal/repository/redis_repo.go ---
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisProfileRepo implements cache-aside for aggregated user profiles.
// Profile previously had no Redis at all, so every GetProfile call hit
// Postgres directly — this was explicitly flagged as the project's biggest
// caching gap.
type RedisProfileRepo struct {
	client *redis.Client
}

func NewRedisProfileRepo(addr string) *RedisProfileRepo {
	client := redis.NewClient(&redis.Options{Addr: addr})
	return &RedisProfileRepo{client: client}
}

func (r *RedisProfileRepo) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisProfileRepo) Close() error {
	return r.client.Close()
}

func cacheKey(userID string) string {
	return fmt.Sprintf("profile:%s", userID)
}

func (r *RedisProfileRepo) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	val, err := r.client.Get(ctx, cacheKey(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var profile UserProfile
	if err := json.Unmarshal(val, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *RedisProfileRepo) SetProfile(ctx context.Context, profile *UserProfile, ttl time.Duration) error {
	payload, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, cacheKey(profile.UserID), payload, ttl).Err()
}

func (r *RedisProfileRepo) InvalidateProfile(ctx context.Context, userID string) error {
	return r.client.Del(ctx, cacheKey(userID)).Err()
}
===== converter =====
--- services/profile/internal/converter/profile.go ---
package converter

import (
	"github.com/Eastwesser/event-horizon/services/profile/internal/model"
	pb "github.com/Eastwesser/event-horizon/services/profile/proto"
)

func ProfileToProto(p *model.Profile) *pb.GetProfileResponse {
	if p == nil {
		return nil
	}
	return &pb.GetProfileResponse{
		UserId:     p.UserID,
		Email:      p.Email,
		Nickname:   p.Nickname,
		TotalScore: p.TotalScore,
		BestScores: p.BestScores,
		Lamps:      p.Lamps,
		Tickets:    p.Tickets,
		// Achievements are loaded separately in the gRPC handler.
	}
}

func ProfileFromUpdateRequest(req *pb.UpdateProfileRequest) *model.Profile {
	if req == nil {
		return nil
	}
	p := &model.Profile{
		UserID:     req.UserId,
		Nickname:   req.Nickname,
		BestScores: req.BestScores,
	}
	if req.TotalScore != nil {
		p.TotalScore = *req.TotalScore
	}
	return p
}
--- services/profile/internal/converter/profile_test.go ---
package converter

import (
	"testing"

	"github.com/Eastwesser/event-horizon/services/profile/internal/model"
)

func TestProfileToProto(t *testing.T) {
	p := ProfileToProto(&model.Profile{UserID: "u1", Email: "e", Nickname: "n", Lamps: 2})
	if p.UserId != "u1" || p.Lamps != 2 {
		t.Fatalf("%+v", p)
	}
}
[denismatveev@c0der event_horizon]$ 
