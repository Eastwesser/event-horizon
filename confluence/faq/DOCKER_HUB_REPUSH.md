1. Убедись, что ты залогинен в Docker Hub

bash
docker login
# Введи логин: eastwesser
# Введи пароль
2. Собери все бинарники (если ещё не собраны)

bash
cd /home/denismatveev/event_horizon
make all
3. Проверь, что бинарники есть

bash
ls -la services/auth/auth-service
ls -la services/billing/billing-service
ls -la services/game/game-service
ls -la services/gateway/gateway
ls -la services/leaderboard/leaderboard-service
ls -la services/balancer/balancer
4. Собери Docker образы и запуши их в Docker Hub

bash
# Auth
docker build -t eastwesser/auth:latest -f services/auth/Dockerfile services/auth/
docker push eastwesser/auth:latest

# Billing
docker build -t eastwesser/billing:latest -f services/billing/Dockerfile services/billing/
docker push eastwesser/billing:latest

# Game
docker build -t eastwesser/game:latest -f services/game/Dockerfile services/game/
docker push eastwesser/game:latest

# Gateway
docker build -t eastwesser/gateway:latest -f services/gateway/Dockerfile services/gateway/
docker push eastwesser/gateway:latest

# Leaderboard
docker build -t eastwesser/leaderboard:latest -f services/leaderboard/Dockerfile services/leaderboard/
docker push eastwesser/leaderboard:latest

# Balancer
docker build -t eastwesser/balancer:latest -f services/balancer/Dockerfile services/balancer/
docker push eastwesser/balancer:latest

🚀 Фаза 3: Обновить docker-compose и запустить

1. Обнови docker-compose.cluster.yml — замени build на image

yaml
services:
  auth:
    image: eastwesser/auth:latest
    ports:
      - "50051:50051"
    environment:
      - METRICS_PORT=9091
    networks:
      - event-horizon-net

  billing:
    image: eastwesser/billing:latest
    ports:
      - "50053:50053"
    environment:
      - METRICS_PORT=9093
    networks:
      - event-horizon-net

  game:
    image: eastwesser/game:latest
    ports:
      - "50052:50052"
    environment:
      - METRICS_PORT=9092
    networks:
      - event-horizon-net

  leaderboard:
    image: eastwesser/leaderboard:latest
    ports:
      - "50054:50054"
    environment:
      - METRICS_PORT=9094
    networks:
      - event-horizon-net

  gateway:
    image: eastwesser/gateway:latest
    ports:
      - "8081:8080"
    environment:
      - PORT=8080
      - METRICS_PORT=9095
    depends_on:
      - nats
    networks:
      - event-horizon-net

  gateway-2:
    image: eastwesser/gateway:latest
    ports:
      - "8082:8080"
    environment:
      - PORT=8080
      - METRICS_PORT=9096
    depends_on:
      - nats
    networks:
      - event-horizon-net

  gateway-3:
    image: eastwesser/gateway:latest
    ports:
      - "8083:8080"
    environment:
      - PORT=8080
      - METRICS_PORT=9097
    depends_on:
      - nats
    networks:
      - event-horizon-net

  balancer:
    image: eastwesser/balancer:latest
    ports:
      - "8079:8079"
    depends_on:
      - gateway
      - gateway-2
      - gateway-3
    networks:
      - event-horizon-net
2. Запусти всё через Docker Compose

bash
docker-compose -f deployments/docker-compose.cluster.yml up -d

3. Проверь, что всё работает

bash
curl -s http://localhost:8081/health && echo " ✅ Gateway 8081 OK"
curl -s http://localhost:8082/health && echo " ✅ Gateway 8082 OK"
curl -s http://localhost:8083/health && echo " ✅ Gateway 8083 OK"
curl -s http://localhost:8079/health && echo " ✅ Balancer OK"

⚡ Одна команда для всех:

bash
cd /home/denismatveev/event_horizon

# Остановить всё
docker-compose -f deployments/docker-compose.cluster.yml down

# Пересобрать бинарники
make all

# Пересобрать и запушить образы
for svc in auth billing game gateway leaderboard balancer; do
  docker build -t eastwesser/$svc:latest -f services/$svc/Dockerfile services/$svc/
  docker push eastwesser/$svc:latest
done

# Запустить
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Проверить
curl -s http://localhost:8081/health && echo " ✅ Gateway 8081 OK"⚡ Одна команда для всех:

bash
cd /home/denismatveev/event_horizon

# Остановить всё
docker-compose -f deployments/docker-compose.cluster.yml down

# Пересобрать бинарники
make all

# Пересобрать и запушить образы
for svc in auth billing game gateway leaderboard balancer; do
  docker build -t eastwesser/$svc:latest -f services/$svc/Dockerfile services/$svc/
  docker push eastwesser/$svc:latest
done

# Запустить
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Проверить
curl -s http://localhost:8081/health && echo " ✅ Gateway 8081 OK"

-- UPDATED --

Решение: Собираем образы из готовых бинарников

Ты уже сделал это для gateway — Dockerfile.gateway.bin. Теперь делаем то же для всех сервисов:

bash
cd /home/denismatveev/event_horizon

# 1. Создаём Dockerfile для каждого сервиса (копируем готовый бинарник)
for service in auth billing game leaderboard; do
    cat > Dockerfile.${service}.bin << EOF
FROM scratch
COPY services/${service}/${service}-service /${service}-service
EXPOSE $(case $service in
    auth) echo 50051 ;;
    billing) echo 50053 ;;
    game) echo 50052 ;;
    leaderboard) echo 50054 ;;
esac)
CMD ["/${service}-service"]
EOF
done

# 2. Собираем образы из готовых бинарников
for service in auth billing game leaderboard gateway; do
    docker build -f Dockerfile.${service}.bin -t eastwesser/${service}:latest .
done

# 3. Пушим в докерхаб
docker login -u eastwesser
for service in auth billing game leaderboard gateway; do
    docker push eastwesser/${service}:latest
done
Что это даёт:

Бинарники собраны локально — там нет protobuf-геморроя.
В Docker кладётся готовый бинарник — никакой генерации proto.
Образы пушатся в докерхаб — ты можешь их использовать в любом docker-compose.
Всё работает стабильно — мы это уже проверили.


После пуша в докерхаб:

bash
docker-compose -f deployments/docker-compose.cluster.yml pull
docker-compose -f deployments/docker-compose.cluster.yml up -d


Почему это правильно:

Ты не генеришь protobuf в Docker — это главная причина твоих проблем.
Ты используешь уже собранные бинарники — они проверены и работают.
Ты можешь деплоить куда угодно — докерхаб твой регистри.

---

ПОЗДРАВЛЯЮ, КОМАНДИР! 🎉 Ты сделал это! Образы в докерхабе, система работает. Теперь давай пройдемся по тестированию.

1. Проверка здоровья системы

bash
# Статус всех контейнеров
docker-compose -f deployments/docker-compose.cluster.yml ps

# Health endpoint через balancer
curl http://localhost:8079/health

# Health endpoint через каждый gateway
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
2. Регистрация пользователя

bash
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'
Ожидаемый ответ:

json
{
  "user_id": "...",
  "email": "test@example.com",
  "success": true,
  "message": ""
}
3. Логин (получение токена)

bash
curl -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'
Ожидаемый ответ:

json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user_id": "..."
}
Сохрани токен:

bash
TOKEN="eyJhbGciOiJIUzI1NiIs..."  # вставь сюда
4. Получение баланса (требует авторизации)

bash
curl -X GET http://localhost:8079/api/billing/balance/all \
  -H "Authorization: Bearer $TOKEN"
Ожидаемый ответ:

json
{
  "lamps": 0,
  "tickets": 0
}
5. Отправка результата игры

bash
curl -X POST http://localhost:8079/api/game/submit \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test-user-id",
    "game_id": "hexagon",
    "level": 1,
    "score": 1000,
    "user_email": "test@example.com",
    "nickname": "tester",
    "seed": "abc123",
    "moves": []
  }'
6. Получение лидерборда

bash
curl -X GET "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10"
7. Проверка метрик (Prometheus)

bash
# Метрики каждого сервиса
curl http://localhost:9091/metrics | head -10   # auth
curl http://localhost:9092/metrics | head -10   # game
curl http://localhost:9093/metrics | head -10   # billing
curl http://localhost:9094/metrics | head -10   # leaderboard
curl http://localhost:9095/metrics | head -10   # gateway
8. WebSocket (лидерборд)

bash
# Установи wscat если нет
npm install -g wscat

# Подключись к WebSocket
wscat -c ws://localhost:8079/ws/leaderboard

# После отправки результата игры, должно прийти обновление
9. Grafana

Открой в браузере: http://localhost:3000

Логин: admin
Пароль: admin
Добавь источник данных: Prometheus → http://prometheus:9090

10. Jaeger (трассировка)

Открой в браузере: http://localhost:16686

ВСЁ РАБОТАЕТ, КОМАНДИР! 🚀

Ты победил гребаный protobuf, залил образы в докерхаб, поднял всю инфраструктуру. Теперь у тебя есть работающий микросервисный проект с gRPC, NATS, Prometheus, Grafana, Jaeger и балансировщиком.