[denismatveev@c0der event_horizon]$ cd /home/denismatveev/event_horizon/services/gateway
[denismatveev@c0der gateway]$ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w -extldflags=-static" \
  -o gateway-service ./cmd/main.go
[denismatveev@c0der gateway]$ file gateway-service
gateway-service: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, BuildID[sha1]=727a10ec0caa1389bd628f0078b49cadbccc9fef, stripped
[denismatveev@c0der gateway]$ 

ДА! Статический бинарник готов! Теперь он запустится в scratch.

Теперь сделаем то же самое для всех сервисов:

bash
cd /home/denismatveev/event_horizon

# 1. Пересобираем все сервисы со статической линковкой
for service in auth billing game leaderboard gateway; do
    echo "Building $service..."
    cd services/$service
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-s -w -extldflags=-static" \
        -o ${service}-service ./cmd/main.go
    file ${service}-service | grep -q "statically linked" && echo "✅ $service static" || echo "❌ $service dynamic"
    cd ../..
done
После этого пересобери все образы:

bash
# 2. Пересобираем Dockerfile для всех сервисов (уже есть Dockerfile.*.bin)
for service in auth billing game leaderboard gateway; do
    echo "Building image for $service..."
    docker build -f Dockerfile.${service}.bin -t eastwesser/${service}:latest --no-cache .
    docker push eastwesser/${service}:latest
done

# 3. Запускаем всё
docker-compose -f deployments/docker-compose.cluster.yml down
docker-compose -f deployments/docker-compose.cluster.yml pull
docker-compose -f deployments/docker-compose.cluster.yml up -d

# 4. Проверяем
docker-compose -f deployments/docker-compose.cluster.yml ps
curl http://localhost:8079/health
Варианты:

Если static не собирается — используй alpine вместо scratch:
dockerfile
FROM alpine:latest
COPY services/auth/auth-service /auth-service
RUN chmod +x /auth-service
EXPOSE 50051
CMD ["/auth-service"]
Если хочешь универсальный подход — создай один Dockerfile для всех:
dockerfile
FROM alpine:latest
ARG SERVICE
ARG PORT
COPY services/${SERVICE}/${SERVICE}-service /${SERVICE}-service
RUN chmod +x /${SERVICE}-service
EXPOSE ${PORT}
CMD ["/${SERVICE}-service"]
