# WebSocket в Gateway — ГОТОВО! 🕸️

**Дата:** 9 мая 2026

## Что сделано

- [x] WebSocket Hub (управление клиентами)
- [x] Эндпоинт `/ws/leaderboard`
- [x] NATS subscriber → WebSocket broadcast
- [x] Graceful shutdown для WebSocket соединений

## Тест

```bash
# Терминал 1: Подключиться через WebSocket
npx wscat -c ws://localhost:8080/ws/leaderboard

# Терминал 2: Отправить рекорд
grpcurl -plaintext -d '{
  "user_id": "ws-test",
  "game_id": "hexagon",
  "level": 3,
  "seed": "ws_test_seed",
  "moves": []
}' localhost:50052 game.GameService/SubmitScore

# В терминале 1 появится обновление топа
```

# Архитектура

```text
curl → Gateway → Game → NATS → Leaderboard
                                    ↓
                                WebSocket ← Gateway ← NATS subscriber
                                    ↓
                                Клиент (React)
```

Real-time leaderboard работает!

WEBSOCKET РАБОТАЕТ!

text
✅ WebSocket client connected. Total: 1
✅ Broadcasted score update to WebSocket clients
✅ wscat получил сообщение: {"user_id":"restart-test","score":8,...}
Что мы видим в логах

Gateway:

text
🟢 WebSocket client connected. Total: 1
📡 Broadcasted score update to WebSocket clients
wscat (клиент):

text
< {"game_id":"hexagon","is_record":true,"level":3,"score":8,...,"user_id":"restart-test"}
Leaderboard:

text
📡 Received score update via NATS: game=hexagon user=restart-test score=8
✅ Score updated for restart-test, new rank: 2
Полный рабочий пайплайн

text
curl/grpcurl → Game (50052) → валидация → NATS (4222) → Leaderboard (50054) → Redis (6382)
                                    ↓
                              Gateway subscriber
                                    ↓
                              WebSocket broadcast
                                    ↓
                              wscat (клиент) ← реальное обновление!
Что у нас есть сейчас

Компонент	Статус
Auth	✅
Gateway	✅ HTTP + WebSocket
Game	✅ Честная валидация
Leaderboard	✅ Redis Sorted Set
NATS	✅ JetStream
WebSocket	✅ Real-time broadcast
Graceful shutdown	✅


# WebSocket в Gateway — ГОТОВО! 🕸️

**Дата:** 9 мая 2026  
**Время:** 19:58

## Тест

```bash
# Терминал 1: wscat клиент
npx wscat -c ws://localhost:8080/ws/leaderboard

# Терминал 2: отправка рекорда
grpcurl -plaintext -d '{
  "user_id": "restart-test",
  "game_id": "hexagon",
  "level": 3,
  "seed": "restart_seed",
  "moves": []
}' localhost:50052 game.GameService/SubmitScore

# Результат в wscat:
< {"game_id":"hexagon","is_record":true,"level":3,"score":8,...,"user_id":"restart-test"}
Логи Gateway

text
🟢 WebSocket client connected. Total: 1
📡 Broadcasted score update to WebSocket clients