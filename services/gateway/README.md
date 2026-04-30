# 2. API GATEWAY SERVICE

```bash
.
├── cmd
│   └── main.go
├── Dockerfile
├── go.mod
├── internal
├── proto
└── README.md
```


Следующий шаг: gateway (REST → gRPC прокси)

Теперь нам нужен HTTP gateway, чтобы фронтенд (Angular) мог отправлять JSON, а мы внутри ходили в gRPC.

План для gateway:

Создадим структуру services/gateway/
Напишем HTTP хендлеры (Gin или стандартный net/http)
Подключимся к auth по gRPC (клиент)
Проксируем запросы: POST /api/auth/register → auth.AuthService/Register

--

2. Gateway: самописный + Envoy потом

Отличный план. Go gateway даст тебе:

Rate limiting (покажешь код на собеседовании)
Маршрутизация HTTP → gRPC
WebSocket прокси для leaderboard
CORS, заголовки, логирование
Envoy (или Traefik) потом — когда понадобится:

L7 балансировка с retries/circuit breakers
gRPC load balancing без дополнительного hop
Let's Encrypt автоматический
Для старта: Go gateway на 300 строк. Envoy добавим как "production hardening" позже.

ОБНОВЛЕНО:

# API Gateway Service

HTTP → gRPC прокси с поддержкой NATS JetStream и WebSocket (в плане).

## Структура
.
├── cmd/
│ └── main.go # Точка входа, HTTP сервер на Gin
├── internal/
│ └── client/
│ └── auth_client.go # gRPC клиент для Auth сервиса
├── Dockerfile
├── go.mod
├── go.sum
└── README.md

text

## Функциональность

### Реализовано ✅

| Метод | HTTP метод | Эндпоинт | Назначение |
|-------|------------|----------|------------|
| Register | POST | `/api/auth/register` | Регистрация пользователя |
| Login | POST | `/api/auth/login` | Логин, выдача JWT |
| Health | GET | `/health` | Проверка работоспособности |
| NATS | - | - | Публикация событий `user.registered`, `user.logged_in` |

### В плане 📋

- Rate limiting (самописный, для собеседований)
- WebSocket прокси для leaderboard
- CORS (для Angular фронтенда)
- Метрики для Prometheus
- Graceful shutdown

## Запуск

### Требования

- Go 1.25.6+
- Запущенный Auth сервис (localhost:50051)
- Запущенный NATS (localhost:4222)

### Локальный запуск

```bash
cd services/gateway
go mod tidy
go run cmd/main.go

Переменные окружения

Переменная	Значение по умолчанию	Назначение
AUTH_GRPC_ADDR	localhost:50051	Адрес Auth gRPC сервера
NATS_URL	nats://localhost:4222	Адрес NATS сервера
GIN_MODE	debug	Режим Gin (release для продакшена)

API Endpoints

POST /api/auth/register

Регистрация нового пользователя.

Тело запроса:

json
{
    "email": "user@example.com",
    "password": "secret123"
}
Ответ (успех):

json
{
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "success": true,
    "message": "user registered successfully"
}
POST /api/auth/login

Аутентификация пользователя.

Тело запроса:

json
{
    "email": "user@example.com",
    "password": "secret123"
}
Ответ (успех):

json
{
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "tokenType": "Bearer",
    "expiresIn": 86400
}
GET /health

Проверка работоспособности.

Ответ:

json
{
    "status": "ok"
}
NATS События

Gateway публикует следующие события в NATS JetStream:

Событие	Топик	Условие
Регистрация	event.user.registered	После успешной регистрации
Логин	event.user.logged_in	После успешного логина
Подписка на события (для отладки)

bash
nats sub "event.>" --server localhost:4222
Тестирование

Через curl

bash
# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Логин
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Health check
curl http://localhost:8080/health
Метрики производительности

Операция	Время (первый запрос)	Примечание
Регистрация	~295ms	Включает bcrypt хэширование
Логин	~121ms	Только проверка пароля + JWT
Холодный старт после docker-compose up может показывать большие значения.

Docker

Сборка образа

bash
docker build -t eventhorizon/gateway services/gateway/
Запуск в Docker Compose

Gateway интегрирован в основной docker-compose.cluster.yml:

bash
docker-compose -f deployments/docker-compose.cluster.yml up gateway
Архитектура

text
Client (HTTP/JSON)
    │
    ▼
Gateway (Gin)
    ├──► Auth (gRPC) ──► PostgreSQL
    └──► NATS ──► JetStream ──► Subscribers (Leaderboard, Analytics, Notification)
Планируемые улучшения (Техдолг)

Rate limiting — реализовать собственный middleware для защиты от DDoS
CORS — настроить для Angular фронтенда
WebSocket — добавить поддержку для real-time leaderboard
Prometheus метрики — количество запросов, ошибки, задержки
Graceful shutdown — корректное завершение при SIGTERM
Circuit breaker — защита от падающих бэкендов
Request ID — сквозная трассировка запросов

Связанные сервисы

Auth Service — gRPC сервер аутентификации
Leaderboard Service — будет подписан на NATS события
NATS JetStream — событийная шина


ЕЩЕ БОЛЬШЕ ИНФЫ:

# API Gateway Service

HTTP → gRPC прокси с поддержкой NATS JetStream. Точка входа для всех клиентов (Angular фронтенд, мобильные приложения, тесты).

## Статус

✅ **Реализовано:** Регистрация, логин, NATS события  
🔄 **В разработке:** WebSocket для leaderboard  
📋 **В плане:** Rate limiting, CORS, метрики  

## Структура
.
├── cmd/
│ └── main.go # Точка входа, HTTP сервер на Gin
├── internal/
│ └── client/
│ └── auth_client.go # gRPC клиент для Auth сервиса
├── Dockerfile
├── go.mod
├── go.sum
└── README.md

text

*Примечание: папка `proto` отсутствует намеренно — gateway импортирует proto из сервиса auth.*

## Функциональность

### Реализовано ✅

| Метод | HTTP | Эндпоинт | Назначение |
|-------|------|----------|------------|
| Register | POST | `/api/auth/register` | Регистрация пользователя |
| Login | POST | `/api/auth/login` | Логин, выдача JWT |
| Health | GET | `/health` | Проверка работоспособности |
| NATS | async | — | Публикация событий в JetStream |

### В плане 📋

| Функция | Приоритет | Назначение |
|---------|-----------|------------|
| Rate limiting | Высокий | Защита от DDoS, покажет код на собеседовании |
| WebSocket | Высокий | Real-time leaderboard |
| CORS | Средний | Для Angular фронтенда |
| Prometheus метрики | Средний | Мониторинг RPS, ошибок, задержек |
| Graceful shutdown | Средний | SIGTERM → корректное завершение |
| Circuit breaker | Низкий | Защита от падающих бэкендов |
| Request ID | Низкий | Сквозная трассировка |

## Запуск

### Требования

- Go 1.25.6+
- Запущенный Auth сервис (`localhost:50051`)
- Запущенный NATS (`localhost:4222`)
- (Опционально) Docker Compose кластер

### Локальный запуск

```bash
cd services/gateway
go mod tidy
go run cmd/main.go
Переменные окружения

Переменная	По умолчанию	Назначение
AUTH_GRPC_ADDR	localhost:50051	Адрес Auth gRPC сервера
NATS_URL	nats://localhost:4222	Адрес NATS сервера
GIN_MODE	debug	Режим Gin (release для продакшена)
PORT	8080	HTTP порт (в плане)
API Endpoints

POST /api/auth/register

Регистрация нового пользователя.

Тело запроса:

json
{
    "email": "user@example.com",
    "password": "secret123"
}
Ответ (успех, 200):

json
{
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "success": true,
    "message": "user registered successfully"
}
Ответ (ошибка валидации, 400):

json
{
    "error": "Key: 'RegisterRequest.Email' Error:Field validation for 'Email' failed on the 'required' tag"
}
Ответ (пользователь уже существует, 200 с success=false):

json
{
    "success": false,
    "message": "user already exists"
}
Ответ (внутренняя ошибка, 500):

json
{
    "error": "rpc error: code = Unavailable desc = connection error: ..."
}
POST /api/auth/login

Аутентификация пользователя.

Тело запроса:

json
{
    "email": "user@example.com",
    "password": "secret123"
}
Ответ (успех, 200):

json
{
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "tokenType": "Bearer",
    "expiresIn": 86400
}
Ответ (неверные учётные данные, 401):

json
{
    "error": "rpc error: code = Unauthenticated desc = invalid credentials"
}
Ответ (пустой email/password, 400):

json
{
    "error": "Key: 'LoginRequest.Email' Error:Field validation for 'Email' failed on the 'required' tag"
}
GET /health

Проверка работоспособности (без аутентификации).

Ответ (200):

json
{
    "status": "ok"
}
NATS События

Gateway публикует следующие события в NATS JetStream:

Событие	Топик	Условие	Payload
Регистрация	event.user.registered	После успешной регистрации	{"event":"user.registered","user_id":"...","email":"..."}
Логин	event.user.logged_in	После успешного логина	{"event":"user.logged_in","email":"..."}
Подписка на события (для отладки)

bash
# Установка NATS CLI (если нет)
go install github.com/nats-io/natscli/nats@latest

# Подписка на все события
nats sub "event.>" --server localhost:4222

# Подписка только на регистрации
nats sub "event.user.registered" --server localhost:4222
Архитектура (Диаграмма)

text
┌─────────────────────────────────────────────────────────────────────┐
│                          Клиент (Angular/curl)                       │
│                               HTTP/JSON                              │
└─────────────────────────────────┬───────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         GATEWAY (Gin)                                │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │  /api/auth/register  →  authClient.Register()               │    │
│  │  /api/auth/login     →  authClient.Login()                  │    │
│  │  /health             →  ok                                  │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                    │                              │                  │
│                    │ gRPC                         │ NATS Publish     │
│                    ▼                              ▼                  │
│            ┌─────────────┐                 ┌─────────────┐          │
│            │ Auth (gRPC) │                 │   NATS      │          │
│            │  :50051     │                 │ JetStream   │          │
│            └─────────────┘                 │  :4222      │          │
│                 │                          └─────────────┘          │
│                 │ PostgreSQL                     │                  │
│                 ▼                                ▼                  │
│            ┌─────────────┐                 ┌─────────────┐          │
│            │ PostgreSQL  │                 │ Subscriber  │          │
│            │   :5460     │                 │ (Leaderboard│          │
│            └─────────────┘                 │  и др.)     │          │
│                                            └─────────────┘          │
└─────────────────────────────────────────────────────────────────────┘
Тестирование

Через curl

bash
# Health check
curl http://localhost:8080/health

# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Логин
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'
Полный сквозной тест (с NATS)

bash
# Терминал 1: NATS subscriber
nats sub "event.>" --server localhost:4222

# Терминал 2: Запустить auth
cd services/auth && ./auth-service

# Терминал 3: Запустить gateway
cd services/gateway && go run cmd/main.go

# Терминал 4: Отправить запрос
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'

# Ожидаем в терминале 1: 
# [#1] Received on "event.user.registered"
# {"email":"alice@example.com","event":"user.registered","user_id":"..."}
Метрики производительности

Операция	Время (первый запрос)	Время (последующие)	Примечание
Регистрация	~295ms	~50-80ms	bcrypt (cost=10) — самый дорогой
Логин	~121ms	~15-30ms	Только проверка пароля + JWT
Health	~0.5ms	~0.3ms	Без логики
Холодный старт после docker-compose up может показывать большие значения из-за инициализации соединений.

Как ускорить в продакшене

bash
# Включить release режим
export GIN_MODE=release

# Уменьшить cost bcrypt (по умолчанию 10 → 8, но менее безопасно)
# В коде auth: bcrypt.GenerateFromPassword(password, 8)

# Добавить кеш JWT в Redis (снизит нагрузку на PostgreSQL)
Docker

Сборка образа

bash
docker build -t eventhorizon/gateway services/gateway/
Запуск в Docker Compose

Gateway интегрирован в основной docker-compose.cluster.yml:

bash
# Запуск всех сервисов
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Запуск только gateway
docker-compose -f deployments/docker-compose.cluster.yml up gateway
Инструкция по отладке

Проблема: Gateway не запускается

bash
# Проверка логов
cd services/gateway
go run cmd/main.go 2>&1 | tee gateway.log

# Типичные ошибки:
# - "connection refused" → Auth не запущен
# - "no such file or directory" → go.mod проблемы
Проблема: Auth connection refused

bash
# Проверить, что Auth слушает
ss -tlnp | grep 50051

# Запустить Auth
cd services/auth && ./auth-service

# Или пересобрать
go build -o auth-service ./cmd/main.go && ./auth-service
Проблема: NATS не публикует события

bash
# Проверить, что NATS работает
docker exec event-horizon-nats nats-server --version

# Проверить JetStream
docker logs event-horizon-nats | grep -i jetstream

# Вручную подписаться
nats sub "event.>" --server localhost:4222

# Отправить тестовое событие (в другом терминале)
nats pub "event.test" "hello" --server localhost:4222
Проблема: Порты заняты

bash
# Найти процесс на порту 8080
sudo lsof -i :8080

# Убить процесс
sudo kill -9 <PID>

# Или сменить порт в коде
# В cmd/main.go: r.Run(":8081")
Планируемые улучшения (Техдолг)

Rate limiting — реализовать собственный middleware для защиты от DDoS (покажет код на собеседовании)
CORS — настроить для Angular фронтенда
WebSocket — добавить поддержку для real-time leaderboard
Prometheus метрики — количество запросов, ошибки, задержки
Graceful shutdown — корректное завершение при SIGTERM
Circuit breaker — защита от падающих бэкендов
Request ID — сквозная трассировка запросов
Envoy — замена самописному gateway для production (L7 балансировка, retries, circuit breakers)
Связанные сервисы

Auth Service — gRPC сервер аутентификации
Leaderboard Service — будет подписан на NATS события
NATS JetStream — событийная шина
