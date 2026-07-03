Перезапускаем Docker Compose

bash
cd /home/denismatveev/event_horizon

# Останавливаем всё
docker-compose -f deployments/docker-compose.cluster.yml down

# Запускаем заново
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Проверяем статус
docker-compose -f deployments/docker-compose.cluster.yml ps
Проверяем логи gateway

bash
# Смотрим логи первого gateway
docker-compose -f deployments/docker-compose.cluster.yml logs gateway --tail=30

# Смотрим логи всех gateway сразу
docker-compose -f deployments/docker-compose.cluster.yml logs | grep gateway
Проверяем работоспособность

bash
# Проверяем health каждого gateway
curl -s http://localhost:8081/health && echo " ✅ Gateway 8081 OK" || echo " ❌ Gateway 8081 FAIL"
curl -s http://localhost:8082/health && echo " ✅ Gateway 8082 OK" || echo " ❌ Gateway 8082 FAIL"
curl -s http://localhost:8083/health && echo " ✅ Gateway 8083 OK" || echo " ❌ Gateway 8083 FAIL"

# Проверяем balancer
curl -s http://localhost:8079/health