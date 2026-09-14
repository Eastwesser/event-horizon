# Быстрый старт

## Запуск

```bash
go mod download
go run main.go
```

## Использование

### Web UI

Откройте http://localhost:8080

Возможности:
- Создание комментариев
- Ответы на комментарии (кнопка "Ответить")
- Удаление комментариев
- Поиск по тексту
- Сортировка (по дате, по автору)
- Постраничная навигация

### API

```bash
# Создать комментарий
curl -X POST http://localhost:8080/comments \
  -H "Content-Type: application/json" \
  -d '{"author": "User", "content": "Test comment"}'

# Получить комментарии
curl "http://localhost:8080/comments?sort=date_desc&page=1"

# Поиск
curl "http://localhost:8080/comments/search?q=test"

# Удалить
curl -X DELETE http://localhost:8080/comments/{id}
```

### PowerShell тест

```powershell
.\test-api.ps1
```

## Структура комментариев

Комментарии могут иметь неограниченную вложенность:
- Комментарий 1
  - Ответ 1.1
    - Ответ 1.1.1
  - Ответ 1.2
- Комментарий 2

При удалении комментария удаляются все его дочерние комментарии.

