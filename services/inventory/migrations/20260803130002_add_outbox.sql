-- +goose Up
CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    processed BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMP
);

CREATE INDEX idx_outbox_processed ON outbox(processed, created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_outbox_processed;
DROP TABLE IF EXISTS outbox;