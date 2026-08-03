# Event Horizon — улучшения для production

## 1. Конкурентные обновления (Optimistic Locking)
- [ ] Добавить `version` поле в таблицы users, teams, tasks
- [ ] Обновлять с проверкой `WHERE id = ? AND version = ?`
- [ ] Возвращать ошибку при конфликте

## 2. Транзакции
- [ ] Оборачивать составные операции в BEGIN/COMMIT
- [ ] Использовать `defer tx.Rollback()`

## 3. Индексы
- [ ] Добавить индексы на внешние ключи
- [ ] Добавить индексы на `status`, `created_at`

## 4. Redis
- [ ] Кешировать список задач команды
- [ ] Инвалидировать кеш при обновлении

## 5. Миграции
- [ ] Добавить `+goose Down` для всех миграций
- [ ] Проверить откат

## 6. Health check
- [ ] Добавить проверку зависимостей (БД, Redis)
- [ ] `/health` → 200 если всё ок, 503 если нет
Капитан, ты составил боевой список улучшений для production. Это именно то, что превращает пет-проект в систему, которую можно показать на собеседовании и не краснеть.

📋 Разбор по пунктам
1. 🔒 Optimistic Locking — это круто
«Добавить version поле, обновлять с проверкой WHERE id = ? AND version = ?»

Это защита от lost updates (потерянных обновлений), когда два пользователя одновременно меняют один ресурс.

Как это работает в Go:

go
type User struct {
    ID      string
    Balance int
    Version int  // ← добавляем поле
}

// При обновлении
result, err := db.Exec(`
    UPDATE users 
    SET balance = $1, version = version + 1 
    WHERE id = $2 AND version = $3
`, newBalance, userID, currentVersion)

if rowsAffected == 0 {
    return errors.New("конфликт: кто-то уже обновил данные")
}
Где применить в Event Horizon:

Billing — обновление баланса (критично!)

Shop — покупка товара (чтобы дважды не списать билетики)

Profile — обновление ника (редко, но тоже)

2. 📦 Транзакции — уже должно быть
«Оборачивать составные операции в BEGIN/COMMIT, использовать defer tx.Rollback()»

Это база. Проверь, что у тебя это есть во всех критических местах.

Пример из Shop (если ещё нет):

go
func (s *shopService) PurchaseItem(ctx context.Context, userID, itemID string) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback() // если не вызовем Commit — откатится
    
    // 1. Списываем билетики
    // 2. Добавляем в инвентарь
    // 3. Записываем в историю
    
    return tx.Commit()
}
Где уже есть: В Shop есть PurchaseItem с транзакцией. Проверь, что в Billing и Game тоже.

3. 🗄️ Индексы — срочно
«Добавить индексы на внешние ключи, status, created_at»

Твой список индексов для Event Horizon:

sql
-- Auth
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);

-- Game
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_highscores_game_score ON highscores(game_id, score DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_highscores_user_game ON highscores(user_id, game_id);

-- Billing
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_currencies_user ON user_currencies(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_user ON transactions(user_id, created_at DESC);

-- Shop
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_user ON inventory(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_purchases_user ON purchases(user_id, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_items_category ON items(category, available);
4. ⚡ Redis — кеширование
«Кешировать список задач команды, инвалидировать при обновлении»

Как это сделать в Event Horizon:

go
// Кешируем лидерборд
func (s *leaderboardService) GetTop(ctx context.Context, gameID string, limit int) ([]Entry, error) {
    cacheKey := fmt.Sprintf("leaderboard:%s:%d", gameID, limit)
    
    // Пытаемся получить из Redis
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var entries []Entry
        json.Unmarshal([]byte(cached), &entries)
        return entries, nil
    }
    
    // Если нет — идём в БД
    entries, err := s.repo.GetTop(ctx, gameID, limit)
    if err != nil {
        return nil, err
    }
    
    // Сохраняем в Redis на 30 секунд
    data, _ := json.Marshal(entries)
    s.redis.Set(ctx, cacheKey, data, 30*time.Second)
    
    return entries, nil
}
5. 🦆 Миграции — добавить Down
«Добавить +goose Down для всех миграций, проверить откат»

Пример миграции:

sql
-- +goose Up
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE items;
Проверка отката:

bash
# Накатить
goose up

# Откатить последнюю
goose down

# Откатить все
goose reset
6. ❤️ Health check
«Добавить проверку зависимостей (БД, Redis), /health → 200 если всё ок, 503 если нет»

Пример для Gateway:

go
// healthcheck.go
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    status := map[string]string{
        "status": "ok",
        "services": "all healthy",
    }
    
    // Проверяем БД
    if err := db.Ping(); err != nil {
        status["status"] = "degraded"
        status["db"] = "unhealthy"
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    
    // Проверяем Redis
    if _, err := redis.Ping().Result(); err != nil {
        status["status"] = "degraded"
        status["redis"] = "unhealthy"
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    
    json.NewEncoder(w).Encode(status)
}
🎯 Приоритеты на сегодня
Приоритет	Задача	Время
🔥 1	Индексы (пункт 3)	20 мин
🔥 2	Health check (пункт 6)	20 мин
🔥 3	Транзакции (пункт 2) — проверить	15 мин
🟡 4	Redis кеширование (пункт 4)	30 мин
🟡 5	Optimistic Locking (пункт 1)	1 час
🟢 6	Миграции Down (пункт 5)	15 мин