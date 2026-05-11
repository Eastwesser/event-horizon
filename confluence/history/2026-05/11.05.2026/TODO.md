# TODO — 11 мая 2026

## Нагрузочное тестирование

### Этапы
- [ ] 10 одновременных пользователей
- [ ] 100 пользователей
- [ ] 200 пользователей
- [ ] 400 пользователей
- [ ] 600 пользователей
- [ ] 800 пользователей
- [ ] 1000 пользователей

### Высокая нагрузка
- [ ] 2000 пользователей
- [ ] 4000 пользователей
- [ ] 6000 пользователей
- [ ] 8000 пользователей
- [ ] 10000 пользователей

### Экстремальная нагрузка (цель)
- [ ] 20000 пользователей
- [ ] 40000 пользователей
- [ ] 60000 пользователей
- [ ] 80000 пользователей
- [ ] 100000 пользователей

### Что измеряем
- [ ] RPS на Gateway
- [ ] Задержка NATS (публикация → получение)
- [ ] Leaderboard latency (Redis Sorted Set)
- [ ] Billing latency (PostgreSQL)
- [ ] Потребление памяти Go сервисами
- [ ] CPU usage
- [ ] Количество горутин

### Инструменты
- [ ] k6 + NATS extension
- [ ] wrk / bombardier для HTTP
- [ ] pprof для Go сервисов

## Оставшиеся сервисы
- [ ] Notification (push/email)
- [ ] Analytics (ClickHouse/PostgreSQL)
- [ ] Payment (Boosty/Stripe)
- [ ] Social (друзья)

## Фронтенд (потом)
- [ ] React + TypeScript
- [ ] Drag-n-drop гексагоны
- [ ] Подключение к WebSocket
- [ ] Интеграция с Billing API

## Команды для быстрого старта после перезагрузки

```bash
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
```

# AFTER HIGH LOAD

# TODO — 11 мая 2026 (20:21)


# 10 пользователей
bombardier -c 10 -n 100 -m POST -H "Content-Type: application/json" \
  -b '{"user_id":"load-10","game_id":"hexagon","level":3,"seed":"load10","moves":[]}' \
  http://localhost:8080/api/game/submit

# 100 пользователей
bombardier -c 100 -n 1000 -m POST -H "Content-Type: application/json" \
  -b '{"user_id":"load-100","game_id":"hexagon","level":3,"seed":"load100","moves":[]}' \
  http://localhost:8080/api/game/submit

# 200 пользователей
bombardier -c 200 -n 2000 -m POST -H "Content-Type: application/json" \
  -b '{"user_id":"load-200","game_id":"hexagon","level":3,"seed":"load200","moves":[]}' \
  http://localhost:8080/api/game/submit

# 400 пользователей
bombardier -c 400 -n 4000 -m POST -H "Content-Type: application/json" \
  -b '{"user_id":"load-400","game_id":"hexagon","level":3,"seed":"load400","moves":[]}' \
  http://localhost:8080/api/game/submit

# 600 пользователей
bombardier -c 600 -n 6000 -m POST -H "Content-Type: application/json" \
  -b '{"user_id":"load-600","game_id":"hexagon","level":3,"seed":"load600","moves":[]}' \
  http://localhost:8080/api/game/submit

# 800 пользователей
bombardier -c 800 -n 8000 -m POST -H "Content-Type: application/json" \
  -b '{"user_id":"load-800","game_id":"hexagon","level":3,"seed":"load800","moves":[]}' \
  http://localhost:8080/api/game/submit

# 1000 пользователей
bombardier -c 1000 -n 10000 -m POST -H "Content-Type: application/json" \
  -b '{"user_id":"load-1000","game_id":"hexagon","level":3,"seed":"load1000","moves":[]}' \
  http://localhost:8080/api/game/submit

# 2000 пользователей
bombardier -c 2000 -n 20000 -m POST -H "Content-Type: application/json" \
  -b '{"user_id":"load-2000","game_id":"hexagon","level":3,"seed":"load2000","moves":[]}' \
  http://localhost:8080/api/game/submit

# 4000 пользователей (ПИК ПРОИЗВОДИТЕЛЬНОСТИ)
bombardier -c 4000 -n 40000 -m POST -H "Content-Type: application/json" \
  -b '{"user_id":"load-4000","game_id":"hexagon","level":3,"seed":"load4000","moves":[]}' \
  http://localhost:8080/api/game/submit

# 6000 пользователей (НАЧАЛО ДЕГРАДАЦИИ)
bombardier -c 6000 -n 60000 -m POST -H "Content-Type: application/json" \
  -b '{"user_id":"load-6000","game_id":"hexagon","level":3,"seed":"load6000","moves":[]}' \
  http://localhost:8080/api/game/submit


## Нагрузочное тестирование — ВЫПОЛНЕНО ✅

### Команды bombardier для всех этапов

#### Установка
```bash
go install github.com/codesenberg/bombardier@latest
export PATH=$PATH:$(go env GOPATH)/bin
Результаты

Пользователи	RPS	P95 latency	Ошибки	Статус
10	643	31ms	0	✅
100	1691	120ms	0	✅
200	2474	257ms	0	✅
400	2271	432ms	0	✅
600	2309	700ms	0	✅
800	3060	790ms	0	✅
1000	9756	110ms	0	✅
2000	20852	107ms	0	✅
4000	25395	185ms	0	✅ ПИК
6000	19213	380ms	1	⚠️ ДЕГРАДАЦИЯ
Точка отказа

8000 пользователей → 1194 ошибки, система начала коллапс
10000 пользователей → 50% ошибок, полный коллапс
Вертикальный скейлинг (можно ли?)

Да, частично:

Компонент	Можно увеличить	Предел
RAM	до 64-128GB	✅
CPU	до 32 ядер	✅
Open files limit (ulimit -n)	до 65535	✅
TCP порты	до 65535	✅
PostgreSQL max_connections	до 500-1000	✅
Что нельзя решить вертикально:

NATS slow consumer (нужен кластер)
Redis single-threaded (нужен кластер)
PostgreSQL write bottleneck (нужны реплики)
Вывод: Вертикальный скейлинг поможет дойти до ~10-15k пользователей, но для 100k нужен горизонтальный.

Горизонтальный скейлинг (для 100k+)

5-10 Gateway
5-10 Game
3-5 Leaderboard + Redis Cluster
3-5 Billing + PostgreSQL Cluster
3-5 NATS нод
Load balancer (nginx/haproxy)
Оставшиеся сервисы (после фронтенда)

Notification (push/email)
Analytics (ClickHouse/PostgreSQL)
Payment (Boosty/Stripe)
Social (друзья)
Фронтенд (следующий этап)

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

# Мониторинг NATS
watch -n 1 'curl -s http://localhost:8222/varz | jq ".now, .in_msgs, .out_msgs, .slow_consumers"'


Отвечая на твой вопрос про вертикальный скейлинг

"Можно ли заскейлиться вертикально? Или уже все?"
Ответ: Частично ДА, но только до ~10-15k пользователей.

Что даст вертикальный скейлинг (увеличение ресурсов 1 сервера):

Увеличение	Эффект
RAM 8GB → 64GB	Выше выдержит до 8000-10000 пользователей
CPU 4 ядра → 16 ядер	Увеличит RPS на 50-100%
ulimit -n 1024 → 65535	Позволит держать 10000+ TCP соединений
net.core.somaxconn 128 → 4096	Увеличит очередь соединений
Что НЕ даст вертикальный скейлинг:

NATS останется single point
Redis останется single-threaded
PostgreSQL станет bottleneck на запись

Итог: Максимум на 1 мощном сервере — ~10-15k пользователей. 
Дальше только горизонтальный скейлинг.
