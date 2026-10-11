package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

// OutboxWorker publishes unpublished game outbox rows to NATS JetStream.
// Rows are locked with SKIP LOCKED so two processes cannot publish the same score twice.
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

type outboxRow struct {
	id        string
	eventType string
	payload   []byte
}

// processBatch locks a page of unpublished rows. A second worker skips those locks
// and cannot publish the same score.updated twice.
func (w *OutboxWorker) processBatch(ctx context.Context) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("game outbox: begin", "err", err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
        SELECT id, event_type, payload FROM outbox
        WHERE processed = false
        ORDER BY created_at
        LIMIT 20
        FOR UPDATE SKIP LOCKED
    `)
	if err != nil {
		slog.Error("game outbox: query", "err", err)
		return
	}
	var batch []outboxRow
	for rows.Next() {
		var row outboxRow
		if err := rows.Scan(&row.id, &row.eventType, &row.payload); err != nil {
			slog.Error("game outbox: scan", "err", err)
			continue
		}
		batch = append(batch, row)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		slog.Error("game outbox: rows", "err", err)
		return
	}

	for _, row := range batch {
		if w.js == nil {
			slog.Error("game outbox: jetstream is not connected", "id", row.id)
			continue
		}
		if _, err := w.js.Publish(row.eventType, row.payload); err != nil {
			slog.Error("game outbox: publish", "event_type", row.eventType, "id", row.id, "err", err)
			continue
		}
		if _, err := tx.ExecContext(ctx, `
            UPDATE outbox SET processed = true, processed_at = NOW() WHERE id = $1
        `, row.id); err != nil {
			slog.Error("game outbox: mark processed", "id", row.id, "err", err)
			return
		}
		slog.Info("game outbox: published", "event_type", row.eventType, "id", row.id)
	}
	if err := tx.Commit(); err != nil {
		slog.Error("game outbox: commit", "err", err)
	}
}
