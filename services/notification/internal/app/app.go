package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/services/notification/internal/config"
	"github.com/Eastwesser/event-horizon/services/notification/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
	cfg      *config.Config
	log      *slog.Logger
	closer   *closer.Closer
	consumer kafka.Consumer
	notifier *service.Notifier
	nc       *nats.Conn
	js       nats.JetStreamContext
	http     *http.Server
}

func New(ctx context.Context) (*App, error) {
	_ = ctx
	cfg := config.Load()
	log := logger.NewFrom(getenv("LOG_LEVEL", "info"), getenv("LOG_FORMAT", "text"))
	a := &App{cfg: cfg, log: log, closer: closer.New(log)}

	kcfg := kafka.LoadConfig()
	if cfg.KafkaBrokers != "" {
		parts := strings.Split(cfg.KafkaBrokers, ",")
		var brokers []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				brokers = append(brokers, p)
			}
		}
		kcfg = kafka.Config{Brokers: brokers, Enabled: len(brokers) > 0}
	}
	a.consumer = kafka.NewConsumer(kcfg, kafka.ConsumerGroupNotify,
		[]string{kafka.TopicPurchasePaid, kafka.TopicPurchaseFulfilled}, log)
	a.closer.AddNamed("kafka consumer", func(context.Context) error { return a.consumer.Close() })
	a.notifier = service.NewNotifier(log, cfg.TelegramToken, cfg.TelegramChatID)

	if cfg.NATSURL != "" {
		nc, err := nats.Connect(cfg.NATSURL)
		if err != nil {
			a.log.Warn("nats connect failed; Kafka-only path", "err", err)
		} else {
			js, err := nc.JetStream()
			if err != nil {
				a.log.Warn("jetstream failed", "err", err)
				_ = nc.Drain()
			} else {
				a.nc = nc
				a.js = js
				a.closer.AddNamed("nats", func(context.Context) error { return nc.Drain() })
				a.log.Info("nats connected", "url", cfg.NATSURL)
			}
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "notification"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "notification"})
	})
	mux.HandleFunc("/start", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Welcome to Event Horizon notifications bot",
		})
	})
	a.http = &http.Server{Addr: ":" + cfg.MetricsPort, Handler: mux}
	a.closer.AddNamed("http", func(ctx context.Context) error { return a.http.Shutdown(ctx) })
	return a, nil
}

func (a *App) RunUntilSignal(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		a.log.Info("notification http listening", "addr", a.http.Addr)
		if err := a.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("http error", "err", err)
		}
	}()

	if a.js != nil {
		go a.consumeNATS(runCtx)
	}
	if a.cfg.KafkaBrokers != "" {
		go func() {
			a.log.Info("notification consuming Kafka purchase topics")
			if err := a.consumer.Consume(runCtx, a.notifier.Handle); err != nil && runCtx.Err() == nil {
				a.log.Error("kafka consume stopped", "err", err)
			}
		}()
	}

	if err := a.closer.WaitSignal(ctx); err != nil {
		cancel()
		return fmt.Errorf("shutdown: %w", err)
	}
	cancel()
	return nil
}

func (a *App) consumeNATS(ctx context.Context) {
	subjects := []string{kafka.TopicPurchasePaid, kafka.TopicPurchaseFulfilled}
	for _, subj := range subjects {
		s := subj
		durable := "notification-" + strings.ReplaceAll(s, ".", "-")
		_, err := a.js.Subscribe(s, func(msg *nats.Msg) {
			err := a.notifier.Handle(ctx, kafka.Message{
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

func (a *App) Close(ctx context.Context) error { return a.closer.CloseAll(ctx) }

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
