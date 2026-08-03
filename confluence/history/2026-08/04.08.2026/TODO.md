

🎯 ЧТО ДЕЛАЕМ СЕГОДНЯ (по приоритету)
Приоритет	Задача	Время	Статус
🔥 1	Индексы в БД	30 мин	⏳
🔥 2	Rate Limiter в Gateway	15 мин	⏳
🔥 3	K6 тесты (p95 < 500ms)	1 час	⏳
🟡 4	README для каждого сервиса	1 час	⏳
🟡 5	Антифрод (1 фенечка/месяц)	1 час	⏳
🟢 6	OpenAPI / Swagger	30 мин	⏳
🚀 ПЕРВЫЙ ШАГ — ИНДЕКСЫ В БД
Вот готовый SQL для добавления недостающих индексов:

sql
-- Shop
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_user_item ON inventory(user_id, item_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_purchases_user ON purchases(user_id, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_items_category ON items(category, available);

-- Billing
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_currencies_user_currency ON user_currencies(user_id, currency_type);

-- Game
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_highscores_game_score ON highscores(game_id, score DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_highscores_user_game ON highscores(user_id, game_id);
Выполни их в каждой БД:

bash
# Shop
docker exec -it event-horizon-postgres-shop psql -U eventhorizon -d eventhorizon_shop -c "..."
