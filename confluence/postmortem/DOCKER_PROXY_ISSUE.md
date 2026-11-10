1. Почему это произошло (коротко)
Exited (255) + docker-proxy, который не умер.

255 — это код выхода «контейнер убит сигналом» (не 0, не 1, не 137). Jaeger не «упал с ошибкой», его прибило извне — либо хост ребутнулся, либо Docker daemon перезапускался, либо кто-то сделал docker kill. При жёстком убийстве контейнера Docker не всегда чистит docker-proxy — процесс остаётся висеть, держит :4317, и Docker считает порт занятым, хотя контейнера уже нет .

kill 1810 1819 не помог, потому что docker-proxy пересоздаётся Docker daemon'ом, пока тот считает, что порт должен быть занят. systemctl restart docker — единственное, что гарантированно чистит всех docker-proxy и iptables-состояние .

То есть: Exited (255) — причина, «зомби docker-proxy» — следствие, restart docker — лечение.

2. Можно ли make deploy прямо сейчас?
Да, но не make deploy целиком. Смотри, что у тебя в docker ps:

deployments-notification-1 → Restarting (1) — это не поднялось после restart docker. Exit code 1 = приложение упало с ошибкой при старте (скорее всего, не достучалось до Postgres/NATS, потому что те поднимались параллельно). Оно само перезапустится, но лучше проверить логи.

Остальные — Up или Up (health: starting). Через 10-30 секунд станут healthy.

Что делать сейчас:

bash
# 1. Проверить, что notification ожил (или почему нет)
docker logs --tail 50 deployments-notification-1

# 2. Если notification в Restarting больше минуты — перезапустить его
docker compose --env-file .env -f deployments/docker-compose.cluster.yml restart notification

# 3. Убедиться, что jaeger поднялся
docker compose --env-file .env -f deployments/docker-compose.cluster.yml ps jaeger

# 4. Докомпоновать то, чего не хватает (если что-то не поднялось после restart docker)
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d
make deploy целиком сейчас — не надо. Он:

Пересоберёт nats-hub (30 секунд, не нужно — он уже собран).

Поднимет Postgres, подождёт pg_isready (все уже healthy — пропустит).

Прогонит migrate-all (no migrations to run — пропустит).

docker compose up -d — вот это и есть то, что тебе нужно. Шаг 4 = docker compose up -d. Сделай его напрямую (команда 4 выше).

3. После ребута — что будет?
Зависит от restart: политики у сервисов.

Смотри свой docker-compose.cluster.yml:

notification, fulfillment, payment, authors, history, analytics, nginx — у них restart: unless-stopped. После ребута поднимутся сами.

jaeger, prometheus, grafana, alertmanager, auth, game, billing, leaderboard, profile, shop, inventory, gateway, balancer — у них нет restart: в compose (или no по умолчанию). После ребута НЕ поднимутся.

То есть после ребута:

bash
# Docker daemon поднимется (systemctl enable docker — проверь, что enabled)
sudo systemctl is-enabled docker

# Часть сервисов поднимется сама (те, у кого restart: unless-stopped)
# Остальные — вручную:
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d
Или — make deploy после ребута (он сделает всё: Postgres, миграции, up -d).

Но make deploy после ребута — правильнее, потому что:

Postgres после ребута может долго восстанавливаться (unclean shutdown).

make deploy ждёт pg_isready до запуска остальных.

Без make deploy ты можешь запустить up -d слишком рано, и сервисы упадут с «connection refused» к Postgres.

Вывод: после ребута — make deploy. Сейчас (стек уже частично жив) — docker compose up -d, без make.

4. Итог — что делать прямо сейчас
bash
# 1. Проверить notification (единственный в Restarting)
docker logs --tail 50 deployments-notification-1
docker compose --env-file .env -f deployments/docker-compose.cluster.yml restart notification

# 2. Докомпоновать то, чего не хватает
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d

# 3. Проверить, что всё живо
docker compose --env-file .env -f deployments/docker-compose.cluster.yml ps
# ждём: все Up или Up (healthy), нет Restarting, нет Exited

# 4. Проверить nginx → balancer → gateway
curl -s -o /dev/null -w "%{http_code}\n" http://localhost/health
Скинь docker compose ps после этого — скажу, всё ли ок.

