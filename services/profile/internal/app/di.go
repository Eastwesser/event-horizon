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

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
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
	cacheTTL := time.Duration(a.cfg.ProfileCacheTTLMin) * time.Minute
	a.di.svc = service.NewProfileService(a.di.repo, a.di.redisRepo, cacheTTL)
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
		}
		_ = msg.Ack()
	}, nats.Durable("profile-score-updated"), nats.ManualAck())
	if err != nil {
		a.log.Warn("failed to subscribe to score.updated", "err", err)
	} else {
		a.log.Info("subscribed to nats", "subject", "score.updated")
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
