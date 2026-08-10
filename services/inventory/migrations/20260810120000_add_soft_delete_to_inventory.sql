-- +goose Up
-- Добавляем поле deleted_at для мягкого удаления товаров
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP DEFAULT NULL;

-- Создаем индекс для ускорения запросов с фильтрацией по deleted_at
CREATE INDEX IF NOT EXISTS idx_inventory_deleted_at ON inventory_items(deleted_at);

-- Комментарий для документации
COMMENT ON COLUMN inventory_items.deleted_at IS 'Время мягкого удаления товара. NULL — товар активен.';

-- +goose Down
-- Удаляем индекс и поле при откате
DROP INDEX IF EXISTS idx_inventory_deleted_at;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS deleted_at;