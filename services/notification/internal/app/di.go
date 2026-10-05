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
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/services/notification/internal/config"
	"github.com/Eastwesser/event-horizon/services/notification/internal/handler"
	"github.com/Eastwesser/event-horizon/services/notification/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/notification/internal/repository"
	"github.com/Eastwesser/event-horizon/services/notification/internal/service"
	"github.com/Eastwesser/event-horizon/services/notification/internal/worker"
	"github.com/Eastwesser/event-horizon/services/notification/migrations"
	pb "github.com/Eastwesser/event-horizon/services/notification/proto"
)

type diContainer struct {
	cfg      *config.Config
	dbpool   *pgxpool.Pool
	js       nats.JetStreamContext
	nc       *nats.Conn
	repo     *repository.PostgresRepo
	inbox    *service.InboxService
	notifier *service.Notifier
	api      pb.NotificationServiceServer
	consumer kafka.Consumer
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
	a.di.inbox = service.NewInbox(a.di.repo, a.log)
	a.di.notifier = service.NewNotifier(a.log, a.cfg.TelegramToken, a.cfg.TelegramChatID)
	a.di.api = handler.NewGRPCHandler(a.di.inbox)

	kcfg := kafka.LoadConfig()
	if a.cfg.KafkaBrokers != "" {
		parts := strings.Split(a.cfg.KafkaBrokers, ",")
		var brokers []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				brokers = append(brokers, p)
			}
		}
		kcfg = kafka.Config{Brokers: brokers, Enabled: len(brokers) > 0}
	}
	a.di.consumer = kafka.NewConsumer(kcfg, kafka.ConsumerGroupNotify,
		[]string{kafka.TopicPurchasePaid, kafka.TopicPurchaseFulfilled}, a.log)
	a.closer.AddNamed("kafka consumer", func(context.Context) error { return a.di.consumer.Close() })
	return nil
}

func (a *App) initWorkers(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(context.Background())
	go worker.NewInboxConsumer(a.di.js, a.di.inbox, a.log).Start(runCtx)
	go a.consumePurchaseNATS(runCtx)
	if a.cfg.KafkaBrokers != "" {
		go func() {
			a.log.Info("notification consuming Kafka purchase topics")
			if err := a.di.consumer.Consume(runCtx, a.di.notifier.Handle); err != nil && runCtx.Err() == nil {
				a.log.Error("kafka consume stopped", "err", err)
			}
		}()
	}
	a.closer.AddNamed("workers", func(context.Context) error { cancel(); return nil })
	_ = ctx
	return nil
}

func (a *App) consumePurchaseNATS(ctx context.Context) {
	subjects := []string{kafka.TopicPurchasePaid, kafka.TopicPurchaseFulfilled}
	for _, subj := range subjects {
		s := subj
		durable := "notification-" + strings.ReplaceAll(s, ".", "-")
		_, err := a.di.js.Subscribe(s, func(msg *nats.Msg) {
			err := a.di.notifier.Handle(ctx, kafka.Message{
				Topic: msg.Subject,
				Value: msg.Data,
			})
			if err != nil {
				a.log.Error("notify handle", "subject", msg.Subject, "err", err)
				_ = msg.Nak()
				return
			}
			_ = msg.Ack()
		}, nats.Durable(durable), nats.ManualAck())
		if err != nil {
			a.log.Error("nats subscribe failed", "subject", s, "err", err)
		} else {
			a.log.Info("notification consuming NATS", "subject", s, "durable", durable)
		}
	}
	<-ctx.Done()
}

func (a *App) initGRPC(_ context.Context) error {
	lis, err := net.Listen("tcp", ":"+a.cfg.GRPCPort)
	if err != nil {
		return err
	}
	a.listener = lis
	a.closer.AddNamed("tcp", func(context.Context) error { return lis.Close() })
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor.Recovery(), interceptor.Logger(), interceptor.Validate()))
	pb.RegisterNotificationServiceServer(s, a.di.api)
	reflection.Register(s)
	a.grpcServer = s
	a.closer.AddNamed("grpc", func(context.Context) error { s.GracefulStop(); return nil })
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "notification"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.di.dbpool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "notification"})
	})
	mux.HandleFunc("/start", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Welcome to Event Horizon notifications bot",
		})
	})
	srv := &http.Server{Addr: ":" + a.cfg.MetricsPort, Handler: mux}
	a.metricsServer = srv
	go func() { _ = srv.ListenAndServe() }()
	a.closer.AddNamed("metrics", func(ctx context.Context) error { return srv.Shutdown(ctx) })
	return nil
}
