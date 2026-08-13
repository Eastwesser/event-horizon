# HTTP-запросы: пагинация, фильтры, идемпотентность, CORS — шпаргалка

## Пагинация

Два основных стиля:

| Стиль | Параметры | Плюсы | Минусы |
|-------|-----------|-------|--------|
| Offset/limit | `?page=2&limit=20` или `offset` | Просто | Медленно на больших offset, drift при вставках |
| Cursor/keyset | `?cursor=eyJ...&limit=20` | Стабильно под нагрузкой | Сложнее, курсор непрозрачный |

Senior-дефолт для лент и shop catalog: **keyset** по `(created_at, id)`. Offset — админки и маленькие таблицы.

Контракт ответа: `items`, `next_cursor` / `total` (total — дорогой COUNT, не всегда нужен). Лимит сверху (например max 100), дефолт 20.

В EH-фронте shop — infinite scroll; на бэке формулируйте пагинацию явно в OpenAPI Gateway.

## Фильтрация и сортировка

- Whitelist полей сортировки (`sort=price`, `order=asc|desc`) — никакого сырого SQL из query string.
- Фильтры: равенство, диапазоны, enum категорий (`category=merch`).
- Индексы под частые фильтры; составные под keyset.
- Пустой результат — **200 + []**, не 404 (404 — нет конкретного ресурса `/items/{id}`).

## Идемпотентный POST

Проблема: клиент ретраит POST purchase/payment — риск двойного списания.

Паттерны:
1. Заголовок `Idempotency-Key` (UUID): сервер хранит ключ → ответ; повтор — тот же результат.
2. Natural key: `event_id` / `order_id` UNIQUE в БД.
3. Outbox + идемпотентный consumer на стороне обработчика событий.

POST создания ≠ всегда неидемпотентен: правильно спроектированный POST с ключом — безопасен к ретраям. PUT/DELETE по id обычно идемпотентны по определению.

Для EH: покупка и payment confirm — зона, где на собесе обязаны упомянуть идемпотентность и outbox.

## CORS (кратко, но точно)

Браузер блокирует cross-origin XHR/fetch без заголовков от сервера. Gateway должен отдавать `Access-Control-Allow-Origin` (конкретный origin, не `*` с credentials), `Allow-Methods`, `Allow-Headers`, отвечать на **OPTIONS** preflight.

Замечания:
- CORS — защита браузера, не auth. curl/postman CORS не ограничивает.
- С credentials нужны явный origin и `Allow-Credentials: true`.
- Не путать с CSRF: для cookie-session нужна antiforgery; при Bearer JWT в header CSRF слабее, но XSS всё ещё смертелен.

## Прочие query-привычки

- Не класть секреты в query (логи прокси).
- Булевы и enum — строгая валидация (proto `validate.rules` / Gin binding).
- Тот же контракт держать в `docs/openapi.yaml` и Gin gateway.

## Антипаттерны

- `SELECT *` + фильтрация в приложении на миллионах строк.
- Отрицательный limit / limit=999999.
- POST `/transfer` без идемпотентности.
- `Access-Control-Allow-Origin: *` + cookies.

## Типичные вопросы на собесе

1. Offset vs cursor — когда что выбираете?
2. Почему COUNT(*) на каждую страницу — плохая идея?
3. Как сделать POST покупки идемпотентным?
4. Чем идемпотентность HTTP-метода отличается от Idempotency-Key?
5. Зачем whitelist для sort/filter?
6. Почему CORS не заменяет аутентификацию?
7. Что вернуть, если фильтр валиден, а данных нет — 200 или 404?
8. Как описать пагинацию в OpenAPI так, чтобы фронт не гадал?
