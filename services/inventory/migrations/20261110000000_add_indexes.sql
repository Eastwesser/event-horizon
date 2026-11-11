-- +goose Up
-- Шаг 5 (10.11.2026): GIN/btree-индексы для inventory/items.
-- toPostgresFilter (internal/repository/postgres_repo.go) фильтрует по
-- author_id, type, price, attributes @>, name ILIKE. Без индексов — Seq Scan
-- на каждом browse-запросе. Под 140 VU (P=200) это открывало CB Inventory.
-- См. confluence/history/2026-11/10.11.2026/GIN_INDEXATION/GIT_INDEXES_UPDATE.md.

-- JSONB attributes: @> (containment) для element/class/rarity фильтров.
CREATE INDEX IF NOT EXISTS idx_inventory_attributes_gin
    ON inventory_items USING GIN (attributes)
    WHERE deleted_at IS NULL;

-- type (creature/artifact/...) — частый фильтр.
CREATE INDEX IF NOT EXISTS idx_inventory_type
    ON inventory_items (type)
    WHERE deleted_at IS NULL;

-- price — фильтр по цене (shop catalog).
CREATE INDEX IF NOT EXISTS idx_inventory_price
    ON inventory_items (price)
    WHERE deleted_at IS NULL;

-- author_id — каталог автора (GET /authors/me/sales, /inventory/items?author_id=).
CREATE INDEX IF NOT EXISTS idx_inventory_author
    ON inventory_items (author_id)
    WHERE deleted_at IS NULL;

-- name ILIKE '%query%' — pg_trgm для LIKE-поиска (MCP postgres_query, admin).
-- pg_trgm — стандартное расширение Postgres, есть в postgres:16-alpine.
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_inventory_name_trgm
    ON inventory_items USING GIN (name gin_trgm_ops)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_inventory_name_trgm;
DROP INDEX IF EXISTS idx_inventory_author;
DROP INDEX IF EXISTS idx_inventory_price;
DROP INDEX IF EXISTS idx_inventory_type;
DROP INDEX IF EXISTS idx_inventory_attributes_gin;