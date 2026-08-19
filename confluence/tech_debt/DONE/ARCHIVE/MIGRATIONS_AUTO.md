# 🗄️ Автоматизация миграций

## Проблема
- При `docker-compose down -v` данные теряются

## Решение
- Использовать `docker-compose down` (без `-v`) для сохранения volumes
- В entrypoint каждого сервиса запускать `goose up`
- Разделить dev / prod volumes
