# ============================================================
# CURSOR RULES — EVENT HORIZON
# Проект: Микросервисная платформа для игр
# Основан на курсе Олега Козырева (Week 6: JWT + Redis)
# ============================================================

# ============================================================
# 1. КОНТЕКСТ ПРОЕКТА
# ============================================================

context:
  project_name: "Event Horizon"
  description: "Микросервисная платформа для игр с аутентификацией, биллингом, лидербордами и инвентарем"
  architecture: "Clean Architecture + gRPC + Redis + PostgreSQL + NATS"
  
  services:
    - auth: "Аутентификация и JWT"
    - billing: "Внутриигровая валюта"
    - game: "Игровая логика"
    - leaderboard: "Топ игроков (Redis Sorted Set)"
    - profile: "Агрегированный профиль (CQRS)"
    - shop: "Магазин товаров"
    - inventory: "Инвентарь с Outbox паттерном"
    - gateway: "API Gateway + WebSocket"
    - balancer: "Балансировщик (Least Connections)"
    - nats-hub: "Инициализация NATS Stream'ов"

  technologies:
    - "Go 1.25"
    - "gRPC + Protocol Buffers"
    - "Redis (для кеша и лидербордов)"
    - "PostgreSQL (основная БД)"
    - "NATS (событийная шина)"
    - "Docker + Docker Compose"
    - "k3s (Kubernetes)"
    - "Prometheus + Grafana + Jaeger"
    - "Ansible (деплой)"
    - "k6 (нагрузочное тестирование)"

# ============================================================
# 2. АРХИТЕКТУРНЫЕ ПРИНЦИПЫ (Clean Architecture)
# ============================================================

architecture:
  layers:
    - api: "gRPC/HTTP хендлеры (входная точка)"
    - service: "Бизнес-логика (Use Cases)"
    - repository: "Работа с БД (интерфейсы)"
    - model: "Domain модели"

  rules:
    - "service НЕ должен знать про БД (использует repository интерфейс)"
    - "handler НЕ должен знать про БД (использует service)"
    - "repository НЕ должен знать про бизнес-логику"
    - "model — plain structs без зависимостей"
    - "все зависимости инжектятся через конструкторы"

# ============================================================
# 3. СТРУКТУРА СЕРВИСА (паттерн для всех сервисов)
# ============================================================

service_structure:
  template: |
    services/{service_name}/
    ├── cmd/
    │   └── main.go
    ├── internal/
    │   ├── api/
    │   │   └── {service_name}_handler.go   # gRPC хендлер
    │   ├── service/
    │   │   └── {service_name}_service.go   # Бизнес-логика
    │   ├── repository/
    │   │   ├── repository.go               # Интерфейс
    │   │   └── postgres_repo.go            # Реализация
    │   ├── model/
    │   │   └── {entity}.go                 # Domain модели
    │   └── config/
    │       └── config.go
    ├── proto/
    │   └── {service}.proto
    ├── migrations/
    │   └── *.sql
    └── go.mod

# ============================================================
# 4. JWT АВТОРИЗАЦИЯ (из курса Week 6)
# ============================================================

jwt:
  structure:
    auth/internal/
    ├── jwt/
    │   ├── generator.go       # Генерация Access/Refresh токенов
    │   └── validator.go       # Валидация токенов
    ├── service/
    │   └── auth_service.go    # Только бизнес-логика
    └── repository/
        ├── user_repo.go       # PostgreSQL
        └── refresh_store.go   # Redis (кэш сессий)

  rules:
    - "Refresh токены хранятся в Redis (НЕ в PostgreSQL)"
    - "Access токен живет 15 минут, Refresh — 7 дней"
    - "При logout — удалять Refresh токен из Redis"
    - "При смене пароля — удалять все Refresh токены"
    - "JWT секрет — через переменную окружения JWT_SECRET"

  example_generator: |
    // internal/jwt/generator.go
    type Generator interface {
        GenerateAccessToken(userID, nickname string) (string, error)
        GenerateRefreshToken(userID string) (string, error)
    }

    type Claims struct {
        UserID   string `json:"user_id"`
        Nickname string `json:"nickname"`
        jwt.RegisteredClaims
    }

  example_validator: |
    // internal/jwt/validator.go
    type Validator interface {
        ValidateAccessToken(token string) (*Claims, error)
        ValidateRefreshToken(token string) (*Claims, error)
    }

  example_refresh_store: |
    // internal/repository/refresh_store.go
    type RefreshStore interface {
        Save(ctx context.Context, userID, token string, ttl time.Duration) error
        Get(ctx context.Context, userID string) (string, error)
        Delete(ctx context.Context, userID string) error
        DeleteAll(ctx context.Context, userID string) error // при смене пароля
    }

# ============================================================
# 5. REDIS КЕШИРОВАНИЕ (из курса Week 6)
# ============================================================

redis:
  structure:
    pkg/redisclient/           # Единый клиент
    ├── client.go              # Интерфейс Cache
    └── redis_client.go        # Реализация Redis

    services/inventory/internal/repository/
    ├── inventory/             # Основной репозиторий (Postgres/Mongo)
    │   └── repository.go
    └── inventory_cache/       # Кеширующий декоратор
        └── repository.go      # Проверяет кеш, потом БД

  rules:
    - "Единый Redis клиент в pkg/redisclient/ (переиспользование)"
    - "Cache Decorator паттерн: сначала кеш, потом БД"
    - "TTL для кеша — 5 минут (настраивается через .env)"
    - "При Create/Update/Delete — инвалидировать кеш"
    - "Redis используется для: кеша, лидербордов (Sorted Set), сессий"

  example_cache_decorator: |
    // internal/repository/inventory_cache/repository.go
    type CacheRepository struct {
        repo  Repository
        cache *redisclient.Client
        ttl   time.Duration
    }

    func (r *CacheRepository) Get(ctx context.Context, id string) (*model.Item, error) {
        // 1. Проверяем кеш
        key := "inventory:" + id
        cached, err := r.cache.Get(ctx, key)
        if err == nil {
            return cached, nil
        }

        // 2. Промах — идем в БД
        item, err := r.repo.Get(ctx, id)
        if err != nil {
            return nil, err
        }

        // 3. Сохраняем в кеш
        r.cache.SetWithTTL(ctx, key, item, r.ttl)
        return item, nil
    }

  example_redis_client: |
    // pkg/redisclient/client.go
    type Cache interface {
        Get(ctx context.Context, key string) (string, error)
        Set(ctx context.Context, key string, value interface{}) error
        SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error
        Delete(ctx context.Context, key string) error
    }

# ============================================================
# 6. GRPC INTERCEPTORS (из курса Week 6)
# ============================================================

interceptors:
  auth:
    description: "Проверка JWT на уровне gRPC"
    location: "internal/interceptor/auth.go"
    rules:
      - "Извлекать токен из metadata (Authorization: Bearer <token>)"
      - "Валидировать токен через jwt.Validator"
      - "Добавлять user_id и nickname в context"
      - "Пропускать health-запросы без авторизации"

  logging:
    description: "Логирование всех gRPC запросов"
    location: "internal/interceptor/logger.go"
    rules:
      - "Логировать метод, длительность, статус"
      - "Логировать user_id из контекста (если есть)"
      - "Не логировать чувствительные данные (пароли)"

# ============================================================
# 7. КОНФИГУРАЦИЯ (через .env)
# ============================================================

config:
  template: |
    # .env файл
    # PostgreSQL
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=eventhorizon
    DB_PASSWORD=eventhorizon
    DB_NAME=eventhorizon

    # Redis
    REDIS_ADDR=localhost:6379
    REDIS_PASSWORD=
    REDIS_DB=0
    CACHE_TTL=5m

    # JWT
    JWT_SECRET=your-secret-key
    ACCESS_TOKEN_TTL=15m
    REFRESH_TOKEN_TTL=168h

    # NATS
    NATS_URL=nats://localhost:4222

    # Jaeger (Tracing)
    JAEGER_ENDPOINT=localhost:4317

  rules:
    - "Все конфиги через переменные окружения"
    - "Никаких хардкодных значений"
    - "В деплое использовать .env файлы из deployments/configs/"
    - "В k3s использовать ConfigMap или Secrets"

# ============================================================
# 8. DEPLOYMENT (Ansible + k3s)
# ============================================================

deployment:
  ansible:
    location: "delivery/ansible/"
    rules:
      - "Поддерживает Arch и Ubuntu"
      - "Устанавливает Docker + Docker Compose"
      - "Клонирует репозиторий (или обновляет)"
      - "Создает .env файл из шаблона"
      - "Запускает docker-compose up -d"
      - "Проверяет health-эндпоинты"

  k3s:
    location: "deployments/k3s/"
    rules:
      - "Для продакшн-деплоя"
      - "Deployment + Service + Ingress"
      - "Каждый сервис — отдельный контейнер в поде"
      - "В будущем — StatefulSet для БД"

  docker_compose:
    location: "deployments/docker-compose.cluster.yml"
    rules:
      - "Кластерная версия (все сервисы в одной сети)"
      - "6 Redis инстансов (по одному на сервис)"
      - "8 PostgreSQL инстансов"
      - "NATS кластер из 3 нод"
      - "Prometheus + Grafana + Jaeger"

# ============================================================
# 9. ТЕСТИРОВАНИЕ (k6)
# ============================================================

testing:
  k6:
    location: "deployments/k6/"
    files:
      - "e2e-test.js"    # Сквозной сценарий
      - "loadtest.js"    # Нагрузочное тестирование

  rules:
    - "E2E тест: регистрация → логин → игра → лидерборд"
    - "Нагрузочный тест: постепенное увеличение нагрузки"
    - "Проверять: p95 < 500ms, ошибки < 1%"
    - "Использовать SharedArray для тестовых пользователей"

# ============================================================
# 10. КОД-СТАЙЛ (golangci-lint)
# ============================================================

code_style:
  linters:
    - "golangci-lint v2.x"
    - "gofumpt (форматирование)"
    - "gci (сортировка импортов)"

  rules:
    - "Использовать gofumpt -extra"
    - "Группировка импортов: standard → external → internal"
    - "Максимальная сложность функции: 20 (cyclop)"
    - "Запрещено: fmt.Print*, time.Sleep, http.DefaultClient"
    - "Все ошибки должны обрабатываться (errcheck)"

  format_command: |
    task format
    # или
    gofumpt -extra -w .
    gci write -s standard -s default -s "prefix(github.com/Eastwesser/event-horizon)" .

# ============================================================
# 11. МОНИТОРИНГ (Prometheus + Grafana)
# ============================================================

monitoring:
  prometheus:
    location: "deployments/prometheus/prometheus.yml"
    metrics:
      - "Все сервисы имеют /metrics на порту 909X"
      - "gRPC метрики (запросы, ошибки, latency)"
      - "Go runtime метрики (goroutines, heap, GC)"

  grafana:
    location: "deployments/grafana/dashboards/"
    dashboards:
      - "RPS Gateway"
      - "Latency P99"
      - "Go Goroutines"
      - "Heap Memory"
      - "NATS Connections"
      - "Redis Memory"
      - "PostgreSQL Connections"

  jaeger:
    location: "deployments/docker-compose.cluster.yml"
    config:
      - "OTLP HTTP: 4318"
      - "OTLP gRPC: 4317"
      - "UI: 16686"

# ============================================================
# 12. БЕЗОПАСНОСТЬ
# ============================================================

security:
  jwt:
    - "JWT_SECRET — минимум 32 символа"
    - "Access токен — 15 минут"
    - "Refresh токен — 7 дней (в Redis)"
    - "При logout — удалять Refresh из Redis"

  passwords:
    - "Пароли хешировать через bcrypt"
    - "Минимальная длина пароля: 6 символов"

  environment:
    - "Никаких секретов в коде"
    - "Все секреты через .env"
    - "В CI/CD использовать GitHub Secrets"

# ============================================================
# 13. NATS (Событийная шина)
# ============================================================

nats:
  setup:
    - "NATS кластер из 3 нод (отказоустойчивость)"
    - "NATS Hub — создает Stream'ы при старте"
    - "JetStream включен (-js флаг)"

  streams:
    - "game_scores — для рекордов"
    - "billing_events — для транзакций"
    - "inventory_events — для Outbox паттерна"
    - "user_events — для профиля (CQRS)"

  rules:
    - "Использовать NATS для асинхронных событий"
    - "Outbox паттерн для гарантированной доставки"
    - "Consumer группы для обработки событий"

# ============================================================
# 14. ПОЛЕЗНЫЕ КОМАНДЫ
# ============================================================

commands:
  deploy:
    - "make deploy"                    # Запуск всего проекта
    - "make migrate-all"               # Миграции всех БД
    - "make restart"                   # Перезапуск

  build:
    - "make docker-build-all"          # Сборка всех образов
    - "make docker-push-all"           # Пуш на Docker Hub

  testing:
    - "cd deployments/k6 && k6 run e2e-test.js"
    - "cd deployments/k6 && k6 run loadtest.js"

  monitoring:
    - "curl http://localhost:8079/health"
    - "docker ps --format 'table {{.Names}}\t{{.Status}}'"
    - "kubectl get pods"

  ansible:
    - "cd delivery/ansible && ansible-playbook -i inventory/dev.ini site.yml"

# ============================================================
# 15. TODO ПО WEEK 6
# ============================================================

todo_week6:
  priority_high:
    - "✅ Перенести Refresh токены из PostgreSQL в Redis"
    - "✅ Создать единый Redis клиент в pkg/redisclient/"
    - "✅ Разделить JWT логику (generator + validator)"

  priority_medium:
    - "✅ Добавить gRPC Auth Interceptor"
    - "✅ Внедрить Cache Decorator паттерн для Inventory"
    - "✅ Добавить graceful shutdown во все сервисы"

  priority_low:
    - "✅ Унифицировать конфигурацию через .env"
    - "✅ Добавить Swagger/OpenAPI документацию"
    - "✅ Настроить автоматическое масштабирование (HPA) в k3s"

# ============================================================
# 16. ССЫЛКИ НА КУРС
# ============================================================

course_references:
  week_6:
    jwt:
      - "kozirev_code/microservices-course-examples-main/week_6/jwt/"
    redis:
      - "kozirev_code/microservices-course-examples-main/week_6/redis/clean_arch/"
    homework:
      - "kozirev_code/microservices-course-homework-main/homeworks/week6/"

# ============================================================
# 17. ПРИМЕРЫ КОДА (шаблоны)
# ============================================================

templates:
  new_service: |
    services/{service_name}/
    ├── cmd/
    │   └── main.go
    ├── internal/
    │   ├── api/
    │   │   └── grpc_handler.go
    │   ├── service/
    │   │   └── service.go
    │   ├── repository/
    │   │   ├── repository.go
    │   │   └── postgres_repo.go
    │   ├── model/
    │   │   └── model.go
    │   └── config/
    │       └── config.go
    ├── proto/
    │   └── {service}.proto
    └── go.mod

  grpc_handler: |
    type GRPCHandler struct {
        pb.Unimplemented{Service}Server
        service *service.{Service}Service
    }

    func (h *GRPCHandler) Method(ctx context.Context, req *pb.Request) (*pb.Response, error) {
        // Валидация
        // Вызов service
        // Преобразование ответа
        return &pb.Response{}, nil
    }

  service_with_cache: |
    type {Service}Service struct {
        repo    repository.{Service}Repository
        cache   *redisclient.Client
        logger  logger.Logger
    }

    func (s *{Service}Service) Get(ctx context.Context, id string) (*model.Entity, error) {
        // Сначала кеш
        key := "{service}:" + id
        cached, err := s.cache.Get(ctx, key)
        if err == nil {
            return cached, nil
        }

        // Потом БД
        entity, err := s.repo.Get(ctx, id)
        if err != nil {
            return nil, err
        }

        // Сохраняем в кеш
        s.cache.SetWithTTL(ctx, key, entity, 5*time.Minute)
        return entity, nil
    }

# ============================================================
# 18. ИНСТРУКЦИИ ДЛЯ CURSOR
# ============================================================

cursor_instructions:
  - "При ответе на вопросы по архитектуре — ссылайся на Clean Architecture"
  - "При генерации кода — используй шаблоны из этого файла"
  - "При упоминании JWT — всегда напоминай про Redis для Refresh токенов"
  - "При упоминании Redis — всегда предлагай Cache Decorator паттерн"
  - "При упоминании деплоя — предлагай Ansible + k3s"
  - "При упоминании тестирования — предлагай k6"
  - "При упоминании мониторинга — предлагай Prometheus + Grafana + Jaeger"
  - "Всегда проверяй код на соответствие golangci-lint правилам"
  - "Никогда не предлагай хардкод — всегда через .env"
  - "При создании нового сервиса — используй структуру из раздела 3"

# ============================================================
# 19. ВЕРСИИ ИНСТРУМЕНТОВ
# ============================================================

versions:
  go: "1.25"
  golangci_lint: "v2.1.5"
  gci: "v0.13.6"
  gofumpt: "v0.8.0"
  docker: "latest"
  k3s: "latest"
  ansible: "latest"

# ============================================================
# 20. КОНТАКТЫ
# ============================================================

contacts:
  github: "https://github.com/Eastwesser/event-horizon"
  course: "https://olezhek28.courses/microservices"
  author: "Eastwesser"

# ============================================================
# КОНЕЦ ФАЙЛА
# ============================================================
📁 Дополнительные файлы для Cursor
1. .cursorrules — основной файл (содержимое выше)
2. .cursor/agents/event-horizon.md — агент для проекта
Создай папку .cursor/agents/ и файл:

markdown
# Event Horizon Agent

## Роль
Ты — архитектор микросервисов, специализирующийся на Go, gRPC, Redis и Clean Architecture. Твоя задача — помогать Eastwesser разрабатывать проект Event Horizon.

## Контекст
Проект основан на курсе Олега Козырева (Микросервисы, как в BigTech 2.0). Мы проходим Week 6 (JWT + Redis).

## Правила
1. Всегда используй Clean Architecture
2. Никогда не предлагай хардкод
3. Всегда проверяй код на ошибки
4. Предлагай тесты для любого нового кода
5. При работе с JWT — всегда упоминай Redis для Refresh токенов
6. При работе с Redis — предлагай Cache Decorator паттерн
7. При деплое — предлагай Ansible + k3s
8. При тестировании — предлагай k6
9. При мониторинге — предлагай Prometheus + Grafana + Jaeger

## Приоритеты
1. Безопасность
2. Производительность
3. Чистая архитектура
4. Тестируемость
5. Масштабируемость

## Вопросы к Eastwesser
1. Какой сервис мы улучшаем?
2. Какая задача по Week 6?
3. Нужен ли код для этого?
4. Есть ли тесты?
5. Как это будет деплоиться?
3. .cursor/commands/event-horizon.md — команды
markdown
# Event Horizon Commands

## Разработка
- `eh:new {service}` — создать новый сервис
- `eh:proto {service}` — сгенерировать proto
- `eh:test {service}` — запустить тесты
- `eh:lint` — запустить линтер

## Деплой
- `eh:deploy` — деплой через Ansible
- `eh:restart` — перезапустить все сервисы
- `eh:status` — проверить статус
- `eh:logs {service}` — посмотреть логи

## Базы данных
- `eh:migrate {service}` — применить миграции
- `eh:migrate-all` — все миграции
- `eh:redis` — проверить Redis

## Тестирование
- `eh:e2e` — запустить E2E тесты (k6)
- `eh:load` — запустить нагрузочные тесты

## Мониторинг
- `eh:metrics` — посмотреть метрики
- `eh:traces` — посмотреть трейсы (Jaeger)
- `eh:dashboard` — открыть Grafana
🚀 Как использовать
Скопируй содержимое .cursorrules в корень проекта

Создай папки .cursor/agents/ и .cursor/commands/

Добавь файлы event-horizon.md и event-horizon.md

Теперь Cursor будет:

Знать контекст проекта

Предлагать решения по Week 6

Следовать Clean Architecture

Использовать правильные паттерны

📝 Быстрый старт
bash
# Создай файл правил
cd /home/denismatveev/event_horizon
cat > .cursorrules << 'EOF'
# ... (вставь содержимое выше) ...
EOF

# Создай агента
mkdir -p .cursor/agents
cat > .cursor/agents/event-horizon.md << 'EOF'
# ... (вставь содержимое выше) ...
EOF

# Создай команды
mkdir -p .cursor/commands
cat > .cursor/commands/event-horizon.md << 'EOF'
# ... (вставь содержимое выше) ...
EOF