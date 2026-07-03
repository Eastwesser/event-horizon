Event Horizon — Success.md
📦 Что поднимается
Инфраструктура (Docker)
Сервис	Порт	Назначение
PostgreSQL (4 шт)	5460-5463	БД для auth, billing, game, leaderboard
Redis (4 шт)	6379-6382	Кэш и сессии для каждого сервиса
NATS	4222, 8222	Message bus + JetStream
Jaeger	16686	Трассировка (UI)
Prometheus	9090	Сбор метрик
Grafana	3000	Дашборды
Микросервисы (Docker)
Сервис	gRPC порт	Metrics порт	Назначение
Auth	50051	9091	Регистрация, JWT
Billing	50053	9093	Балансы, транзакции
Game	50052	9092	Игровая логика
Leaderboard	50054	9094	Рейтинги
Gateway	8081-8083	9095-9097	API Gateway (3 инстанса)
Balancer	8079	9098	Load Balancer
Архитектура
text
[Frontend] → [Balancer :8079]
                ↓
         [Gateway :8081/8082/8083]
                ↓
    ┌───────────┼───────────┐
    ↓           ↓           ↓
 [Auth]    [Billing]   [Game]    [Leaderboard]
    ↓           ↓           ↓           ↓
 [PG:5460] [PG:5462]  [PG:5461]   [PG:5463]
 [Redis]    [Redis]     [Redis]     [Redis]
                      ↕
                   [NATS :4222]
                      ↕
          [Jaeger :16686] [Prometheus :9090]
                ↓
           [Grafana :3000]
🚀 Как поднять систему
1. Клонировать репозиторий
bash
git clone https://github.com/Eastwesser/event-horizon.git
cd event-horizon
2. Поднять всё одной командой
bash
docker-compose -f deployments/docker-compose.cluster.yml up -d
3. Проверить статус
bash
docker-compose -f deployments/docker-compose.cluster.yml ps
4. Проверить health
bash
# Через balancer
curl http://localhost:8079/health

# Напрямую через gateway
curl http://localhost:8081/health
5. Накатить миграции (если БД пустые)
bash
cd services/auth && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5460/eventhorizon?sslmode=disable" up
cd services/billing && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5462/eventhorizon_billing?sslmode=disable" up
cd services/game && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5461/eventhorizon_game?sslmode=disable" up
cd services/leaderboard && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5463/eventhorizon_leaderboard?sslmode=disable" up
6. Проверить API
bash
# Регистрация
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'

# Логин
curl -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'
7. Мониторинг
Grafana: http://localhost:3000 (admin/admin)

Jaeger: http://localhost:16686

Prometheus: http://localhost:9090

🔥 Постмортем (что сломали и как починили)
День 1: Начало ада
Проблема: Gateway падает с panic: runtime error: slice bounds out of range в protobuf.

Диагноз: protoc генерирует битый RawDescriptor для .pb.go файлов.

Что перепробовали:

✅ Все версии protoc (25.1, 24.4)

✅ Все версии protoc-gen-go (1.28.0 → 1.36.11)

✅ Все версии google.golang.org/protobuf (1.33.0 → 1.34.2 → 1.36.11)

✅ Добавляли enum-заглушки в proto

✅ Меняли go_package на все возможные варианты

✅ Перегенерировали _grpc.pb.go

✅ Пробовали Dockerfile с полным контекстом

Решение: Собрать бинарник локально со статической линковкой и скопировать в Docker-образ:

dockerfile
FROM scratch
COPY services/gateway/gateway-service /gateway
CMD ["/gateway"]
И флаги сборки:

bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -extldflags=-static" -o gateway-service ./cmd/main.go
День 2: DNS-ад
Проблема: Сервисы в Docker не видят друг друга по имени.

Диагноз: В docker-compose.cluster.yml не было сетей для всех сервисов.

Решение: Добавить networks: - event-horizon-net для всех сервисов.

День 3: NATS-нил
Проблема: Gateway падает с nil pointer dereference в NATS.

Диагноз: defer nc.Drain() вызывался на nil-объекте, когда NATS не подключался.

Решение: Обернуть NATS-блок в if nc != nil, а js объявить на уровне функции.

День 4: Balancer-слепота
Проблема: Balancer не видит gateway.

Диагноз: Balancer искал 127.0.0.1:8081, а нужно gateway:8080 (имена в Docker-сети).

Решение: Исправить бекенды в balancer и пересобрать статически.

День 5: Метрики-молчание
Проблема: Prometheus не видит метрики.

Диагноз: У balancer не было обработчика /metrics, а NATS отдаёт JSON, а не Prometheus-формат.

Решение:

Balancer: добавить promhttp.Handler() и пробросить порт 9098

NATS: забить и не мониторить через Prometheus

Итоговый урок
Корень всех проблем: protoc + protobuf + gRPC в Docker-контексте — это ад. Особенно если версии разъезжаются.

Рабочая стратегия:

Собирать бинарники локально со статической линковкой

Копировать готовые бинарники в Docker-образы (FROM scratch)

Использовать имена сервисов как DNS (в одной Docker-сети)

Не генерировать protobuf в процессе сборки Docker-образа

✅ Что работает сейчас
Все 5 микросервисов в Docker

3 gateway + balancer

PostgreSQL, Redis, NATS, Jaeger, Prometheus, Grafana

Регистрация и логин

Health checks

Метрики (кроме NATS)

WebSocket для лидерборда

Трассировка в Jaeger

Образы в докерхабе (eastwesser/*:latest)

📝 Команды для быстрого старта
bash
# Всё поднять
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Проверить здоровье
curl http://localhost:8079/health

# Посмотреть метрики
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# Логи всех сервисов
docker-compose -f deployments/docker-compose.cluster.yml logs -f

# Пересобрать один сервис
docker build -f Dockerfile.auth.bin -t eastwesser/auth:latest .
docker push eastwesser/auth:latest
docker-compose -f deployments/docker-compose.cluster.yml up -d auth