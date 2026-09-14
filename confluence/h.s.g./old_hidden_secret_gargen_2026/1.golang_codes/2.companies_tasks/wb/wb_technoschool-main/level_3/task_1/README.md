# DelayedNotifier

Сервис отложенных уведомлений с поддержкой RabbitMQ, Redis и множественных каналов отправки.

## Описание

Сервис принимает запросы на создание уведомлений, складывает их в очередь и отправляет в указанное время. При ошибке автоматически повторяет попытки с экспоненциальной задержкой.

## Функциональность

**HTTP API:**
- `POST /notify` - создание уведомления
- `GET /notify/:id` - получение статуса
- `DELETE /notify/:id` - отмена уведомления
- `GET /notify` - список всех уведомлений

**Каналы отправки:**
- Console (тестирование)
- Email (SMTP)
- Telegram (Bot API)

**Дополнительно:**
- Redis для кэширования
- RabbitMQ для очередей
- Retry с экспоненциальной задержкой (до 5 попыток)
- Web UI для управления
- Работает без Docker (in-memory режим)

## Быстрый старт

### С Docker:

```bash
# Запуск инфраструктуры
docker compose up -d

# Установка зависимостей
go mod download

# Запуск
go run main.go
```

### Без Docker:

```bash
go mod download
go run main.go
```

Приложение автоматически использует in-memory компоненты если RabbitMQ/Redis недоступны.

## Настройка

Создайте `.env` (опционально):

```env
PORT=8080
RABBITMQ_URL=amqp://guest:guest@localhost:5872/
REDIS_ADDR=localhost:6379

# Email (опционально)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@email.com
SMTP_PASSWORD=your-password
SMTP_FROM=your@email.com

# Telegram (опционально)
TELEGRAM_BOT_TOKEN=your-token
```

## Использование

### Web UI

Откройте http://localhost:8080

### API

```bash
# Создать уведомление
curl -X POST http://localhost:8080/notify \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "console",
    "recipient": "test-user",
    "message": "Тестовое уведомление",
    "scheduled_at": "2025-11-01T15:00:00Z"
  }'

# Получить статус
curl http://localhost:8080/notify/{id}

# Отменить
curl -X DELETE http://localhost:8080/notify/{id}
```

### PowerShell тест

```powershell
.\test-api.ps1
```

## Структура

```
task_1/
├── cache/          - Redis кэширование
├── config/         - Конфигурация
├── handlers/       - HTTP handlers
├── models/         - Модели данных
├── notifiers/      - Каналы отправки
├── queue/          - RabbitMQ/in-memory очередь
├── service/        - Бизнес-логика
├── storage/        - In-memory хранилище
├── worker/         - Фоновый обработчик
├── static/         - Web UI
└── main.go         - Точка входа
```

## Технологии

- Go 1.21+
- Gin Web Framework
- RabbitMQ
- Redis
- Docker Compose

## Статусы уведомлений

- `pending` - ожидает
- `scheduled` - запланировано
- `sending` - отправляется
- `sent` - отправлено
- `failed` - ошибка после всех попыток
- `cancelled` - отменено

## Troubleshooting

**Порт 5672 занят (Windows):**
```bash
# Уже изменено на 5872 в docker-compose.yml
```

**Приложение не запускается:**
```bash
# Проверьте, что порт 8080 свободен
# Или измените через переменную окружения:
set PORT=8081
go run main.go
```

**RabbitMQ недоступен:**
Приложение автоматически использует in-memory очередь.

**Redis недоступен:**
Приложение автоматически работает без кэша.
