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
Billing слушает событие и начисляет валюту.