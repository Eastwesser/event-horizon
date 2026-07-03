# Останавливаем всё
docker-compose -f deployments/docker-compose.cluster.yml down

# Собираем Game
docker build -t eastwesser/game:latest -f services/game/Dockerfile .

# Собираем Leaderboard
docker build -t eastwesser/leaderboard:latest -f services/leaderboard/Dockerfile .

# Собираем Gateway
docker build -t eastwesser/gateway:latest -f services/gateway/Dockerfile .

# Пушим в Docker Hub
docker push eastwesser/game:latest
docker push eastwesser/leaderboard:latest
docker push eastwesser/gateway:latest

# Запускаем всё заново
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Проверяем статус
docker-compose -f deployments/docker-compose.cluster.yml ps

# Проверяем NATS Exporter
curl http://localhost:7777/metrics | head -20


# Почему это работает

Флаг	Что делает
CGO_ENABLED=0	Отключает CGO → бинарник не зависит от системных библиотек
-ldflags="-s -w"	Удаляет отладочную информацию → уменьшает размер на 30-40%
FROM scratch	Создаёт минимальный образ без лишних слоёв (всего ~10-15 МБ)

# 🚀 После сборки и запуска

Ты получишь:
Метрики NATS на http://localhost:7777/metrics (автоматически собираются Prometheus)
Game и Leaderboard с обновлённым кодом (с context.WithTimeout и Retry)
E2E-тест можно запустить командой k6 run deployments/k6/e2e-test.js

📝 Проверка после запуска
bash
# Проверить, что NATS Exporter работает
curl -s http://localhost:7777/metrics | grep -E "nats_conns|nats_msgs"

# Проверить, что Prometheus видит NATS
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job=="nats") | .health'

# Проверить логи Game
docker logs event-horizon-game --tail=20

# Проверить логи Leaderboard
docker logs event-horizon-leaderboard --tail=20


