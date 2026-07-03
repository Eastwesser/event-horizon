🔍 БАЗОВЫЕ CURL-ПРОВЕРКИ:
1. Healthcheck всего кластера
bash
# Балансировщик
curl -s http://localhost:8079/health | jq '.'

# Gateway напрямую (все три)
curl -s http://localhost:8081/health | jq '.'
curl -s http://localhost:8082/health | jq '.'
curl -s http://localhost:8083/health | jq '.'

# NATS (проверка, что работает)
curl -s http://localhost:8222/varz | jq '.server_id'
2. Регистрация пользователя
bash
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "secret123",
    "nickname": "TestPlayer"
  }' | jq '.'
3. Логин (получить токен)
bash
# Сохраняем токен в переменную
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "secret123"
  }' | jq -r '.access_token')

echo "Токен: $TOKEN"
4. Проверка баланса (Billing)
bash
curl -X GET http://localhost:8079/api/billing/balance/all \
  -H "Authorization: Bearer $TOKEN" | jq '.'
5. Отправка рекорда (Game)
bash
curl -X POST http://localhost:8079/api/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "user_id": "test-user",
    "game_id": "hexagon",
    "level": 1,
    "score": 150,
    "seed": "test_seed_123",
    "moves": []
  }' | jq '.'
6. Проверка лидерборда (Leaderboard)
bash
# Всегда без токена (публичный эндпоинт)
curl -s "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10" | jq '.'
7. Проверка NATS (сообщения)
bash
# Проверить, что NATS видит подключения
curl -s http://localhost:8222/varz | jq '.connections'
8. Проверка Prometheus
bash
# Проверить, что все таргеты UP
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
9. Проверка Jaeger
bash
# Проверить сервисы в Jaeger
curl -s http://localhost:16686/api/services | jq '.'

# Проверить количество трейсов для gateway
curl -s "http://localhost:16686/api/traces?service=gateway&limit=1" | jq '.data | length'
🚀 ЗАПУСК ФРОНТЕНДА:
bash
cd ~/event_horizon/frontend
npm start
# или
yarn start
# или
pnpm start
Фронтенд обычно запускается на http://localhost:3000 (или 5173 для Vite).

🧪 K6 ТЕСТ (после того как всё проверишь):
bash
cd ~/event_horizon/deployments/k6
k6 run e2e-test.js
📋 ЧЕК-ЛИСТ (чтобы убедиться, что всё работает):
Проверка	Команда	Ожидаемый результат
Health	curl localhost:8079/health	{"status":"ok"}
Регистрация	POST /register	{"success":true}
Логин	POST /login	Токен
Баланс	GET /balance	{"lamps":0,"tickets":0}
Рекорд	POST /game/submit	{"success":true}
Лидерборд	GET /leaderboard	{"entries":[]} или с данными
NATS	/varz	{"server_id":...}
Prometheus	/api/v1/targets	Все "health":"up"
Jaeger	/api/services	["gateway", "jaeger-all-in-one"]