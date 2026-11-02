# 🥶 Event Horizon — Cold Start Playbook

> **Цель:** Пошаговая инструкция, как поднять проект с нуля на пустом сервере (или после полного вайпа).
> **Предположение:** У тебя есть доступ к GitHub и последние бэкапы в S3.

## 1. Подготовка сервера (Ubuntu/Arch)

```bash
# 1. Установить Docker
sudo apt update && sudo apt install -y docker.io docker-compose-plugin git make

# 2. Добавить пользователя в группу docker
sudo usermod -aG docker $USER
newgrp docker

# 3. Проверить
docker --version
make --version
2. Клонирование и настройка
bash
# 1. Клонировать репозиторий
git clone git@github.com:Eastwesser/event-horizon.git
cd event-horizon

# 2. Восстановить .env (из Vault или из архива)
cp .env.example .env
nano .env
# Вставить: JWT_SECRET, POSTGRES_PASSWORD, GRAFANA_ADMIN_PASSWORD, TELEGRAM_BOT_TOKEN
3. Запуск пустого стека (без данных)
bash
# 1. Поднять контейнеры
make deploy

# 2. Проверить, что всё живо
make status
curl http://localhost:8079/health
4. Восстановление данных из бэкапа
bash
# 1. Скачать дампы из S3
aws s3 sync s3://eh-backups/postgres/ /tmp/restore/

# 2. Восстановить базы
for db in auth game billing leaderboard shop inventory profile payment authors history notification; do
  echo "Restoring $db..."
  docker exec -i event-horizon-postgres psql -U eventhorizon -d eventhorizon_$db < /tmp/restore/$db.sql
done

# 3. Восстановить ClickHouse (если нужно)
aws s3 sync s3://eh-backups/clickhouse/ /tmp/clickhouse-restore/
docker exec -it event-horizon-clickhouse clickhouse-backup restore <backup_name>
5. Проверка после восстановления
bash
# 1. Smoke-тесты
make test-smoke

# 2. Проверить, что пользователи на месте
docker exec -it event-horizon-postgres psql -U eventhorizon -d eventhorizon -c "SELECT count(*) FROM users;"

# 3. Проверить, что магазин работает
curl -s http://localhost:8079/api/v1/shop/items | jq '. | length'
6. Переключение DNS (если нужно)
Зайти в Cloudflare.

Поменять A-запись на новый IP.

Подождать TTL.

7. Что делать, если бэкапа нет
Если бэкапа нет (полная потеря данных):

Запустить make deploy (пустая система).

Засеять админа: make seed-admin.

Засеять магазин: make seed-v110.

Сказать пользователям, что данные потеряны (если они были).
