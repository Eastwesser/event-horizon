-- +goose Up
ALTER TABLE user_currencies ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE user_currencies DROP COLUMN IF EXISTS version;
