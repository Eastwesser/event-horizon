cd ~/event_horizon

# Останавливаем
make stop-services

# Пересобираем leaderboard (с правильным интерфейсом)
cd services/leaderboard
go build -o leaderboard-service ./cmd/main.go

# Пересобираем gateway
cd ../gateway
go build -o gateway ./cmd/main.go

# Пересобираем billing
cd ../billing
go build -o billing-service ./cmd/main.go

# Пересобираем game (на всякий случай)
cd ../game
go build -o game-service ./cmd/main.go

# Пересобираем auth
cd ../auth
go build -o auth-service ./cmd/main.go

# Запускаем всё
cd ~/event_horizon

make down

make up

make all

# Проверяем статус
make status

# FULL RESTART (clears db!)

cd ~/event_horizon

# 1. Остановить все Go сервисы
make stop-services

# 2. Остановить и удалить Docker контейнеры
docker-compose -f deployments/docker-compose.cluster.yml down -v

# 3. Удалить временные логи (опционально)
rm -f /tmp/*.log

# 4. Поднять Docker контейнеры заново
docker-compose -f deployments/docker-compose.cluster.yml up -d

# 5. Проверить, что контейнеры поднялись
docker ps | grep event-horizon

# 6. Пересобрать и запустить все Go сервисы
make all

# 7. Проверить статус
make status

# 8. Проверить, что Gateway отвечает
curl http://localhost:8080/health