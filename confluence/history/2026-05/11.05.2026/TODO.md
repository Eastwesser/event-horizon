# TODO — 12 мая 2026

## Нагрузочное тестирование

### Этапы
- [ ] 10 одновременных пользователей
- [ ] 100 пользователей
- [ ] 200 пользователей
- [ ] 400 пользователей
- [ ] 600 пользователей
- [ ] 800 пользователей
- [ ] 1000 пользователей

### Высокая нагрузка
- [ ] 2000 пользователей
- [ ] 4000 пользователей
- [ ] 6000 пользователей
- [ ] 8000 пользователей
- [ ] 10000 пользователей

### Экстремальная нагрузка (цель)
- [ ] 20000 пользователей
- [ ] 40000 пользователей
- [ ] 60000 пользователей
- [ ] 80000 пользователей
- [ ] 100000 пользователей

### Что измеряем
- [ ] RPS на Gateway
- [ ] Задержка NATS (публикация → получение)
- [ ] Leaderboard latency (Redis Sorted Set)
- [ ] Billing latency (PostgreSQL)
- [ ] Потребление памяти Go сервисами
- [ ] CPU usage
- [ ] Количество горутин

### Инструменты
- [ ] k6 + NATS extension
- [ ] wrk / bombardier для HTTP
- [ ] pprof для Go сервисов

## Оставшиеся сервисы
- [ ] Notification (push/email)
- [ ] Analytics (ClickHouse/PostgreSQL)
- [ ] Payment (Boosty/Stripe)
- [ ] Social (друзья)

## Фронтенд (потом)
- [ ] React + TypeScript
- [ ] Drag-n-drop гексагоны
- [ ] Подключение к WebSocket
- [ ] Интеграция с Billing API

## Команды для быстрого старта после перезагрузки

```bash
cd ~/event_horizon
make all          # запустить всё
make stop         # остановить всё
make restart      # перезапустить всё
make ps           # статус Docker контейнеров

# Проверить логи
tail -f /tmp/auth.log
tail -f /tmp/game.log
tail -f /tmp/billing.log
tail -f /tmp/leaderboard.log
tail -f /tmp/gateway.log