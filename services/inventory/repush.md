# CORE

cd /home/denismatveev/event_horizon/services/inventory

# 1. Очищаем кеш (на всякий случай)
go clean -cache

# 2. Собираем
go build -o inventory-service ./cmd/main.go

# 3. Если всё ок — финальная сборка
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o inventory-service ./cmd/main.go

# 4. Собираем образ
docker build -t eastwesser/inventory:latest -f Dockerfile.inventory.bin .

# 5. Пушим
docker push eastwesser/inventory:latest

# ALTERNATIVE:

cd /home/denismatveev/event_horizon

# 1. Останавливаем всё
docker-compose -f deployments/docker-compose.cluster.yml down

# 2. Пересобираем бинарник
cd services/inventory
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o inventory-service ./cmd/main.go
cd ../..

# 3. Пушим через make (если есть)
make deploy

# Или вручную поднимаем
docker-compose -f deployments/docker-compose.cluster.yml up -d

## COMMON

cd /home/denismatveev/event_horizon

# Убедись, что бинарник собран
cd services/inventory
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o inventory-service ./cmd/main.go
cd ../..

# Собери образ из бинарника
docker build -t eastwesser/inventory:latest -f Dockerfile.inventory.bin .

# Пуш
docker push eastwesser/inventory:latest

# Перезапуск
docker-compose -f deployments/docker-compose.cluster.yml down
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Проверка
curl http://localhost:9096/health