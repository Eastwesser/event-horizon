[denismatveev@c0der event_horizon]$ find services/fulfillment -type f -name '*.go' | sort
echo "===== main ====="
cat services/fulfillment/cmd/main.go
echo "===== app ====="
for f in services/fulfillment/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/fulfillment/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
services/fulfillment/cmd/main.go
services/fulfillment/internal/app/app.go
services/fulfillment/internal/config/config.go
services/fulfillment/internal/config/config_test.go
services/fulfillment/internal/service/fulfillment.go
services/fulfillment/internal/service/fulfillment_test.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/fulfillment/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init fulfillment app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("fulfillment stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/fulfillment/internal/app/app.go ---
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
	"github.com/Eastwesser/event-horizon/services/fulfillment/internal/config"
	"github.com/Eastwesser/event-horizon/services/fulfillment/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
	cfg      *config.Config
	log      *slog.Logger
	closer   *closer.Closer
	consumer kafka.Consumer
	producer kafka.Producer
	svc      *service.FulfillmentService
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
		kcfg = kafka.Config{Brokers: splitCSV(cfg.KafkaBrokers), Enabled: true}
	}
	a.producer = kafka.NewProducer(kcfg, kafka.TopicPurchaseFulfilled, log)
	a.consumer = kafka.NewConsumer(kcfg, kafka.ConsumerGroupFulfillment, []string{kafka.TopicPurchasePaid}, log)
	a.closer.AddNamed("kafka producer", func(context.Context) error { return a.producer.Close() })
	a.closer.AddNamed("kafka consumer", func(context.Context) error { return a.consumer.Close() })

	a.svc = service.New(a.producer, log, time.Duration(cfg.AssembleDelaySec)*time.Second)

	if cfg.NATSURL != "" {
		nc, err := nats.Connect(cfg.NATSURL)
		if err != nil {
			a.log.Warn("nats connect failed; Kafka-only path", "err", err)
		} else {
			a.nc = nc
			js, err := nc.JetStream()
			if err != nil {
				a.log.Warn("jetstream failed", "err", err)
				_ = nc.Drain()
			} else {
				a.js = js
				a.svc.SetJetStream(js)
				a.closer.AddNamed("nats", func(context.Context) error { return nc.Drain() })
				a.log.Info("nats connected", "url", cfg.NATSURL)
			}
		}
	}

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

	// Primary path: NATS JetStream (thin deploy).
	if a.js != nil {
		go func() {
			a.log.Info("fulfillment consuming NATS", "subject", kafka.TopicPurchasePaid)
			_, err := a.js.Subscribe(kafka.TopicPurchasePaid, func(msg *nats.Msg) {
				err := a.svc.HandlePurchasePaid(runCtx, kafka.Message{
					Topic: msg.Subject,
					Value: msg.Data,
				})
				if err != nil {
					a.log.Error("handle purchase.paid", "err", err)
					_ = msg.Nak()
					return
				}
				_ = msg.Ack()
			}, nats.Durable("fulfillment-purchase-paid"), nats.ManualAck())
			if err != nil {
				a.log.Error("nats subscribe failed", "err", err)
			}
			<-runCtx.Done()
		}()
	}

	// Kafka only when JetStream is down. Both transports would fulfill the same order twice.
	if a.js == nil && a.cfg.KafkaBrokers != "" {
		go func() {
			a.log.Info("fulfillment consuming Kafka", "topic", kafka.TopicPurchasePaid)
			if err := a.consumer.Consume(runCtx, a.svc.HandlePurchasePaid); err != nil && runCtx.Err() == nil {
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
===== service =====
--- services/fulfillment/internal/service/fulfillment.go ---
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/Eastwesser/event-horizon/contracts/events"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
)

// FulfillmentService consumes PurchasePaid, waits, produces PurchaseFulfilled.
// Publishes to NATS JetStream when js is set; Kafka producer kept for optional heavy path.
type FulfillmentService struct {
	out   kafka.Producer
	js    nats.JetStreamContext
	log   *slog.Logger
	delay time.Duration
}

func New(out kafka.Producer, log *slog.Logger, delay time.Duration) *FulfillmentService {
	if delay <= 0 {
		delay = 10 * time.Second
	}
	return &FulfillmentService{out: out, log: log, delay: delay}
}

func (s *FulfillmentService) SetJetStream(js nats.JetStreamContext) {
	s.js = js
}

func (s *FulfillmentService) HandlePurchasePaid(ctx context.Context, msg kafka.Message) error {
	paid, err := events.UnmarshalPurchasePaid(msg.Value)
	if err != nil {
		return err
	}
	s.log.Info("purchase paid received; assembling", "purchase", paid.PurchaseUUID, "delay", s.delay)
	start := time.Now()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(s.delay):
	}
	metrics.ObserveAssembly(time.Since(start).Seconds())
	out := events.PurchaseFulfilled{
		EventUUID:    newUUID(),
		PurchaseUUID: paid.PurchaseUUID,
		UserUUID:     paid.UserUUID,
		ItemUUID:     paid.ItemUUID,
	}
	body, err := out.Marshal()
	if err != nil {
		return err
	}
	if s.js != nil {
		if _, err := s.js.Publish(kafka.TopicPurchaseFulfilled, body); err != nil {
			s.log.Error("nats publish purchase.fulfilled", "err", err)
			return err
		}
		metrics.PurchasesFulfilledTotal.Inc()
		s.log.Info("purchase fulfilled published", "purchase", out.PurchaseUUID)
		return nil
	}
	if s.out != nil {
		if err := s.out.Send(ctx, []byte(out.PurchaseUUID), body); err != nil {
			return err
		}
	}
	metrics.PurchasesFulfilledTotal.Inc()
	s.log.Info("purchase fulfilled published", "purchase", out.PurchaseUUID)
	return nil
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
--- services/fulfillment/internal/service/fulfillment_test.go ---
package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/contracts/events"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
)

type stubProducer struct {
	sent [][]byte
}

func (s *stubProducer) Send(_ context.Context, _, value []byte) error {
	s.sent = append(s.sent, append([]byte(nil), value...))
	return nil
}

func (s *stubProducer) Close() error { return nil }

func TestHandlePurchasePaid_PublishesFulfilled(t *testing.T) {
	prod := &stubProducer{}
	svc := New(prod, slog.Default(), time.Millisecond)

	body, err := events.PurchasePaid{
		EventUUID:    "e1",
		PurchaseUUID: "p1",
		UserUUID:     "u1",
		ItemUUID:     "i1",
		Price:        42,
	}.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.HandlePurchasePaid(context.Background(), kafka.Message{Value: body}); err != nil {
		t.Fatalf("HandlePurchasePaid: %v", err)
	}
	if len(prod.sent) != 1 {
		t.Fatalf("sent=%d want 1", len(prod.sent))
	}
	got, err := events.UnmarshalPurchaseFulfilled(prod.sent[0])
	if err != nil {
		t.Fatal(err)
	}
	if got.PurchaseUUID != "p1" || got.UserUUID != "u1" || got.ItemUUID != "i1" {
		t.Fatalf("fulfilled=%+v", got)
	}
}

func TestHandlePurchasePaid_BadPayload(t *testing.T) {
	svc := New(&stubProducer{}, slog.Default(), time.Millisecond)
	err := svc.HandlePurchasePaid(context.Background(), kafka.Message{Value: []byte("not-json")})
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}
[denismatveev@c0der event_horizon]$ 
