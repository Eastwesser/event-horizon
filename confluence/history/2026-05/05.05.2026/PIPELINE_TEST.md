Сквозной тест всей системы

Теперь у тебя есть полная цепочка:

curl → Gateway (8080) → Game (50052) → NATS (4222) → Leaderboard (50054) → Redis (6382)

Хочешь протестировать весь пайплайн? 🚀

# Терминал 1: Docker инфраструктура
cd ~/event_horizon
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Терминал 2: Auth
cd services/auth && ./auth-service

# Терминал 3: Leaderboard
cd services/leaderboard && ./leaderboard-service

# Терминал 4: Game
cd services/game && ./game-service

# Терминал 5: Gateway
cd services/gateway && go run cmd/main.go

Тест через Gateway (HTTP → gRPC)

bash
# 1. Регистрация пользователя
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"player@example.com","password":"secret123"}'

# 2. Логин (получить JWT)
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"player@example.com","password":"secret123"}'

# 3. Отправить рекорд через Gateway (нужен эндпоинт в gateway)
# TODO: добавить POST /api/game/submit в gateway

# AFTER ADDING POST IN GATEWAY:

# 1. Убедись, что все сервисы запущены:
#    - Auth (:50051)
#    - Game (:50052)
#    - Leaderboard (:50054)
#    - Gateway (:8080)

# 2. Отправить рекорд через HTTP
curl -X POST http://localhost:8080/api/game/submit \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-001",
    "game_id": "hexagon",
    "level": 3,
    "score": 250,
    "seed": "test_seed_123",
    "moves": [{"fromX":0,"fromY":0,"toX":1,"toY":1,"timestamp":1000}]
  }'

# 3. Проверить топ
grpcurl -plaintext -d '{"game_id":"hexagon","limit":10}' \
  localhost:50054 leaderboard.LeaderboardService/GetTopScores


# Try again:

Тест полного пайплайна

В разных терминалах:

bash
# Терминал 1: Docker
cd ~/event_horizon
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Терминал 2: Auth
cd services/auth && ./auth-service

# Терминал 3: Leaderboard
cd services/leaderboard && ./leaderboard-service

# Терминал 4: Game
cd services/game && ./game-service

# Терминал 5: Gateway
cd services/gateway && ./gateway
Отправка рекорда через HTTP

bash
curl -X POST http://localhost:8080/api/game/submit \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-from-http",
    "game_id": "hexagon",
    "level": 3,
    "score": 500,
    "seed": "http_test_seed",
    "moves": [
      {"fromX": 0, "fromY": 0, "toX": 1, "toY": 1, "timestamp": 1000}
    ]
  }'
Проверка топа

bash
grpcurl -plaintext -d '{"game_id":"hexagon","limit":10}' \
  localhost:50054 leaderboard.LeaderboardService/GetTopScores
