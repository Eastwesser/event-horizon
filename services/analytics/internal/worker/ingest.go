package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/service"
)

var subjects = []string{
	"payment.completed",
	"author.upserted",
	"shop.purchased",
	"score.updated",
	"user.registered",
}

type IngestWorker struct {
	js  nats.JetStreamContext
	svc *service.AnalyticsService
}

func NewIngestWorker(js nats.JetStreamContext, svc *service.AnalyticsService) *IngestWorker {
	return &IngestWorker{js: js, svc: svc}
}

func (w *IngestWorker) Start(ctx context.Context) {
	if w.js == nil {
		return
	}
	for _, subj := range subjects {
		s := subj
		_, err := w.js.Subscribe(s, func(msg *nats.Msg) {
			userID := ""
			var raw map[string]any
			if err := json.Unmarshal(msg.Data, &raw); err == nil {
				if v, ok := raw["user_id"].(string); ok {
					userID = v
				}
			}
			if err := w.svc.RecordEvent(context.Background(), userID, s, string(msg.Data)); err != nil {
				log.Printf("analytics ingest %s: %v", s, err)
				_ = msg.Nak()
				return
			}
			_ = msg.Ack()
		}, nats.Durable("analytics-"+s), nats.ManualAck())
		if err != nil {
			log.Printf("analytics subscribe %s: %v", s, err)
		}
	}
	<-ctx.Done()
}
