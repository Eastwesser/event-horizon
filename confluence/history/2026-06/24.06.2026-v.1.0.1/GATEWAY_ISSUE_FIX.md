# 1. Собрать бинарник локально
cd /home/denismatveev/event_horizon/services/gateway
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway ./cmd/main.go

# 2. Собрать Docker образ (из папки services/gateway)
docker build -t eastwesser/gateway:latest .

# 3. Запушить
docker push eastwesser/gateway:latest

# 4. Перезапустить
cd /home/denismatveev/event_horizon
docker-compose -f deployments/docker-compose.cluster.yml up -d gateway gateway-2 gateway-3