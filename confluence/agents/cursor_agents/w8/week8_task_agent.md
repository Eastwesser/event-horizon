# ============================================================
# Cursor Rules — Event Horizon Project
# Для курса "Микросервисы, как в BigTech 2.0" (Олег Козырев)
# ============================================================

# ============================================================
# 1. ОПИСАНИЕ ПРОЕКТА
# ============================================================

project:
  name: "Event Horizon"
  description: "Платформа для онлайн-игр с микросервисной архитектурой"
  type: "microservices"
  course: "Микросервисы, как в BigTech 2.0"
  author: "Oleg Kozyrev (olezhek28)"
  student: "Eastwesser"

# ============================================================
# 2. ТЕКУЩАЯ НЕДЕЛЯ (WEEK 8 — OBSERVABILITY)
# ============================================================

current_week: 8
focus: "observability"
tasks:
  - "structured_logging"      # Замена на Zap
  - "elk_stack"               # Elasticsearch + Logstash + Kibana
  - "business_metrics"        # Prometheus метрики
  - "telegram_alerts"         # Alertmanager + Telegram
  - "distributed_tracing"     # OpenTelemetry + Jaeger

# ============================================================
# 3. СТРУКТУРА ПРОЕКТА
# ============================================================

structure:
  root: "/home/denismatveev/event_horizon"
  
  services:
    path: "services/"
    list:
      - auth
      - billing
      - game
      - gateway
      - inventory
      - leaderboard
      - profile
      - shop
      - balancer
      - nats-hub
  
  platform:
    path: "pkg/"
    components:
      - logger
      - metrics
      - tracing
      - closer
      - redisclient
  
  deployments:
    path: "deployments/"
    compose: "deployments/docker-compose.cluster.yml"
    prometheus: "deployments/prometheus/prometheus.yml"
    grafana: "deployments/grafana/"
    k6: "deployments/k6/"
  
  delivery:
    path: "delivery/"
    ansible: "delivery/ansible/"
    cicd: "delivery/ci-cd/.github/workflows/"

# ============================================================
# 4. ТЕХНИЧЕСКИЙ СТЕК
# ============================================================

tech_stack:
  language: "Go 1.24"
  
  frameworks:
    grpc: "google.golang.org/grpc"
    http: "chi"
    websocket: "gorilla/websocket"
  
  databases:
    postgres: "postgres:16-alpine"
    redis: "redis:7-alpine"
    mongo: "mongo:6"  # только для inventory
  
  message_broker:
    nats: "nats:2.10-alpine"
  
  observability:
    logging: "uber-go/zap"
    metrics: "prometheus/client_golang"
    tracing: "go.opentelemetry.io/otel"
    jaeger: "jaegertracing/all-in-one"
    grafana: "grafana/grafana"
    elk: "elasticsearch + logstash + kibana"
  
  testing:
    load: "k6"
    unit: "testing + testify"
  
  deployment:
    docker: "Docker Compose"
    kubernetes: "k3s"
    cicd: "GitHub Actions + Ansible"

# ============================================================
# 5. ПАТТЕРНЫ ПРОЕКТИРОВАНИЯ
# ============================================================

patterns:
  service_layer:
    - cmd/main.go       # Точка входа
    - internal/config/  # Конфигурация
    - internal/handler/ # gRPC/HTTP handlers
    - internal/service/ # Бизнес-логика
    - internal/repository/ # Работа с БД
    - internal/model/   # DTO/Entities
    - proto/            # .proto файлы
  
  dependency_injection:
    style: "wire"  # или ручной DI через конструкторы
    pattern: "New{Component}(cfg Config, deps ...) *Component"
  
  error_handling:
    style: "wrapped errors"
    package: "github.com/pkg/errors"
    pattern: "errors.Wrap(err, \"context\")"

# ============================================================
# 6. КОД-СТАНДАРТЫ (из .golangci.yml)
# ============================================================

code_standards:
  lint_config: ".golangci.yml"
  max_complexity: 20
  max_same_issues: 0
  uniq_by_line: true
  
  enabled_linters:
    - errcheck
    - staticcheck
    - govet
    - gocritic
    - revive
    - unused
    - gosec
    - depguard
    - bodyclose
    - cyclop
    - dupl
    - ineffassign
    - unparam
    - errorlint
  
  disabled_linters:
    - gocyclo  # заменён на cyclop
    - lll      # шумный
  
  formatting:
    - gofumpt
    - gci
    - imports_order: "standard, default, prefix(github.com/Eastwesser/event-horizon)"

# ============================================================
# 7. ДОМАШНЕЕ ЗАДАНИЕ (WEEK 8)
# ============================================================

homework_week8:
  source: "/home/denismatveev/event_horizon/kozirev_code/microservices-course-homework-main/homeworks/week8/hw.md"
  
  requirements:
    structured_logging:
      status: "todo"
      tool: "zap"
      package: "pkg/logger"
      action: "Заменить все log.Println на logger.Info/Error"
    
    elk_stack:
      status: "todo"
      components: ["elasticsearch", "logstash", "kibana"]
      action: "Добавить в docker-compose.cluster.yml"
    
    business_metrics:
      status: "todo"
      metrics:
        - name: "orders_total"
          type: "counter"
          service: "shop"
        - name: "orders_revenue_total"
          type: "counter"
          service: "shop"
        - name: "assembly_duration_seconds"
          type: "histogram"
          service: "inventory"
      action: "Добавить в pkg/metrics"
    
    alerts:
      status: "todo"
      tool: "Alertmanager + Telegram"
      condition: "rate(orders_total[1m]) > 10"
      action: "Настроить алерты"
    
    tracing:
      status: "in_progress"
      tool: "OpenTelemetry + Jaeger"
      package: "pkg/tracing"
      action: "Интегрировать во все сервисы"

# ============================================================
# 8. ССЫЛКИ НА ПРИМЕРЫ КОДА (Козырев)
# ============================================================

examples:
  week7_tracing: "/home/denismatveev/event_horizon/kozirev_code/microservices-course-examples-main/week_7/tracing/"
  week7_metrics: "/home/denismatveev/event_horizon/kozirev_code/microservices-course-examples-main/week_7/metrics/"
  boilerplate: "/home/denismatveev/event_horizon/kozirev_code/microservices-course-boilerplate-main/"

# ============================================================
# 9. ПОЛЕЗНЫЕ КОМАНДЫ
# ============================================================

commands:
  deploy: "make deploy"
  restart: "make restart"
  status: "make status"
  logs: "make logs"
  migrate: "make migrate-all"
  test: "make test-all"
  build: "make docker-build-all"
  push: "make docker-push-all"
  
  k6_test: "cd deployments/k6 && k6 run e2e-test.js"
  k6_load: "cd deployments/k6 && k6 run loadtest.js"
  
  jaeger: "http://localhost:16686"
  grafana: "http://localhost:3000 (admin/admin)"
  kibana: "http://localhost:5601"

# ============================================================
# 10. ИНСТРУКЦИИ ДЛЯ CURSOR AI
# ============================================================

ai_instructions:
  - "Всегда учитывай контекст WEEK 8 — observability"
  - "При ответе ссылайся на примеры Козырева из kozirev_code/"
  - "Предлагай код, который соответствует структуре проекта"
  - "Используй pkg/ для платформенных компонентов"
  - "Для новых фич создавай отдельные пакеты в pkg/"
  - "В сервисах используй internal/ для бизнес-логики"
  - "Все логи должны быть структурированными (Zap)"
  - "Метрики регистрируй через promauto"
  - "Трейсы инициализируй через OpenTelemetry"
  - "Конфигурацию читай из переменных окружения"
  - "Для ошибок используй errors.Wrap с контекстом"
  - "Тесты пиши в *_test.go рядом с кодом"

# ============================================================
# 11. ENV ПЕРЕМЕННЫЕ
# ============================================================

env_vars:
  required:
    - JAEGER_ENDPOINT: "jaeger:4317"
    - METRICS_PORT: "9091-9099"
    - NATS_URL: "nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222"
    - DB_HOST: "postgres"
    - DB_PORT: "5432"
    - DB_USER: "eventhorizon"
    - DB_PASSWORD: "eventhorizon"
    - DB_NAME: "eventhorizon"
    - REDIS_ADDR: "redis:6379"
  
  optional:
    - LOG_LEVEL: "info"  # debug, info, warn, error
    - TELEGRAM_BOT_TOKEN: "your_token"
    - TELEGRAM_CHAT_ID: "your_chat_id"

# ============================================================
# 12. МИКРОСЕРВИСЫ И ИХ ПОРТЫ
# ============================================================

service_ports:
  auth:
    grpc: 50051
    metrics: 9091
  game:
    grpc: 50052
    metrics: 9092
  billing:
    grpc: 50053
    metrics: 9093
  leaderboard:
    grpc: 50054
    metrics: 9094
  shop:
    grpc: 50055
    metrics: 9095
  inventory:
    grpc: 50059
    metrics: 9096
  profile:
    grpc: 50060
    metrics: 9099
  gateway:
    http: 8080
    metrics: 9095-9097
  balancer:
    http: 8079
    metrics: 9098
  nats-hub:
    metrics: 9097

# ============================================================
# 13. МИГРАЦИИ
# ============================================================

migrations:
  tool: "goose"
  path: "services/{service}/migrations/"
  command: "goose -dir migrations postgres \"postgres://user:pass@host:port/db?sslmode=disable\" up"

# ============================================================
# 14. ОТВЕТЫ НА ЧАСТЫЕ ВОПРОСЫ
# ============================================================

faq:
  - question: "Где лежат примеры Козырева?"
    answer: "/home/denismatveev/event_horizon/kozirev_code/microservices-course-examples-main/"
  
  - question: "Как добавить новый сервис?"
    answer: "Создать папку в services/, добавить cmd/main.go, internal/, proto/, go.mod"
  
  - question: "Как поднять всё окружение?"
    answer: "make deploy"
  
  - question: "Где смотреть логи?"
    answer: "make logs или docker logs event-horizon-{service}"
  
  - question: "Как запустить нагрузочный тест?"
    answer: "cd deployments/k6 && k6 run loadtest.js"

# ============================================================
# 15. КОНТАКТЫ
# ============================================================

contacts:
  course_author: "Олег Козырев (olezhek28)"
  telegram: "http://t.me/olezhek28go"
  course_url: "https://olezhek28.courses/microservices"
  student: "Eastwesser"
  github: "https://github.com/Eastwesser/event-horizon"

# ============================================================
# КОНЕЦ ФАЙЛА
# ============================================================
🚀 Как использовать
Сохрани файл как .cursorrules в корне проекта

Перезапусти Cursor IDE (или закрой и открой проект)

Теперь при общении с Cursor AI он будет:

Знать структуру твоего проекта

Понимать, что ты на 8-й неделе

Давать советы в контексте курса Козырева

Ссылаться на примеры из kozirev_code/

Использовать правильные пакеты и паттерны

🔧 Дополнительно: Cursor Settings
Можешь также добавить в ~/.cursor/settings.json:

json
{
  "projectType": "go",
  "goPath": "/home/denismatveev/event_horizon",
  "linting": {
    "enabled": true,
    "config": ".golangci.yml"
  },
  "formatting": {
    "enabled": true,
    "tool": "gofumpt"
  },
  "ai": {
    "contextFiles": [
      ".cursorrules",
      "README.md",
      "Makefile",
      "deployments/docker-compose.cluster.yml"
    ]
  }
}