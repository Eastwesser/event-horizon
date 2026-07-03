🚀 ВЫПОЛНЯЕМ ПО ИНСТРУКЦИИ:
bash
cd ~/event_horizon/services/gateway

# 1. Собираем статический бинарник (уже есть, но пересоберем для чистоты)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags=-static" \
    -o gateway-service ./cmd/main.go

# 2. Проверяем, что он статический
file gateway-service
# Должно быть: statically linked

# 3. Возвращаемся в корень
cd ~/event_horizon

# 4. Собираем образ из готового бинарника
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .

# 5. Пушим в Docker Hub
docker push eastwesser/gateway:latest

# 6. Перезапускаем
docker-compose -f deployments/docker-compose.cluster.yml down
docker-compose -f deployments/docker-compose.cluster.yml up -d

# 7. Проверяем логи
docker logs deployments-gateway-1 --tail=30 | grep -i jaeger

🎯 ПОСЛЕ ЭТОГО ДОЛЖНО ПОЯВИТЬСЯ:
text
🔄 Initializing Jaeger tracer with endpoint: jaeger:4317
✅ Jaeger tracer initialized for Gateway

🔍 Проверяем трейсы:
bash
curl "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10"
curl -s http://localhost:16686/api/services | jq '.'
curl -s "http://localhost:16686/api/traces?service=gateway&limit=1" | jq '.data | length'