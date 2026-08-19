-- +goose Up
CREATE TABLE IF NOT EXISTS outbox (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type    TEXT NOT NULL,
    payload       JSONB NOT NULL,
    processed     BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_shop_outbox_processed ON outbox(processed, created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_shop_outbox_processed;
DROP TABLE IF EXISTS outbox;
