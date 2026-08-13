package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/Eastwesser/event-horizon/services/history/internal/service"
)

// NATS subjects mirrored into history trail.
var subjects = []string{
	"payment.completed",
	"author.upserted",
	"shop.purchased",
	"score.updated",
	"user.registered",
}

type IngestWorker struct {
	js  nats.JetStreamContext
	svc *service.HistoryService
}

func NewIngestWorker(js nats.JetStreamContext, svc *service.HistoryService) *IngestWorker {
	return &IngestWorker{js: js, svc: svc}
}

func (w *IngestWorker) Start(ctx context.Context) {
	if w.js == nil {
		return
	}
	for _, subj := range subjects {
		s := subj
		_, err := w.js.Subscribe(s, func(msg *nats.Msg) {
			var raw map[string]any
			userID := ""
			if err := json.Unmarshal(msg.Data, &raw); err == nil {
				if v, ok := raw["user_id"].(string); ok {
					userID = v
				}
			}
			if _, err := w.svc.RecordEvent(context.Background(), userID, s, string(msg.Data)); err != nil {
				log.Printf("history ingest %s: %v", s, err)
				_ = msg.Nak()
				return
			}
			_ = msg.Ack()
		}, nats.Durable("history-"+s), nats.ManualAck())
		if err != nil {
			log.Printf("history subscribe %s: %v", s, err)
		}
	}
	<-ctx.Done()
}

type RetentionWorker struct {
	svc *service.HistoryService
}

func NewRetentionWorker(svc *service.HistoryService) *RetentionWorker {
	return &RetentionWorker{svc: svc}
}

func (w *RetentionWorker) Start(ctx context.Context) {
	t := time.NewTicker(1 * time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n, err := w.svc.PurgeExpired(ctx)
			if err != nil {
				log.Printf("history retention: %v", err)
			} else if n > 0 {
				log.Printf("history retention: deleted %d rows", n)
			}
		}
	}
}
