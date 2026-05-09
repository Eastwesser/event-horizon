# ФАЗА 3: ЧЕСТНАЯ ВАЛИДАЦИЯ ИГРЫ — ГОТОВО! 🎉

**Дата:** 9 мая 2026  
**Время:** 16:30  
**Версия:** v0.3.0

---

## Что сделано

| Компонент | Статус | Детали |
|-----------|--------|--------|
| **Auth** | ✅ | JWT, регистрация, логин, bcrypt |
| **Gateway** | ✅ | HTTP → gRPC прокси, NATS publisher |
| **Game** | ✅ | Честная валидация, подсчёт очков, solvability |
| **Leaderboard** | ✅ | Redis Sorted Set, NATS subscriber |
| **NATS JetStream** | ✅ | Событийная шина, durable subscription |
| **Graceful shutdown** | ✅ | Game, Gateway, Leaderboard |

---

## Полный рабочий пайплайн
curl → Gateway (8080) → Game (50052) → валидация + подсчёт → NATS (4222) → Leaderboard (50054) → Redis (6382)

text

---

## Ключевые технические решения

### 1. Серверная валидация (Путь В)
Клиент отправляет `seed + moves`, сервер:
- Восстанавливает доску из `seed`
- Применяет все ходы
- Честно подсчитывает очки

**Протокол (game.proto):**
```protobuf
message SubmitScoreRequest {
    string user_id = 1;
    string game_id = 2;
    int32 level = 3;
    string seed = 4;           // начальное состояние
    repeated Move moves = 5;   // все ходы игрока
    // Поле score УДАЛЕНО — сервер сам вычисляет!
}
2. Генерация solvable доски

go
func GenerateInitialBoard(seed string, level int) (*Board, error) {
    // 1. Создаём все клетки как "empty"
    // 2. Заполняем случайные клетки плитками
    // 3. Проверяем IsSolvable()
    // 4. Если нет — регенерируем (максимум 5 попыток)
}
3. Graceful shutdown для всех сервисов

go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

log.Println("Shutting down gracefully...")
grpcServer.GracefulStop()
nc.Drain()
Тестовый запуск

1. Поднять инфраструктуру

bash
cd ~/event_horizon
docker-compose -f deployments/docker-compose.cluster.yml up -d
2. Запустить сервисы

bash
# Терминал 1: Auth
cd services/auth && ./auth-service

# Терминал 2: Leaderboard
cd services/leaderboard && ./leaderboard-service

# Терминал 3: Game
cd services/game && ./game-service

# Терминал 4: Gateway
cd services/gateway && ./gateway
3. Отправить рекорд

bash
grpcurl -plaintext -d '{
  "user_id": "test-1",
  "game_id": "hexagon",
  "level": 3,
  "seed": "test_seed_123",
  "moves": []
}' localhost:50052 game.GameService/SubmitScore
Ответ:

json
{
  "success": true,
  "newHighscore": 8,
  "message": "score submitted successfully",
  "lampsEarned": 10
}
4. Проверить топ

bash
grpcurl -plaintext -d '{"game_id":"hexagon","limit":10}' \
  localhost:50054 leaderboard.LeaderboardService/GetTopScores
Логи работающей системы

text
✅ Solvable board generated on attempt 0
🎮 Game service listening on :50052
📡 Published score.updated: user=test-1, game=hexagon, score=8, is_record=true

📡 Received score update via NATS: game=hexagon user=test-1 score=8
✅ Score updated for test-1, new rank: 5
Инсайты и выводы

Что узнали про NATS

JetStream даёт persistence без боли
Durable subscription переживает рестарт consumer'а
msg.Ack() гарантирует обработку
Что узнали про генерацию доски

Нужно сначала создать ВСЕ клетки как "empty"
Проверка solvability — только после этого
5 попыток достаточно для генерации (обычно с 1-й работает)
Что узнали про graceful shutdown

nc.Drain() важнее nc.Close()
gRPC сервер требует GracefulStop()
Следующие шаги (TODO на 9 мая)

WebSocket в Gateway (real-time leaderboard)
React фронтенд (drag-n-drop гексагоны)
Billing сервис (лампочки, билетики)
Prometheus + Grafana мониторинг
NATS кластер из 3 нод
Итог

Система готова к демонстрации:

text
curl → Gateway → Game → NATS → Leaderboard → Redis

Всё работает, не падает, считает честно. 🚀
Маленькими шагами — к большой цели.
