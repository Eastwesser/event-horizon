📋 Мануальная проверка после ребута и make deploy
Выполняй по порядку, отмечай ✅ или ❌

1. Проверка Docker-контейнеров
bash
docker ps --format "table {{.Names}}\t{{.Status}}" | grep event-horizon
Ожидаем: Все 22 контейнера в статусе Up или healthy.

2. Проверка Gateway метрик (RPS, Latency)
bash
# Проверяем, что метрики регистрируются
curl -s http://localhost:9095/metrics | grep gateway_requests_total
Ожидаем: Строки типа gateway_requests_total{method="GET",path="/health",status="200"} 123

3. Проверка Balancer метрик
bash
curl -s http://localhost:9098/metrics | grep balancer
Ожидаем: balancer_active_connections 0 и balancer_requests_total 0

4. Проверка Game метрик
bash
curl -s http://localhost:9092/metrics | grep game_submits_total
Ожидаем: game_submits_total{game_id="hexagon",status="success"} 122

5. Проверка NATS метрик
bash
curl -s http://localhost:7777/metrics | grep gnatsd_varz_connections
Ожидаем: gnatsd_varz_connections{server_id="..."} 6

6. Проверка Redis Exporter
bash
curl -s http://localhost:9121/metrics | grep redis_memory_used_bytes
Ожидаем: redis_memory_used_bytes 123456

7. Проверка PostgreSQL Exporter
bash
curl -s http://localhost:9187/metrics | grep pg_stat_database_numbackends
Ожидаем: pg_stat_database_numbackends{datname="eventhorizon"} 1

8. Проверка API (бизнес-логика)
bash
# Health
curl -s http://localhost:8079/health | jq '.'

# Регистрация
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","nickname":"Test"}' | jq '.'

# Логин и получение токена
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')
echo "Token: $TOKEN"

# Отправка рекорда
curl -X POST http://localhost:8079/api/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"user_id":"test-user","game_id":"hexagon","level":1,"score":150,"seed":"test_seed","moves":[]}' | jq '.'

# Лидерборд
curl -s "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10" | jq '.'
9. Проверка Jaeger (трейсинг)
bash
curl -s http://localhost:16686/api/services | jq '.data | length'
Ожидаем: >= 2 (должны быть gateway и jaeger-all-in-one)

10. Проверка Grafana (дашборды)
Открой в браузере: http://localhost:3000 (admin/admin). Проверь, что дашборд "Event Horizon" отображается с панелями.

📝 После проверки — запускай скрипт с нагрузкой
bash
cd ~/event_horizon
./metrics_snapshot.sh 60 50
Он соберёт все метрики и сохранит их в metrics_snapshots/.



---
