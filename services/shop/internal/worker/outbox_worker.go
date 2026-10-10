package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// CompensationEvent is an outbox row the shop worker settles by calling billing,
// not by publishing to NATS.
const CompensationEvent = "billing.refund"

// CompensateFunc retries a ticket refund. A duplicate reference_id is success.
type CompensateFunc func(ctx context.Context, payload []byte) error

type OutboxWorker struct {
	db         *sql.DB
	js         nats.JetStreamContext
	compensate CompensateFunc
}

func NewOutboxWorker(db *sql.DB, js nats.JetStreamContext, compensate CompensateFunc) *OutboxWorker {
	return &OutboxWorker{db: db, js: js, compensate: compensate}
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

func (w *OutboxWorker) processBatch(ctx context.Context) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("shop outbox: begin", "err", err)
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
		slog.Error("shop outbox: query", "err", err)
		return
	}

	var batch []outboxRow
	for rows.Next() {
		var row outboxRow
		if err := rows.Scan(&row.id, &row.eventType, &row.payload); err != nil {
			slog.Error("shop outbox: scan", "err", err)
			continue
		}
		batch = append(batch, row)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		slog.Error("shop outbox: rows", "err", err)
		return
	}

	for _, row := range batch {
		if err := w.deliver(ctx, row); err != nil {
			slog.Error("shop outbox: deliver", "event_type", row.eventType, "id", row.id, "err", err)
			continue
		}
		if _, err := tx.ExecContext(ctx, `
            UPDATE outbox SET processed = true, processed_at = NOW()
            WHERE id = $1
        `, row.id); err != nil {
			slog.Error("shop outbox: mark processed", "id", row.id, "err", err)
			return
		}
		slog.Info("shop outbox: published", "event_type", row.eventType, "id", row.id)
	}
	if err := tx.Commit(); err != nil {
		slog.Error("shop outbox: commit", "err", err)
	}
}

func (w *OutboxWorker) deliver(ctx context.Context, row outboxRow) error {
	if row.eventType == CompensationEvent {
		if w.compensate == nil {
			return errNoCompensator
		}
		err := w.compensate(ctx, row.payload)
		if err != nil && !isDuplicateRef(err) {
			return err
		}
		return nil
	}
	if w.js == nil {
		return errNoJetStream
	}
	_, err := w.js.Publish(row.eventType, row.payload)
	return err
}

var (
	errNoCompensator = errString("shop compensation handler is not wired")
	errNoJetStream   = errString("jetstream is not connected")
)

type errString string

func (e errString) Error() string { return string(e) }

func isDuplicateRef(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate") || strings.Contains(s, "23505") || strings.Contains(s, "unique constraint")
}
