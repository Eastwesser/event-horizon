# TODO - 6 мая 2026

## Статус на сегодня (5 мая)
✅ Полный пайплайн работает: curl → Gateway → Game → NATS → Leaderboard → Redis
✅ Auth (JWT, регистрация, логин)
✅ Gateway (HTTP → gRPC прокси, NATS publisher)
✅ Game (валидация заглушка, NATS publisher)
✅ Leaderboard (Redis Sorted Set, NATS subscriber)

## Задачи на 6 мая

### Приоритет 1: Исправить валидацию в Game
- [ ] Реализовать полноценную валидацию ходов в `games/hexagons/validator.go`
  - [ ] Восстановление состояния по seed
  - [ ] Применение всех ходов из moves
  - [ ] Проверка, что полученный счёт совпадает с finalScore
  - [ ] Проверка solvability начальной конфигурации (уровни 1-5)

### Приоритет 2: WebSocket для реального времени
- [ ] Добавить WebSocket в Gateway
  - [ ] Эндпоинт `/ws/leaderboard?game_id=hexagon`
  - [ ] Подписка на NATS `score.updated`
  - [ ] Рассылка обновлений всем подключённым клиентам
- [ ] Клиентская часть (пока через `wscat` для теста)

### Приоритет 3: Фронтенд (React)
- [ ] Начать проект `frontend/` с React + TypeScript
- [ ] Drag-n-drop для гексагонального пазла
- [ ] Подключение к Gateway API
- [ ] WebSocket для live leaderboard

### Техдолг (можно доделать)
- [ ] Graceful shutdown для всех сервисов (у Auth уже есть, остальным добавить)
- [ ] NATS кластер из 3 нод (сейчас 1)
- [ ] Prometheus + Grafana метрики
- [ ] `GIN_MODE=release` для продакшен-замеров

### Игровая логика (из манифеста)
- [ ] Выбор уровня сложности (3/5) с пропорциональной наградой
- [ ] Подсказка за лампочки
- [ ] Сброс партии за лампочки
- [ ] Защита от нерешаемой партии

## Быстрые команды для запуска всех сервисов

```bash
# Терминал 1: Docker
cd ~/event_horizon && docker-compose -f deployments/docker-compose.cluster.yml up -d

# Терминал 2: Auth
cd services/auth && ./auth-service

# Терминал 3: Leaderboard
cd services/leaderboard && ./leaderboard-service

# Терминал 4: Game
cd services/game && ./game-service

# Терминал 5: Gateway
cd services/gateway && ./gateway
Тестирование после запуска

bash
# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Логин
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Отправить рекорд
curl -X POST http://localhost:8080/api/game/submit \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-test",
    "game_id": "hexagon",
    "level": 3,
    "score": 1000,
    "seed": "test_seed",
    "moves": [{"fromX":0,"fromY":0,"toX":1,"toY":1,"timestamp":1000}]
  }'

# Проверить топ
grpcurl -plaintext -d '{"game_id":"hexagon","limit":10}' \
  localhost:50054 leaderboard.LeaderboardService/GetTopScores
Следующий большой этап

Полноценная игра с честной валидацией
WebSocket leaderboard
React фронтенд
Помнить: Мы строим игру, а не казино. 🎮

text

---

## Коммит

```bash
cd ~/event_horizon
git add .
git commit -m "✅ Full pipeline working: Auth + Gateway + Game + NATS + Leaderboard

- Auth: JWT, registration, login
- Gateway: HTTP → gRPC proxy, NATS publisher
- Game: SubmitScore, NATS publisher, validation stub
- Leaderboard: Redis Sorted Set, NATS subscriber
- Full integration: curl → Gateway → Game → NATS → Leaderboard → Redis

Tech debt recorded, TODO for tomorrow created"



Отличный вопрос! Ты мыслишь как senior engineer — после того как всё заработало, спросить про утечки горутин.

Быстрая проверка (без паники)

Коротко по каждому сервису:

Сервис	Потенциальная утечка	Статус
Auth	gRPC сервер	✅ Graceful shutdown есть
Gateway	Gin + gRPC conn + NATS conn	🟡 Нет graceful shutdown
Game	gRPC + NATS publish	🟡 Нет graceful shutdown
Leaderboard	gRPC + NATS subscribe	✅ Graceful shutdown есть
Что работает хорошо

В Leaderboard мы добавили graceful shutdown:

go
go func() {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    log.Println("Shutting down gracefully...")
    grpcServer.GracefulStop()
    nc.Drain()
    os.Exit(0)
}()
Что нужно добавить

В Gateway и Game нет graceful shutdown. При Ctrl+C контексты не закрываются, соединения не дрейнятся.

Проверим прямо сейчас (лёгкий тест)

bash
# Запусти gateway, подожди пару секунд, нажми Ctrl+C
# Посмотри, есть ли "Shutting down gracefully..." — скорее всего нет, сервис просто упадёт.
Добавляем в TODO на завтра

/home/denismatveev/event_horizon/confluence/history/2026-05/06.05.2026/TODO.md — дополнение:

markdown
## Приоритет 0 (критично)

### Graceful shutdown для всех сервисов

- [ ] **Auth** — ✅ уже есть
- [ ] **Gateway** — добавить:
  - [ ] Ловим SIGINT/SIGTERM
  - [ ] `grpcServer.GracefulStop()` для gRPC клиента
  - [ ] `nc.Drain()` для NATS
  - [ ] `gin` сервер — `srv.Shutdown(ctx)`
- [ ] **Game** — добавить (аналогично Gateway)
- [ ] **Leaderboard** — ✅ уже есть

### Проверка утечек горутин

```bash
# Запустить сервис, под нагрузкой посмотреть число горутин
curl -s http://localhost:8080/debug/pprof/goroutine?debug=1

# Или через pprof
go tool pprof http://localhost:6060/debug/pprof/goroutine
Добавить pprof эндпоинт для каждого сервиса (на отдельном порту)
Провести нагрузочный тест на 1000 запросов и проверить, что горутины не копятся