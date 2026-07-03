версия protobuf v1.35.1 тоже имеет этот баг! Давай понизим до v1.34.2 (стабильная версия):

Фикс для всех сервисов:

bash
cd /home/denismatveev/event_horizon

# Понижаем protobuf до v1.34.2 во всех сервисах
for svc in auth billing game leaderboard gateway; do
  echo "Fixing $svc..."
  cd services/$svc
  go get google.golang.org/protobuf@v1.34.2
  go mod tidy
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $svc-service ./cmd/main.go
  cd ../..
done

# Собираем бинарник для gateway
cd services/gateway
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway ./cmd/main.go

# Собираем Docker образ
docker build -t eastwesser/gateway:latest -f Dockerfile.simple .

# Пушим
docker push eastwesser/gateway:latest

# Перезапускаем
cd /home/denismatveev/event_horizon
docker-compose -f deployments/docker-compose.cluster.yml down
docker-compose -f deployments/docker-compose.cluster.yml up -d

sleep 5
docker ps | grep gateway
curl -s http://localhost:8081/health && echo " ✅ Gateway 8081 OK"