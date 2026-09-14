# URL Shortener

Сервис сокращения URL с аналитикой переходов.

## Описание

Сервис генерирует короткие ссылки, перенаправляет пользователей на оригинальные URL и собирает статистику по переходам.

## Функциональность

**HTTP API:**
- `POST /shorten` - создание короткой ссылки
- `GET /s/{short_url}` - редирект на оригинальный URL
- `GET /analytics/{short_url}` - получение аналитики

**Аналитика:**
- Количество переходов
- Группировка по дням и месяцам
- Анализ User-Agent
- История переходов

**Дополнительно:**
- Redis кэширование популярных ссылок
- Кастомные имена для ссылок
- Web UI для управления

## Быстрый старт

### С Docker

```bash
docker compose up -d
go mod download
go run main.go
```

### Без Docker

```bash
go mod download
go run main.go
```

Откройте http://localhost:8080

## Использование

### Web UI

Создайте ссылку через форму, затем просматривайте аналитику.

### API

```bash
# Создать короткую ссылку
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.google.com", "custom": "google"}'

# Получить аналитику
curl http://localhost:8080/analytics/google
```

### Тестирование

```powershell
.\test-api.ps1
```

## Настройка

Создайте `.env` (опционально):

```env
PORT=8080
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
```

## Структура

```
task_2/
├── cache/          - Redis кэширование
├── config/         - Конфигурация
├── handlers/       - HTTP handlers
├── models/         - Модели данных
├── service/        - Бизнес-логика
├── storage/        - In-memory хранилище
├── static/         - Web UI
└── main.go         - Точка входа
```

## Технологии

- Go 1.21+
- Gin Web Framework
- Redis (опционально)