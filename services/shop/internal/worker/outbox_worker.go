package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

type OutboxWorker struct {
	db *sql.DB
	js nats.JetStreamContext
}

func NewOutboxWorker(db *sql.DB, js nats.JetStreamContext) *OutboxWorker {
	return &OutboxWorker{db: db, js: js}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	rows, err := w.db.QueryContext(ctx, `
        SELECT id, event_type, payload FROM outbox
        WHERE processed = false
        ORDER BY created_at
        LIMIT 100
    `)
	if err != nil {
		slog.Error("shop outbox: query", "err", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, eventType string
		var payload []byte
		if err := rows.Scan(&id, &eventType, &payload); err != nil {
			slog.Error("shop outbox: scan", "err", err)
			continue
		}
		if w.js != nil {
			if _, err := w.js.Publish(eventType, payload); err != nil {
				slog.Error("shop outbox: publish", "event_type", eventType, "err", err)
				continue
			}
		}
		if _, err := w.db.ExecContext(ctx, `
            UPDATE outbox SET processed = true, processed_at = NOW()
            WHERE id = $1
        `, id); err != nil {
			slog.Error("shop outbox: mark processed", "id", id, "err", err)
			continue
		}
		slog.Info("shop outbox: published", "event_type", eventType, "id", id)
	}
}
