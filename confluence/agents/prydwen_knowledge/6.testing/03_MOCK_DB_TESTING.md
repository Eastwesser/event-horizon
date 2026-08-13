# Моки и тестирование БД

## Стратегии
1. **Интерфейс + mock** (gomock/mockery) — usecase без Postgres
2. **sqlmock** — проверка SQL/аргументов, хрупко к тексту запроса
3. **Реальная БД** (testcontainers) — правда для транзакций/outbox/индексов
4. **Фейк in-memory** — быстро, но врёт на SQL-диалекте

## Когда что
| Слой | Подход |
|------|--------|
| Service / usecase | mock repository |
| Repository SQL | testcontainers или sqlmock |
| Converter | чистые unit без моков |
| Outbox+tx | только интеграция с PG |

## EH
- Auth: mock `UserRepository` в `auth_service_test`
- Inventory: cache decorator unit-тесты
- Billing/Shop: placeholder `-tags=integration` — добить testcontainers

## Правила
- Не мокай то, чем не владеешь (stdlib HTTP — лучше httptest)
- Assert поведение, не число вызовов без нужды
- После смены SQL — интеграционный тест важнее sqlmock

## Типичные вопросы на собесе
- Почему mock БД не ловит deadlock/unique violation?
- Чем sqlmock хуже testcontainers для Outbox?
- Как тестировать Redis cache-aside без флаков?
