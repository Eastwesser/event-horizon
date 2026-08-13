-- +goose Up
-- Outbox pattern for billing, mirroring inventory (the project's reference
-- implementation). Without this, balance.updated events could only be
-- published in-process with no guaranteed delivery if NATS was down.
CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_billing_outbox_processed ON outbox(processed, created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_billing_outbox_processed;
DROP TABLE IF EXISTS outbox;
