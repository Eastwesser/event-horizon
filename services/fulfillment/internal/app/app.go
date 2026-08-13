package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/Eastwesser/event-horizon/services/fulfillment/internal/config"
	"github.com/Eastwesser/event-horizon/services/fulfillment/internal/service"
)

type App struct {
	cfg      *config.Config
	log      *slog.Logger
	closer   *closer.Closer
	consumer kafka.Consumer
	producer kafka.Producer
	svc      *service.FulfillmentService
	http     *http.Server
}

func New(ctx context.Context) (*App, error) {
	_ = ctx
	cfg := config.Load()
	log := logger.NewFrom(getenv("LOG_LEVEL", "info"), getenv("LOG_FORMAT", "text"))
	a := &App{cfg: cfg, log: log, closer: closer.New(log)}

	kcfg := kafka.LoadConfig()
	if cfg.KafkaBrokers != "" {
		kcfg = kafka.Config{Brokers: splitCSV(cfg.KafkaBrokers), Enabled: true}
	}
	a.producer = kafka.NewProducer(kcfg, kafka.TopicPurchaseFulfilled, log)
	a.consumer = kafka.NewConsumer(kcfg, kafka.ConsumerGroupFulfillment, []string{kafka.TopicPurchasePaid}, log)
	a.closer.AddNamed("kafka producer", func(context.Context) error { return a.producer.Close() })
	a.closer.AddNamed("kafka consumer", func(context.Context) error { return a.consumer.Close() })

	a.svc = service.New(a.producer, log, time.Duration(cfg.AssembleDelaySec)*time.Second)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "fulfillment"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "fulfillment"})
	})
	a.http = &http.Server{Addr: ":" + cfg.MetricsPort, Handler: mux}
	a.closer.AddNamed("metrics http", func(ctx context.Context) error { return a.http.Shutdown(ctx) })
	metrics.SetHealthy("fulfillment")
	return a, nil
}

func (a *App) RunUntilSignal(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		a.log.Info("fulfillment health listening", "addr", a.http.Addr)
		if err := a.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("http error", "err", err)
		}
	}()
	go func() {
		a.log.Info("fulfillment consuming", "topic", kafka.TopicPurchasePaid)
		if err := a.consumer.Consume(runCtx, a.svc.HandlePurchasePaid); err != nil && runCtx.Err() == nil {
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

func splitCSV(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			p := s[start:i]
			// trim spaces
			for len(p) > 0 && p[0] == ' ' {
				p = p[1:]
			}
			for len(p) > 0 && p[len(p)-1] == ' ' {
				p = p[:len(p)-1]
			}
			if p != "" {
				out = append(out, p)
			}
			start = i + 1
		}
	}
	return out
}
