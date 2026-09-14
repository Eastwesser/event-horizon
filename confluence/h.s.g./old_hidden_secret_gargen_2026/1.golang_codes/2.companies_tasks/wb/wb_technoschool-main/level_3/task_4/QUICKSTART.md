# Быстрый старт ImageProcessor

## Вариант 1: С Docker (Kafka)

1. Запустите Kafka и Zookeeper:
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

5. Откройте браузер: http://localhost:8080

## Вариант 2: Без Docker (In-memory очередь)

1. Установите зависимости:
```bash
make install
```

2. Запустите приложение:
```bash
make run
```

Приложение автоматически использует in-memory очередь, если Kafka недоступен.

## Использование

### Через веб-интерфейс

1. Откройте http://localhost:8080
2. Перетащите изображение в область загрузки или нажмите для выбора
3. Нажмите "Загрузить"
4. Дождитесь обработки (статус обновляется автоматически)
5. Нажмите "Просмотр" для просмотра обработанного изображения

### Через API

**Загрузить изображение:**
```bash
curl -X POST -F "file=@image.jpg" http://localhost:8080/upload
```

**Получить информацию об изображении:**
```bash
curl http://localhost:8080/image/{id}
```

**Получить обработанное изображение:**
```bash
curl http://localhost:8080/image/{id}/file?type=processed
curl http://localhost:8080/image/{id}/file?type=thumbnail
curl http://localhost:8080/image/{id}/file?type=original
```

**Удалить изображение:**
```bash
curl -X DELETE http://localhost:8080/image/{id}
```

**Получить все изображения:**
```bash
curl http://localhost:8080/images
```

## Остановка

Для остановки сервиса нажмите `Ctrl+C`.

Для остановки Docker контейнеров:
```bash
make docker-down
```

## Устранение неполадок

**Порт 8080 занят:**
- Измените `PORT` в `.env` файле или через переменную окружения

**Kafka не подключается:**
- Проверьте, что контейнеры запущены: `docker ps`
- Приложение автоматически переключится на in-memory очередь

**Ошибки обработки изображений:**
- Убедитесь, что формат поддерживается (JPG, PNG, GIF)
- Проверьте логи приложения

