Путь запроса в Event Horizon
text
[React Client] (браузер, порт 3000/5173)
       │
       │ HTTP запрос (JSON)
       ▼
[Balancer :8079]  ←── 1. Принимает запрос
       │
       │ 2. Выбирает gateway (round-robin)
       │    (без Consul — просто циклически)
       ▼
[Gateway-1 :8081]  ←── 3. Принимает HTTP запрос
[Gateway-2 :8082]
[Gateway-3 :8083]
       │
       │ 4. Проверяет JWT (если нужно)
       │ 5. Преобразует HTTP → gRPC
       │ 6. Определяет, какой сервис нужен
       ▼
┌──────────────┼──────────────┐
│              │              │
▼              ▼              ▼
[Auth :50051]  [Game :50052]  [Billing :50053]  [Leaderboard :50054]
       │              │              │              │
       │              │              │              │
       ▼              ▼              ▼              ▼
[PG :5460]     [PG :5461]     [PG :5462]     [PG :5463] + [Redis :6382]
  (users)        (scores)       (balances)       (leaderboard)
       │              │              │              │
       │              │              │              │
       └──────────────┼──────────────┴──────────────┘
                      │
                      ▼
                 [NATS :4222]  ←── События между сервисами
                 (score.updated, user.registered)
                      │
                      ▼
           [Leaderboard] подписан на NATS
                      │
                      ▼
           [Redis :6382] обновляет топ
                      │
                      ▼
           [WebSocket] отправляет обновление клиенту
Детально по каждому этапу
1. Клиент → Balancer
Порт: 8079

Протокол: HTTP

Что делает: балансирует нагрузку между 3 gateway (round-robin)

Как узнаёт о живых gateway: без Consul — просто циклически перебирает, если один недоступен — ошибка

2. Balancer → Gateway
Порты: 8081, 8082, 8083

Протокол: HTTP

Что делает:

Проверяет JWT (если есть)

Определяет, какой сервис нужен по пути (/api/auth/..., /api/game/...)

Преобразует HTTP → gRPC

3. Gateway → Сервисы
Порты: 50051 (auth), 50052 (game), 50053 (billing), 50054 (leaderboard)

Протокол: gRPC

Что делает: вызывает нужный метод сервиса

4. Сервисы → БД
PostgreSQL: 5460-5463 (основные данные)

Redis: 6379-6382 (кэш, сессии, leaderboard)

5. Сервисы → NATS
Порт: 4222

Что делает: отправляет события (user.registered, score.updated)

6. NATS → Leaderboard
Что делает: leaderboard подписан на score.updated, получает события и обновляет Redis

7. Leaderboard → WebSocket
Что делает: отправляет обновления всем подключённым клиентам

8. Мониторинг (параллельно)
Prometheus :9090 — собирает метрики со всех сервисов

Grafana :3000 — дашборды

Jaeger :16686 — трассировка

---

AFTER VOICE MESSAGE

---

1. Путь запроса — HTTP → Balancer → Gateway → gRPC → NATS
Ты правильно понял, но есть нюанс:

text
Клиент (React :5173)
    │ HTTP
    ▼
Balancer (:8079) ← самописный Least Connections
    │ HTTP
    ▼
Gateway-1 (:8081) или Gateway-2 (:8082) или Gateway-3 (:8083)
    │ HTTP → gRPC (преобразование)
    ▼
┌──────────────┼──────────────┐
│              │              │
▼              ▼              ▼
Auth (:50051)  Game (:50052)  Billing (:50053)  Leaderboard (:50054)
│              │              │                  │
│ gRPC         │ gRPC         │ gRPC             │ gRPC
▼              ▼              ▼                  ▼
PostgreSQL     PostgreSQL     PostgreSQL         Redis + PostgreSQL
(5460)         (5461)         (5462)             (5463 + 6382)
Gateway НЕ ОБРАЩАЕТСЯ к NATS напрямую. Gateway — это входная точка (API Gateway). Он:

Принимает HTTP-запрос

Проверяет JWT

Вызывает gRPC-метод нужного сервиса

Сервис уже сам решает — публиковать событие в NATS или нет

NATS — это шина для событий МЕЖДУ сервисами, а не для gateway.

2. Как общаются сервисы между собой
Правильно: Сервисы общаются через NATS (событийная шина), а не напрямую по gRPC.

Пример:

Пользователь отправил рекорд → Game сервис сохраняет в БД

Game сервис публикует событие score.updated в NATS

Billing подписан на score.updated → начисляет лампочки/билетики

Leaderboard подписан на score.updated → обновляет топ

Gateway подписан на score.updated → отправляет WebSocket клиентам

text
[Game] → NATS (:4222) → [Billing]
                      → [Leaderboard]
                      → [Gateway] → WebSocket → Клиент
Сервисы НЕ общаются напрямую по gRPC между собой. Только через NATS.

3. NATS — репликация и кластер
Как работает NATS кластер:

NATS поддерживает кластеризацию «из коробки»

Настраивается через -cluster флаг с указанием адресов других нод

Если 1 нода падает — другие подхватывают трафик

Данные хранятся в JetStream с репликацией (можно настроить replicas: 3)

Для 2 серверов:

yaml
nats-1:
  command: ["-js", "-cluster", "nats://0.0.0.0:6222", "-routes", "nats://nats-2:6222"]

nats-2:
  command: ["-js", "-cluster", "nats://0.0.0.0:6222", "-routes", "nats://nats-1:6222"]
Для k3s — проще через Helm:

bash
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm install nats nats/nats --set cluster.enabled=true --set cluster.replicas=3
4. Rate Limiter — когда включать
Сейчас: закомментирован → правильно, не мешает разработке.

Когда включить:

На проде (или стейджинге)

Когда будет нагрузка > 100 RPS

Когда появятся реальные пользователи

Как настроить:

go
// В services/gateway/cmd/main.go
// Раскомментировать:
limiter := ratelimit.NewRateLimiter(rdb)
r.Use(middleware.RateLimitMiddleware(limiter))
Настройки в internal/ratelimit/limiter.go:

AllowSubmit — 10 запросов/сек на пользователя

AllowLogin — 5 запросов/сек с IP

AllowWebSocket — 100 соединений/мин с IP

5. Redis и PostgreSQL — структура
Сейчас:

Каждый сервис имеет свой Redis (6379-6382)

Каждый сервис имеет свой PostgreSQL (5460-5463)

Правильно ли это? Да, для микросервисов — идеально. Каждый сервис владеет своей БД.

Мастер + слейвы:

1 мастер на запись

3 слейва на чтение

Это имеет смысл при нагрузке > 1000 RPS

Пока что 1 БД на сервис — достаточно

Redis кэш:

Сервис сначала проверяет Redis

Если нет — идёт в PostgreSQL

Результат кэшируется

6. 50 серверов — это много или нормально?
Для старта: 1 сервер с Docker Compose — достаточно.

50 серверов — это когда у тебя 1 000 000+ пользователей в месяц.

Реалистичный план для MVP:

1-3 сервера (dev/stage/prod)

Docker Compose на каждом

Потом переезд на k3s (3-5 нод)

Не нужно 50 серверов. Это оверхед.

7. Мониторинг и метрики
Сервис	gRPC	Metrics	Что мониторим
Auth	50051	9091	JWT ошибки, регистрации
Game	50052	9092	Количество игр, очки
Billing	50053	9093	Транзакции, балансы
Leaderboard	50054	9094	Обновления топа
Gateway	8081-8083	9095-9097	RPS, latency, ошибки
Balancer	8079	9098	Активные соединения
NATS порт 8222 — это HTTP-мониторинг, но он отдаёт JSON, не Prometheus-формат. Поэтому мы его не используем в Prometheus.

8. Дополнительные сервисы
Сервис	Назначение	БД
Notification	Push-уведомления, email	Redis + Firebase/FCM
Analytics	DAU, MAU, Retention	ClickHouse или PostgreSQL с ивентами
Payment	Реальные деньги	PostgreSQL + Redis
9. Вопросы, которые нужно уточнить
Как информация возвращается из кэша/БД?
→ Тем же маршрутом: сервис → gateway → balancer → клиент

Сервисы общаются только через NATS?
→ Да, только через NATS. Прямого gRPC между сервисами нет.

Ретраи с джиттером уже есть?
→ В billing добавили retry при подключении к NATS. Для gateway и других — нужно проверить.

Consul для сервис-дискавери?
→ Можно добавить, но пока не критично. Balancer использует round-robin.

---

✅ Вопросы, на которые я ответил:
№	Вопрос	Ответ
1	Какой путь запроса от клиента до БД?	✅ Расписал всю цепочку: React → Balancer → Gateway → gRPC → Service → БД
2	Balancer — round-robin или least connections?	✅ Least Connections (самописный)
3	Как настроить Rate Limiter?	✅ Раскомментировать в gateway, настройки в limiter.go
4	Когда включать Rate Limiter?	✅ На проде, когда >100 RPS
5	Gateway обращается к NATS или к сервисам?	✅ Gateway → gRPC → сервисы, не NATS
6	Как общаются сервисы между собой?	✅ Через NATS (событийная шина)
7	Как реплицируется NATS?	✅ Через кластеризацию (-cluster, -routes)
8	Что за порт 8222 у NATS?	✅ HTTP-мониторинг (JSON), не Prometheus
9	Нужно ли 50 серверов?	✅ Нет, для старта 1-3, потом k3s
10	Мастер + слейвы для БД — имеет смысл?	✅ Да, при нагрузке >1000 RPS
11	Как информация возвращается из кэша/БД?	✅ Тем же маршрутом обратно
12	Сервисы только через NATS общаются?	✅ Да, только через NATS
13	Ретраи с джиттером уже есть?	✅ В billing добавили, в других надо проверить
14	Добавлять Consul?	✅ Можно, но пока не критично

❓ Если я что-то упустил — скажи, я дополню.

А теперь коротко, как это всё работает (для памяти):

text
Клиент (React)
    │ HTTP
    ▼
Balancer (:8079) — Least Connections
    │ HTTP
    ▼
Gateway (:8081/8082/8083) — JWT проверка, HTTP→gRPC
    │ gRPC
    ▼
Сервис (Auth/Game/Billing/Leaderboard)
    │
    ├──► PostgreSQL (своя БД)
    ├──► Redis (кэш)
    └──► NATS (публикация событий)
            │
            ▼
         Другие сервисы (подписка на события)
            │
            ▼
         Gateway → WebSocket → Клиент


 УТОЧНЕНО:

                    [React Client] (браузер, порт 3000/5173)
                        │
                        │ HTTP запрос (JSON)
                        ▼
                    [Balancer :8079]  ←── 1. Принимает запрос
                        │
                        │ 2. Выбирает gateway с НАИМЕНЬШИМ КОЛИЧЕСТВОМ АКТИВНЫХ СОЕДИНЕНИЙ
                        │    (Least Connections, самописный)
                        ▼
                    [Gateway-1 :8081]  ←── 3. Принимает HTTP запрос
                    [Gateway-2 :8082]
                    [Gateway-3 :8083]
                        │
                        │ 4. Проверяет JWT (если нужно)
                        │ 5. Преобразует HTTP → gRPC
                        │ 6. Определяет, какой сервис нужен
                        ▼
        ┌──────────────┼──────────────┐
        │              │              │
        ▼              ▼              ▼
[Auth :50051]  [Game :50052]  [Billing :50053]  [Leaderboard :50054]
       │              │              │              │
       │              │              │              │
       ▼              ▼              ▼              ▼
[PG :5460]     [PG :5461]     [PG :5462]     [PG :5463] + [Redis :6382]
  (users)        (scores)       (balances)       (leaderboard)
       │              │              │              │
       │              │              │              │
       └──────────────┼──────────────┴──────────────┘
                      │
                      ▼
                 [NATS :4222]  ←── События между сервисами
                 (score.updated, user.registered)
                      │
                      ▼
           [Leaderboard] подписан на NATS
                      │
                      ▼
           [Redis :6382] обновляет топ
                      │
                      ▼
           [WebSocket] отправляет обновление клиенту
Код Balancer — Least Connections:
go
func (lb *LeastConnBalancer) getLeastConnBackend() *Backend {
    var selected *Backend
    var minConns int32 = 2147483647

    for _, b := range lb.backends {
        conns := atomic.LoadInt32(&b.ActiveConns)
        if conns < minConns {
            minConns = conns
            selected = b
        }
    }
    return selected
}
Как работает:

При каждом запросе balancer смотрит ActiveConns у всех бекендов

Выбирает тот, у которого меньше всего активных соединений

Увеличивает счётчик на 1 перед проксированием

Уменьшает на 1 после завершения запроса        