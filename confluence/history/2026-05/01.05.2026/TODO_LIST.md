## План на 30.04.2026 - ВЫПОЛНЕНО ✅

1. ✅ Добавить PostgreSQL/Redis для Game, Billing, Leaderboard в docker-compose
2. ✅ Gateway: базовая структура, HTTP → Auth gRPC прокси
3. ✅ NATS тест: отправить событие из Gateway, поймать в отдельном воркере
4. ⏳ Leaderboard: Redis Sorted Set + подписка на NATS (в процессе)

## План на 01.05.2026

### Утро
- [ ] Leaderboard сервис (скелет + подписка на `score.updated`)
- [ ] Redis Sorted Set для топа-10

### День
- [ ] Game сервис (заглушка с `SubmitScore`)
- [ ] Публикация рекордов в NATS (`score.updated`)

### Вечер
- [ ] Сквозной тест: Game → NATS → Leaderboard
- [ ] WebSocket в Gateway для реального времени

## Техдолг (актуальный)

1. **Graceful shutdown** — все сервисы (SIGTERM)
2. **NATS кластер** — 3 ноды (нагрузочное тестирование)
3. **Мониторинг** — Prometheus + Grafana для NATS (порт 8222)
4. **Порты** — задокументировать диапазон 5460-5463
5. **Load balancer** — свой + nginx (после MVP)
6. **CQRS** — улучшение для собеседования (после MVP)
7. **Envoy** — замена самописному gateway (опционально)
8. **GIN_MODE=release** — для production-замеров

## NEXT TIME:

Следующая цель: Leaderboard

Теперь, когда документация готова и NATS работает, идём к leaderboard.

Что нужно сделать:

Структура leaderboard (уже есть скелет)
Подписка на NATS (слушаем score.updated)
Redis Sorted Set (хранение топа-10)
gRPC метод GetTop10