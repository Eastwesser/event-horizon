# Comment Tree

Сервис древовидных комментариев с поиском и навигацией.

## Описание

Сервис для работы с древовидными комментариями с неограниченной вложенностью, поиском, сортировкой и постраничной навигацией.

## Функциональность

**HTTP API:**
- `POST /comments` - создание комментария (с указанием родительского)
- `GET /comments?parent={id}` - получение комментариев и всех вложенных
- `GET /comments/search?q={query}` - полнотекстовый поиск
- `DELETE /comments/{id}` - удаление комментария и всех вложенных

**Дополнительно:**
- Постраничная навигация (`page`, `page_size`)
- Сортировка (`sort`: `date_asc`, `date_desc`, `author`)
- Web UI с визуальной древовидной структурой
- Полнотекстовый поиск по содержимому и автору

## Быстрый старт

```bash
go mod download
go run main.go
```

Откройте http://localhost:8080

## Использование

### Web UI

Создавайте комментарии, отвечайте на них, удаляйте и ищите через интерфейс.

### API

```bash
# Создать комментарий
curl -X POST http://localhost:8080/comments \
  -H "Content-Type: application/json" \
  -d '{"author": "Alice", "content": "Hello!"}'

# Создать ответ
curl -X POST http://localhost:8080/comments \
  -H "Content-Type: application/json" \
  -d '{"author": "Bob", "content": "Hi!", "parent_id": "comment-id"}'

# Получить комментарии
curl "http://localhost:8080/comments?sort=date_desc&page=1&page_size=10"

# Поиск
curl "http://localhost:8080/comments/search?q=hello"

# Удалить
curl -X DELETE http://localhost:8080/comments/{id}
```

### Тестирование

```powershell
.\test-api.ps1
```

## Структура

```
task_3/
├── config/         - Конфигурация
├── handlers/       - HTTP handlers
├── models/         - Модели данных
├── service/        - Бизнес-логика (построение дерева)
├── storage/        - In-memory хранилище
├── static/         - Web UI
└── main.go         - Точка входа
```

## Технологии

- Go 1.21+
- Gin Web Framework

## Параметры запросов

**GET /comments:**
- `parent` - ID родительского комментария (необязательно)
- `sort` - сортировка: `date_asc`, `date_desc`, `author`
- `page` - номер страницы (по умолчанию 1)
- `page_size` - размер страницы (по умолчанию 10)

**GET /comments/search:**
- `q` - поисковый запрос (обязательно)
