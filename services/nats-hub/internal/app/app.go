package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
)

// App bootstraps JetStream EVENTS stream and optional event logging.
type App struct {
	log    *slog.Logger
	closer *closer.Closer
	nc     *nats.Conn
	http   *http.Server
}

func New(ctx context.Context) (*App, error) {
	_ = ctx
	level := getenv("LOG_LEVEL", "info")
	format := getenv("LOG_FORMAT", "text")
	log := logger.NewFrom(level, format)
	a := &App{log: log, closer: closer.New(log)}

	natsURL := getenv("NATS_URL", "nats://localhost:4222")
	a.log.Info("connecting to nats", "url", natsURL)

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	a.nc = nc
	a.closer.AddNamed("nats", func(context.Context) error {
		return nc.Drain()
	})

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}

	streamConfig := &nats.StreamConfig{
		Name: "EVENTS",
		Subjects: []string{
			"event.>",
			"score.updated",
			"user.registered",
			"shop.purchased",
			"payment.completed",
			"inventory.item.created",
			"inventory.item.updated",
			"inventory.item.deleted",
		},
		Storage:  nats.FileStorage,
		MaxAge:   7 * 24 * time.Hour,
		MaxBytes: 1024 * 1024 * 1024,
	}

	if info, err := js.AddStream(streamConfig); err != nil {
		a.log.Warn("stream create (may exist)", "err", err)
		if info, err := js.StreamInfo("EVENTS"); err != nil {
			a.log.Error("stream info failed", "err", err)
		} else {
			a.log.Info("stream EVENTS exists", "messages", info.State.Msgs)
		}
	} else {
		a.log.Info("stream EVENTS created", "name", info.Config.Name)
	}

	if _, err := js.Subscribe("event.>", func(msg *nats.Msg) {
		a.log.Info("event", "subject", msg.Subject, "bytes", len(msg.Data))
		_ = msg.Ack()
	}, nats.Durable("nats-hub-listener")); err != nil {
		a.log.Warn("subscribe failed", "err", err)
	} else {
		a.log.Info("subscribed to event.>")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if nc.IsConnected() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok","service":"nats-hub"}`))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"degraded","service":"nats-hub"}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		if nc.IsConnected() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ready","service":"nats-hub"}`))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"not_ready","service":"nats-hub"}`))
	})
	metricsAddr := getenv("METRICS_PORT", "9097")
	a.http = &http.Server{Addr: ":" + metricsAddr, Handler: mux}
	a.closer.AddNamed("metrics http", func(ctx context.Context) error {
		return a.http.Shutdown(ctx)
	})

	return a, nil
}

func (a *App) RunUntilSignal(ctx context.Context) error {
	go func() {
		a.log.Info("nats-hub health listening", "addr", a.http.Addr)
		if err := a.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("health server error", "err", err)
		}
	}()
	a.log.Info("nats-hub running")
	if err := a.closer.WaitSignal(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

func (a *App) Close(ctx context.Context) error {
	return a.closer.CloseAll(ctx)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
