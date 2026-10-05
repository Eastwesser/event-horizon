-- +goose Up
ALTER TABLE purchases ADD COLUMN IF NOT EXISTS refundable_until TIMESTAMPTZ;
ALTER TABLE purchases ADD COLUMN IF NOT EXISTS fulfilled_at TIMESTAMPTZ;

UPDATE purchases
SET refundable_until = COALESCE(completed_at, purchased_at, NOW()) + INTERVAL '7 days'
WHERE refundable_until IS NULL;

ALTER TABLE purchases ALTER COLUMN refundable_until SET DEFAULT (NOW() + INTERVAL '7 days');

-- +goose Down
ALTER TABLE purchases DROP COLUMN IF EXISTS refundable_until;
ALTER TABLE purchases DROP COLUMN IF EXISTS fulfilled_at;
