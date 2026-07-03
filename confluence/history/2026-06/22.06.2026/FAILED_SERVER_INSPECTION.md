Leaderboard падает с lookup leaderboard on 127.0.0.11:53: no such host. Это значит, что leaderboard не запущен или у него нет сети.

Проверяем:
bash
# 1. Жив ли leaderboard?
docker-compose -f deployments/docker-compose.cluster.yml ps | grep leaderboard

# 2. Смотрим логи
docker-compose -f deployments/docker-compose.cluster.yml logs leaderboard --tail=30

# 3. Проверяем есть ли контейнер
docker ps | grep leaderboard

# 4. Проверяем сеть
docker network inspect event-horizon-net | grep -A5 leaderboard

Если leaderboard не запущен — поднимаем:
bash
docker-compose -f deployments/docker-compose.cluster.yml up -d leaderboard
docker-compose -f deployments/docker-compose.cluster.yml logs leaderboard --tail=30

Если leaderboard запущен, но не виден в сети:

Проблема: у leaderboard нет networks: - event-horizon-net.

Решение: добавь в docker-compose.cluster.yml:

yaml
leaderboard:
  networks:
    - event-horizon-net
И перезапусти:

bash
docker-compose -f deployments/docker-compose.cluster.yml up -d leaderboard
