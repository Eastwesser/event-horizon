// services/billing/internal/worker/outbox_worker.go
package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

// OutboxWorker polls the billing outbox table and republishes any events that
// haven't made it to NATS yet, guaranteeing at-least-once delivery for
// balance.updated even if NATS was briefly unavailable when the balance changed.
type OutboxWorker struct {
	db *pgxpool.Pool
	js nats.JetStreamContext
}

func NewOutboxWorker(db *pgxpool.Pool, js nats.JetStreamContext) *OutboxWorker {
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
	rows, err := w.db.Query(ctx, `
        SELECT id, event_type, payload FROM outbox
        WHERE processed = false
        ORDER BY created_at
        LIMIT 100
    `)
	if err != nil {
		slog.Error("billing outbox: query", "err", err)
		return
	}
	defer rows.Close()

	type pending struct {
		id, eventType string
		payload       []byte
	}
	var batch []pending

	for rows.Next() {
		var p pending
		if err := rows.Scan(&p.id, &p.eventType, &p.payload); err != nil {
			slog.Error("billing outbox: scan", "err", err)
			continue
		}
		batch = append(batch, p)
	}
	rows.Close()

	for _, p := range batch {
		if _, err := w.js.Publish(p.eventType, p.payload); err != nil {
			slog.Error("billing outbox: publish", "event_type", p.eventType, "err", err)
			continue
		}

		if _, err := w.db.Exec(ctx, `
            UPDATE outbox SET processed = true, processed_at = NOW()
            WHERE id = $1
        `, p.id); err != nil {
			slog.Error("billing outbox: mark processed", "err", err)
		} else {
			slog.Info("billing outbox: published", "event_type", p.eventType, "id", p.id)
		}
	}
}
