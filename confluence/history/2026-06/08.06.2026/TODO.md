# TODO — 8 июня 2026
# ФОКУС: БЭКЕНД

## ✅ ВЫПОЛНЕНО (фронтенд)
- [x] Все 4 игры готовы (Блинопёк, Мемония, Flappy, Башенки)
- [x] Лидерборд поддерживает 4 игры
- [x] Профиль показывает статистику по 4 играм

---

## 🔴 КРИТИЧНО (бэкенд — сделать сегодня)

### 1. Кастомный ник в лидерборде
**Проблема:** Сейчас в лидерборде отображается `email.split('@')[0]`
**Требование:** Отображать nickname, который игрок задал в профиле

#### Фронтенд (уже частично есть):
- [ ] Добавить `nickname` в `submitScore` запрос

#### Бэкенд (Game сервис):
- [ ] Принимать `nickname` в `SubmitScoreRequest`
- [ ] Пробрасывать в NATS событие `score.updated`

#### Бэкенд (Leaderboard сервис):
- [ ] Сохранять nickname в Redis: `HSET user:profile:<user_id> nickname <nickname> email <email>`
- [ ] При запросе топа возвращать nickname из профиля

---

### 2. Суммирование очков в лидерборде
**Проблема:** Redis заменяет счёт игрока, а должен прибавлять

#### Бэкенд (Leaderboard сервис):
- [ ] В `redis_repo.go` в `UpdateScore` получать текущий счёт через `ZSCORE`
- [ ] Прибавлять новый счёт к существующему: `newScore = currentScore + newScore`
- [ ] Использовать `ZADD` с новым суммированным значением
- [ ] Проверить, что после нескольких партий счёт накапливается

---

### 3. Goose миграции для всех сервисов
**Проблема:** Миграции есть только для Auth

#### Нужно добавить:
- [ ] **Game сервис** — миграции для таблицы `highscores`
- [ ] **Billing сервис** — миграции для `user_currencies`, `transactions`
- [ ] **Leaderboard сервис** — миграции для `leaderboard_backup` (PostgreSQL)

#### Автоматизация:
- [ ] Добавить `migrate-all` команду в `Makefile`
- [ ] Автоматический запуск миграций при `make all`

---

## 🟡 ВАЖНО (бэкенд — на этой неделе)

### 4. Billing сервис — улучшения
- [ ] Добавить эндпоинт `GET /api/billing/balance/all` (если нет)
- [ ] Добавить эндпоинт `POST /api/billing/transfer` (перевод лампочек в билетики)
- [ ] Добавить транзакции для покупок (подсказки, сброс партии)

### 5. Game сервис — уровни сложности
- [ ] Принимать `level` от фронтенда (сейчас всегда 1)
- [ ] Множитель очков в зависимости от уровня
- [ ] Валидация, что уровень не превышает максимальный для игры

### 6. WebSocket лидерборд
- [ ] Проверить, что WebSocket обновляет топ после каждого рекорда
- [ ] Добавить фильтрацию по `game_id` в WebSocket

---

## 🟢 ПРИЯТНО (бэкенд — долгосрочно)

### 7. NATS кластер из 3 нод
- [ ] Добавить 3 ноды в `docker-compose.cluster.yml`
- [ ] Проверить отказоустойчивость

### 8. Prometheus + Grafana мониторинг
- [ ] Добавить метрики для NATS (порт 8222)
- [ ] Сбор метрик Go сервисов (память, goroutines, RPS)

### 9. CQRS для leaderboard
- [ ] Запись рекордов через NATS (уже есть)
- [ ] Чтение топа из Redis (уже есть)
- [ ] Разделить на команды и запросы

### 10. Envoy как API gateway
- [ ] Заменить самописный Gateway на Envoy
- [ ] Настроить маршрутизацию, rate limiting, TLS

### 11. Нагрузочное тестирование
- [ ] Go producer + consumer для NATS
- [ ] k6 + NATS extension
- [ ] Мониторинг очереди JetStream

### 12. Graceful shutdown (проверить все сервисы)
- [ ] Auth
- [ ] Game
- [ ] Billing
- [ ] Leaderboard
- [ ] Gateway

### 13. GIN_MODE=release для всех сервисов

---

## 📝 Игровой дизайн (этика) — бэкенд
- [ ] Подсказки за лампочки (API для покупки)
- [ ] Сброс партии за лампочки (API для оплаты)
- [ ] Прогресс не обнуляется между сезонами

---

## 🐛 Баги (бэкенд)
- [ ] Проверить почему иногда `user=undefined` в логах game.log
- [ ] WebSocket: новые рекорды не всегда показываются

---

## Команды для работы с бэкендом

```bash
# Полный старт
cd ~/event_horizon
make all

# Пересобрать конкретный сервис
cd services/game && go build -o game-service ./cmd/main.go
pkill -f game-service && ./game-service > /tmp/game.log 2>&1 &

# Логи
tail -f /tmp/game.log
tail -f /tmp/leaderboard.log
tail -f /tmp/gateway.log
tail -f /tmp/billing.log

# NATS подписка
nats sub "score.updated" --server localhost:4222

# PostgreSQL
docker exec -it event-horizon-postgres-game psql -U eventhorizon -d eventhorizon_game

# Redis
docker exec -it event-horizon-redis-leaderboard redis-cli
Приоритеты на сегодня (8 июня)

Приоритет	Задача	Время
1	Суммирование очков в лидерборде (Redis)	1 час
2	Кастомный ник (фронт + бэк)	2-3 часа
3	Goose миграции	2 часа
Создано: 8 июня 2026
Фокус: бэкенд, 4 игры уже готовы на фронтенде 🚀