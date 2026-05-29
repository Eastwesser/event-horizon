# SUCCESS — 28-29 мая 2026

## Победа! Лидерборд работает!

### Проблема
Фронтенд отправлял `score`, но бэкенд его не видел — NATS получал `score: 0`.

### Причина
В `gateway/cmd/main.go` в обработчике `/api/game/submit` отсутствовало поле `Score` в структуре запроса.

### Решение
Добавлено поле `Score` в структуру и передача в gRPC вызов.

### Результат
📥 Gateway received: ... "score":38 ...
📥 Game received: score=38
📡 Published score.updated: score=38
📡 Leaderboard received: score=38

text

**Лидерборд показывает 38 очков!** ✅

## Что сделано за эти дни

1. ✅ Починен `userId` в localStorage (логин сохраняет)
2. ✅ Добавлено поле `Score` во все слои (proto → gateway → handler → service)
3. ✅ Лидерборд получает и отображает реальные очки
4. ✅ WebSocket подключается и обновляет топ (временно)
5. ✅ Игра полностью функциональна

## Технические детали

### Изменённые файлы
- `services/game/proto/game.proto` — добавлено поле `score`
- `services/game/internal/handler/grpc_handler.go` — передача `Score`
- `services/game/internal/service/game_service.go` — приём `Score`
- `services/gateway/cmd/main.go` — добавлено поле `Score` в структуру
- `frontend/src/store/gameStore.ts` — отправка `score`

### Архитектура потока очков
Фронтенд → Gateway (HTTP/JSON) → Game (gRPC) → NATS → Leaderboard → Redis
↓
score передаётся напрямую

text

## Команда для проверки

```bash
# Слушать NATS
nats sub "score.updated" --server localhost:4222

# Логи Game
tail -f /tmp/game.log

# Логи Leaderboard
tail -f /tmp/leaderboard.log
Итог

«Никуся — Блинопёк» v0.8 — полностью играбельный прототип!

✅ Авторизация (JWT)
✅ Гексагональное поле (drag-n-drop)
✅ Поднос с 3 стопками
✅ Сложение стопок и очки
✅ Уровни (1-5+)
✅ Leaderboard с реальными очками
✅ WebSocket обновления
Следующая цель: WebSocket персистентность, выбор уровня, блинопекарня.