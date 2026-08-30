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
	}
	if s.out != nil {
		if err := s.out.Send(ctx, []byte(out.PurchaseUUID), body); err != nil {
			return err
		}
	}
	s.log.Info("purchase fulfilled published", "purchase", out.PurchaseUUID)
	return nil
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
