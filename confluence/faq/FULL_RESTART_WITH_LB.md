Чистый запуск

bash
cd /home/denismatveev/event_horizon

# 1. Убиваем все локальные процессы
pkill -f "auth-service" 2>/dev/null
pkill -f "billing-service" 2>/dev/null
pkill -f "game-service" 2>/dev/null
pkill -f "leaderboard-service" 2>/dev/null
pkill -f "gateway-service" 2>/dev/null
pkill -f "gateway" 2>/dev/null

# 2. Гасим Docker контейнеры
docker-compose -f deployments/docker-compose.cluster.yml down

# 3. Проверяем что порты свободны
sudo ss -tlnp | grep -E "50051|50052|50053|50054|8080|8079"

# 4. Поднимаем инфраструктуру в Docker (без gateway)
docker-compose -f deployments/docker-compose.cluster.yml up -d postgres redis postgres-game redis-game postgres-billing redis-billing postgres-leaderboard redis-leaderboard nats prometheus grafana jaeger balancer

# 5. Ждем 5 секунд, пока базы поднимутся
sleep 5

# 6. Запускаем микросервисы локально (по одному, в фоне)
cd services/auth && ./auth-service > /tmp/auth.log 2>&1 &
cd ../billing && ./billing-service > /tmp/billing.log 2>&1 &
cd ../game && ./game-service > /tmp/game.log 2>&1 &
cd ../leaderboard && ./leaderboard-service > /tmp/leaderboard.log 2>&1 &
cd ../gateway && ./gateway-service > /tmp/gateway.log 2>&1 &

6.5

Терминал 1 — Auth
bash
cd /home/denismatveev/event_horizon/services/auth
./auth-service

Терминал 2 — Billing
bash
cd /home/denismatveev/event_horizon/services/billing
./billing-service

Терминал 3 — Game
bash
cd /home/denismatveev/event_horizon/services/game
./game-service

Терминал 4 — Leaderboard
bash
cd /home/denismatveev/event_horizon/services/leaderboard
./leaderboard-service

Терминал 5 — Gateway
bash
cd /home/denismatveev/event_horizon/services/gateway
./gateway-service


Если какой-то сервис упадет — ты сразу увидишь ошибку и сможешь быстро починить.

Но если хочешь запустить всё в фоне с логами в файлы (как ты предложил):

bash
cd /home/denismatveev/event_horizon

# Очищаем старые логи
rm -f /tmp/auth.log /tmp/billing.log /tmp/game.log /tmp/leaderboard.log /tmp/gateway.log

# Запускаем все в фоне
cd services/auth && nohup ./auth-service > /tmp/auth.log 2>&1 &
cd ../billing && nohup ./billing-service > /tmp/billing.log 2>&1 &
cd ../game && nohup ./game-service > /tmp/game.log 2>&1 &
cd ../leaderboard && nohup ./leaderboard-service > /tmp/leaderboard.log 2>&1 &
cd ../gateway && nohup ./gateway-service > /tmp/gateway.log 2>&1 &

# Смотрим логи в реальном времени
tail -f /tmp/auth.log /tmp/billing.log /tmp/game.log /tmp/leaderboard.log /tmp/gateway.log

# 7. Проверяем что все запустились
sleep 3
make status

# 8. Проверяем health
curl http://localhost:8080/health
curl http://localhost:8079/health
Если auth или gateway всё ещё падают — проверь логи:

bash
tail -20 /tmp/auth.log
tail -20 /tmp/gateway.log
Покажи что там. Если всё ок — будет {"status":"ok"}.