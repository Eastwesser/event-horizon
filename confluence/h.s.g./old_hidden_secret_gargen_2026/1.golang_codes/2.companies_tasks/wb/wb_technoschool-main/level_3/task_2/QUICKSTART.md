# Быстрый старт

## Вариант 1: С Docker

```bash
docker compose up -d
go mod download
go run main.go
```

## Вариант 2: Без Docker

```bash
go mod download
go run main.go
```

Приложение автоматически использует in-memory компоненты если Redis недоступен.

## Использование

### Web UI

Откройте http://localhost:8080

### API

```bash
# Создать короткую ссылку
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.google.com"}'

# Перейти по короткой ссылке
curl -L http://localhost:8080/s/abc123

# Получить аналитику
curl http://localhost:8080/analytics/abc123
```

### PowerShell тест

```powershell
.\test-api.ps1
```

## Настройка Redis (опционально)

Создайте `.env`:

```env
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
PORT=8080
```

## Endpoints

- `POST /shorten` - создать короткую ссылку
  ```json
  {"url": "https://example.com", "custom": "my-link"}
  ```

- `GET /s/{short_url}` - редирект на оригинальный URL

- `GET /analytics/{short_url}` - получить аналитику

