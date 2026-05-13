cd ~/event_horizon

# 1. Убедимся, что NATS работает
docker ps | grep nats
docker logs event-horizon-nats --tail 10

# 2. Запускаем сервисы (каждый в своей подоболочке, с правильным путём)
cd services/auth && go build -o auth-service ./cmd/main.go && ./auth-service > /tmp/auth.log 2>&1 &
cd ~/event_horizon

cd services/leaderboard && go build -o leaderboard-service ./cmd/main.go && ./leaderboard-service > /tmp/leaderboard.log 2>&1 &
cd ~/event_horizon

cd services/game && go build -o game-service ./cmd/main.go && ./game-service > /tmp/game.log 2>&1 &
cd ~/event_horizon

cd services/billing && go build -o billing-service ./cmd/main.go && ./billing-service > /tmp/billing.log 2>&1 &
cd ~/event_horizon

cd services/gateway && go build -o gateway ./cmd/main.go && ./gateway > /tmp/gateway.log 2>&1 &
cd ~/event_horizon

# 3. Проверяем логи через 5 секунд
sleep 5
tail -5 /tmp/auth.log
tail -5 /tmp/game.log
tail -5 /tmp/billing.log
tail -5 /tmp/leaderboard.log
tail -5 /tmp/gateway.log

ss -tlnp | grep -E "50051|50052|50053|50054|8080"