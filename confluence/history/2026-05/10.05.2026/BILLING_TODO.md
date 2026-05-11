# Billing Service — план

## Задачи

- [ ] gRPC контракт (billing.proto)
  - GetBalance(user_id, currency_type)
  - AddLamps(user_id, amount)
  - AddTickets(user_id, amount)
  - SpendLamps(user_id, amount)

- [ ] PostgreSQL схема (user_currencies, user_daily_revenue)
- [ ] Redis кеш для балансов
- [ ] NATS subscriber для событий (начисление за рекорды)
- [ ] Graceful shutdown

## Интеграция с Game

При рекорде Game публикует:
```json
{
  "user_id": "...",
  "game_id": "hexagon",
  "score": 100,
  "is_record": true,
  "lamps_earned": 10,
  "tickets_earned": 5
}
```

Billing слушает событие и начисляет валюту.

# PHASES

Фаза 1: Billing сервис
  gRPC контракт (billing.proto)
  GetBalance(user_id, currency_type)
  AddLamps(user_id, amount)
  AddTickets(user_id, amount)
  SpendLamps(user_id, amount)
  PostgreSQL схема
  user_currencies (user_id, currency_type, balance)
  user_daily_revenue (пассивный доход с этажей)
  Redis кеш для балансов (чтобы не дёргать БД при каждом запросе)
  NATS subscriber — слушает score.updated и начисляет награды
  Graceful shutdown

Фаза 2: Интеграция с Game
  Game уже публикует lamps_earned и tickets_earned в событии score.updated. Billing должен это слушать.

Фаза 3: Обновить Gateway (если нужно)
  Возможно, добавить эндпоинты:
  GET /api/billing/balance?user_id=...&currency=lamps
  POST /api/billing/spend (покупка подсказок, сброса партии и т.д.)
  