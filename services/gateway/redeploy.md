cd /home/denismatveev/event_horizon

# Собираем Gateway
cd services/gateway
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go
cd ../..

# Собираем Docker образ
docker build -t eastwesser/gateway:latest -f Dockerfile.gateway.bin .

# Пушим
docker push eastwesser/gateway:latest

# Перезапускаем Gateway
docker-compose -f deployments/docker-compose.cluster.yml stop gateway gateway-2 gateway-3
docker-compose -f deployments/docker-compose.cluster.yml up -d gateway gateway-2 gateway-3

# Проверяем
curl http://localhost:8081/health