6. Inventory Seq Scan — индексы
postgres_repo.go:toPostgresFilter — фильтры по author_id, type, price, attributes @>, name ILIKE. Под 140 VU — Seq Scan на каждом. Нет индексов на attributes (GIN), на type, на price.

Фикс: добавить индексы:

sql
CREATE INDEX IF NOT EXISTS idx_inventory_attributes_gin ON inventory_items USING GIN (attributes);
CREATE INDEX IF NOT EXISTS idx_inventory_type ON inventory_items (type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_inventory_price ON inventory_items (price) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_inventory_name_trgm ON inventory_items USING GIN (name gin_trgm_ops);
Приоритет: P2. Без этого inventory/items 6.9s.


Патч Шага 5 (индексы Inventory)
Файл: /home/denismatveev/event_horizon/services/inventory/migrations/20261110000000_add_indexes.sql (новый)

sql
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
Применить:

bash
# 1. Миграция применится автоматически при старте Inventory
#    (initPostgres → migrator.Up). Просто ребилд + рестарт:
bash scripts/rebuild-services.sh inventory
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  up -d inventory

# 2. Проверить, что индексы созданы:
docker exec -it event-horizon-postgres-inventory psql -U eventhorizon -d eventhorizon_inventory \
  -c "SELECT indexname FROM pg_indexes WHERE tablename = 'inventory_items' ORDER BY indexname;"
# Должно быть 5 новых: idx_inventory_attributes_gin, idx_inventory_type,
# idx_inventory_price, idx_inventory_author, idx_inventory_name_trgm.

# 3. Проверить, что планировщик использует GIN на фильтре:
docker exec -it event-horizon-postgres-inventory psql -U eventhorizon -d eventhorizon_inventory \
  -c "EXPLAIN ANALYZE SELECT * FROM inventory_items WHERE attributes @> '{\"element\":\"woods\"}' AND deleted_at IS NULL LIMIT 20;"
# Должно быть "Bitmap Index Scan on idx_inventory_attributes_gin",
# а не "Seq Scan on inventory_items".
Затем — P=200 (90s) — проверить, что circuit inventory исчез:

bash
EH_K6_USERS_JSON="$(cat load_test_results/k6_users.json)" \
  EH_PROFILE=200 EH_DURATION=90s EH_RAMP=30s \
  env -u K6_VUS -u K6_DURATION -u K6_RAMP -u K6_SCENARIOS \
  k6 run deployments/k6/highload/10k-probe.js