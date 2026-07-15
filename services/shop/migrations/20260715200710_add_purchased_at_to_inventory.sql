-- +goose Up
ALTER TABLE inventory ADD COLUMN IF NOT EXISTS purchased_at TIMESTAMP DEFAULT NOW();

-- +goose Down
ALTER TABLE inventory DROP COLUMN IF EXISTS purchased_at;