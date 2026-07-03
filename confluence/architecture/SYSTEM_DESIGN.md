# Архитектура Event Horizon

> **Версия:** 1.0  
> **Дата:** 25.06.2026  
> **Автор:** Денис Матвеев (Eastwesser)  
> **Статус:** Актуально, система в разработке

---

## 📌 Основные принципы

- **Микросервисная архитектура** — каждый сервис отвечает за свою бизнес-логику.
- **Событийно-ориентированная** — сервисы общаются через NATS (асинхронно).
- **Синхронное взаимодействие** — через gRPC (клиент → сервис).
- **Каждый сервис владеет своей БД** — никаких общих таблиц.
- **Кеширование** — Redis (Cache-Aside паттерн).
- **Мониторинг** — Prometheus + Grafana + Jaeger.

---

## 🧩 Общая архитектура (все компоненты + порты)

```mermaid
graph TD
    Client[React Client :5173] -->|HTTP| Balancer[Balancer :8079]
    Balancer -->|Least Connections| G1[Gateway-1 :8081]
    Balancer -->|Least Connections| G2[Gateway-2 :8082]
    Balancer -->|Least Connections| G3[Gateway-3 :8083]
    
    G1 -->|gRPC| Auth[Auth :5051]
    G1 -->|gRPC| Game[Game :5052]
    G1 -->|gRPC| Billing[Billing :5053]
    G1 -->|gRPC| LB[Leaderboard :5054]
    
    Auth --> PG1[PostgreSQL :5460]
    Game --> PG2[PostgreSQL :5461]
    Billing --> PG3[PostgreSQL :5462]
    LB --> PG4[PostgreSQL :5463]
    
    Auth --> R1[Redis :6379]
    Game --> R2[Redis :6380]
    Billing --> R3[Redis :6381]
    LB --> R4[Redis :6382]
    
    Game -->|publish score.updated| NATS[NATS :4222]
    NATS -->|subscribe| LB
    NATS -->|subscribe| Billing
    NATS -->|subscribe| G1
    
    LB -->|WebSocket| Client
    
    Auth -->|:9091| Prometheus[Prometheus :9090]
    Game -->|:9092| Prometheus
    Billing -->|:9093| Prometheus
    LB -->|:9094| Prometheus
    G1 -->|:9095| Prometheus
    G2 -->|:9096| Prometheus
    G3 -->|:9097| Prometheus
    Balancer -->|:9098| Prometheus
    
    Prometheus --> Grafana[Grafana :3000]
    Services -->|OTLP| Jaeger[Jaeger :16686]
📦 Компоненты и их порты
Компонент	Протокол	Порт	Назначение
Balancer	HTTP	8079	Least Connections, самописный
Gateway-1	HTTP	8081	Входная точка API, JWT, роутинг
Gateway-2	HTTP	8082	Входная точка API, JWT, роутинг
Gateway-3	HTTP	8083	Входная точка API, JWT, роутинг
Auth	gRPC	5051	Регистрация, логин, JWT
Game	gRPC	5052	Игровая логика, очки, рекорды
Billing	gRPC	5053	Внутриигровая валюта (лампочки/билетики)
Leaderboard	gRPC	5054	Топ игроков
NATS	TCP	4222	Событийная шина (JetStream)
NATS (мониторинг)	HTTP	8222	JSON-метрики (не Prometheus)
PostgreSQL (Auth)	TCP	5460	Пользователи
PostgreSQL (Game)	TCP	5461	Рекорды
PostgreSQL (Billing)	TCP	5462	Балансы, транзакции
PostgreSQL (Leaderboard)	TCP	5463	Топ игроков
Redis (Auth)	TCP	6379	Кеш JWT, сессии
Redis (Game)	TCP	6380	Кеш игровых данных
Redis (Billing)	TCP	6381	Кеш балансов
Redis (Leaderboard)	TCP	6382	Топ-10 в реальном времени
Prometheus	HTTP	9090	Сбор метрик
Grafana	HTTP	3000	Дашборды
Jaeger	HTTP	16686	Трассировка
OTLP (HTTP)	HTTP	4318	OpenTelemetry
OTLP (gRPC)	gRPC	4317	OpenTelemetry
🔄 Путь запроса (синхронный)
Клиент → Balancer → Gateway → Сервис → БД → ответ тем же маршрутом.

Что делает Gateway:

Проверяет JWT (если требуется)

Определяет сервис по URL (/api/auth/* → Auth, /api/game/* → Game и т.д.)

Превращает HTTP → gRPC

Вызывает нужный метод сервиса

⚡ Асинхронное взаимодействие (NATS)
Сервисы НЕ общаются напрямую по gRPC. Только через NATS.

События в NATS (сейчас):

user.registered — Auth → другие сервисы

score.updated — Game → Leaderboard, Billing, Gateway (WebSocket)

События в NATS (в планах):

shop.purchase — Shop → Billing, Analytics

payment.completed — Payment → Billing, Analytics

notification.send — Notification → пользователь

🗄️ Базы данных
PostgreSQL (каждому сервису — своя)
Сервис	БД	Порт	Таблицы
Auth	users	5460	users, sessions
Game	scores	5461	highscores
Billing	balances	5462	user_currencies, transactions
Leaderboard	leaderboard	5463	leaderboard_backup
Репликация (в будущем):

1 мастер на запись

3 слейва на чтение

Включить при нагрузке > 1000 RPS

Redis (кеш, сессии, топ)
Сервис	Порт	TTL	Назначение
Auth	6379	15 мин	JWT, сессии
Game	6380	5 мин	Кеш игровых данных
Billing	6381	5 мин	Кеш балансов
Leaderboard	6382	1 мин	Топ-10 в реальном времени
Схема кеширования (Cache-Aside):

Сервис проверяет Redis

Если есть — возвращает

Если нет — идёт в PostgreSQL

Записывает в Redis

При обновлении — инвалидирует кеш

🖥️ Мониторинг
Метрики (Prometheus)
Сервис	Порт	Что собираем
Auth	9091	JWT errors, registrations, logins
Game	9092	Games played, scores
Billing	9093	Transactions, balances
Leaderboard	9094	Top updates, WS connections
Gateway-1	9095	RPS, latency, HTTP errors
Gateway-2	9096	RPS, latency, HTTP errors
Gateway-3	9097	RPS, latency, HTTP errors
Balancer	9098	Active connections
Важно: NATS на 8222 отдаёт JSON, а не Prometheus-формат. Поэтому NATS не включён в Prometheus (пока).

Инструменты
Инструмент	Порт	Назначение
Prometheus	9090	Сбор метрик
Grafana	3000	Дашборды
Jaeger	16686	Трассировка
OTLP (HTTP)	4318	OpenTelemetry
OTLP (gRPC)	4317	OpenTelemetry
🔄 Балансировщик (самописный)
Алгоритм: Least Connections (наименьшее количество активных соединений).

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
Особенности:

Самописный, без Consul (пока)

Метрики на порту :9098

Работает с 3 gateway инстансами

🧠 Планируемые сервисы (будущее)
Сервис	gRPC	Метрики	БД	Назначение
Shop	5055	9099	PostgreSQL + Redis	Магазин (трата билетиков)
Notification	5056	9100	Firebase + Redis	Push, email, SMS, Telegram
Analytics	5057	9101	ClickHouse / PG	DAU, MAU, Retention
Payment	5058	9102	PostgreSQL + Redis	Реальные деньги, Boosty
Все новые сервисы:

Общаются через NATS

Имеют свою БД и Redis

Запускаются в 2 экземплярах (основной + резервный)

🚀 Развертывание
Сейчас (разработка)
bash
docker-compose -f deployments/docker-compose.cluster.yml up -d
make start-services
План (продакшен)
Ansible — деплой бинарников на VM

k3s — оркестрация

Helm — управление сервисами

GitHub Actions — CI/CD

📎 Ссылки
Репозиторий: https://github.com/Eastwesser/event-horizon

Документация: /confluence/

TODO: /confluence/history/2026-06/25.06.2026/TODO_LIST.md

## Архитектура

\```mermaid
graph TD
    Client[React Client :5173] -->|HTTP| Balancer[Balancer :8079]
    Balancer -->|Least Connections| G1[Gateway-1 :8081]
    Balancer -->|Least Connections| G2[Gateway-2 :8082]
    Balancer -->|Least Connections| G3[Gateway-3 :8083]
    
    G1 -->|gRPC| Auth[Auth :5051]
    G1 -->|gRPC| Game[Game :5052]
    G1 -->|gRPC| Billing[Billing :5053]
    G1 -->|gRPC| LB[Leaderboard :5054]
    
    Auth --> PG1[PostgreSQL :5460]
    Game --> PG2[PostgreSQL :5461]
    Billing --> PG3[PostgreSQL :5462]
    LB --> PG4[PostgreSQL :5463]
    
    Auth --> R1[Redis :6379]
    Game --> R2[Redis :6380]
    Billing --> R3[Redis :6381]
    LB --> R4[Redis :6382]
    
    Game -->|publish score.updated| NATS[NATS :4222]
    NATS -->|subscribe| LB
    NATS -->|subscribe| Billing
    NATS -->|subscribe| G1
    
    LB -->|WebSocket| Client
\```

\```mermaid
sequenceDiagram
    participant Client
    participant Balancer
    participant Gateway
    participant Service
    participant DB
    
    Client->>Balancer: HTTP /api/auth/register
    Balancer->>Gateway: HTTP (Least Connections)
    Gateway->>Gateway: JWT (если есть)
    Gateway->>Service: gRPC (Auth/Game/Billing/LB)
    Service->>DB: SQL (PostgreSQL / Redis)
    DB-->>Service: данные
    Service-->>Gateway: gRPC ответ
    Gateway-->>Balancer: HTTP ответ
    Balancer-->>Client: ответ
\```

\```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant Game
    participant NATS
    participant Leaderboard
    participant Billing
    participant WS
    
    Client->>Gateway: POST /game/submit
    Gateway->>Game: gRPC SubmitScore
    Game->>Game: сохранить рекорд
    Game->>NATS: publish score.updated
    NATS->>Leaderboard: subscribe
    NATS->>Billing: subscribe
    NATS->>Gateway: subscribe (WS)
    Leaderboard->>Leaderboard: обновить топ (Redis)
    Billing->>Billing: начислить валюту
    Gateway->>WS: broadcast update
    WS-->>Client: real-time обновление
\```

\```mermaid
graph LR
    subgraph Services
        Auth
        Game
        Billing
        LB[Leaderboard]
        G1[Gateway-1]
        G2[Gateway-2]
        G3[Gateway-3]
        Balancer
    end
    
    Auth -->|:9091| Prometheus[Prometheus :9090]
    Game -->|:9092| Prometheus
    Billing -->|:9093| Prometheus
    LB -->|:9094| Prometheus
    G1 -->|:9095| Prometheus
    G2 -->|:9096| Prometheus
    G3 -->|:9097| Prometheus
    Balancer -->|:9098| Prometheus
    
    Prometheus --> Grafana[Grafana :3000]
    Prometheus --> Alertmanager[Alertmanager :9093]
    Alertmanager --> Telegram[Telegram Bot]
    
    Services -->|OTLP| Jaeger[Jaeger :16686]
\```