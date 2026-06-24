cd ~/event_horizon

# 1. Полный стоп ВСЕГО (контейнеры, сети, volumes)
docker-compose -f deployments/docker-compose.cluster.yml down -v

# 2. Проверяем, что всё чисто
docker ps -a | grep event-horizon

# 3. Поднимаем ЗАНОВО
docker-compose -f deployments/docker-compose.cluster.yml up -d

# 4. Смотрим, как взлетает
docker-compose -f deployments/docker-compose.cluster.yml logs -f