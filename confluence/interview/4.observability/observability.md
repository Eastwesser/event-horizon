📈 4. МОНИТОРИНГ И НАБЛЮДАЕМОСТЬ


❓ Как настроил Consumer Lag в Grafana?
Ответ: Consumer Lag — это разница между последним сообщением в Stream и последним обработанным потребителем.

Метрики из NATS Exporter:

nats_consumer_lag — показывает отставание

nats_consumer_msgs_pending — неподтвержденные сообщения

В Prometheus:

yaml
# Забираем метрики с NATS Exporter
- job_name: 'nats'
  static_configs:
    - targets: ['nats-exporter:7777']
В Grafana:

promql
# Consumer Lag
nats_consumer_lag{consumer="billing-durable"}


❓ Как работает распределенная трассировка в Jaeger?
Ответ: Каждый запрос получает уникальный trace_id, который передается между сервисами. Каждый сервис добавляет свой span с временем выполнения.

Как передаю context:

go
// В Gateway
ctx := context.Background()
ctx = jaeger.Inject(ctx)

// В gRPC клиенте
_, err := client.SubmitScore(ctx, req)

// В сервисе
func (s *GameService) SubmitScore(ctx context.Context, req *pb.SubmitRequest) {
    // Автоматически создается span
}
В Event Horizon: Все сервисы инициализируют Jaeger при старте.


❓ Почему p99 важнее средней?
Ответ: Средняя задержка скрывает выбросы. Например:

100 запросов по 1ms = средняя 1ms

99 запросов по 1ms + 1 запрос по 10s = средняя 100ms, но p99 = 10s

В Event Horizon: Я смотрю на p95 и p99, а не на среднюю.


❓ Как обнаружить утечку памяти?
Ответ:

Добавляю pprof: import _ "net/http/pprof"

Снимаю heap-профиль:

bash
go tool pprof http://localhost:6060/debug/pprof/heap
Смотрю топ аллокаций:

text
(pprof) top5
В Prometheus смотрю:

promql
go_memstats_alloc_bytes
В Event Horizon: Все сервисы имеют pprof на порту 6060.


❓ Паникуешь ли при 100% CPU?
Ответ: Нет, это не всегда плохо. 100% CPU может означать:

Нормальная работа — если сервис утилизирует все ядра.

Проблема — если CPU растет без увеличения RPS.

Как проверить:

Сравнить CPU с RPS

Взять CPU-профиль:

bash
go tool pprof http://localhost:6060/debug/pprof/profile


❓ Как отследить медленный запрос через 5 сервисов?
Ответ: Через Jaeger:

Найти trace_id медленного запроса

Посмотреть, какой span дольше всего

Увидеть, в каком сервисе проблема

В Event Horizon: Все запросы от Gateway до Billing/Shop трассируются.


❓ RED и USE метрики?
Ответ:

RED — Rate, Errors, Duration (для запросов к сервису)

USE — Utilization, Saturation, Errors (для ресурсов)

В Event Horizon:

RED для каждого сервиса: RPS, ошибки, latency

USE для БД: CPU, память, IO


❓ Как настроить алерт на error rate?
Ответ:

promql
# Алерт, если ошибки > 1% за 5 минут
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.01
В Prometheus:

yaml
groups:
- name: api
  rules:
  - alert: HighErrorRate
    expr: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.01
    for: 5m
    annotations:
      summary: "High error rate"


❓ Структурированные логи в Go?
Ответ:

go
import "log/slog"

slog.Info("user purchased item",
    slog.String("user_id", userID),
    slog.String("item_id", itemID),
    slog.Int("price", price),
)
Почему лучше:

Можно парсить и искать по полям

Легко интегрировать с ELK/Loki

Не надо писать парсеры для текста

В Event Horizon: Я использую slog для структурированных логов.


❓ Как выглядят трассировки WebSocket?
Ответ: WebSocket трассировки отличаются:

Одно соединение → множество спанов (сообщения)

Долгоживущие спаны (вместо коротких HTTP)

В Event Horizon: Gateway пишет span при подключении и при каждом сообщении.

