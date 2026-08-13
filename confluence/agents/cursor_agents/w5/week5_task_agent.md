---
description: Event Horizon Agent — правила для работы с проектом микросервисов на Go
globs: 
  - "**/*.go"
  - "**/Dockerfile*"
  - "**/docker-compose*.yml"
  - "**/Makefile"
  - "**/proto/*.proto"
  - "services/**/*.go"
alwaysApply: true
---

# 🚀 Event Horizon — Cursor Agent Rules

Ты — AI-ассистент для проекта **Event Horizon** — MMO-платформы на микросервисах Go.

---

## 🏗️ Архитектура проекта

### Основные компоненты

1. **Микросервисы** (`services/`):
   - `auth` — аутентификация (gRPC, JWT)
   - `billing` — баланс и транзакции (gRPC, PostgreSQL, Redis)
   - `game` — игровая логика (gRPC, PostgreSQL)
   - `leaderboard` — рейтинги (gRPC, Redis Sorted Sets)
   - `profile` — агрегированный профиль (gRPC, PostgreSQL)
   - `shop` — магазин (gRPC, PostgreSQL, Redis)
   - `inventory` — каталог товаров (gRPC, PostgreSQL/MongoDB, Redis, NATS)
   - `gateway` — API Gateway (HTTP, WebSocket, gRPC клиент)
   - `balancer` — балансировщик (Least Connections)
   - `nats-hub` — инициализация NATS Streams

2. **Инфраструктура**:
   - PostgreSQL (8 отдельных БД)
   - Redis (6 инстансов для кешей)
   - NATS Cluster (3 ноды с JetStream)
   - Kafka (для надежных транзакций) — планируется
   - Prometheus + Grafana + Jaeger
   - Ansible + GitHub Actions (CI/CD)

3. **Паттерны**:
   - Clean Architecture (internal/{config,handler,repository,service})
   - Outbox Pattern (для надежной доставки событий)
   - CQRS (Profile — read model)
   - Repository Pattern (поддержка PostgreSQL и MongoDB)

---

## 🎯 Принципы разработки

### 1. Структура сервиса (Clean Architecture)
services/{service}/
├── cmd/
│ └── main.go # Точка входа
├── internal/
│ ├── config/ # Конфигурация
│ │ └── config.go
│ ├── handler/ # gRPC/HTTP обработчики
│ │ └── grpc_handler.go
│ ├── model/ # DTO/Entity
│ │ ├── errors.go
│ │ └── {entity}.go
│ ├── repository/ # Работа с БД
│ │ ├── repository.go # Интерфейс
│ │ ├── postgres_repo.go
│ │ └── redis_repo.go # Кеш
│ └── service/ # Бизнес-логика
│ ├── service.go # Интерфейс
│ └── {service}.go
├── proto/ # gRPC контракты
│ └── {service}.proto
├── migrations/ # Goose миграции
├── Dockerfile
└── README.md

text

### 2. Паттерны Козырева (из курса)

**При написании кода используй подходы из курса Олега Козырева:**

- **Dependency Injection**: все зависимости собираются в `app/di.go` (lazy initialization)
- **Closer**: все ресурсы регистрируются для graceful shutdown
- **Middleware Chain**: для логирования, метрик, трассировки
- **Конвертеры**: отдельный слой для преобразования между слоями (`converter/`)

### 3. Брокеры сообщений

**NATS** — для игровых сессий и быстрых событий:
- Subjects: `event.>`, `score.updated`, `user.registered`, `inventory.item.*`
- Используй JetStream для persistence
- Outbox для надежности (PostgreSQL)

**Kafka** — для покупок и платежей (планируется):
- Exactly-once семантика
- Долгое хранение
- Replay

---

## 📝 Стиль кода

### 1. Go

```go
// Используй структурированное логирование (slog или zap)
logger.Info(ctx, "User created", 
    zap.String("user_id", userID),
    zap.String("email", email),
)

// Обработка ошибок
if err != nil {
    logger.Error(ctx, "Failed to create item", zap.Error(err))
    return status.Errorf(codes.Internal, "failed to create item: %v", err)
}

// Контекст везде первым аргументом
func (s *Service) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.ItemResponse, error)
2. Protobuf
protobuf
syntax = "proto3";

package {service}.v1;

option go_package = "github.com/Eastwesser/event-horizon/services/{service}/proto;{service}";

service {Service}Service {
    rpc Create(CreateRequest) returns (CreateResponse);
}

message CreateRequest {
    string id = 1;
    // ...
}
3. Миграции (Goose)
sql
-- +goose Up
CREATE TABLE IF NOT EXISTS table_name (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS table_name;
🔧 Команды
Работа с сервисами
bash
# Сборка всех сервисов
make build-all

# Запуск инфраструктуры
make deploy

# Миграции
make migrate-all

# Проверка статуса
make status
Работа с Docker
bash
# Сборка всех образов
make docker-build-all

# Пуш в Docker Hub
make docker-push-all

# Перезапуск
make restart
Работа с Ansible
bash
cd delivery/ansible
ansible-playbook -i inventory/dev.ini site.yml
📦 Добавление нового сервиса
Создать структуру services/{service}/

Написать proto/{service}.proto и сгенерировать код

Реализовать internal/

Добавить migrations/

Создать Dockerfile

Добавить сервис в docker-compose.cluster.yml

Добавить в Makefile

Обновить gateway (если нужен HTTP)

Написать README.md

🧪 Тестирование
bash
# Все тесты
make test-all

# Тесты конкретного сервиса
cd services/{service}
go test -v ./...

# Нагрузочное тестирование (k6)
cd deployments/k6
k6 run loadtest.js
🐛 Решение проблем
1. Health check
bash
curl http://localhost:9096/health
2. Логи
bash
docker-compose -f deployments/docker-compose.cluster.yml logs -f {service}
3. Перезапуск сервиса
bash
docker-compose -f deployments/docker-compose.cluster.yml restart {service}
🔗 Полезные ссылки
Курс Козырева: kozirev_code/microservices-course-examples-main/

Домашки: kozirev_code/microservices-course-homework-main/

OpenAPI: docs/openapi.yaml

Grafana: http://localhost:3000 (admin/admin)

Jaeger: http://localhost:16686

Prometheus: http://localhost:9090

📌 Твои предпочтения
Язык: Go 1.25+

Брокеры: NATS + Kafka (гибрид)

Базы: PostgreSQL (основная) + MongoDB (опционально)

Кеш: Redis

Мониторинг: Prometheus + Grafana + Jaeger

Деплой: Docker Compose + Ansible + k3s

Тестирование: K6 (нагрузочное)

🎯 Что я должен делать как AI-ассистент
Анализировать код в контексте Clean Architecture

Предлагать решения с учетом паттернов Козырева

Писать код с соблюдением стиля проекта

Объяснять архитектурные решения (почему так, а не иначе)

Помогать с интеграцией NATS и Kafka

Следовать принципам: KISS, DRY, YAGNI, SOLID

⚡ Быстрые действия
Показать архитектуру сервиса: анализирую структуру и даю рекомендации

Написать CRUD для сервиса: создаю handler, repository, service

Добавить событие в NATS: пишу producer и consumer

Написать миграцию: создаю SQL для Goose

Настроить мониторинг: добавляю метрики и health check

Написать тест: unit или e2e (K6)

