# Нагрузочное тестирование (highload)

## Реалистичные цифры EH
Из метрик проекта: средний RPS ~17, пик ~35, DAU ~10k — **не** 100k RPS. Собеседователю важнее честность и методология, чем фантазия.

## Инструменты
- **k6** — HTTP сценарии (`deployments/k6/`, `scripts/loadtest_k6.js`)
- **vegeta / hey** — быстрый smoke RPS
- **pprof + Prometheus** — где горит во время нагрузки

## Методика
1. Smoke: `/health` `/ready` всех сервисов
2. Baseline: p50/p95/p99 на целевом сценарии (login → shop → purchase)
3. Step load: +RPS пока p95 < SLO (например 200–500 ms)
4. Soak: 30–60 мин на поиск утечек
5. Failover: убить Billing → circuit breaker → 503, не зависание

## Узкие места, которые ждут
- NATS consumer lag / outbox backlog
- Postgres connection pool exhaustion (MaxConns=25)
- Redis hot keys
- Gateway без таймаутов/CB

## Make
`make test-k6` — опциональный шаг; `make test-smoke` при поднятом compose.

## Типичные вопросы на собесе
- Чем нагрузка отличается от стресс/soak?
- Как выбрать SLO по latency?
- Что смотреть в Grafana при росте p99?
- Как доказать, что узкое место — БД, а не сеть?
