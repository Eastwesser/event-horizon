package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/services/notification/internal/config"
	"github.com/Eastwesser/event-horizon/services/notification/internal/service"
)

type App struct {
	cfg      *config.Config
	log      *slog.Logger
	closer   *closer.Closer
	consumer kafka.Consumer
	notifier *service.Notifier
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

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "notification"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "notification"})
	})
	// Telegram bot /start (simple webhook/polls placeholder HTTP)
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
	go func() {
		a.log.Info("notification consuming purchase topics")
		if err := a.consumer.Consume(runCtx, a.notifier.Handle); err != nil && runCtx.Err() == nil {
			a.log.Error("consume stopped", "err", err)
		}
	}()
	if err := a.closer.WaitSignal(ctx); err != nil {
		cancel()
		return fmt.Errorf("shutdown: %w", err)
	}
	cancel()
	return nil
}

func (a *App) Close(ctx context.Context) error { return a.closer.CloseAll(ctx) }

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
