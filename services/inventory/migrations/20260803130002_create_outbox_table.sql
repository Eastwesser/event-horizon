-- +goose Up
-- The outbox table is required by postgres_repo.go (CreateOutboxEvent, CreateItemWithOutbox)
-- and internal/worker/outbox_worker.go, but was never actually created — only an
-- index-only migration existed on top of it, so a fresh DB deploy would fail.
CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS outbox;
