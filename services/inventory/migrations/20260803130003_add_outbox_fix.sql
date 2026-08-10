-- +goose Up
CREATE INDEX IF NOT EXISTS idx_outbox_processed ON outbox(processed, created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_outbox_processed;
