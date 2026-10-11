package worker

import (
	"context"
	"database/sql"
	"encoding/json"
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
			w.resumeSagas(ctx)
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

type sagaResume struct {
	id         string
	userID     string
	itemID     string
	purchaseID string
	refundRef  string
	price      int
}

// resumeSagas finishes a purchase that died after billing spent tickets
// and before the shop row was committed. The row lock is released before the billing call.
func (w *OutboxWorker) resumeSagas(ctx context.Context) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("shop saga: begin", "err", err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id::text, user_id, item_id, purchase_id::text, refund_ref, price
		FROM purchase_sagas
		WHERE state IN ('spent', 'compensating')
		  AND updated_at < NOW() - INTERVAL '30 seconds'
		ORDER BY updated_at
		LIMIT 10
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		slog.Error("shop saga: query", "err", err)
		return
	}

	var refunds []sagaResume
	for rows.Next() {
		var s sagaResume
		if err := rows.Scan(&s.id, &s.userID, &s.itemID, &s.purchaseID, &s.refundRef, &s.price); err != nil {
			slog.Error("shop saga: scan", "err", err)
			continue
		}
		var stocked bool
		if err := tx.QueryRowContext(ctx, `
			SELECT EXISTS(SELECT 1 FROM purchases WHERE id = $1::uuid)
		`, s.purchaseID).Scan(&stocked); err != nil {
			slog.Error("shop saga: purchase lookup", "id", s.id, "err", err)
			continue
		}
		if stocked {
			if _, err := tx.ExecContext(ctx, `
				UPDATE purchase_sagas SET state = 'stocked', updated_at = NOW() WHERE id = $1::uuid
			`, s.id); err != nil {
				slog.Error("shop saga: mark stocked", "id", s.id, "err", err)
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE purchase_sagas SET state = 'compensating', updated_at = NOW() WHERE id = $1::uuid
		`, s.id); err != nil {
			slog.Error("shop saga: mark compensating", "id", s.id, "err", err)
			continue
		}
		refunds = append(refunds, s)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		slog.Error("shop saga: rows", "err", err)
		return
	}
	if err := tx.Commit(); err != nil {
		slog.Error("shop saga: commit", "err", err)
		return
	}

	for _, s := range refunds {
		if w.compensate == nil {
			slog.Error("shop saga: compensator is not wired", "id", s.id)
			continue
		}
		body, _ := json.Marshal(map[string]any{
			"user_id":      s.userID,
			"item_id":      s.itemID,
			"amount":       s.price,
			"reference_id": s.refundRef,
			"error":        "saga resume",
		})
		if err := w.compensate(ctx, body); err != nil && !isDuplicateRef(err) {
			slog.Error("shop saga: refund", "id", s.id, "err", err)
			continue
		}
		if _, err := w.db.ExecContext(ctx, `
			UPDATE purchase_claims SET cycle = cycle + 1
			WHERE user_id = $1::uuid AND item_id = $2::uuid
		`, s.userID, s.itemID); err != nil {
			slog.Error("shop saga: bump cycle", "id", s.id, "err", err)
			continue
		}
		if _, err := w.db.ExecContext(ctx, `
			UPDATE purchase_sagas SET state = 'compensated', updated_at = NOW() WHERE id = $1::uuid
		`, s.id); err != nil {
			slog.Error("shop saga: mark compensated", "id", s.id, "err", err)
		}
	}
}
