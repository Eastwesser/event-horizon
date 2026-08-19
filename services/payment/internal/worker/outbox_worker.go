package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

type OutboxWorker struct {
	db *pgxpool.Pool
	js nats.JetStreamContext
}

func NewOutboxWorker(db *pgxpool.Pool, js nats.JetStreamContext) *OutboxWorker {
	return &OutboxWorker{db: db, js: js}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	rows, err := w.db.Query(ctx, `
		SELECT id, event_type, payload FROM outbox
		WHERE processed = false
		ORDER BY created_at ASC
		LIMIT 50`)
	if err != nil {
		slog.Error("payment outbox: query", "err", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, eventType string
		var payload []byte
		if err := rows.Scan(&id, &eventType, &payload); err != nil {
			slog.Error("payment outbox: scan", "err", err)
			continue
		}
		if w.js != nil {
			if _, err := w.js.Publish(eventType, payload); err != nil {
				slog.Error("payment outbox: publish", "event_type", eventType, "err", err)
				continue
			}
		}
		if _, err := w.db.Exec(ctx, `
			UPDATE outbox SET processed = true, processed_at = NOW() WHERE id = $1`, id); err != nil {
			slog.Error("payment outbox: mark processed", "err", err)
			continue
		}
		slog.Info("payment outbox: published", "event_type", eventType, "id", id)
	}
}
