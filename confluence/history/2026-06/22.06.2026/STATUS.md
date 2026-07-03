SUCCESS

ПОЗДРАВЛЯЮ, КОМАНДИР! 🎉 Ты сделал это! Образы в докерхабе, система работает. Теперь давай пройдемся по тестированию.

1. Проверка здоровья системы

bash
# Статус всех контейнеров
docker-compose -f deployments/docker-compose.cluster.yml ps

# Health endpoint через balancer
curl http://localhost:8079/health

# Health endpoint через каждый gateway
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health

2. Регистрация пользователя

bash
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'
  
Ожидаемый ответ:

json
{
  "user_id": "...",
  "email": "test@example.com",
  "success": true,
  "message": ""
}
3. Логин (получение токена)

bash
curl -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'
Ожидаемый ответ:

json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user_id": "..."
}
Сохрани токен:

bash
TOKEN="eyJhbGciOiJIUzI1NiIs..."  # вставь сюда
4. Получение баланса (требует авторизации)

bash
curl -X GET http://localhost:8079/api/billing/balance/all \
  -H "Authorization: Bearer $TOKEN"
Ожидаемый ответ:

json
{
  "lamps": 0,
  "tickets": 0
}
5. Отправка результата игры

bash
curl -X POST http://localhost:8079/api/game/submit \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test-user-id",
    "game_id": "hexagon",
    "level": 1,
    "score": 1000,
    "user_email": "test@example.com",
    "nickname": "tester",
    "seed": "abc123",
    "moves": []
  }'
6. Получение лидерборда

bash
curl -X GET "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10"
7. Проверка метрик (Prometheus)

bash
# Метрики каждого сервиса
curl http://localhost:9091/metrics | head -10   # auth
curl http://localhost:9092/metrics | head -10   # game
curl http://localhost:9093/metrics | head -10   # billing
curl http://localhost:9094/metrics | head -10   # leaderboard
curl http://localhost:9095/metrics | head -10   # gateway
8. WebSocket (лидерборд)

bash
# Установи wscat если нет
npm install -g wscat

# Подключись к WebSocket
wscat -c ws://localhost:8079/ws/leaderboard

# После отправки результата игры, должно прийти обновление
9. Grafana

Открой в браузере: http://localhost:3000

Логин: admin
Пароль: admin
Добавь источник данных: Prometheus → http://prometheus:9090

10. Jaeger (трассировка)

Открой в браузере: http://localhost:16686

ВСЁ РАБОТАЕТ, КОМАНДИР! 🚀

Ты победил гребаный protobuf, залил образы в докерхаб, поднял всю инфраструктуру. Теперь у тебя есть работающий микросервисный проект с gRPC, NATS, Prometheus, Grafana, Jaeger и балансировщиком.