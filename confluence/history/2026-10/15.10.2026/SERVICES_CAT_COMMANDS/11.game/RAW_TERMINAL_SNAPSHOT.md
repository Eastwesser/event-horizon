[denismatveev@c0der event_horizon]$ find services/game -type f -name '*.go' | sort
echo "===== main ====="
cat services/game/cmd/main.go
echo "===== app ====="
for f in services/game/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/game/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/game/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/game/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== converter ====="
for f in services/game/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/game/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
services/game/cmd/main.go
services/game/games/hexagons/board.go
services/game/games/hexagons/validator.go
services/game/games/memory/game.go
services/game/internal/app/app.go
services/game/internal/app/di.go
services/game/internal/config/config.go
services/game/internal/config/interfaces.go
services/game/internal/converter/score.go
services/game/internal/converter/score_test.go
services/game/internal/handler/grpc_handler.go
services/game/internal/interceptor/logger.go
services/game/internal/interceptor/recovery.go
services/game/internal/interceptor/validate.go
services/game/internal/model/errors.go
services/game/internal/model/score.go
services/game/internal/repository/game_repo.go
services/game/internal/service/game_service.go
services/game/internal/service/game_service_test.go
services/game/internal/worker/outbox_worker.go
services/game/migrations/embed.go
services/game/proto/game_grpc.pb.go
services/game/proto/game.pb.go
services/game/proto/validate.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/game/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init game app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("game stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/game/internal/app/app.go ---
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/game/internal/config"
)

// App is the runnable game service process (gRPC + metrics HTTP).
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
		a.log.Info("game gRPC listening", "port", a.cfg.GRPCConfig().Port())
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
--- services/game/internal/app/di.go ---
package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
	"github.com/Eastwesser/event-horizon/services/game/internal/config"
	"github.com/Eastwesser/event-horizon/services/game/internal/handler"
	"github.com/Eastwesser/event-horizon/services/game/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/game/internal/repository"
	"github.com/Eastwesser/event-horizon/services/game/internal/service"
	"github.com/Eastwesser/event-horizon/services/game/internal/worker"
	"github.com/Eastwesser/event-horizon/services/game/migrations"
	pb "github.com/Eastwesser/event-horizon/services/game/proto"
)

type diContainer struct {
	cfg config.ConfigProvider

	db            *sql.DB
	nc            *nats.Conn
	js            nats.JetStreamContext
	billingClient billingPb.BillingServiceClient
	repo          repository.GameRepository
	svc           service.GameService
	api           pb.GameServiceServer
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
			semconv.ServiceNameKey.String("game"),
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

func (a *App) initPostgres(_ context.Context) error {
	db, err := sql.Open("postgres", a.cfg.PostgresConfig().DSN())
	if err != nil {
		return fmt.Errorf("postgres open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("postgres ping: %w", err)
	}
	if err := migrator.Up(db, migrations.FS); err != nil {
		_ = db.Close()
		return fmt.Errorf("migrations: %w", err)
	}

	a.di.db = db
	a.closer.AddNamed("postgres", func(context.Context) error {
		return db.Close()
	})
	a.log.Info("postgres ready, migrations applied")
	return nil
}

func (a *App) initNATS(_ context.Context) error {
	url := a.cfg.NATSConfig().URL()
	var nc *nats.Conn
	var lastErr error
	for i := 0; i < 30; i++ {
		nc, lastErr = nats.Connect(url)
		if lastErr == nil {
			break
		}
		a.log.Warn("nats connection attempt failed", "attempt", i+1, "of", 30, "err", lastErr)
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
	a.log.Info("nats connected", "url", url)
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	addr := a.cfg.BillingAddr()
	billingConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		a.log.Warn("billing dial failed; boosts disabled", "addr", addr, "err", err)
	} else {
		a.di.billingClient = billingPb.NewBillingServiceClient(billingConn)
		a.closer.AddNamed("billing grpc", func(context.Context) error {
			return billingConn.Close()
		})
		a.log.Info("billing client ready", "addr", addr)
	}

	a.di.repo = repository.NewPostgresGameRepo(a.di.db)
	a.di.svc = service.NewGameService(a.di.repo, a.di.js, a.di.billingClient)
	a.di.api = handler.NewGameHandler(a.di.svc)
	return nil
}

func (a *App) initOutbox(_ context.Context) error {
	outboxCtx, cancelOutbox := context.WithCancel(context.Background())
	outboxWorker := worker.NewOutboxWorker(a.di.db, a.di.js)
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
	pb.RegisterGameServiceServer(grpcServer, a.di.api)
	reflection.Register(grpcServer)

	a.grpcServer = grpcServer
	a.closer.AddNamed("grpc server", func(context.Context) error {
		grpcServer.GracefulStop()
		return nil
	})
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	db := a.di.db
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "game"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded", "reason": "postgres unreachable"})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "game"})
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
===== handler =====
--- services/game/internal/handler/grpc_handler.go ---
package handler

import (
    "context"
    "log"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"

    pb "github.com/Eastwesser/event-horizon/services/game/proto"
    "github.com/Eastwesser/event-horizon/services/game/internal/service"
    hexagonValidator "github.com/Eastwesser/event-horizon/services/game/games/hexagons"
)

var (
    gameSubmitsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "game_submits_total",
            Help: "Total number of score submissions",
        },
        []string{"game_id", "status"},
    )

    gameScoreHistogram = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "game_score_histogram",
            Help:    "Distribution of scores by game",
            Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000, 5000},
        },
        []string{"game_id"},
    )
)

type GameHandler struct {
    pb.UnimplementedGameServiceServer
    gameService service.GameService
}

func NewGameHandler(svc service.GameService) *GameHandler {
    return &GameHandler{
        gameService: svc,
    }
}

func (h *GameHandler) SubmitScore(ctx context.Context, req *pb.SubmitScoreRequest) (*pb.SubmitScoreResponse, error) {
    log.Printf("📥 HANDLER RAW: %+v", req)
    
    if req.UserId == "" || req.GameId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id and game_id are required")
    }
    
    log.Printf("📥 HANDLER: user=%s, game=%s, level=%d, score=%d, moves=%d, email=%s", 
        req.UserId, req.GameId, req.Level, req.Score, len(req.Moves), req.UserEmail)

    // Конвертируем moves из protobuf в hexagonValidator.Move
    moves := make([]hexagonValidator.Move, len(req.Moves))
    for i, m := range req.Moves {
        moves[i] = hexagonValidator.Move{
            FromX:     int(m.FromX),
            FromY:     int(m.FromY),
            ToX:       int(m.ToX),
            ToY:       int(m.ToY),
            Timestamp: m.Timestamp,
        }
    }

    resp, err := h.gameService.SubmitScore(ctx, &service.SubmitScoreRequest{
        UserID:    req.UserId,
        GameID:    req.GameId,
        Level:     int(req.Level),
        Score:     int(req.Score),
        UserEmail: req.UserEmail,
        Nickname:  req.Nickname,
        Seed:      req.Seed,
        Moves:     moves,
        BoostID:   req.BoostId,
    })
    if err != nil {
        log.Printf("SubmitScore error: %v", err)
        return &pb.SubmitScoreResponse{
            Success: false,
            Message: err.Error(),
            Ranked:  false,
        }, nil
    }

    // Метрики после успешного ответа
    if resp.Success {
        gameSubmitsTotal.WithLabelValues(req.GameId, "success").Inc()
        gameScoreHistogram.WithLabelValues(req.GameId).Observe(float64(resp.NewHighscore))
    } else {
        gameSubmitsTotal.WithLabelValues(req.GameId, "error").Inc()
    }

    return &pb.SubmitScoreResponse{
        Success:       resp.Success,
        NewHighscore:  int32(resp.NewHighscore),
        Rank:          int32(resp.Rank),
        Message:       resp.Message,
        LampsEarned:   int32(resp.LampsEarned),
        TicketsEarned: int32(resp.TicketsEarned),
        Ranked:        resp.Ranked,
    }, nil
}

func (h *GameHandler) StartBoost(ctx context.Context, req *pb.StartBoostRequest) (*pb.StartBoostResponse, error) {
    if req == nil || req.UserId == "" || req.GameId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id and game_id are required")
    }
    resp, err := h.gameService.StartBoost(ctx, &service.StartBoostRequest{
        UserID: req.UserId,
        GameID: req.GameId,
    })
    if err != nil {
        return nil, err
    }
    return &pb.StartBoostResponse{
        BoostId:    resp.BoostID,
        Boosted:    resp.Boosted,
        Cost:       int32(resp.Cost),
        NewBalance: int32(resp.NewBalance),
        Message:    resp.Message,
    }, nil
}

func (h *GameHandler) GetGameInfo(ctx context.Context, req *pb.GetGameInfoRequest) (*pb.GetGameInfoResponse, error) {
    if req.GameId == "" {
        return nil, status.Error(codes.InvalidArgument, "game_id is required")
    }

    info, err := h.gameService.GetGameInfo(ctx, req.GameId)
    if err != nil {
        return nil, status.Error(codes.NotFound, err.Error())
    }

    levels := make([]*pb.LevelInfo, len(info.Levels))
    for i, l := range info.Levels {
        levels[i] = &pb.LevelInfo{
            Level:         int32(l.Level),
            TargetScore:   int32(l.TargetScore),
            RewardLamps:   int32(l.RewardLamps),
            RewardTickets: int32(l.RewardTickets),
        }
    }

    return &pb.GetGameInfoResponse{
        GameId:      info.GameID,
        Name:        info.Name,
        Description: info.Description,
        Levels:      levels,
    }, nil
}
===== service =====
--- services/game/internal/service/game_service.go ---
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/nats-io/nats.go"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
    "github.com/Eastwesser/event-horizon/services/game/internal/repository"
    hexagonValidator "github.com/Eastwesser/event-horizon/services/game/games/hexagons"
    "github.com/Eastwesser/event-horizon/services/game/games/memory"
)

const (
    boostCostLamps = 10
)

// boostAllowedGames — games that can arm a lamp boost (boosted runs are not ranked).
var boostAllowedGames = map[string]struct{}{
    "hexagon":   {},
    "flappy":    {},
    "memory":    {},
    "towers":    {},
    "hanoi":     {},
    "twenty48":  {},
    "gears":     {},
    "companion": {},
}

type SubmitScoreRequest struct {
    UserID    string
    GameID    string
    Level     int
    Score     int
    UserEmail string
    Nickname  string
    Seed      string
    Moves     []hexagonValidator.Move
    BoostID   string
}

type SubmitScoreResponse struct {
    Success       bool
    NewHighscore  int
    Rank          int
    Message       string
    LampsEarned   int
    TicketsEarned int
    Ranked        bool
}

type StartBoostRequest struct {
    UserID string
    GameID string
}

type StartBoostResponse struct {
    BoostID    string
    Boosted    bool
    Cost       int
    NewBalance int
    Message    string
}

type GameInfo struct {
    GameID      string
    Name        string
    Description string
    Levels      []LevelInfo
}

type LevelInfo struct {
    Level         int
    TargetScore   int
    RewardLamps   int
    RewardTickets int
}

type GameService interface {
    SubmitScore(ctx context.Context, req *SubmitScoreRequest) (*SubmitScoreResponse, error)
    GetGameInfo(ctx context.Context, gameID string) (*GameInfo, error)
    StartBoost(ctx context.Context, req *StartBoostRequest) (*StartBoostResponse, error)
}

type gameService struct {
    repo          repository.GameRepository
    js            nats.JetStreamContext
    validator     *hexagonValidator.Validator
    billingClient billingPb.BillingServiceClient
}

func NewGameService(repo repository.GameRepository, js nats.JetStreamContext, billingClient billingPb.BillingServiceClient) GameService {
    return &gameService{
        repo:          repo,
        js:            js,
        validator:     hexagonValidator.NewValidator(),
        billingClient: billingClient,
    }
}

func (s *gameService) SubmitScore(ctx context.Context, req *SubmitScoreRequest) (*SubmitScoreResponse, error) {
    var lampsEarned, ticketsEarned int
    validatedScore := req.Score

    log.Printf("📥 Service received: game_id=%s, user=%s, score=%d", req.GameID, req.UserID, req.Score)

    // В зависимости от игры — своя логика наград
    switch req.GameID {
    case "hexagon":
        valid, valScore, err := s.validator.ValidateMoves(req.Seed, req.Moves, req.Score)
        if err != nil {
            log.Printf("❌ Hexagon validation error: %v", err)
            return &SubmitScoreResponse{Success: false, Message: "validation error"}, nil
        }
        if !valid {
            log.Printf("⚠️ Invalid hexagon game state for user %s", req.UserID)
            return &SubmitScoreResponse{Success: false, Message: "invalid game state or moves"}, nil
        }
        validatedScore = valScore
        lampsEarned = 10
        ticketsEarned = 0
        if validatedScore > 0 {
            ticketsEarned = validatedScore / 100
            if ticketsEarned > 100 {
                ticketsEarned = 100
            }
        }

    case "memory":
        game := memory.NewMemoryGame(req.Seed)
        var memoryMoves []memory.Move
        for _, m := range req.Moves {
            memoryMoves = append(memoryMoves, memory.Move{
                CardIndex1: int(m.FromX),
                CardIndex2: int(m.ToX),
            })
        }

        valid, valScore, err := game.ValidateMoves(req.Seed, memoryMoves, req.Score)
        if err != nil {
            log.Printf("❌ Memory validation error: %v", err)
            return &SubmitScoreResponse{
                Success: false,
                Message: "validation error",
            }, nil
        }
        if !valid {
            log.Printf("⚠️ Invalid memory game state for user %s", req.UserID)
            return &SubmitScoreResponse{
                Success: false,
                Message: "invalid game state or moves",
            }, nil
        }

        validatedScore = valScore
        lampsEarned, ticketsEarned = game.CalculateRewards(validatedScore)

    case "flappy":
        if req.Score < 0 || req.Score > 10000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        if req.Level < 1 || req.Level > 10 {
            return &SubmitScoreResponse{Success: false, Message: "flappy level must be 1-10"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 5 + (req.Level - 1)
        ticketsEarned = req.Score/10 + (req.Level-1)*5
        if ticketsEarned > 100 {
            ticketsEarned = 100
        }
    case "towers":
        if req.Score < 0 || req.Score > 10000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 5
        ticketsEarned = req.Score / 20
        if ticketsEarned > 50 {
            ticketsEarned = 50
        }

    case "hanoi":
        // Score is computed client-side as 1000 - (excess moves × 20), floored at 100
        // (same formula as Memory), so the valid range is always [100, 1000].
        if req.Score < 100 || req.Score > 1000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 5
        ticketsEarned = 5
        if validatedScore > 500 {
            ticketsEarned += (validatedScore - 500) / 100
        }
        if ticketsEarned > 20 {
            ticketsEarned = 20
        }
        if validatedScore >= 900 {
            lampsEarned += 5
        }

    case "twenty48":
        if req.Score < 0 || req.Score > 100000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 5
        ticketsEarned = req.Score / 200
        if ticketsEarned > 80 {
            ticketsEarned = 80
        }

    case "gears":
        if req.Score < 0 || req.Score > 50000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 5
        ticketsEarned = req.Score / 50
        if ticketsEarned > 60 {
            ticketsEarned = 60
        }

    case "companion":
        if req.Score < 0 || req.Score > 10000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 3
        ticketsEarned = req.Score / 100
        if ticketsEarned > 30 {
            ticketsEarned = 30
        }

    default:
        return &SubmitScoreResponse{
            Success: false,
            Message: fmt.Sprintf("unknown game_id: %s", req.GameID),
        }, nil
    }

    log.Printf("📥 Validated score: %d", validatedScore)

    // Kids-safe: any unconsumed boost for this user+game skips leaderboard + rewards.
    boostID := strings.TrimSpace(req.BoostID)
    if boostID == "" {
        if activeID, ok, lookupErr := s.repo.GetActiveRunBoost(ctx, req.UserID, req.GameID); lookupErr != nil {
            log.Printf("lookup boost failed: %v", lookupErr)
        } else if ok {
            boostID = activeID
        }
    }
    consumedID, boosted, err := s.repo.ConsumeActiveRunBoost(ctx, req.UserID, req.GameID, boostID)
    if err != nil {
        log.Printf("consume boost failed: %v", err)
    }
    // If client sent a stale boost_id, still block on any remaining active boost.
    if !boosted {
        if activeID, ok, lookupErr := s.repo.GetActiveRunBoost(ctx, req.UserID, req.GameID); lookupErr == nil && ok {
            if cid, cok, cerr := s.repo.ConsumeActiveRunBoost(ctx, req.UserID, req.GameID, activeID); cerr == nil && cok {
                consumedID, boosted = cid, true
            }
        }
    }
    if boosted {
        log.Printf("🚫 Boosted run not ranked: user=%s game=%s boost=%s score=%d",
            req.UserID, req.GameID, consumedID, validatedScore)
        return &SubmitScoreResponse{
            Success:       true,
            NewHighscore:  validatedScore,
            Rank:          0,
            Message:       "boosted run — not ranked",
            LampsEarned:   0,
            TicketsEarned: 0,
            Ranked:        false,
        }, nil
    }

    level := req.Level
    if level < 1 {
        level = 1
    }
    // Only Flappy partitions the leaderboard by difficulty level.
    // Other games' in-run "level" must not fan out into empty boards (UI reads level=1).
    if req.GameID != "flappy" {
        level = 1
    }

    // Получаем текущий рекорд (per-level)
    currentHighscore, err := s.repo.GetHighscore(ctx, req.UserID, req.GameID, level)
    if err != nil {
        log.Printf("Failed to get highscore: %v", err)
    }

    isNewRecord := validatedScore > currentHighscore

    log.Printf("🎯 isNewRecord=%v, validatedScore=%d, level=%d, lamps=%d, tickets=%d",
        isNewRecord, validatedScore, level, lampsEarned, ticketsEarned)

    // Event for Leaderboard / Billing (same payload shape as before).
    event := map[string]interface{}{
        "user_id":         req.UserID,
        "game_id":         req.GameID,
        "user_email":      req.UserEmail,
        "nickname":        req.Nickname,
        "score":           validatedScore,
        "is_record":       isNewRecord,
        "level":           level,
        "lamps_earned":    lampsEarned,
        "tickets_earned":  ticketsEarned,
        "timestamp":       time.Now().Unix(),
    }
    eventData, _ := json.Marshal(event)

    // Prefer transactional outbox (reliable). Fall back to legacy direct NATS publish
    // if outbox insert fails (e.g. migration not yet applied).
    if isNewRecord {
        if err := s.repo.SaveHighscoreAndEnqueueOutbox(ctx, req.UserID, req.GameID, validatedScore, level, "score.updated", eventData); err != nil {
            log.Printf("outbox+highscore failed, falling back to SaveHighscore+Publish: %v", err)
            if saveErr := s.repo.SaveHighscore(ctx, req.UserID, req.GameID, validatedScore, level); saveErr != nil {
                log.Printf("Failed to save highscore: %v", saveErr)
            }
            s.publishScoreUpdatedDirect(eventData, req.UserID, req.GameID, validatedScore, isNewRecord, lampsEarned, ticketsEarned)
        } else {
            log.Printf("📬 Enqueued score.updated (with highscore): user=%s game=%s level=%d score=%d", req.UserID, req.GameID, level, validatedScore)
        }
    } else {
        if err := s.repo.EnqueueOutbox(ctx, "score.updated", eventData); err != nil {
            log.Printf("outbox enqueue failed, falling back to direct Publish: %v", err)
            s.publishScoreUpdatedDirect(eventData, req.UserID, req.GameID, validatedScore, isNewRecord, lampsEarned, ticketsEarned)
        } else {
            log.Printf("📬 Enqueued score.updated: user=%s game=%s level=%d score=%d", req.UserID, req.GameID, level, validatedScore)
        }
    }

    return &SubmitScoreResponse{
        Success:       true,
        NewHighscore:  validatedScore,
        Rank:          0,
        Message:       "score submitted successfully",
        LampsEarned:   lampsEarned,
        TicketsEarned: ticketsEarned,
        Ranked:        true,
    }, nil
}

func (s *gameService) StartBoost(ctx context.Context, req *StartBoostRequest) (*StartBoostResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "request is required")
    }
    gameID := strings.TrimSpace(req.GameID)
    userID := strings.TrimSpace(req.UserID)
    if userID == "" || gameID == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id and game_id are required")
    }
    if _, ok := boostAllowedGames[gameID]; !ok {
        return nil, status.Errorf(codes.FailedPrecondition, "boosts not available for game %s", gameID)
    }
    if s.billingClient == nil {
        return nil, status.Error(codes.Unavailable, "billing unavailable")
    }

    // Reuse an unpaid-for active boost (crash / reconnect) without double-charging.
    if existing, ok, err := s.repo.GetActiveRunBoost(ctx, userID, gameID); err != nil {
        return nil, status.Errorf(codes.Internal, "lookup boost: %v", err)
    } else if ok {
        return &StartBoostResponse{
            BoostID:    existing,
            Boosted:    true,
            Cost:       0,
            NewBalance: -1,
            Message:    "existing boost reused",
        }, nil
    }

    boostID := uuid.NewString()
    spendResp, err := s.billingClient.SpendCurrency(ctx, &billingPb.SpendCurrencyRequest{
        UserId:      userID,
        Currency:    billingPb.CurrencyType_LAMPS,
        Amount:      boostCostLamps,
        Reason:      "game_boost",
        ReferenceId: "boost:" + boostID,
    })
    if err != nil {
        if st, ok := status.FromError(err); ok && st.Code() == codes.FailedPrecondition {
            return nil, status.Error(codes.FailedPrecondition, st.Message())
        }
        if strings.Contains(err.Error(), "insufficient") {
            return nil, status.Error(codes.FailedPrecondition, err.Error())
        }
        return nil, status.Errorf(codes.Internal, "spend lamps: %v", err)
    }
    if spendResp != nil && !spendResp.GetSuccess() {
        msg := spendResp.GetMessage()
        if msg == "" {
            msg = "failed to spend lamps"
        }
        return nil, status.Error(codes.FailedPrecondition, msg)
    }

    if err := s.repo.CreateRunBoost(ctx, boostID, userID, gameID); err != nil {
        return nil, status.Errorf(codes.Internal, "record boost: %v", err)
    }

    s.publishBoostStarted(boostID, userID, gameID)

    newBalance := 0
    if spendResp != nil {
        newBalance = int(spendResp.GetNewBalance())
    }
    return &StartBoostResponse{
        BoostID:    boostID,
        Boosted:    true,
        Cost:       boostCostLamps,
        NewBalance: newBalance,
        Message:    "boost armed — this run will not be ranked",
    }, nil
}

func (s *gameService) publishBoostStarted(boostID, userID, gameID string) {
    if s.js == nil {
        return
    }
    body, err := json.Marshal(map[string]any{
        "event_uuid": uuid.NewString(),
        "event_type": "BoostStarted",
        "user_uuid":  userID,
        "game_id":    gameID,
        "boost_uuid": boostID,
        "cost_lamps": boostCostLamps,
    })
    if err != nil {
        log.Printf("boost.started marshal: %v", err)
        return
    }
    if _, err := s.js.Publish("boost.started", body); err != nil {
        log.Printf("boost.started publish: %v", err)
        return
    }
    log.Printf("📡 Published boost.started: user=%s game=%s boost=%s", userID, gameID, boostID)
}

// publishScoreUpdatedDirect is the legacy path (kept for fallback / tests).
func (s *gameService) publishScoreUpdatedDirect(eventData []byte, userID, gameID string, score int, isRecord bool, lamps, tickets int) {
    if s.js == nil {
        log.Printf("NATS JetStream nil; skip direct score.updated publish")
        return
    }
    _, err := s.js.Publish("score.updated", eventData)
    if err != nil {
        log.Printf("Failed to publish to NATS: %v", err)
        return
    }
    log.Printf("📡 Published score.updated (direct): user=%s, game=%s, score=%d, is_record=%v, lamps=%d, tickets=%d",
        userID, gameID, score, isRecord, lamps, tickets)
}

func (s *gameService) GetGameInfo(ctx context.Context, gameID string) (*GameInfo, error) {
    switch gameID {
    case "hexagon":
        return &GameInfo{
            GameID:      "hexagon",
            Name:        "Блинопёк",
            Description: "Гексагональный пазл с блинами",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 100, RewardLamps: 10, RewardTickets: 0},
                {Level: 2, TargetScore: 200, RewardLamps: 15, RewardTickets: 0},
                {Level: 3, TargetScore: 350, RewardLamps: 20, RewardTickets: 5},
                {Level: 4, TargetScore: 550, RewardLamps: 30, RewardTickets: 10},
                {Level: 5, TargetScore: 800, RewardLamps: 50, RewardTickets: 20},
            },
        }, nil
    case "memory":
        return &GameInfo{
            GameID:      "memory",
            Name:        "Меморина",
            Description: "Найди пары фруктов",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 500, RewardLamps: 5, RewardTickets: 5},
                {Level: 2, TargetScore: 800, RewardLamps: 10, RewardTickets: 10},
                {Level: 3, TargetScore: 1000, RewardLamps: 15, RewardTickets: 15},
            },
        }, nil
    case "flappy":
        flappyLevels := make([]LevelInfo, 0, 10)
        for lv := 1; lv <= 10; lv++ {
            flappyLevels = append(flappyLevels, LevelInfo{
                Level:         lv,
                TargetScore:   10 * lv,
                RewardLamps:   5 + (lv - 1),
                RewardTickets: (lv - 1) * 5,
            })
        }
        return &GameInfo{
            GameID:      "flappy",
            Name:        "Flappy Bird",
            Description: "Трубы, птичка, полёт",
            Levels:      flappyLevels,
        }, nil
    case "towers":
        return &GameInfo{
            GameID:      "towers",
            Name:        "Башенки",
            Description: "Строй башню из падающих блоков",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 100, RewardLamps: 5, RewardTickets: 0},
                {Level: 2, TargetScore: 200, RewardLamps: 10, RewardTickets: 5},
                {Level: 3, TargetScore: 350, RewardLamps: 15, RewardTickets: 10},
            },
        }, nil
    case "hanoi":
        return &GameInfo{
            GameID:      "hanoi",
            Name:        "Ханойская башня",
            Description: "Переместите все кольца на третий стержень за минимум ходов",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 500, RewardLamps: 5, RewardTickets: 5},
                {Level: 2, TargetScore: 800, RewardLamps: 8, RewardTickets: 10},
                {Level: 3, TargetScore: 1000, RewardLamps: 10, RewardTickets: 20},
            },
        }, nil
    case "twenty48":
        return &GameInfo{
            GameID:      "twenty48",
            Name:        "Горизонт 2048",
            Description: "Сдвигай плитки, собери 2048",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 500, RewardLamps: 5, RewardTickets: 5},
                {Level: 2, TargetScore: 2000, RewardLamps: 10, RewardTickets: 15},
                {Level: 3, TargetScore: 5000, RewardLamps: 15, RewardTickets: 30},
            },
        }, nil
    case "gears":
        return &GameInfo{
            GameID:      "gears",
            Name:        "Орбиты",
            Description: "Сливай шестерёнки до восьмой",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 80, RewardLamps: 5, RewardTickets: 5},
                {Level: 2, TargetScore: 200, RewardLamps: 10, RewardTickets: 15},
                {Level: 3, TargetScore: 400, RewardLamps: 15, RewardTickets: 25},
            },
        }, nil
    case "companion":
        return &GameInfo{
            GameID:      "companion",
            Name:        "Компаньон",
            Description: "Мягкий тамагочи без FOMO-смерти",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 50, RewardLamps: 3, RewardTickets: 3},
                {Level: 2, TargetScore: 150, RewardLamps: 5, RewardTickets: 8},
                {Level: 3, TargetScore: 400, RewardLamps: 8, RewardTickets: 15},
            },
        }, nil
    default:
        return nil, fmt.Errorf("game not found: %s", gameID)
    }
}
--- services/game/internal/service/game_service_test.go ---
package service

import (
	"context"
	"errors"
	"testing"

	billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
	"github.com/Eastwesser/event-horizon/services/game/internal/repository"
	"google.golang.org/grpc"
)

type mockGameRepo struct {
	highscore   int
	saveHS      bool
	enqueue     bool
	saveEnqueue bool
	activeBoost string
	consumeOK   bool
	consumeID   string
	createBoost bool
	err         error
}

func (m *mockGameRepo) GetHighscore(context.Context, string, string, int) (int, error) {
	return m.highscore, m.err
}
func (m *mockGameRepo) SaveHighscore(context.Context, string, string, int, int) error {
	m.saveHS = true
	return m.err
}
func (m *mockGameRepo) EnqueueOutbox(context.Context, string, []byte) error {
	m.enqueue = true
	return m.err
}
func (m *mockGameRepo) SaveHighscoreAndEnqueueOutbox(context.Context, string, string, int, int, string, []byte) error {
	m.saveEnqueue = true
	return m.err
}
func (m *mockGameRepo) CreateRunBoost(context.Context, string, string, string) error {
	m.createBoost = true
	return m.err
}
func (m *mockGameRepo) GetActiveRunBoost(context.Context, string, string) (string, bool, error) {
	if m.activeBoost == "" {
		return "", false, nil
	}
	return m.activeBoost, true, nil
}
func (m *mockGameRepo) ConsumeActiveRunBoost(context.Context, string, string, string) (string, bool, error) {
	if m.consumeOK {
		return m.consumeID, true, nil
	}
	return "", false, nil
}

type stubBilling struct {
	spendOK  bool
	spendBal int32
	spendErr error
}

func (b *stubBilling) GetBalance(context.Context, *billingPb.GetBalanceRequest, ...grpc.CallOption) (*billingPb.GetBalanceResponse, error) {
	return nil, errors.New("unused")
}
func (b *stubBilling) GetAllBalances(context.Context, *billingPb.GetAllBalancesRequest, ...grpc.CallOption) (*billingPb.GetAllBalancesResponse, error) {
	return nil, errors.New("unused")
}
func (b *stubBilling) AddCurrency(context.Context, *billingPb.AddCurrencyRequest, ...grpc.CallOption) (*billingPb.AddCurrencyResponse, error) {
	return nil, errors.New("unused")
}
func (b *stubBilling) SpendCurrency(context.Context, *billingPb.SpendCurrencyRequest, ...grpc.CallOption) (*billingPb.SpendCurrencyResponse, error) {
	if b.spendErr != nil {
		return nil, b.spendErr
	}
	return &billingPb.SpendCurrencyResponse{Success: b.spendOK, NewBalance: b.spendBal}, nil
}
func (b *stubBilling) GetTransactionHistory(context.Context, *billingPb.GetTransactionHistoryRequest, ...grpc.CallOption) (*billingPb.GetTransactionHistoryResponse, error) {
	return nil, errors.New("unused")
}

func TestGetGameInfo(t *testing.T) {
	svc := NewGameService(&mockGameRepo{}, nil, nil)
	for _, id := range []string{"hexagon", "memory", "flappy", "towers", "hanoi", "twenty48", "gears", "companion"} {
		info, err := svc.GetGameInfo(context.Background(), id)
		if err != nil || info == nil || info.GameID != id {
			t.Fatalf("%s: %+v err=%v", id, info, err)
		}
	}
	_, err := svc.GetGameInfo(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected not found")
	}
}

func TestSubmitScore_UnknownAndFlappyValidation(t *testing.T) {
	svc := NewGameService(&mockGameRepo{}, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{GameID: "nope", UserID: "u"})
	if err != nil || resp.Success {
		t.Fatalf("%+v err=%v", resp, err)
	}
	resp, err = svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		GameID: "flappy", UserID: "u", Score: -1, Level: 1,
	})
	if err != nil || resp.Success {
		t.Fatalf("bad score: %+v", resp)
	}
	resp, err = svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		GameID: "flappy", UserID: "u", Score: 10, Level: 99,
	})
	if err != nil || resp.Success {
		t.Fatalf("bad level: %+v", resp)
	}
}

func TestSubmitScore_FlappyNewRecord(t *testing.T) {
	repo := &mockGameRepo{highscore: 5}
	svc := NewGameService(repo, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u1", GameID: "flappy", Score: 50, Level: 2, UserEmail: "a@b.c", Nickname: "n",
	})
	if err != nil || !resp.Success || !resp.Ranked || resp.LampsEarned <= 0 {
		t.Fatalf("%+v err=%v", resp, err)
	}
	if !repo.saveEnqueue {
		t.Fatal("expected SaveHighscoreAndEnqueueOutbox")
	}
}

func TestSubmitScore_FlappyNonRecord(t *testing.T) {
	repo := &mockGameRepo{highscore: 100}
	svc := NewGameService(repo, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u1", GameID: "flappy", Score: 10, Level: 1,
	})
	if err != nil || !resp.Success || !repo.enqueue {
		t.Fatalf("%+v enqueue=%v err=%v", resp, repo.enqueue, err)
	}
}

func TestSubmitScore_BoostedNotRanked(t *testing.T) {
	repo := &mockGameRepo{consumeOK: true, consumeID: "b1"}
	svc := NewGameService(repo, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u1", GameID: "flappy", Score: 20, Level: 1, BoostID: "b1",
	})
	if err != nil || !resp.Success || resp.Ranked || resp.LampsEarned != 0 {
		t.Fatalf("%+v err=%v", resp, err)
	}
}

func TestSubmitScore_TowersAndHanoi(t *testing.T) {
	repo := &mockGameRepo{highscore: 0}
	svc := NewGameService(repo, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "towers", Score: 100, Level: 1,
	})
	if err != nil || !resp.Success || resp.TicketsEarned != 5 {
		t.Fatalf("towers %+v err=%v", resp, err)
	}
	resp, err = svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "hanoi", Score: 950, Level: 1,
	})
	if err != nil || !resp.Success || resp.LampsEarned < 5 {
		t.Fatalf("hanoi %+v err=%v", resp, err)
	}
}

func TestSubmitScore_MoreGamesAndOutboxFallback(t *testing.T) {
	repo := &mockGameRepo{highscore: 0, err: errors.New("outbox down")}
	// err on SaveHighscoreAndEnqueueOutbox triggers fallback SaveHighscore
	svc := NewGameService(repo, nil, nil)
	for _, tc := range []struct {
		game  string
		score int
	}{
		{"twenty48", 400},
		{"gears", 100},
		{"companion", 50},
	} {
		repo.saveEnqueue, repo.saveHS, repo.enqueue = false, false, false
		repo.err = nil // happy enqueue for these
		resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
			UserID: "u", GameID: tc.game, Score: tc.score, Level: 1,
		})
		if err != nil || !resp.Success {
			t.Fatalf("%s %+v err=%v", tc.game, resp, err)
		}
	}
	// force outbox failure on new record → SaveHighscore fallback (+ nil js publish skip)
	repo = &mockGameRepo{highscore: 0}
	repo.err = errors.New("outbox down")
	// SaveHighscoreAndEnqueueOutbox returns err; SaveHighscore also uses m.err — need split
	// use custom: first call fails enqueue+hs, SaveHighscore succeeds if we clear err after
	failOnce := &failEnqueueRepo{mockGameRepo: mockGameRepo{highscore: 0}}
	svc = NewGameService(failOnce, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "flappy", Score: 30, Level: 1,
	})
	if err != nil || !resp.Success || !failOnce.savedHS {
		t.Fatalf("fallback %+v saved=%v err=%v", resp, failOnce.savedHS, err)
	}
	// non-record outbox fail → direct publish skip
	failEnq := &failEnqueueRepo{mockGameRepo: mockGameRepo{highscore: 1000}}
	svc = NewGameService(failEnq, nil, nil)
	resp, err = svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "flappy", Score: 1, Level: 1,
	})
	if err != nil || !resp.Success {
		t.Fatalf("non-record fallback %+v err=%v", resp, err)
	}
}

type failEnqueueRepo struct {
	mockGameRepo
	savedHS bool
}

func (f *failEnqueueRepo) SaveHighscoreAndEnqueueOutbox(context.Context, string, string, int, int, string, []byte) error {
	return errors.New("outbox down")
}
func (f *failEnqueueRepo) SaveHighscore(context.Context, string, string, int, int) error {
	f.savedHS = true
	return nil
}
func (f *failEnqueueRepo) EnqueueOutbox(context.Context, string, []byte) error {
	return errors.New("outbox down")
}

func TestSubmitScore_ActiveBoostLookup(t *testing.T) {
	repo := &mockGameRepo{activeBoost: "ab1", consumeOK: true, consumeID: "ab1"}
	resp, err := NewGameService(repo, nil, nil).SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "flappy", Score: 10, Level: 1, // no BoostID — lookup path
	})
	if err != nil || resp.Ranked {
		t.Fatalf("%+v err=%v", resp, err)
	}
}

func TestStartBoost_BillingUnavailableAndSpendFail(t *testing.T) {
	_, err := NewGameService(&mockGameRepo{}, nil, nil).StartBoost(context.Background(), &StartBoostRequest{UserID: "u", GameID: "flappy"})
	if err == nil {
		t.Fatal("expected billing unavailable")
	}
	_, err = NewGameService(&mockGameRepo{}, nil, &stubBilling{spendErr: errors.New("insufficient balance")}).
		StartBoost(context.Background(), &StartBoostRequest{UserID: "u", GameID: "flappy"})
	if err == nil {
		t.Fatal("expected spend fail")
	}
	_, err = NewGameService(&mockGameRepo{}, nil, &stubBilling{spendOK: false}).
		StartBoost(context.Background(), &StartBoostRequest{UserID: "u", GameID: "flappy"})
	if err == nil {
		t.Fatal("expected success=false")
	}
}

func TestStartBoost_AllowedGames(t *testing.T) {
	repo := &mockGameRepo{}
	billing := &stubBilling{spendOK: true, spendBal: 90}
	svc := NewGameService(repo, nil, billing)
	resp, err := svc.StartBoost(context.Background(), &StartBoostRequest{UserID: "u1", GameID: "flappy"})
	if err != nil || !resp.Boosted || resp.Cost != 10 || !repo.createBoost {
		t.Fatalf("flappy: %+v err=%v", resp, err)
	}
	repo2 := &mockGameRepo{}
	svc2 := NewGameService(repo2, nil, billing)
	resp, err = svc2.StartBoost(context.Background(), &StartBoostRequest{UserID: "u1", GameID: "towers"})
	if err != nil || !resp.Boosted || resp.Cost != 10 || !repo2.createBoost {
		t.Fatalf("towers: %+v err=%v", resp, err)
	}
	_, err = svc.StartBoost(context.Background(), &StartBoostRequest{UserID: "u1", GameID: "unknown_game"})
	if err == nil {
		t.Fatal("expected unknown game rejected")
	}
}

func TestStartBoost_ReusesActive(t *testing.T) {
	repo := &mockGameRepo{activeBoost: "existing"}
	svc := NewGameService(repo, nil, &stubBilling{spendOK: true})
	resp, err := svc.StartBoost(context.Background(), &StartBoostRequest{UserID: "u1", GameID: "flappy"})
	if err != nil || !resp.Boosted || resp.BoostID != "existing" || repo.createBoost {
		t.Fatalf("%+v create=%v err=%v", resp, repo.createBoost, err)
	}
}

// silence unused import if repository only used via interface satisfaction
var _ repository.GameRepository = (*mockGameRepo)(nil)

func TestSubmitScore_HexagonMemorySmoke(t *testing.T) {
	svc := NewGameService(&mockGameRepo{highscore: 0}, nil, nil)
	// Exercise validator branches (empty seed/moves — outcome depends on validators).
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "hexagon", Score: 10, Seed: "", Moves: nil,
	})
	if err != nil || resp == nil {
		t.Fatalf("%+v err=%v", resp, err)
	}
	resp, err = svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "memory", Score: 10, Seed: "x", Moves: nil,
	})
	if err != nil || resp == nil {
		t.Fatalf("%+v err=%v", resp, err)
	}
}

func TestStartBoost_NilRequest(t *testing.T) {
	_, err := NewGameService(&mockGameRepo{}, nil, &stubBilling{spendOK: true}).StartBoost(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
===== repository =====
--- services/game/internal/repository/game_repo.go ---
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
)

type GameRepository interface {
	GetHighscore(ctx context.Context, userID, gameID string, level int) (int, error)
	SaveHighscore(ctx context.Context, userID, gameID string, score, level int) error
	// EnqueueOutbox inserts a NATS event for the outbox worker (no highscore change).
	EnqueueOutbox(ctx context.Context, eventType string, payload []byte) error
	// SaveHighscoreAndEnqueueOutbox writes highscore + outbox row in one transaction.
	SaveHighscoreAndEnqueueOutbox(ctx context.Context, userID, gameID string, score, level int, eventType string, payload []byte) error

	CreateRunBoost(ctx context.Context, id, userID, gameID string) error
	GetActiveRunBoost(ctx context.Context, userID, gameID string) (id string, ok bool, err error)
	// ConsumeActiveRunBoost marks the active boost consumed. If boostID is non-empty,
	// only that row is consumed; otherwise any active boost for the pair is consumed.
	ConsumeActiveRunBoost(ctx context.Context, userID, gameID, boostID string) (consumedID string, ok bool, err error)
}

// OutboxRecord is optional metadata for callers that build payloads separately.
type OutboxRecord struct {
	EventType string
	Payload   json.RawMessage
}

type PostgresGameRepo struct {
	db *sql.DB
}

func NewPostgresGameRepo(db *sql.DB) *PostgresGameRepo {
	return &PostgresGameRepo{db: db}
}

func normalizeHighscoreLevel(level int) int {
	if level < 1 {
		return 1
	}
	return level
}

func (r *PostgresGameRepo) GetHighscore(ctx context.Context, userID, gameID string, level int) (int, error) {
	level = normalizeHighscoreLevel(level)
	var score int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(score, 0) FROM highscores WHERE user_id = $1 AND game_id = $2 AND level = $3",
		userID, gameID, level,
	).Scan(&score)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return score, err
}

func (r *PostgresGameRepo) SaveHighscore(ctx context.Context, userID, gameID string, score, level int) error {
	level = normalizeHighscoreLevel(level)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO highscores (user_id, game_id, level, score, updated_at)
         VALUES ($1, $2, $3, $4, NOW())
         ON CONFLICT (user_id, game_id, level) DO UPDATE
         SET score = EXCLUDED.score, updated_at = NOW()`,
		userID, gameID, level, score,
	)
	return err
}

func (r *PostgresGameRepo) EnqueueOutbox(ctx context.Context, eventType string, payload []byte) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`,
		eventType, payload,
	)
	return err
}

func (r *PostgresGameRepo) SaveHighscoreAndEnqueueOutbox(ctx context.Context, userID, gameID string, score, level int, eventType string, payload []byte) error {
	level = normalizeHighscoreLevel(level)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO highscores (user_id, game_id, level, score, updated_at)
         VALUES ($1, $2, $3, $4, NOW())
         ON CONFLICT (user_id, game_id, level) DO UPDATE
         SET score = EXCLUDED.score, updated_at = NOW()`,
		userID, gameID, level, score,
	); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`,
		eventType, payload,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresGameRepo) CreateRunBoost(ctx context.Context, id, userID, gameID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO run_boosts (id, user_id, game_id, created_at)
         VALUES ($1::uuid, $2, $3, NOW())`,
		id, userID, gameID,
	)
	return err
}

func (r *PostgresGameRepo) GetActiveRunBoost(ctx context.Context, userID, gameID string) (string, bool, error) {
	var id string
	err := r.db.QueryRowContext(ctx,
		`SELECT id::text FROM run_boosts
         WHERE user_id = $1 AND game_id = $2 AND consumed_at IS NULL
         ORDER BY created_at DESC
         LIMIT 1`,
		userID, gameID,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

func (r *PostgresGameRepo) ConsumeActiveRunBoost(ctx context.Context, userID, gameID, boostID string) (string, bool, error) {
	var id string
	var err error
	if boostID != "" {
		err = r.db.QueryRowContext(ctx,
			`UPDATE run_boosts
             SET consumed_at = NOW()
             WHERE id = $1::uuid AND user_id = $2 AND game_id = $3 AND consumed_at IS NULL
             RETURNING id::text`,
			boostID, userID, gameID,
		).Scan(&id)
	} else {
		err = r.db.QueryRowContext(ctx,
			`UPDATE run_boosts
             SET consumed_at = NOW()
             WHERE id = (
                 SELECT id FROM run_boosts
                 WHERE user_id = $1 AND game_id = $2 AND consumed_at IS NULL
                 ORDER BY created_at DESC
                 LIMIT 1
             )
             RETURNING id::text`,
			userID, gameID,
		).Scan(&id)
	}
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}
===== converter =====
--- services/game/internal/converter/score.go ---
package converter

import (
	"github.com/Eastwesser/event-horizon/services/game/internal/model"
	pb "github.com/Eastwesser/event-horizon/services/game/proto"
)

func SubmitScoreFromProto(req *pb.SubmitScoreRequest) model.ScoreSubmission {
	if req == nil {
		return model.ScoreSubmission{}
	}
	return model.ScoreSubmission{
		UserID:   req.UserId,
		GameID:   req.GameId,
		Level:    req.Level,
		Score:    req.Score,
		Nickname: req.Nickname,
	}
}
--- services/game/internal/converter/score_test.go ---
package converter

import (
	"testing"

	pb "github.com/Eastwesser/event-horizon/services/game/proto"
)

func TestSubmitScoreFromProto(t *testing.T) {
	m := SubmitScoreFromProto(&pb.SubmitScoreRequest{UserId: "u1", GameId: "hexagon", Level: 2, Score: 100})
	if m.UserID != "u1" || m.GameID != "hexagon" || m.Score != 100 {
		t.Fatalf("%+v", m)
	}
}
===== worker =====
--- services/game/internal/worker/outbox_worker.go ---
package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

// OutboxWorker publishes unpublished game outbox rows to NATS JetStream.
// Rows are locked with SKIP LOCKED so two processes cannot publish the same score twice.
type OutboxWorker struct {
	db *sql.DB
	js nats.JetStreamContext
}

func NewOutboxWorker(db *sql.DB, js nats.JetStreamContext) *OutboxWorker {
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

type outboxRow struct {
	id        string
	eventType string
	payload   []byte
}

// processBatch locks a page of unpublished rows. A second worker skips those locks
// and cannot publish the same score.updated twice.
func (w *OutboxWorker) processBatch(ctx context.Context) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("game outbox: begin", "err", err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
        SELECT id, event_type, payload FROM outbox
        WHERE processed = false
        ORDER BY created_at
        LIMIT 20
        FOR UPDATE SKIP LOCKED
    `)
	if err != nil {
		slog.Error("game outbox: query", "err", err)
		return
	}
	var batch []outboxRow
	for rows.Next() {
		var row outboxRow
		if err := rows.Scan(&row.id, &row.eventType, &row.payload); err != nil {
			slog.Error("game outbox: scan", "err", err)
			continue
		}
		batch = append(batch, row)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		slog.Error("game outbox: rows", "err", err)
		return
	}

	for _, row := range batch {
		if w.js == nil {
			slog.Error("game outbox: jetstream is not connected", "id", row.id)
			continue
		}
		if _, err := w.js.Publish(row.eventType, row.payload); err != nil {
			slog.Error("game outbox: publish", "event_type", row.eventType, "id", row.id, "err", err)
			continue
		}
		if _, err := tx.ExecContext(ctx, `
            UPDATE outbox SET processed = true, processed_at = NOW() WHERE id = $1
        `, row.id); err != nil {
			slog.Error("game outbox: mark processed", "id", row.id, "err", err)
			return
		}
		slog.Info("game outbox: published", "event_type", row.eventType, "id", row.id)
	}
	if err := tx.Commit(); err != nil {
		slog.Error("game outbox: commit", "err", err)
	}
}
[denismatveev@c0der event_horizon]$ 
