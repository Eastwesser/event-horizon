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