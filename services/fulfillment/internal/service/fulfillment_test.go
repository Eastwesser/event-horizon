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
