-- +goose Up
ALTER TABLE purchases ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_purchases_user_item_status
    ON purchases (user_id, item_id, status);

-- +goose Down
DROP INDEX IF EXISTS idx_purchases_user_item_status;
ALTER TABLE purchases DROP COLUMN IF EXISTS refunded_at;
