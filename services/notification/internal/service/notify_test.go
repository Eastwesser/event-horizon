package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/Eastwesser/event-horizon/contracts/events"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
)

func TestHandle_IgnoredTopic(t *testing.T) {
	n := NewNotifier(slog.Default(), "", "")
	if err := n.Handle(context.Background(), kafka.Message{Topic: "unrelated"}); err != nil {
		t.Fatalf("ignored topic: %v", err)
	}
}

func TestHandle_PurchasePaid_NoTelegram(t *testing.T) {
	n := NewNotifier(slog.Default(), "", "")
	body, err := events.PurchasePaid{
		EventUUID:    "e1",
		PurchaseUUID: "p1",
		UserUUID:     "u1",
		ItemUUID:     "i1",
		Price:        10,
	}.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := n.Handle(context.Background(), kafka.Message{
		Topic: kafka.TopicPurchasePaid,
		Value: body,
	}); err != nil {
		t.Fatalf("Handle: %v", err)
	}
}

func TestHandle_PurchasePaid_BadJSON(t *testing.T) {
	n := NewNotifier(slog.Default(), "", "")
	err := n.Handle(context.Background(), kafka.Message{
		Topic: kafka.TopicPurchasePaid,
		Value: []byte("not-json"),
	})
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}
