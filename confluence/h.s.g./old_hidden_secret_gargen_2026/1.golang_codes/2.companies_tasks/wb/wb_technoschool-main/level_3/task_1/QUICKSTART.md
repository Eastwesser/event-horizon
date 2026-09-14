# Быстрый старт

## Запуск за 3 минуты

### Вариант 1: С Docker

```bash
docker compose up -d
go mod download
go run main.go
```

### Вариант 2: Без Docker

```bash
go mod download
go run main.go
```

Приложение автоматически работает с in-memory компонентами если Docker недоступен.

## Использование

### Web UI

Откройте http://localhost:8080

Создайте уведомление:
- Канал: Console
- Получатель: test-user  
- Сообщение: Тест
- Время: через 1 минуту

### API тест

```powershell
.\test-api.ps1
```

## Настройка каналов

### Email

Создайте `.env`:

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@email.com
SMTP_PASSWORD=app-password
SMTP_FROM=your@email.com
```

Для Gmail создайте App Password: https://myaccount.google.com/apppasswords

### Telegram

Добавьте в `.env`:

```env
TELEGRAM_BOT_TOKEN=your-bot-token
```

Создайте бота: @BotFather → /newbot

## Troubleshooting

**Порт занят:**
```bash
set PORT=8081
go run main.go
```

**Windows порт 5672 зарезервирован:**
Уже исправлено - используется порт 5872 в `docker-compose.yml`

