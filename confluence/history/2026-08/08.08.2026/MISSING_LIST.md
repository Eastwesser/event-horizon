Делаем то, чего не хватает:

🔍 Детальный анализ

1. ❌ Optimistic Locking — 0
Где должно быть: Billing (баланс), Shop (инвентарь), Game (рекорды).

Что нужно сделать:

go
// Добавить поле Version в структуры
type UserCurrency struct {
    UserID   string
    Currency string
    Balance  int
    Version  int  // ← добавить
}

// При обновлении проверять версию
UPDATE user_currencies 
SET balance = $1, version = version + 1 
WHERE user_id = $2 AND currency = $3 AND version = $4

2. 🟡 Транзакции — только в Shop
Где есть: services/shop/internal/repository/postgres_repo.go

Где НЕТ:

billing — при списании/начислении билетиков

game — при сохранении рекорда и начислении наград

Что нужно сделать:

go
// В Billing при AddBalance
tx, err := r.db.BeginTx(ctx, nil)
defer tx.Rollback()
// обновляем баланс
// записываем транзакцию
tx.Commit()

3. ✅ Индексы — 7 штук
Хорошо: Индексы есть в Auth, Billing, Game, Leaderboard.

Чего не хватает:

sql
-- Shop: индексы для инвентаря и покупок
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_user_item ON inventory(user_id, item_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_purchases_user ON purchases(user_id, created_at DESC);

-- Billing: индекс для быстрых запросов баланса
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_currencies_user_currency ON user_currencies(user_id, currency_type);

-- Game: составной индекс для лидерборда
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_highscores_game_score ON highscores(game_id, score DESC);

4. ✅ Redis кеширование — 2 вызова SetBalance
Что есть: Billing кеширует баланс в Redis на 5 минут.

Что можно улучшить:

Добавить кеширование для Shop (список товаров)

Добавить кеширование для Leaderboard (топ-10)

5. ✅ Миграции Down — 9 штук
Отлично: У всех сервисов есть секция -- +goose Down.

Проверь: Что можно откатить без потери данных (для development).

6. ✅ Health check — 1 (Gateway)
Что есть: /health в Gateway.

Что нужно:

go
// Добавить проверку зависимостей
func HealthCheck(c *gin.Context) {
    status := gin.H{
        "status": "ok",
        "services": gin.H{
            "db": checkDB(),
            "redis": checkRedis(),
            "nats": checkNATS(),
        },
    }
    c.JSON(200, status)
}

🎯 Приоритеты на сегодня

Приоритет	Задача	Время

🔥 1	Optimistic Locking для Billing	1 час
🔥 2	Добавить индексы в Shop и Billing	20 мин
🔥 3	Транзакции в Billing (если нет)	30 мин
🟡 4	Health check с зависимостями	30 мин
🟢 5	Кеширование лидерборда в Redis	30 мин

