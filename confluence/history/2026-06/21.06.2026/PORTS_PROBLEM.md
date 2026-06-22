Balancer ищет gateway на 127.0.0.1:8081, а gateway ВНУТРИ контейнера слушает на 8080, но проброшен на хост как 8081. Внутри Docker-сети gateway доступен по имени сервиса и порту 8080, а не по localhost:8081.

Проблема: Balancer запущен в Docker и пытается достучаться до 127.0.0.1:8081 (это его собственный localhost, а не хост). Ему нужно обращаться к gateway:8080, gateway-2:8080, gateway-3:8080.

Решение: Пересобрать balancer с правильными бекендами

bash
cd /home/denismatveev/event_horizon/services/balancer

# Меняем бекенды на имена сервисов в Docker-сети
sed -i 's|127.0.0.1:8081|gateway:8080|g' cmd/main.go
sed -i 's|127.0.0.1:8082|gateway-2:8080|g' cmd/main.go
sed -i 's|127.0.0.1:8083|gateway-3:8080|g' cmd/main.go

# Собираем статический бинарник
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w -extldflags=-static" \
  -o balancer-service ./cmd/main.go

# Проверяем
file balancer-service

# Создаем Dockerfile для balancer (если нет)
cd /home/denismatveev/event_horizon
cat > Dockerfile.balancer.bin << 'EOF'
FROM scratch
COPY services/balancer/balancer-service /balancer
EXPOSE 8079
CMD ["/balancer"]
EOF

# Собираем и пушим
docker build -f Dockerfile.balancer.bin -t eastwesser/balancer:latest --no-cache .
docker push eastwesser/balancer:latest

# Перезапускаем balancer
docker-compose -f deployments/docker-compose.cluster.yml down balancer
docker-compose -f deployments/docker-compose.cluster.yml pull balancer
docker-compose -f deployments/docker-compose.cluster.yml up -d balancer

# Проверяем
curl http://localhost:8079/health
Почему так: Balancer в Docker должен обращаться к другим сервисам по их именам в Docker-сети (gateway, gateway-2, gateway-3), а не по localhost.