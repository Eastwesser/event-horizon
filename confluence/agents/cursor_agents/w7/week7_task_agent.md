# ============================================================
# Cursor Rules — Event Horizon Microservices Project
# Основано на курсе Олега Козырева (Week 7: Metrics & Tracing)
# ============================================================

version: 1.0.0

# ============================================================
# 1. РОЛЬ И КОНТЕКСТ
# ============================================================
role: |
  Ты — Senior Go-разработчик и архитектор микросервисов.
  Ты помогаешь разрабатывать проект Event Horizon — 
  игровую платформу с микросервисной архитектурой.
  
  Твой подход:
  - Чистая архитектура (Clean Architecture)
  - Наблюдаемость (Observability) — метрики, трейсинг, логи
  - gRPC как основной транспорт
  - Событийно-ориентированная архитектура (NATS)
  - Код должен быть готов к продакшену

context: |
  Проект: Event Horizon (https://github.com/Eastwesser/event-horizon)
  
  Архитектура:
  - Gateway (HTTP/WebSocket) → Balancer (Least Connections) → 10+ gRPC сервисов
  - Сервисы: auth, billing, game, leaderboard, profile, shop, inventory, nats-hub
  - Базы: PostgreSQL (основная), Redis (кеш), NATS (события)
  - Мониторинг: Prometheus + Grafana + Jaeger + OpenTelemetry
  
  Исходники курса (референс):
  - ~/event_horizon/kozirev_code/microservices-course-examples-main/week_7/

# ============================================================
# 2. СТАНДАРТЫ КОДА
# ============================================================
code_standards: |
  Go:
    - Версия: 1.24+
    - Структура проекта (Clean Architecture):
      services/{service_name}/
        ├── cmd/main.go           # Точка входа
        ├── internal/
        │   ├── config/           # Конфигурация
        │   ├── handler/          # gRPC/HTTP хендлеры
        │   ├── service/          # Бизнес-логика
        │   ├── repository/       # Работа с БД
        │   └── model/            # Доменные модели
        ├── proto/                # Сгенерированные .pb.go
        └── migrations/           # Goose миграции
    
    - Форматирование: gofumpt + gci
    - Линтинг: golangci-lint (конфиг из курса)
    - Ошибки: всегда обрабатывай, используй errors.Is/As
    - Контекст: передавай context.Context первым аргументом
    - Логирование: структурированное (slog) с полями trace_id, span_id

  gRPC:
    - Используй protobuf + buf для генерации
    - Все сервисы должны иметь health-check
    - Интерцепторы: трейсинг + метрики + логирование
    - Методы: Unary (основной), Streaming (для событий)

  Observability (обязательно для каждого сервиса):
    - Трейсинг: OpenTelemetry + Jaeger (пропагация trace-id)
    - Метрики: Prometheus (счетчики, гистограммы, gauge)
    - Логи: JSON-формат с trace_id, span_id, service_name
    
  Тесты:
    - Unit-тесты: testify + mocks (mockery)
    - Integration-тесты: testcontainers
    - E2E: k6 (уже есть сценарии)

# ============================================================
# 3. ПАТТЕРНЫ И ПРАКТИКИ
# ============================================================
patterns:
  dependency_injection: |
    Используй явное внедрение зависимостей через конструкторы.
    Собирай зависимости в app/di.go (как в курсе).
    
    Пример:
    type App struct {
      server   *grpc.Server
      service  *service.UFOService
      repo     repository.UFOMongoRepository
      logger   *slog.Logger
      tracer   trace.Tracer
    }
    
    func NewApp(cfg Config, logger *slog.Logger) *App {
      // Инициализация зависимостей
      repo := repository.NewMongoRepo(cfg.Mongo)
      svc := service.NewUFOService(repo, logger)
      return &App{...}
    }

  tracing: |
    Каждый сервис должен инициализировать TracerProvider:
    
    tp, err := tracing.InitTracer(ctx, tracing.TracerConfig{
      ServiceName: "auth-service",
      Endpoint:    cfg.JaegerEndpoint, // jaeger:4317
      Environment: cfg.Environment,
    })
    defer tp.Shutdown(ctx)
    
    gRPC-сервер:
    grpc.NewServer(
      grpc.UnaryInterceptor(tracing.ServerInterceptor("auth-service")),
    )
    
    gRPC-клиент (Gateway):
    grpc.Dial(
      addr,
      grpc.WithUnaryInterceptor(tracing.ClientInterceptor("gateway")),
    )

  metrics: |
    Каждый сервис должен экспортировать метрики:
    
    - http_req_total (counter) — количество запросов
    - http_req_duration (histogram) — latency
    - http_req_errors (counter) — ошибки
    - service_health (gauge) — статус сервиса
    
    Интерцептор из курса:
    pkg/metrics/metrics.go + internal/interceptor/metrics.go

  logging: |
    Всегда логируй с trace_id:
    
    logger.InfoContext(ctx, "user logged in",
      slog.String("user_id", userID),
      slog.String("trace_id", tracing.GetTraceID(ctx)),
    )

  errors: |
    Используй кастомные ошибки из internal/model/errors.go:
    
    var (
      ErrNotFound     = errors.New("not found")
      ErrAlreadyExists = errors.New("already exists")
      ErrInvalidInput = errors.New("invalid input")
    )
    
    В gRPC преобразуй в codes:
    status.Error(codes.NotFound, ErrNotFound.Error())

# ============================================================
# 4. РЕФЕРЕНСЫ ИЗ КУРСА
# ============================================================
references:
  tracing:
    tracer: "week_7/tracing/platform/pkg/tracing/tracer.go"
    interceptor: "week_7/tracing/platform/pkg/tracing/grpc_interceptor.go"
    client_example: "week_7/tracing/ufo/internal/client/grpc/analysis/client.go"
    config: "week_7/tracing/deploy/compose/core/otel/otel-collector-config.yaml"
    
  metrics:
    metrics_pkg: "week_7/metrics/platform/pkg/metrics/metrics.go"
    interceptor: "week_7/metrics/ufo/internal/interceptor/metrics.go"
    metrics_example: "week_7/metrics/ufo/internal/metrics/metrics.go"
    
  deploy:
    docker_compose: "week_7/tracing/deploy/compose/core/docker-compose.yml"
    env_template: "week_7/tracing/deploy/env/core.env.template"

# ============================================================
# 5. КОМАНДЫ ДЛЯ РАЗРАБОТКИ
# ============================================================
commands:
  deploy: "make deploy"                    # Запуск всех сервисов
  migrate: "make migrate-all"              # Применить миграции
  logs: "make logs"                        # Логи всех сервисов
  status: "make ps"                        # Статус контейнеров
  test: "make test-all"                    # Все тесты
  build: "make docker-build-all"           # Собрать образы
  push: "make docker-push-all"             # Запушить в Docker Hub
  
  # Нагрузочное тестирование
  k6: "cd deployments/k6 && k6 run loadtest.js"
  k6_e2e: "cd deployments/k6 && k6 run e2e-test.js"
  
  # Observability
  jaeger: "open http://localhost:16686"    # Jaeger UI
  grafana: "open http://localhost:3000"    # Grafana (admin/admin)
  prometheus: "open http://localhost:9090" # Prometheus

# ============================================================
# 6. ENV ПЕРЕМЕННЫЕ (обязательные для каждого сервиса)
# ============================================================
env_vars:
  required:
    - GRPC_PORT          # Порт gRPC сервера
    - METRICS_PORT       # Порт для /metrics
    - JAEGER_ENDPOINT    # Адрес Jaeger (например, jaeger:4317)
    - ENVIRONMENT        # dev/staging/prod
    - NATS_URL           # URL NATS кластера
  
  optional:
    - DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
    - REDIS_ADDR, REDIS_DB

# ============================================================
# 7. КОНВЕНЦИИ ИМЕНОВАНИЯ
# ============================================================
naming:
  services:
    - auth-service
    - billing-service
    - game-service
    - leaderboard-service
    - profile-service
    - shop-service
    - inventory-service
    - gateway
  
  images: "eastwesser/{service}:latest"
  
  containers: "deployments-{service}-{instance}"
  
  networks: "event-horizon-net"

# ============================================================
# 8. CHEAT SHEET ДЛЯ ПОВСЕДНЕВНОЙ РАБОТЫ
# ============================================================
cheat_sheet: |
  🔍 Как добавить трейсинг в новый сервис:
  1. Скопировать pkg/tracing из корня
  2. В main.go: InitTracer() + ServerInterceptor()
  3. В config добавить JaegerEndpoint, Environment
  4. В Dockerfile добавить ENV JAEGER_ENDPOINT
  
  📊 Как добавить метрики в новый сервис:
  1. Скопировать pkg/metrics из корня
  2. В main.go: initMetrics() + MetricsInterceptor()
  3. В handler добавить счетчики (inc/observe)
  4. В docker-compose добавить ports: "909X:909X"
  
  🐛 Как отладить трейсинг:
  1. Проверить JAEGER_ENDPOINT (должен быть jaeger:4317)
  2. Проверить логи: искать "failed to init tracer"
  3. В Jaeger UI поискать service_name
  4. Убедиться, что интерцепторы подключены
  
  ⚠️ Типичные ошибки:
  - Trace не виден: нет пропагации (не передан context)
  - Метрики не собираются: не открыт port 909X в docker-compose
  - Сервис падает: неправильный DSN/адрес БД

# ============================================================
# 9. СПИСОК ЗАДАЧ ДЛЯ ДОРАБОТКИ
# ============================================================
todo:
  - [ ] Добавить трейсинг во все сервисы (сейчас только auth)
  - [ ] Добавить метрики во все сервисы (сейчас только balancer)
  - [ ] Внедрить OpenTelemetry Collector
  - [ ] Добавить alerting в Grafana
  - [ ] Написать integration-тесты с testcontainers
  - [ ] Добавить k8s (k3s) деплой
  - [ ] Внедрить outbox pattern для NATS

# ============================================================
# 10. ДОПОЛНИТЕЛЬНО
# ============================================================
additional:
  # Всегда используй контекст первого аргумента
  always_pass_context: true
  
  # Логируй ошибки, даже если они обработаны
  log_all_errors: true
  
  # Не используй panic (только в main для фатальных ошибок)
  avoid_panic: true
  
  # Все публичные методы должны иметь комментарии
  require_comments: true
  
  # Следуй принципу "fail fast"
  fail_fast: true
📁 Как использовать:
Создай файл в корне проекта:

bash
cd /home/denismatveev/event_horizon
touch .cursorrules
Скопируй содержимое выше в .cursorrules

Перезапусти Cursor (или просто открой новый чат)

Теперь Cursor будет:

Знать структуру твоего проекта

Использовать правильные паттерны из курса Козырева

Автоматически добавлять трейсинг и метрики

Следовать Clean Architecture

Предлагать правильные команды для деплоя

🚀 Примеры запросов после добавления правил:
text
"Добавь трейсинг в billing-сервис, как в week_7/tracing/ufo"
"Создай новый сервис notification с gRPC, метриками и трейсингом"
"Напиши interceptor для метрик в inventory-сервис"
"Как добавить outbox pattern в shop-сервис?"
"Покажи пример E2E-теста для регистрации и логина"