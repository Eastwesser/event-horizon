🤖 Агент для Cursor IDE: Event Horizon Development
📋 Контекст проекта
Проект: Event Horizon — микросервисная игровая платформа
Архитектура: Clean Architecture + gRPC + Event-Driven (NATS)
Инфраструктура: Docker Compose + k3s + Ansible + GitHub Actions
Базы данных: PostgreSQL (per-service) + Redis (caching)

🎯 Основные правила работы
1. Структура кода (Clean Architecture)
Каждый сервис должен следовать этой структуре:

text
services/{service_name}/
├── cmd/
│   └── main.go                    # Только запуск (тонкий)
├── internal/
│   ├── api/{service}/v1/          # gRPC/HTTP хендлеры
│   │   ├── api.go                 # Интерфейс API
│   │   ├── create.go              # Методы по одному на файл
│   │   ├── get.go
│   │   ├── update.go
│   │   └── delete.go
│   ├── app/                       # DI и сборка приложения
│   │   ├── app.go                 # Структура App
│   │   └── di.go                  # Конструктор NewApp()
│   ├── config/                    # Конфигурация
│   │   ├── config.go              # Структура Config
│   │   ├── interfaces.go          # Интерфейсы для тестов
│   │   └── env/                   # Парсеры переменных
│   │       ├── logger.go
│   │       ├── postgres.go
│   │       └── grpc.go
│   ├── converter/                 # Конвертеры между слоями
│   ├── model/                     # Доменные модели
│   │   ├── errors.go              # Кастомные ошибки
│   │   └── {entity}.go
│   ├── repository/                # Репозитории (слой БД)
│   │   ├── repository.go          # Интерфейс репозитория
│   │   ├── converter/             # Конвертеры model ↔ repo model
│   │   ├── model/                 # Модели БД
│   │   └── {storage}/             # Реализация (postgres/mongo/redis)
│   │       ├── create.go
│   │       ├── get.go
│   │       ├── update.go
│   │       └── delete.go
│   └── service/                   # Бизнес-логика
│       ├── service.go             # Интерфейс сервиса
│       └── {service}/
│           ├── create.go          # Методы по одному на файл
│           ├── get.go
│           ├── update.go
│           └── delete.go
├── migrations/                    # Goose миграции
│   └── {timestamp}_{name}.sql
├── proto/                         # gRPC протофайлы
│   ├── {service}.proto
│   ├── {service}_grpc.pb.go       # Сгенерировано
│   └── {service}.pb.go            # Сгенерировано
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
2. Dependency Injection (DI)
Правило: Никакой логики сборки в main.go!

go
// ✅ ПРАВИЛЬНО: services/auth/cmd/main.go
func main() {
    ctx := context.Background()
    
    app, err := app.NewApp(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer app.Close()
    
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
go
// ❌ НЕПРАВИЛЬНО: Вся логика в main.go
func main() {
    cfg := config.Load()
    db := connectDB(cfg)
    repo := repository.New(db)
    svc := service.New(repo)
    handler := handler.New(svc)
    server := grpc.NewServer()
    // ... и так далее
}
Шаблон DI:

go
// services/auth/internal/app/di.go
func NewApp(ctx context.Context) (*App, error) {
    // 1. Конфиг
    cfg, err := config.NewConfig()
    if err != nil {
        return nil, err
    }
    
    // 2. Логгер
    logger := logger.New(cfg.LoggerConfig())
    
    // 3. Closer (graceful shutdown)
    closer := closer.New()
    
    // 4. База данных
    db, err := postgres.Connect(cfg.PostgresConfig())
    if err != nil {
        return nil, err
    }
    closer.Add(db.Close)
    
    // 5. Репозиторий
    repo := repository.New(db)
    
    // 6. Сервис
    svc := service.New(repo, logger)
    
    // 7. API хендлер
    api := api.New(svc, logger)
    
    // 8. gRPC сервер
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(interceptor.Logger(logger)),
    )
    pb.RegisterAuthServiceServer(grpcServer, api)
    
    return &App{
        logger: logger,
        grpcServer: grpcServer,
        closer: closer,
    }, nil
}
3. Конфигурация через интерфейсы
Правило: Всегда используй интерфейсы для конфигов — это делает тесты возможными.

go
// services/auth/internal/config/interfaces.go
type ConfigProvider interface {
    LoggerConfig() LoggerConfig
    PostgresConfig() PostgresConfig
    GRPCConfig() GRPCConfig
}

type LoggerConfig interface {
    Level() string
    Format() string
}

type PostgresConfig interface {
    DSN() string
    MaxOpenConns() int
    MaxIdleConns() int
}
4. Обработка ошибок
Правило: Используй кастомные ошибки с типами, чтобы клиенты могли их обрабатывать.

go
// internal/model/errors.go
var (
    ErrNotFound = errors.New("entity not found")
    ErrAlreadyExists = errors.New("entity already exists")
    ErrInvalidInput = errors.New("invalid input")
)

type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
5. Тестирование
Правило: Каждый сервис должен иметь три уровня тестов:

Unit-тесты (*_test.go рядом с кодом) — с моками

Integration-тесты — с реальной БД (testcontainers)

E2E-тесты (tests/e2e/) — полный сценарий

Шаблон E2E теста:

go
// tests/e2e/auth/suite_test.go
func TestMain(m *testing.M) {
    ctx := context.Background()
    
    // 1. Поднимаем контейнер с PostgreSQL
    container, err := testcontainers.StartPostgres(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer container.Terminate(ctx)
    
    // 2. Запускаем сервис
    app, err := app.NewApp(ctx)
    if err != nil {
        log.Fatal(err)
    }
    go app.Run()
    
    // 3. Ждем готовности
    time.Sleep(2 * time.Second)
    
    // 4. Запускаем тесты
    code := m.Run()
    
    // 5. Graceful shutdown
    app.Close()
    os.Exit(code)
}
6. Platform пакет
Правило: Общие утилиты выносим в platform/ модуль.

text
platform/
├── go.mod
├── go.sum
└── pkg/
    ├── closer/        # Graceful shutdown
    │   └── closer.go
    ├── logger/        # Структурированный логгер
    │   ├── logger.go
    │   ├── noop_logger.go
    │   └── logger_bench_test.go
    ├── grpc/
    │   └── health/    # Health-check для k8s
    │       └── health.go
    └── testcontainers/# Для E2E тестов
        ├── app/
        ├── mongo/
        └── network/
7. Graceful Shutdown (Closer)
Правило: Все ресурсы должны закрываться корректно при остановке.

go
// platform/pkg/closer/closer.go
type Closer struct {
    mu    sync.Mutex
    funcs []func() error
}

func (c *Closer) Add(fn func() error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.funcs = append(c.funcs, fn)
}

func (c *Closer) Close() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    var errs []error
    for i := len(c.funcs) - 1; i >= 0; i-- {
        if err := c.funcs[i](); err != nil {
            errs = append(errs, err)
        }
    }
    return errors.Join(errs...)
}
8. Логирование
Правило: Используй структурированный логгер (slog) с уровнями.

go
// platform/pkg/logger/logger.go
type Logger struct {
    *slog.Logger
}

func New(cfg LoggerConfig) *Logger {
    var handler slog.Handler
    
    switch cfg.Format() {
    case "json":
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: level,
        })
    default:
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: level,
        })
    }
    
    return &Logger{
        Logger: slog.New(handler),
    }
}
9. Миграции (Goose)
Правило: Миграции должны быть идемпотентными и версионированными.

sql
-- services/auth/migrations/20260530005336_init_users.sql
-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    nickname TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS users;
10. Протофайлы (gRPC)
Правило: Используй buf для генерации кода.

yaml
# shared/proto/buf.gen.yaml
version: v1
plugins:
  - name: go
    out: ../pkg/proto
    opt: paths=source_relative
  - name: go-grpc
    out: ../pkg/proto
    opt: paths=source_relative
  - name: validate
    out: ../pkg/proto
    opt: paths=source_relative,lang=go
🛠️ Инструменты и команды
Taskfile (замена Makefile)
yaml
# Taskfile.yml
version: '3'

vars:
  GO_VERSION: '1.24'
  GOLANGCI_LINT_VERSION: 'v2.1.5'

tasks:
  dev:
    desc: "Запуск в режиме разработки"
    cmds:
      - task: infra-up
      - task: migrate-all
      - task: run-all
  
  infra-up:
    desc: "Запуск инфраструктуры"
    cmds:
      - docker-compose -f deployments/docker-compose.cluster.yml up -d
  
  migrate-all:
    desc: "Применение всех миграций"
    cmds:
      - for: ['auth', 'billing', 'game', 'shop']
        cmd: cd services/{{.ITEM}} && goose -dir migrations postgres "postgres://..." up
  
  test-all:
    desc: "Запуск всех тестов"
    cmds:
      - go test -v -race -cover ./...
  
  lint:
    desc: "Линтинг"
    cmds:
      - golangci-lint run ./...
  
  format:
    desc: "Форматирование"
    cmds:
      - gofumpt -extra -w .
      - gci write -s standard -s default -s "prefix(github.com/Eastwesser/event-horizon)" .
Команды для разработки
bash
# Разработка
task dev              # Запуск всего
task test-all         # Все тесты
task lint             # Линтинг
task format           # Форматирование

# Инфраструктура
make up               # Docker Compose up
make down             # Docker Compose down
make logs             # Логи
make migrate-all      # Миграции

# Деплой
make deploy           # Локальный деплой
make deploy-k3s       # Деплой в k3s
📝 Правила работы с Cursor
1. Генерация кода
Когда просишь сгенерировать новый сервис:

text
Создай новый микросервис {service_name} со структурой Clean Architecture:
- cmd/main.go (тонкий)
- internal/app/ (DI)
- internal/api/{service}/v1/ (хендлеры)
- internal/config/ (конфиг с интерфейсами)
- internal/model/ (доменные модели)
- internal/repository/ (репозитории)
- internal/service/ (бизнес-логика)
- migrations/ (Goose миграции)
- proto/ (gRPC протофайлы)
2. Рефакторинг
Когда просишь рефакторинг:

text
Рефакторинг {service_name}:
1. Вынеси DI в internal/app/di.go
2. Сделай конфиг через интерфейсы
3. Добавь closer для graceful shutdown
4. Разбей handler на отдельные файлы (create.go, get.go, update.go, delete.go)
5. Добавь E2E тесты
3. Тестирование
Когда просишь тесты:

text
Напиши для {service_name}:
1. Unit-тесты для service (с моками repository)
2. Integration-тесты для repository (с testcontainers)
3. E2E тесты (полный сценарий)
4. Код-ревью
При проверке кода проверяй:

✅ Структура соответствует Clean Architecture

✅ DI вынесен в internal/app/

✅ Есть интерфейсы для конфигов

✅ Используется closer для graceful shutdown

✅ Есть тесты (unit + integration)

✅ Ошибки обрабатываются через кастомные типы

✅ Миграции идемпотентны

🔗 Полезные ссылки
Курс Козырева: kozirev_code/microservices-course-examples-main/week_4/

Домашка: kozirev_code/microservices-course-homework-main/homeworks/week4/

Boilerplate: kozirev_code/microservices-course-boilerplate-main/

🎯 Приоритеты
DI в auth — самый простой сервис, делаем первым

Platform пакет — логгер + closer

E2E тесты — для auth

Taskfile — замена Makefile

Раскатка на остальные сервисы

Этот агент теперь знает:

Твою архитектуру

Подход Козырева из Week 4

Твои текущие проблемы (DI, тесты, graceful shutdown)

Что ты хочешь улучшить

Теперь ты можешь просто сказать:

"Сделай DI для billing как в Week 4"

"Напиши E2E тесты для shop"

"Добавь closer в auth"

"Создай новый сервис {name}"

И я буду применять все эти правила автоматически! 🚀