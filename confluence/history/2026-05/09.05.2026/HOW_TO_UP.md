Как переподниматься после перезагрузки:

Быстрый перезапуск всего кластера

bash
cd ~/event_horizon

# 1. Поднять Docker контейнеры (PostgreSQL, Redis, NATS)
docker-compose -f deployments/docker-compose.cluster.yml up -d

# 2. Проверить, что всё поднялось
docker ps | grep event-horizon

# 3. Запустить Auth
cd services/auth && ./auth-service &
cd ~/event_horizon

# 4. Запустить Leaderboard
cd services/leaderboard && ./leaderboard-service &
cd ~/event_horizon

# 5. Запустить Game
cd services/game && ./game-service &
cd ~/event_horizon

# 6. Запустить Gateway
cd services/gateway && ./gateway &
cd ~/event_horizon
Или одной командой (через Makefile)

Добавь в Makefile:

makefile
.PHONY: all
all:
	@echo "🚀 Starting all services..."
	docker-compose -f deployments/docker-compose.cluster.yml up -d
	cd services/auth && go build -o auth-service ./cmd/main.go && ./auth-service &
	cd services/leaderboard && go build -o leaderboard-service ./cmd/main.go && ./leaderboard-service &
	cd services/game && go build -o game-service ./cmd/main.go && ./game-service &
	cd services/gateway && go build -o gateway ./cmd/main.go && ./gateway &
	@echo "✅ All services started"
Затем:

bash
make all
Проверка после запуска

bash
# Проверить все сервисы
ps aux | grep -E "(auth|leaderboard|game|gateway)-service" | grep -v grep

# Проверить порты
ss -tlnp | grep -E "50051|50052|50054|8080"

# Должны быть:
# :50051 (Auth)
# :50052 (Game)
# :50054 (Leaderboard)
# :8080  (Gateway)
Тест WebSocket после перезапуска

bash
# В одном терминале
npx wscat -c ws://localhost:8080/ws/leaderboard

# В другом — отправить рекорд
grpcurl -plaintext -d '{
  "user_id": "restart-test",
  "game_id": "hexagon",
  "level": 3,
  "seed": "restart_seed",
  "moves": []
}' localhost:50052 game.GameService/SubmitScore