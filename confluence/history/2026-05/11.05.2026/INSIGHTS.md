Готов ли проект к нагрузке?

Короткий ответ: Да, архитектурно — да. Но есть нюансы:

Компонент	Готовность	    Что может упасть первым
Gateway	    ✅	            Goroutine leak, WebSocket connections
Game	    ✅	            CPU на валидации, память на хранении досок
Billing	    ✅	            PostgreSQL (если без кеша), Redis memory
Leaderboard	✅	            Redis memory (Sorted Set)
NATS	    ✅	            JetStream storage, ack latency
Auth	    ✅	            PostgreSQL connections, JWT generation

Что может стать узким местом:

PostgreSQL — если не использовать кеш для Billing
Redis memory — если топ-10 хранит много данных
WebSocket — если много одновременных соединений
NATS JetStream — диск может не успеть записывать

# TODO — 11 мая 2026

## Нагрузочное тестирование

### Инструменты для установки

```bash
# k6 для нагрузочного тестирования
sudo pacman -S k6

# bombardier для HTTP тестов
go install github.com/codesenberg/bombardier@latest

# wrk (альтернатива)
sudo pacman -S wrk

# pprof для Go профилирования (уже встроен)
Этапы тестирования

Уровень 1: Базовая проверка
10 одновременных пользователей — sanity check
100 пользователей — базовый порог
200 пользователей

Уровень 2: Средняя нагрузка
400 пользователей
600 пользователей
800 пользователей
1000 пользователей

Уровень 3: Целевая нагрузка (10k RPS)
2000 пользователей
4000 пользователей
6000 пользователей
8000 пользователей
10000 пользователей

Уровень 4: Экстремальная нагрузка (стресс-тест)
20000 пользователей
40000 пользователей
60000 пользователей
80000 пользователей
100000 пользователей

# ============================================
Что измеряем

Метрика	    Инструмент	Целевое значение

RPS         на Gateway	bombardier/wrk	10k+
P95 latency	k6/bombardier	< 100ms
NATS pub → sub latency	NATS metrics	< 10ms
Leaderboard (Redis)	pprof/redis-cli	< 5ms
Billing (PostgreSQL)	pg_stat_statements	< 20ms
Memory usage	ps aux или pprof	< 500MB per service
Goroutines	pprof	< 10000 total
CPU usage	top/htop	< 80% per core
WebSocket connections	Gateway logs	10k+

# ============================================

Тестовые сценарии

Сценарий 1: Регистрация + Логин

bash
# Регистрация нового пользователя
POST /api/auth/register
# Логин для получения JWT
POST /api/auth/login
Сценарий 2: Игровой цикл

bash
# Отправка рекорда → Game → NATS → Leaderboard
POST /api/game/submit
Сценарий 3: WebSocket (real-time)

bash
# Подключение к WebSocket, получение обновлений
ws://localhost:8080/ws/leaderboard
Сценарий 4: Billing запросы

bash
# Получение баланса
GET /api/billing/balance?user_id=...&currency=lamps
Команды для тестов

HTTP нагрузка через bombardier

bash
# Регистрация
bombardier -c 100 -n 10000 -m POST \
  -H "Content-Type: application/json" \
  -b '{"email":"test@example.com","password":"secret123"}' \
  http://localhost:8080/api/auth/register

# Отправка рекорда
bombardier -c 100 -n 10000 -m POST \
  -H "Content-Type: application/json" \
  -b '{"user_id":"test-uuid","game_id":"hexagon","level":3,"seed":"test","moves":[]}' \
  http://localhost:8080/api/game/submit

# =====================================================================================

# k6s to test

# Запуск k6
k6 run scripts/loadtest_k6.js
Мониторинг во время тестов

bash
# В отдельном терминале смотреть логи
tail -f /tmp/gateway.log | grep -E "latency|error|timeout"

# Метрики NATS
curl http://localhost:8222/varz | jq

# Статистика Redis
redis-cli -p 6382 INFO stats

# Статистика PostgreSQL
docker exec event-horizon-postgres-billing psql -U eventhorizon -c "SELECT * FROM pg_stat_database;"
Ожидаемые результаты

Нагрузка	Ожидаемый RPS	P95 latency	CPU (всего)	Memory
100	~500	< 50ms	~20%	~200MB
1000	~5000	< 100ms	~50%	~500MB
10000	~50000	< 500ms	~80%	~1-2GB
Критерии успеха

10 000 одновременных пользователей без падений
P95 latency < 500ms
0 ошибок при нормальной нагрузке
Нет утечек памяти (goroutines не растут бесконечно)
WebSocket держит 1000+ соединений
Оставшиеся сервисы (после тестов)

Notification (push/email)
Analytics (ClickHouse/PostgreSQL)
Payment (Boosty/Stripe)
Social (друзья)
Jaeger + Grafana + Prometheus (мониторинг)
Фронтенд (после бэкенда)

React + TypeScript
Drag-n-drop гексагоны (блинчики!)
Подключение к WebSocket
Интеграция с Billing API
Команды для быстрого старта

bash
cd ~/event_horizon
make all          # запустить всё
make stop         # остановить всё
make restart      # перезапустить всё
make ps           # статус Docker контейнеров

# Проверить логи
tail -f /tmp/auth.log
tail -f /tmp/game.log
tail -f /tmp/billing.log
tail -f /tmp/leaderboard.log
tail -f /tmp/gateway.log

# Запуск нагрузочного теста
k6 run scripts/loadtest_k6.js
