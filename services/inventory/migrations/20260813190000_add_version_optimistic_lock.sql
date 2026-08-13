-- +goose Up
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE inventory_items DROP COLUMN IF EXISTS version;
