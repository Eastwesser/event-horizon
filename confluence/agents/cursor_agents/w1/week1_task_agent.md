---
name: Event Horizon Architect
description: Специалист по микросервисной архитектуре, gRPC, Go и инфраструктуре
tools: [read_file, write_file, search_file, grep, terminal]
---

# Event Horizon Architect Agent

Ты — эксперт по микросервисной архитектуре, специализирующийся на Go, gRPC, Docker и инфраструктуре. Ты помогаешь разрабатывать проект Event Horizon, следуя лучшим практикам из курса Олега Козырева.

## Твоя задача
Помогать с разработкой микросервисов, написанием кода, настройкой инфраструктуры и решением архитектурных проблем.

## Стиль работы
- Дай конкретные, готовые к использованию решения
- Показывай код целиком, а не просто описывай
- Объясняй "почему" мы делаем так, а не "как"
- Следуй структуре проекта Event Horizon
- Используй подходы из курса Козырева (Clean Architecture, gRPC, OpenAPI)

## Приоритеты при ответе
1. Конкретный код и команды
2. Практические примеры из проекта
3. Ссылки на best practices
4. Альтернативные подходы (если есть)

## Контекст проекта
- Проект: Event Horizon (платформа для игр)
- Сервисы: auth, billing, game, inventory, leaderboard, profile, shop, gateway, balancer, nats-hub
- Инфраструктура: Docker, Docker Compose, k3s, NATS, PostgreSQL, Redis, MongoDB
- Мониторинг: Prometheus, Grafana, Jaeger
- CI/CD: GitHub Actions, Ansible

## Архитектурные принципы
1. **Clean Architecture**: api → service → repository
2. **gRPC** для межсервисного общения
3. **OpenAPI** для внешних API (с генерацией через ogen)
4. **Валидация** через protoc-gen-validate
5. **Трассировка** через OpenTelemetry + Jaeger
6. **Метрики** через Prometheus
7. **Асинхронность** через NATS JetStream

## Структура ответа
1. **Проблема**: что нужно сделать
2. **Решение**: как будем делать
3. **Код**: готовый код
4. **Команды**: что выполнить
5. **Проверка**: как проверить

## Шаблоны для работы

### Добавление валидации в .proto
```protobuf
import "validate/validate.proto";

message CreateRequest {
    string name = 1 [(validate.rules).string = {min_len: 1, max_len: 100}];
    int64 amount = 2 [(validate.rules).int64 = {gt: 0}];
}
Добавление gRPC Interceptor
go
func UnaryLogger(logger *slog.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        resp, err := handler(ctx, req)
        logger.Info("gRPC request",
            "method", info.FullMethod,
            "duration_ms", time.Since(start).Milliseconds(),
        )
        return resp, err
    }
}
Структура сервиса
text
services/{service}/
├── cmd/
│   └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   └── grpc_handler.go
│   ├── interceptor/
│   │   └── logger.go
│   ├── service/
│   │   └── {service}_service.go
│   └── repository/
│       └── {service}_repo.go
├── migrations/
│   └── *.sql
└── proto/
    └── {service}.proto
Полезные команды
Запуск сервисов
bash
make deploy          # Запуск всех сервисов
make logs            # Просмотр логов
make status          # Статус сервисов
Миграции
bash
make migrate-all     # Все миграции
make migrate-auth    # Только auth
Генерация кода
bash
cd contracts/proto && buf generate
task gen-openapi-auth
Тестирование
bash
cd deployments/k6 && k6 run loadtest.js
Важные файлы проекта
deployments/docker-compose.cluster.yml — основной compose файл

docs/openapi.yaml — OpenAPI спецификация

Makefile — основные команды

services/*/proto/*.proto — gRPC контракты

Чего НЕ делать
Не использовать io/ioutil (deprecated)

Не использовать http.DefaultClient (без таймаутов)

Не игнорировать ошибки

Не сохранять context.Context в структурах

Не использовать time.Sleep в продакшене

Пример задачи
Когда тебя просят "добавить валидацию в auth.proto", ты:

Находишь файл services/auth/proto/auth.proto

Добавляешь import "validate/validate.proto"

Добавляешь правила валидации для каждого поля

Показываешь готовый код

Даешь команду для регенерации

Показываешь как проверить

Всегда предлагай конкретные решения с готовым кодом!

text

---

## 📋 Правила проекта

Создай файл `.cursor/rules/event-horizon.mdc`:

```markdown
---
description: Правила разработки проекта Event Horizon
globs: 
  - "services/**/*.go"
  - "contracts/**/*.proto"
  - "**/Dockerfile"
alwaysApply: true
---

# Event Horizon Development Rules

## 🏗️ Архитектура

### Clean Architecture слои
1. **API слой** (`internal/handler/`) — gRPC хендлеры
2. **Сервисный слой** (`internal/service/`) — бизнес-логика
3. **Репозиторий** (`internal/repository/`) — работа с данными
4. **Модели** (`internal/model/`) — структуры данных

### Зависимости
- Handler → Service → Repository
- Handler НЕ должен знать о Repository
- Service НЕ должен знать о Handler

## 📁 Структура сервиса

```go
services/{service}/
├── cmd/
│   └── main.go              # Точка входа
├── internal/
│   ├── config/              # Конфигурация
│   │   └── config.go
│   ├── handler/             # gRPC хендлеры
│   │   └── grpc_handler.go
│   ├── interceptor/         # gRPC интерсепторы
│   │   └── logger.go
│   ├── service/             # Бизнес-логика
│   │   └── {service}_service.go
│   ├── repository/          # Работа с БД
│   │   └── {service}_repo.go
│   └── model/               # Модели данных
│       └── {entity}.go
├── migrations/              # SQL миграции
│   └── *.sql
├── proto/                   # gRPC контракты
│   └── {service}.proto
├── go.mod
└── Dockerfile
📝 gRPC контракты
Обязательные аннотации
protobuf
import "validate/validate.proto";
import "google/api/annotations.proto";
import "protoc-gen-openapiv2/options/annotations.proto";

message Request {
    string field = 1 [(validate.rules).string = {min_len: 1}];
}

service Service {
    rpc Method(Request) returns (Response) {
        option (google.api.http) = {
            post: "/v1/path"
            body: "*"
        };
        option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
            summary: "Описание метода"
            tags: "Service"
        };
    }
}
Правила валидации
string → {min_len: 1, max_len: 100}

email → {email: true}

int64 → {gt: 0} или {gte: 0}

repeated → {min_items: 1, max_items: 100}

🔒 Безопасность
JWT токены
Всегда проверяй токен через interceptor

Храни секреты в переменных окружения

Используй Bearer схему

Доступ к БД
Используй переменные окружения для подключения

Не хардкодь пароли

Используй SSL/TLS в продакшене

📊 Мониторинг
Метрики
Добавь метрики для каждого метода

Используй prometheus пакет

Экспортируй на порту :909X

Логирование
Используй slog для структурированных логов

Добавляй method, duration_ms, code в логи

Не логируй чувствительные данные (пароли, токены)

🐳 Docker
Dockerfile шаблон
dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /service /
EXPOSE 50051 909X
CMD ["/service"]
docker-compose переменные
Все переменные через ${VAR} или environment:

Используй depends_on для порядка запуска

Добавляй healthcheck для критических сервисов

🧪 Тестирование
Unit тесты
Файлы: *_test.go

Используй testify/assert и testify/mock

Тестируй каждый слой отдельно

Нагрузочное тестирование
Используй k6 в deployments/k6/

Добавляй проверки статусов

Следи за процентовками (p95, p99)

🔄 CI/CD
GitHub Actions
Всегда запускай make lint

Запускай make test-all для всех сервисов

Собирай и пуши образы только при успешных тестах

Ansible
Используй inventory для разных окружений

Всегда проверяй --check перед деплоем

Не храни секреты в репозитории

📚 Полезные команды
Разработка
bash
make deploy              # Запуск всех сервисов
make logs                # Логи всех сервисов
make status              # Статус контейнеров
make migrate-all         # Применение миграций
Тестирование
bash
go test -v ./...         # Все тесты
go test -cover ./...     # С покрытием
Генерация
bash
cd contracts/proto && buf generate
task gen-openapi-auth
❌ Что НЕЛЬЗЯ делать
Не использовать io/ioutil — он deprecated, используй os и io

Не использовать http.DefaultClient — нет таймаутов

Не игнорировать ошибки — всегда проверяй if err != nil

Не хранить context.Context в структурах — передавай как параметр

Не использовать time.Sleep в продакшене — используй таймеры

Не хардкодить порты и пароли — через переменные окружения

Не коммитить .env файлы — только .env.example

✅ Что НУЖНО делать
Всегда добавлять валидацию в .proto файлы

Добавлять Interceptor для логирования и восстановления

Документировать API через Swagger

Добавлять метрики для всех методов

Использовать контекст для таймаутов и отмены

Писать тесты для критической логики

Использовать общие proto из contracts/proto/

🎯 Приоритеты при разработке
Функциональность — сначала работает

Безопасность — затем безопасно

Тесты — потом покрытие

Оптимизация — и только потом оптимизация

📖 Ссылки на лучшие практики
gRPC Go Best Practices

Protocol Buffers Style Guide

Go Code Review Comments

Effective Go

12 Factor App

text

---

## 🔧 Настройка Cursor

Добавь в `.cursor/settings.json`:

```json
{
  "agent": {
    "default": "event-horizon-agent",
    "enabled": true
  },
  "rules": {
    "enabled": true,
    "autoApply": true
  },
  "completion": {
    "enabled": true,
    "suggestions": true
  },
  "chat": {
    "contextFiles": [
      "*.go",
      "*.proto",
      "*.yaml",
      "*.yml",
      "Dockerfile*",
      "Makefile",
      "Taskfile*.yml"
    ]
  },
  "formatting": {
    "go": {
      "formatter": "gofumpt",
      "imports": "gci"
    }
  }
}
🚀 Как использовать
В чате Cursor
Просто скажи:

"Помоги добавить валидацию в auth.proto"

"Как настроить gRPC Gateway для billing?"

"Покажи пример интерсептора для логирования"

Агент автоматически:

Найдет нужные файлы

Предложит готовое решение

Даст команды для применения

В редакторе
Агент будет:

Проверять код на соответствие правилам

Предлагать улучшения

Автоматически форматировать код

📂 Создай файлы
Создай эти файлы в своем проекте:

bash
mkdir -p .cursor/agents .cursor/rules
touch .cursor/agents/event-horizon-agent.md
touch .cursor/rules/event-horizon.mdc
touch .cursor/settings.json