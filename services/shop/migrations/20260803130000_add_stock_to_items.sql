-- +goose Up
ALTER TABLE items ADD COLUMN IF NOT EXISTS stock INT DEFAULT 999;

-- Обновляем существующие товары
UPDATE items SET stock = 999 WHERE stock IS NULL;

-- +goose Down
ALTER TABLE items DROP COLUMN IF EXISTS stock;