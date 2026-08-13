# Интеграционные, e2e, нагрузка, smoke — шпаргалка senior

## Карта уровней

| Уровень | Что поднимаем | Скорость | Цель |
|---------|---------------|----------|------|
| Unit | ничего | мс | бизнес-логика |
| Integration | Postgres/Redis/NATS в Docker | секунды | SQL, кеш, миграции |
| E2E / smoke | compose cluster / gateway | минуты | happy-path API |
| Load (k6) | живой стенд | минуты+ | latency, RPS, ошибки |

На собесе важно уметь **развести** эти уровни и сказать, какой баг какой уровень ловит.

## Testcontainers (Go)

Идея: в тесте с тегом `integration` стартует контейнер Postgres/Redis/Kafka, прогоняются миграции, репозиторий бьёт в реальный драйвер.

Типичный скелет:
1. `testcontainers.GenericContainer` / Postgres module.
2. DSN из mapped port → `pgxpool`.
3. Apply migrations (тот же migrator, что в сервисе).
4. CRUD + транзакция + outbox insert — assert в таблице.
5. `t.Cleanup` / terminate контейнера.

В EH: pipeline `make test-unit` / `test-smoke` / `test-k6`; testcontainers — placeholder `-tags=integration` до стабильного Docker CI. На собесе говорите честно: «контракт готов, wire когда Docker в CI».

Плюсы vs shared staging DB: изоляция, воспроизводимость, нет «чужих» данных. Минусы: медленнее, нужен Docker socket.

## E2E

Сценарий сквозь Gateway: register/login → JWT → Shop purchase → Billing spend → событие NATS → History/Analytics ingest.

Проверяете контракты HTTP/gRPC и wiring, не каждую ветку service. Хрупкость: таймауты, порядок consumers, eventual consistency — assert с retry/poll, не сразу `SELECT`.

## Smoke

Короткий post-deploy чеклист: `/health`, `/ready` на metrics-портах, один authenticated GET, один критичный POST (покупка/баланс) на staging. Цель — «кластер жив», не coverage.

В EH: metrics HTTP `/health` (liveness) + `/ready` (ping зависимостей). Balancer :8079 → Gateway×3.

## k6 (нагрузка)

Скрипт: VU, stages (ramp-up → plateau → ramp-down), thresholds (`http_req_failed`, `p95`).

Что меряем для EH-подобных API:
- RPS на Gateway login / list items / purchase.
- Ошибки 429/503 (rate limit / circuit open) — ожидаемое поведение под давлением, не всегда баг.
- Деградация: рост p95 при открытии circuit на billing/shop.

Не путать load с soak (долгая устойчивость) и stress (поиск предела).

## Когда какой тест писать

- Баг в SQL/`FOR UPDATE` / optimistic `version` → integration.
- Неверный статус HTTP при RBAC → e2e через Gateway middleware.
- Регресс «сервис не стартует» → smoke.
- «Сколько выдержит shop purchase» → k6 на стенде с реалистичным NATS/PG.

## Антипаттерны

- Весь регресс только e2e (медленно, flake).
- Интеграционный тест без миграций (схема «на глаз»).
- k6 против ноутбука разработчика как «доказательство SLA».
- Sleep 5s вместо poll условия готовности.

## Связь с EH пайплайном

`HARDENING_STATUS`: unit + smoke + k6 в Makefile есть; integration tags — следующий шаг. Inventory — эталон для Outbox/Redis/tx в интеграционных сценариях.

## Типичные вопросы на собесе

1. Чем testcontainers лучше «общего» тестового Postgres в CI?
2. Как тестировать eventual consistency после Outbox → NATS?
3. Что входит в smoke после деплоя микросервисов?
4. Какие thresholds поставить в k6 для API с circuit breaker?
5. Почему e2e не заменяет unit на converters/JWT?
6. Как изолировать integration-теги в Go (`-tags=integration`)?
7. Что делать с flake из-за порядка JetStream consumers?
8. Когда 503 в нагрузке — ожидаемый сигнал, а не падение SLO?
