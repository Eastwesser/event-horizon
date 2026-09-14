# HTTP Protocol

## Грейд
17-19

## Вопрос
Что из себя представляет протокол HTTP? Из каких основных частей состоит HTTP запрос и HTTP ответ?

## Ответ

**HTTP (HyperText Transfer Protocol)** — текстовый протокол, работающий по принципу **запрос-ответ** поверх TCP (обычно порт 80 для HTTP, 443 для HTTPS).

### Структура HTTP запроса
```text
START LINE: POST /api/users HTTP/1.1

HEADERS:
Host: example.com
Content-Type: application/json
Content-Length: 27
Authorization: Bearer token123

EMPTY LINE: \r\n

BODY: {"name": "John", "age": 30}
```

```text

### Структура HTTP ответа
START LINE: HTTP/1.1 200 OK

HEADERS:
Content-Type: application/json
Content-Length: 15

EMPTY LINE: \r\n

BODY: {"status": "ok"}
```

## Дополнительные вопросы

### Какие основные методы HTTP?

| Метод | Назначение | Идемпотентный |
|-------|------------|---------------|
| GET | Получение ресурса | Да |
| HEAD | Получение только заголовков | Да |
| POST | Создание ресурса / отправка данных | Нет |
| PUT | Полная замена ресурса | Да |
| PATCH | Частичное обновление | Нет |
| DELETE | Удаление ресурса | Да |

### Примеры HTTP хедеров

| Хедер | Назначение |
|-------|------------|
| `Content-Type` | Тип тела (application/json, text/html) |
| `Content-Length` | Длина тела в байтах |
| `Cookie` | Отправка cookie на сервер |
| `Set-Cookie` | Установка cookie сервером |
| `Authorization` | Токен авторизации (Bearer, Basic) |
| `User-Agent` | Информация о клиенте |

### Назовите классы кодов состояний HTTP

| Класс | Значение | Примеры |
|-------|----------|---------|
| 1xx | Informational | 100 Continue |
| 2xx | Success | 200 OK, 201 Created, 204 No Content |
| 3xx | Redirect | 301 Moved Permanently, 302 Found |
| 4xx | Client Error | 400 Bad Request, 401 Unauthorized, 404 Not Found |
| 5xx | Server Error | 500 Internal Server Error, 502 Bad Gateway, 503 Service Unavailable |

## Что важно сказать на собеседовании

### Почему статику рекомендуется размещать на отдельном домене?
- Отсутствие кук (меньше данных в каждом запросе)
- Параллельные запросы (браузер ограничивает количество соединений на домен)
- Возможность использовать CDN

### HTTP/1.1 vs HTTP/2 vs HTTP/3
- **HTTP/1.1** — текстовый, одно соединение = один запрос (head-of-line blocking)
- **HTTP/2** — бинарный, мультиплексирование, server push
- **HTTP/3** — поверх UDP (QUIC), нет head-of-line blocking на транспортном уровне
