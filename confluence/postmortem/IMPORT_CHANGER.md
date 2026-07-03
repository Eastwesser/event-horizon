find services -name "*.go" -exec sed -i 's|github.com/Eastwesser/event-horizon/|event_horizon/|g' {} \;

find services -name "go.mod" -exec sed -i 's|module github.com/Eastwesser/event-horizon/|module event_horizon/|g' {} \;

find services -name "go.mod" -exec sed -i 's|github.com/Eastwesser/event-horizon/|event_horizon/|g' {} \;

🚀 После замены — пересобрать всё:

bash
cd /home/denismatveev/event_horizon

# Остановить всё
make stop-services

# Пересобрать бинарники
make all

# Пересобрать Docker образы и запушить
for svc in auth billing game gateway leaderboard balancer; do
  docker build -t eastwesser/$svc:latest -f services/$svc/Dockerfile services/$svc/
  docker push eastwesser/$svc:latest
done

# Запустить через Docker Compose
docker-compose -f deployments/docker-compose.cluster.yml down
docker-compose -f deployments/docker-compose.cluster.yml up -d

--

# 1. Проверить, что контейнеры живы
docker ps | grep -E "auth|billing|game|gateway|leaderboard|balancer"

# 2. Проверить health
curl -s http://localhost:8081/health && echo " ✅ Gateway 8081 OK"
curl -s http://localhost:8082/health && echo " ✅ Gateway 8082 OK"
curl -s http://localhost:8083/health && echo " ✅ Gateway 8083 OK"
curl -s http://localhost:8079/health && echo " ✅ Balancer OK"

# 3. Проверить логи (если что-то не работает)
docker logs deployments-auth-1 --tail 10
docker logs deployments-gateway-1 --tail 10