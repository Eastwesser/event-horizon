# Быстрый старт EventBooker

## Запуск с Docker (PostgreSQL)

1. Запустите PostgreSQL:
```bash
make docker-up
```

Или вручную:
```bash
docker compose up -d
```

2. Установите зависимости:
```bash
make install
```

3. Создайте `.env` файл (опционально):
```bash
cp env.example .env
```

4. Запустите приложение:
```bash
make run
```

5. Откройте браузер:
   - Пользовательская часть: http://localhost:8080
   - Админ-панель: http://localhost:8080/admin

## Использование

### Веб-интерфейс (для пользователей)

1. Откройте http://localhost:8080
2. Введите ваш email
3. Выберите мероприятие и нажмите "Забронировать"
4. Подтвердите бронирование до истечения срока оплаты
5. Ваши бронирования отображаются внизу страницы

### Админ-панель

1. Откройте http://localhost:8080/admin
2. Создайте мероприятие (укажите название, дату, количество мест, срок оплаты)
3. Просматривайте список всех мероприятий
4. Кликните на мероприятие или нажмите "Брони" для просмотра бронирований

### API

**Создать мероприятие:**
```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tech Conference",
    "description": "Annual tech event",
    "event_date": "2025-12-31T18:00:00Z",
    "total_seats": 100,
    "booking_timeout_min": 15
  }'
```

**Получить все мероприятия:**
```bash
curl http://localhost:8080/events
```

**Забронировать место:**
```bash
curl -X POST http://localhost:8080/events/{event_id}/book \
  -H "Content-Type: application/json" \
  -d '{"user_email": "user@example.com"}'
```

**Подтвердить бронирование:**
```bash
curl -X POST http://localhost:8080/bookings/{booking_id}/confirm
```

**Получить бронирования пользователя:**
```bash
curl "http://localhost:8080/bookings?email=user@example.com"
```

**Получить бронирования мероприятия:**
```bash
curl http://localhost:8080/events/{event_id}/bookings
```

## Автоматическая отмена броней

- Scheduler автоматически проверяет просроченные бронирования каждые 30 секунд
- Если бронь не подтверждена в течение указанного времени - она отменяется
- Место автоматически освобождается и становится доступным для других

## Остановка

Для остановки сервиса нажмите `Ctrl+C`.

Для остановки Docker контейнеров:
```bash
make docker-down
```

## Устранение неполадок

**Порт 8080 занят:**
- Измените `port` в `.env` файле

**PostgreSQL не подключается:**
- Убедитесь, что контейнер запущен: `docker ps`
- Проверьте логи: `docker logs eventbooker-postgres`

**Бронирования не отменяются автоматически:**
- Проверьте логи приложения - scheduler должен запуститься
- Проверьте, что указано корректное время в `booking_timeout_min`

