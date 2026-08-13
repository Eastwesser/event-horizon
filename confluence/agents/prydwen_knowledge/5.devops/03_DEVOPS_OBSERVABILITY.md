# Observability: метрики, логи, трейсы (Event Horizon)

Наблюдаемость отвечает на: что сломалось, где, почему, насколько плохо. Три столпа — **metrics, logs, traces**. Senior на собесе red_mad_robot связывает их с SLO и алертами. В EH: Prometheus, Grafana-стек идей, Jaeger/OTel-tracing в platform pkg, `/health` и `/ready` на metrics HTTP.

## Три столпа

- **Metrics**: агрегаты во времени (RPS, latency, error rate, lag). Дёшевы для алертинга.
- **Logs**: детальный контекст события/ошибки. Нужны поля (service, trace_id, user_id hash).
- **Traces**: путь запроса через gateway → gRPC → DB/NATS. Показывают, кто тормозит.
- Без correlation: три столпа превращаются в три свалки. **trace_id / correlation_id** сквозь логи и spans.

## Метрики и Prometheus

- Модель: pull scrape HTTP `/metrics` (Prometheus exposition format).
- Типы: Counter, Gauge, Histogram/Summary (latency buckets).
- RED: Rate, Errors, Duration для сервисов; USE: Utilization, Saturation, Errors для ресурсов.
- Бизнес-метрики: покупки, spend/refund, outbox lag, consumer lag NATS/Kafka.
- EH: у сервисов отдельный `METRICS_PORT`; `platform/pkg/metrics` (grpc/business helpers).
- `deployments/prometheus/prometheus.yml` + `alerts.yml`; Alertmanager — `deployments/alertmanager/`.

## Логи

- Structured JSON предпочтительнее plain text на проде.
- Уровни: debug/info/warn/error; не логировать секреты/PII без маскирования.
- `platform/pkg/logger` — единый подход; interceptor логирует gRPC вызовы (`pkg/interceptor` / per-service).
- Централизация: Loki/ELK; на lab — docker logs + grep по trace_id.
- Cardinality: не делай label/metric из необработанного user_id миллионами значений.

## Трейсы и Jaeger / OpenTelemetry

- Span = единица работы; Trace = дерево span'ов.
- Контекст в gRPC metadata / W3C traceparent.
- `platform/pkg/tracing` — tracer + trace_id helpers.
- Jaeger / Tempo / OTel Collector — бэкенд хранения и UI.
- Сэмплирование: на проде не всегда 100% (стоимость); ошибки — сохранить.
- Главный вопрос трейса: «где время?» — сеть, БД, брокер, сериализация.

## /health и /ready (обязательный стандарт EH)

- **`GET /health`**: liveness — процесс жив, без глубоких проверок.
- **`GET /ready`**: readiness — ping критичных зависимостей (PG, Redis, NATS/Kafka по роли сервиса).
- Слушать на metrics HTTP, не смешивать с бизнес-портом без нужды.
- K8s/compose используют эти эндпоинты для probes/healthcheck.
- Примеры: `services/*/internal/app/app.go` (fulfillment, notification, nats-hub, …).

## Алерты: что реально будить

- Error rate / gRPC `Unavailable` spike.
- Latency p95/p99 выше SLO.
- Consumer lag (Kafka/NATS) растёт.
- Outbox unpublished backlog.
- Saturation: CPU throttling, PG connections near MaxOpen(25).
- Не алертить на всё подряд — fatigue убивает реакцию.

## Как дебажить инцидент (шпаргалка)

1. Алерт / симптом (5xx на gateway, lag fulfillment).
2. Метрики: какой сервис/метод красный.
3. Trace: медленный span / ошибка downstream.
4. Логи по trace_id: корневая причина.
5. Зафиксировать: регресс релиза, зависимость, saturation.

## Антипаттерны

- Только логи без метрик (алертить по grep невозможно стабильно).
- Метрики без labels service/method.
- Логировать каждый successful request на info в hot path без сэмпла.
- Ready зависит от опциональной зависимости → сервис вечно NotReady.
- Нет единого trace_id на async path (после NATS publish — класть id в payload).

## Типичные вопросы на собесе

- Чем отличаются metrics, logs и traces и как их связать?
- Что такое RED и какие метрики повесите на gRPC сервис?
- Зачем histogram buckets для latency?
- Чем `/health` отличается от `/ready` и как это стыкуется с k8s probes?
- Как искать причину медленной покупки end-to-end в EH?
- Какие алерты обязательны для outbox/consumer?
- Как не взорвать cardinality в Prometheus?
