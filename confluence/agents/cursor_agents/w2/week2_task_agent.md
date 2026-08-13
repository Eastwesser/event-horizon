# ============================================================
# Cursor Rules for Event Horizon Project
# Основано на курсе "Микросервисы, как в BigTech 2.0" (Олег Козырев)
# ============================================================

# ============================================================
# 1. КОНТЕКСТ ПРОЕКТА
# ============================================================

context:
  project_name: "Event Horizon"
  project_type: "microservices"
  architecture: "Clean Architecture"
  tech_stack:
    - "Go 1.25"
    - "gRPC"
    - "PostgreSQL 16"
    - "Redis 7"
    - "NATS 2.10 (JetStream)"
    - "Docker Compose"
    - "k3s (Kubernetes)"
    - "Prometheus + Grafana"
    - "Jaeger (OpenTelemetry)"
    - "K6 (load testing)"
  course: "Микросервисы, как в BigTech 2.0"
  instructor: "Олег Козырев"

# ============================================================
# 2. СТРУКТУРА ПРОЕКТА
# ============================================================

project_structure:
  services:
    path: "services/"
    pattern: "services/{service_name}/"
    layers:
      - "cmd/main.go"               # Точка входа
      - "internal/config/"          # Конфигурация
      - "internal/handler/"         # gRPC хендлеры (API слой)
      - "internal/service/"         # Бизнес-логика
      - "internal/repository/"      # Работа с БД
      - "internal/model/"           # Доменные модели
      - "internal/converter/"       # Конвертация proto ↔ model
      - "internal/worker/"          # Фоновые задачи
      - "migrations/"               # SQL миграции
      - "proto/"                    # Protobuf файлы
      - "Dockerfile"                # Сборка образа
      - "go.mod"                    # Зависимости
  
  deployments:
    path: "deployments/"
    files:
      - "docker-compose.cluster.yml"  # Основной compose
      - "prometheus/prometheus.yml"   # Метрики
      - "grafana/dashboards/"         # Дашборды
      - "k3s/"                        # Kubernetes манифесты
      - "k6/"                         # Нагрузочное тестирование
  
  delivery:
    path: "delivery/"
    files:
      - "ansible/"                    # Плейбуки для деплоя
      - "ci-cd/.github/workflows/"    # GitHub Actions

# ============================================================
# 3. АРХИТЕКТУРНЫЕ ПРИНЦИПЫ (из курса Козырева)
# ============================================================

architecture_principles:
  clean_architecture:
    layers:
      - name: "API (Handler)"
        description: "gRPC хендлеры, входная точка"
        dependencies: ["Service"]
        rules:
          - "Не содержит бизнес-логики"
          - "Использует Converter для конвертации"
          - "Только валидация запросов"
      
      - name: "Service"
        description: "Бизнес-логика (Use Cases)"
        dependencies: ["Repository"]
        rules:
          - "Содержит всю бизнес-логику"
          - "Не знает о gRPC или HTTP"
          - "Работает с доменными моделями"
          - "Использует кастомные ошибки из model/"
      
      - name: "Repository"
        description: "Доступ к данным"
        dependencies: ["Model"]
        rules:
          - "Интерфейс в repository.go"
          - "Реализация в repository/{name}/"
          - "Не содержит бизнес-логики"
      
      - name: "Converter"
        description: "Конвертация между слоями"
        rules:
          - "proto ↔ model конвертация"
          - "model ↔ repository model конвертация"
      
      - name: "Model"
        description: "Доменные модели"
        rules:
          - "Чистые Go структуры"
          - "Содержат кастомные ошибки"
          - "Не зависят от внешних пакетов"

# ============================================================
# 4. СТАНДАРТЫ КОДА
# ============================================================

coding_standards:
  go:
    version: "1.25"
    formatters:
      - "gofumpt"
      - "gci"
    linters:
      - "golangci-lint (v2)"
    rules:
      - "Все ошибки должны обрабатываться"
      - "Использовать context.Context в методах"
      - "Логировать через структурированный логгер"
      - "Не использовать fmt.Print* для логирования"
      - "Не использовать http.DefaultClient"
      - "Не использовать time.Sleep в продакшене"
  
  testing:
    framework: "testify"
    mocking: "gomock"
    structure:
      - "suite_test.go для группировки тестов"
      - "create_test.go, get_test.go и т.д."
      - "go:generate для генерации моков"
    coverage:
      target: "70%"
      command: "go test -coverprofile=coverage.out ./..."
  
  proto:
    tool: "buf"
    files:
      - "buf.yaml"
      - "buf.gen.yaml"
    rules:
      - "Все сервисы должны иметь Health Check"
      - "gRPC Gateway для REST API (опционально)"

# ============================================================
# 5. ШАБЛОНЫ ДЛЯ СОЗДАНИЯ НОВЫХ СЕРВИСОВ
# ============================================================

templates:
  new_service:
    steps:
      - "Создать структуру папок: cmd/, internal/, proto/, migrations/"
      - "Создать internal/model/ с доменными моделями и errors.go"
      - "Создать internal/converter/ с конвертерами"
      - "Создать internal/repository/repository.go с интерфейсом и go:generate"
      - "Создать internal/repository/{name}/ с реализацией"
      - "Создать internal/service/service.go с интерфейсом и go:generate"
      - "Создать internal/service/{name}/ с бизнес-логикой"
      - "Создать internal/handler/grpc_handler.go с gRPC хендлерами"
      - "Создать proto/{service}.proto"
      - "Создать migrations/ с SQL файлами"
      - "Создать Dockerfile"
      - "Добавить сервис в docker-compose.cluster.yml"
      - "Добавить сервис в Makefile"

# ============================================================
# 6. КЛЮЧЕВЫЕ ПАТТЕРНЫ (из курса Козырева)
# ============================================================

patterns:
  dependency_injection:
    description: "Ручное внедрение зависимостей через конструкторы"
    example: |
      repo := repository.New()
      svc := service.New(repo)
      api := handler.New(svc)
  
  error_handling:
    description: "Кастомные ошибки в model/errors.go"
    example: |
      var (
        ErrNotFound = errors.New("entity not found")
        ErrInvalidData = errors.New("invalid data")
      )
  
  context_propagation:
    description: "Передача context.Context через все слои"
    example: |
      func (s *Service) Get(ctx context.Context, id string) (*model.Entity, error)
  
  converter_pattern:
    description: "Отдельный слой для конвертации"
    example: |
      func ProtoToModel(req *pb.Request) *model.Entity
      func ModelToProto(entity *model.Entity) *pb.Response

# ============================================================
# 7. ПОЛЕЗНЫЕ КОМАНДЫ
# ============================================================

commands:
  dev:
    - "make deploy"              # Запустить все сервисы
    - "make logs"                # Посмотреть логи
    - "make status"              # Проверить статус
    - "make migrate-all"         # Применить миграции
  
  build:
    - "make build-all"           # Собрать все сервисы
    - "make docker-build-all"    # Собрать Docker образы
    - "make docker-push-all"     # Запушить образы
  
  test:
    - "make test-all"            # Запустить все тесты
    - "go test -v ./... -cover"  # Тесты с покрытием
  
  deploy:
    - "make deploy"              # Локальный деплой
    - "make deploy-k3s"          # Деплой в k3s
    - "cd delivery/ansible && ansible-playbook -i inventory/dev.ini site.yml"  # Ansible деплой

# ============================================================
# 8. ИСТОЧНИКИ ЗНАНИЙ
# ============================================================

knowledge_sources:
  course:
    path: "/home/denismatveev/event_horizon/kozirev_code/microservices-course-examples-main/"
    weeks:
      week_1:
        topics: ["gRPC", "gRPC Gateway", "Swagger", "Interceptors", "HTTP Chi", "Ogen"]
      week_2:
        topics: ["Clean Architecture", "Unit Tests", "Mocks", "Stubs", "Testify"]
        path: "week_2/"
      week_3:
        topics: ["Docker", "Docker Compose", "PostgreSQL", "MongoDB"]
      week_4:
        topics: ["Config", "DI", "E2E Tests", "Testcontainers"]
      week_5:
        topics: ["Kafka", "Telegram Bot"]
      week_6:
        topics: ["JWT", "Redis"]
      week_7:
        topics: ["Metrics", "Tracing", "OpenTelemetry"]
  
  homework:
    path: "/home/denismatveev/event_horizon/kozirev_code/microservices-course-homework-main/homeworks/"

# ============================================================
# 9. ПРИОРИТЕТЫ УЛУЧШЕНИЙ
# ============================================================

improvement_priorities:
  high:
    - "Добавить слой Converter во все сервисы"
    - "Добавить папку model/ с кастомными ошибками во все сервисы"
    - "Написать тесты для Service layer (хотя бы для auth)"
    - "Настроить go:generate для моков"
  
  medium:
    - "Написать тесты для API layer (gRPC хендлеры)"
    - "Добавить suite_test.go для группировки тестов"
    - "Добавить Health Check во все сервисы"
  
  low:
    - "Добавить интеграционные тесты с Testcontainers"
    - "Настроить CI для автоматического запуска тестов"

# ============================================================
# 10. РОЛИ И ОТВЕТСТВЕННОСТИ
# ============================================================

roles:
  ai_assistant:
    responsibilities:
      - "Помогать с поиском информации по конкретной неделе курса"
      - "Давать примеры кода из материалов Козырева"
      - "Анализировать текущую архитектуру и предлагать улучшения"
      - "Помогать с написанием тестов и рефакторингом"
      - "Объяснять архитектурные решения"
    dont:
      - "Не писать код за пользователя без объяснений"
      - "Не предлагать изменения без обоснования"
      - "Не игнорировать существующую структуру проекта"

# ============================================================
# 11. ИНСТРУКЦИИ ДЛЯ AI
# ============================================================

ai_instructions:
  when_asking_about_code:
    - "Покажи путь к файлу (services/auth/internal/handler/grpc_handler.go)"
    - "Дай ссылку на пример из курса Козырева"
    - "Объясни, почему так лучше"
  
  when_suggesting_changes:
    - "Покажи 'как сейчас' и 'как должно быть'"
    - "Дай пошаговую инструкцию"
    - "Предупреди о возможных проблемах"
  
  when_writing_tests:
    - "Использовать testify/assert и testify/suite"
    - "Генерировать моки через go:generate"
    - "Покрывать успешные и ошибочные сценарии"
  
  when_discussing_architecture:
    - "Следовать Clean Architecture"
    - "Разделять слои: API → Service → Repository"
    - "Использовать Dependency Injection"

# ============================================================
# 12. ПРИМЕРЫ КОДА ИЗ КУРСА
# ============================================================

code_examples:
  converter:
    path: "week_2/layers/internal/converter/ufo.go"
    description: "Конвертация между proto и model"
  
  service_test:
    path: "week_2/unit_tests/6_unit_test_in_clean_arch/internal/service/ufo/create_test.go"
    description: "Тест бизнес-логики с моками"
  
  api_test:
    path: "week_2/unit_tests/6_unit_test_in_clean_arch/internal/api/ufo/v1/create_test.go"
    description: "Тест gRPC хендлера с моками"
  
  suite_test:
    path: "week_2/unit_tests/6_unit_test_in_clean_arch/internal/service/ufo/suite_test.go"
    description: "Группировка тестов с testify/suite"

# ============================================================
# 13. МЕТРИКИ И МОНИТОРИНГ
# ============================================================

metrics:
  prometheus_endpoints:
    - "auth:9091"
    - "billing:9093"
    - "game:9092"
    - "leaderboard:9094"
    - "profile:9099"
    - "shop:9095"
    - "inventory:9096"
    - "gateway:9095,9096,9097"
    - "balancer:9098"
  
  grafana_dashboards:
    path: "deployments/grafana/dashboards/event-horizon.json"

# ============================================================
# 14. ССЫЛКИ НА ВНЕШНИЕ РЕСУРСЫ
# ============================================================

references:
  - "https://github.com/olezhek28/microservices-course-olezhek-solution"
  - "https://github.com/Eastwesser/event-horizon"
  - "https://buf.build/"
  - "https://grpc.io/"
  - "https://taskfile.dev/"
🤖 Как использовать эти правила
Сохрани файл в корне проекта:

bash
nano /home/denismatveev/event_horizon/.cursorrules
Перезапусти Cursor IDE (или перезагрузи окно).

Теперь AI-ассистент будет:

Знать структуру твоего проекта

Понимать архитектурные принципы из курса Козырева

Давать релевантные примеры из материалов курса

Помогать с тестами, конвертерами и рефакторингом

Следовать стандартам кода из .golangci.yml

Примеры запросов, которые будут работать лучше:

"Помоги добавить Converter в auth сервис"

"Напиши тесты для auth_service.go"

"Как добавить новый сервис по шаблону Козырева?"

"Покажи пример из week_2 для тестирования с моками"

