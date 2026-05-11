это был тест ТОЛЬКО одной ручки POST /api/game/submit! А если добавить WebSocket, регистрацию, логин, billing запросы, получение топа — нагрузка вырастет кратно.

Давай пересчитаем с учётом полного профиля пользователя

Один пользователь за сессию делает:

Действие	Кол-во	Запросов в час
Авторизация	1 раз в день	0.04
Регистрация	1 раз в жизни	~0
Просмотр топа	10 раз	10
Игры	5-10 партий	5-10
Получение баланса	5 раз	5
WebSocket keepalive	постоянно	~3600 пакетов
Итого на одного активного пользователя: ~15-20 API запросов + WebSocket трафик.

Пересчитанные цифры

Параллельных игроков	RPS на пике (только /submit)	Реальный RPS (все эндпоинты)	Статус
4000	25,000	~5,000-8,000	✅ Устойчиво
6000	19,000	~4,000-6,000	⚠️ Деградация
8000	16,700	~3,000-5,000	❌ Таймауты
10000	3,700	~1,000-2,000	❌ Коллапс
Реалистичная оценка для 1 сервера

text
Максимум одновременных игроков: ~5,000-6,000
с комфортной игрой (без таймаутов)
Почему? Учитываем:

20% игроков одновременно совершают действия
80% просто висят в WebSocket (мало нагрузки на API)
NATS должен обрабатывать события от всех
Команды для повторного тестирования

Добавлю их в FINAL_RESULTS.md:

bash
# Базовые команды для нагрузочного тестирования

## Установка bombardier
```bash
go install github.com/codesenberg/bombardier@latest
export PATH=$PATH:$(go env GOPATH)/bin
Тесты POST /api/game/submit

bash
# 1000 пользователей
bombardier -c 1000 -n 10000 -m POST \
  -H "Content-Type: application/json" \
  -b '{"user_id":"test-1000","game_id":"hexagon","level":3,"seed":"test1000","moves":[]}' \
  http://localhost:8080/api/game/submit

# 2000 пользователей
bombardier -c 2000 -n 20000 -m POST \
  -H "Content-Type: application/json" \
  -b '{"user_id":"test-2000","game_id":"hexagon","level":3,"seed":"test2000","moves":[]}' \
  http://localhost:8080/api/game/submit

# 4000 пользователей (пик)
bombardier -c 4000 -n 40000 -m POST \
  -H "Content-Type: application/json" \
  -b '{"user_id":"test-4000","game_id":"hexagon","level":3,"seed":"test4000","moves":[]}' \
  http://localhost:8080/api/game/submit

# 6000 пользователей (деградация)
bombardier -c 6000 -n 60000 -m POST \
  -H "Content-Type: application/json" \
  -b '{"user_id":"test-6000","game_id":"hexagon","level":3,"seed":"test6000","moves":[]}' \
  http://localhost:8080/api/game/submit

# 8000 пользователей (таймауты)
bombardier -c 8000 -n 80000 -m POST \
  -H "Content-Type: application/json" \
  -b '{"user_id":"test-8000","game_id":"hexagon","level":3,"seed":"test8000","moves":[]}' \
  http://localhost:8080/api/game/submit

# 10000 пользователей (коллапс)
bombardier -c 10000 -n 100000 -m POST \
  -H "Content-Type: application/json" \
  -b '{"user_id":"test-10000","game_id":"hexagon","level":3,"seed":"test10000","moves":[]}' \
  http://localhost:8080/api/game/submit
Тесты регистрации

bash
bombardier -c 1000 -n 1000 -m POST \
  -H "Content-Type: application/json" \
  -b '{"email":"test$i@example.com","password":"secret123"}' \
  http://localhost:8080/api/auth/register
Тесты получения топа

bash
bombardier -c 1000 -n 10000 -m GET \
  -H "Content-Type: application/json" \
  -b '{"game_id":"hexagon","limit":10}' \
  http://localhost:50054 leaderboard.LeaderboardService/GetTopScores
Мониторинг во время тестов

bash
# NATS метрики
watch -n 1 'curl -s http://localhost:8222/varz | jq ".now, .in_msgs, .out_msgs, .slow_consumers"'

# Логи сервисов
tail -f /tmp/leaderboard.log | grep -E "score|error"
tail -f /tmp/game.log | grep -E "Published|error"
tail -f /tmp/billing.log | grep -E "Added|error"
Измерение лимитов системы

bash
# Проверить лимит открытых файлов
ulimit -n

# Проверить порты
sysctl net.ipv4.ip_local_port_range

# Проверить память
free -h

# Проверить CPU
top -p $(pgrep -d',' -f "auth-service|game-service|billing-service|leaderboard-service|gateway")
Результаты (обновлённые)

Пользователи	RPS	P95 latency	Ошибки	Статус
4000	25395	185ms	0	✅ ПИК
6000	19213	380ms	1	⚠️ Деградация
8000	16732	560ms	1194	❌ Таймауты
10000	3695	19s	53034	❌ Коллапс
Вывод

1 сервер (Arch Linux) выдерживает стабильно до 4000 одновременных пользователей
при условии что 20-30% из них активно играют, а остальные просто висят в WebSocket.

Пиковая нагрузка: 25,000 RPS на ручку /submit
Реальная ёмкость: ~5,000-6,000 игроков с комфортной задержкой

Для 100,000 игроков нужно: 5-10 серверов + кластеризация NATS/Redis/PostgreSQL