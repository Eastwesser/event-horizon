# Billing Service 💰

Управление игровыми валютами: лампочки (за активность) и билетики (пассивный доход).

## Структура
services/billing/
├── cmd/
│ └── main.go # точка входа
├── internal/
│ ├── config/
│ │ └── config.go # конфигурация
│ ├── handler/
│ │ └── grpc_handler.go # gRPC хендлер
│ ├── repository/
│ │ ├── postgres_repo.go # PostgreSQL хранилище
│ │ └── redis_repo.go # Redis кеш
│ └── service/
│ └── billing_service.go # бизнес-логика
├── proto/
│ ├── billing.proto # gRPC контракт
│ ├── billing.pb.go
│ └── billing_grpc.pb.go
├── scripts/
│ └── init.sql # схема БД
├── go.mod
├── go.sum
├── Dockerfile
└── README.md

text

## API (gRPC)

| Метод | Назначение |
|-------|------------|
| `GetBalance` | Получить баланс по типу валюты |
| `GetAllBalances` | Получить все балансы пользователя |
| `AddCurrency` | Начислить валюту |
| `SpendCurrency` | Списать валюту |
| `GetTransactionHistory` | История транзакций |

## База данных

### Таблица `user_currencies`
| Колонка | Тип | Назначение |
|---------|-----|------------|
| user_id | UUID | ID пользователя |
| currency_type | TEXT | 'lamps' или 'tickets' |
| balance | INT | Текущий баланс |
| updated_at | TIMESTAMP | Время обновления |

### Таблица `transactions`
| Колонка | Тип | Назначение |
|---------|-----|------------|
| id | UUID | Уникальный ID |
| user_id | UUID | ID пользователя |
| currency_type | TEXT | Тип валюты |
| amount | INT | Сумма (+/−) |
| balance_after | INT | Баланс после операции |
| reason | TEXT | Причина ('game_reward', 'daily_play', 'hint') |
| reference_id | TEXT | Idempotency key |
| created_at | TIMESTAMP | Время операции |

## Запуск

```bash
# Локальный запуск
cd services/billing
go build -o billing-service ./cmd/main.go
./billing-service

# Через Makefile (из корня)
make all
Тестирование

bash
# Получить баланс лампочек
grpcurl -plaintext -d '{"user_id":"<UUID>","currency":1}' \
  localhost:50053 billing.BillingService/GetBalance

# Начислить 100 лампочек
grpcurl -plaintext -d '{
  "user_id": "<UUID>",
  "currency": 1,
  "amount": 100,
  "reason": "test",
  "reference_id": "test-1"
}' localhost:50053 billing.BillingService/AddCurrency

# Получить все балансы
grpcurl -plaintext -d '{"user_id":"<UUID>"}' \
  localhost:50053 billing.BillingService/GetAllBalances
Интеграция с NATS

Billing подписан на топик score.updated и автоматически начисляет награды:

lamps_earned → лампочки
tickets_earned → билетики
Порты

Сервис	Порт
gRPC	50053
PostgreSQL	5462
Redis	6381
```

## UPD INFO 3rd of July, 2026:

```bash
cd ~/event_horizon/services/billing
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o billing-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.billing.bin -t eastwesser/billing:latest .
docker-compose -f deployments/docker-compose.cluster.yml up -d billing
```
