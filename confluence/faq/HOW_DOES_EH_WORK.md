ФИНАЛЬНАЯ ДОКУМЕНТАЦИЯ
Проект Event Horizon (v1.0)
1. Цель проекта
Event Horizon — это pet-проект, который я делаю, чтобы:

Отработать архитектуру микросервисов на реальном проекте.

Разобраться в DevOps-подходе: CI/CD, оркестрация, мониторинг.

Подготовиться к собеседованиям, имея в портфолио живую систему, которую можно показать и объяснить.

Проект — игровой бэкенд с элементами:

Аутентификация (JWT)

Игровая логика (очки, рекорды)

Внутриигровая валюта (лампочки/билетики)

Таблица лидеров (топ-10 в реальном времени)

Уведомления (планируются)

Аналитика (планируется)

2. Архитектура (общая схема)
text
            [React Client :5173]
                │ HTTP (JSON)
                ▼
            [Balancer :8079] — самописный, Least Connections
                │ HTTP
                ▼
            [Gateway 1-3 :8081-8083] — JWT, HTTP→gRPC
                │ gRPC
                ▼
┌──────────────┼──────────────┐
│              │              │
▼              ▼              ▼
Auth :5051     Game :5052     Billing :5053     Leaderboard :5054
│              │              │                  │
▼              ▼              ▼                  ▼
PG :5460       PG :5461       PG :5462          PG :5463 + Redis :6382
(users)        (scores)       (balances)        (leaderboard)
│              │              │                  │
└──────────────┼──────────────┴──────────────────┘
               │
               ▼
          [NATS :4222] — событийная шина (score.updated, user.registered)
               │
               ▼
    Leaderboard подписан → обновляет Redis → WebSocket → клиент

3. Компоненты и порты
Компонент	Протокол	Порт	Назначение
Balancer	HTTP	8079	Least Connections, самописный
Gateway 1	HTTP	8081	Входная точка, JWT, роутинг
Gateway 2	HTTP	8082	Входная точка, JWT, роутинг
Gateway 3	HTTP	8083	Входная точка, JWT, роутинг
Auth	gRPC	5051	Аутентификация, JWT
Game	gRPC	5052	Игровая логика, очки
Billing	gRPC	5053	Внутриигровая валюта
Leaderboard	gRPC	5054	Топ-10, WebSocket
NATS	TCP	4222	Событийная шина (JetStream)
NATS (мониторинг)	HTTP	8222	JSON-метрики (не для Prometheus)
PostgreSQL (Auth)	TCP	5460	Пользователи
PostgreSQL (Game)	TCP	5461	Рекорды, счета
PostgreSQL (Billing)	TCP	5462	Балансы, транзакции
PostgreSQL (Leaderboard)	TCP	5463	Топ-10
Redis (Auth)	TCP	6379	Кеш, JWT сессии
Redis (Game)	TCP	6380	Кеш
Redis (Billing)	TCP	6381	Кеш
Redis (Leaderboard)	TCP	6382	Кеш, топ-10
Prometheus	HTTP	9090	Метрики
Grafana	HTTP	3000	Дашборды
Jaeger	HTTP	16686	Трассировка
OTLP (HTTP)	HTTP	4318	OpenTelemetry
OTLP (gRPC)	gRPC	4317	OpenTelemetry
4. Взаимодействие сервисов
4.1. Клиент → Сервис (синхронный запрос)
text
1. Клиент → Balancer :8079 (HTTP)
2. Balancer выбирает Gateway с наименьшим количеством активных соединений
3. Balancer → Gateway :8081-8083 (HTTP)
4. Gateway проверяет JWT (если есть)
5. Gateway преобразует HTTP → gRPC
6. Gateway → нужный сервис (Auth/Game/Billing/Leaderboard) :5051-5054 (gRPC)
7. Сервис → БД (PostgreSQL) или Redis
8. Ответ → тем же маршрутом обратно
4.2. Сервисы → Сервисы (асинхронные события)
text
1. Game сохраняет рекорд в PostgreSQL
2. Game → NATS :4222 (публикует событие score.updated)
3. NATS → Leaderboard (подписка)
4. Leaderboard обновляет Redis :6382
5. Leaderboard → WebSocket → клиент (push-уведомление)
Сервисы НЕ общаются друг с другом напрямую по gRPC!
Только через NATS.

5. Балансировщик (самописный)
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

Самописный, без Consul (пока).

Метрики на :9098.

6. Gateway
Задачи:

Принимает HTTP-запросы от клиента.

Проверяет JWT (если требуется).

Определяет, какой сервис нужен по пути:

/api/auth/* → Auth

/api/game/* → Game

/api/billing/* → Billing

/api/leaderboard/* → Leaderboard

Преобразует HTTP → gRPC.

Вызывает метод сервиса.

Rate Limiter:

Сейчас закомментирован (не мешает разработке).

Включить на проде, когда нагрузка станет > 100 RPS.

Настройки в internal/ratelimit/limiter.go:

AllowSubmit — 10 запросов/сек на пользователя

AllowLogin — 5 запросов/сек с IP

AllowWebSocket — 100 соединений/мин с IP

7. NATS (событийная шина)
Роль: передача событий между сервисами.

Порт: 4222 (клиентский), 8222 (HTTP-мониторинг).

JetStream: включён (сообщения хранятся персистентно).

Кластер:

NATS поддерживает кластеризацию "из коробки".

Для прода — минимум 3 ноды (RAFT).

Настройка: -cluster nats://0.0.0.0:6222 -routes nats://other-node:6222.

События (примеры):

user.registered — Auth → другие сервисы

score.updated — Game → Leaderboard, Billing

payment.completed — Payment → Billing, Analytics

8. Базы данных
8.1. PostgreSQL (каждому сервису своя)
Сервис	БД	Порт	Таблицы
Auth	users	5460	users, sessions
Game	scores	5461	scores, games
Billing	balances	5462	balances, transactions
Leaderboard	leaderboard	5463	top_players, history
Репликация:

План: 1 мастер на запись + 3 слейва на чтение.

Включить при нагрузке > 1000 RPS.

Сейчас — 1 БД на сервис.

8.2. Redis (кеш)
Сервис	Порт	TTL	Назначение
Auth	6379	15 мин	JWT, сессии
Game	6380	5 мин	Кеш игровых данных
Billing	6381	5 мин	Кеш балансов
Leaderboard	6382	1 мин	Топ-10 (обновляется часто)
Схема кеширования (Cache-Aside):

Сервис проверяет Redis.

Если есть — возвращает из кеша.

Если нет — идёт в PostgreSQL, записывает в Redis.

При обновлении — инвалидирует кеш.

9. Мониторинг
9.1. Метрики
Сервис	Метрики	Что собираем
Auth	:9091	JWT errors, registration, login
Game	:9092	Score updates, games played
Billing	:9093	Transactions, balances
Leaderboard	:9094	Top updates, WS connections
Gateway 1-3	:9095-9097	RPS, latency, HTTP errors
Balancer	:9098	Active connections
NATS	:8222	JSON (не используется в Prometheus)
9.2. Инструменты
Инструмент	Порт	Назначение
Prometheus	9090	Сбор метрик
Grafana	3000	Дашборды
Jaeger	16686	Трассировка
OTLP (HTTP)	4318	OpenTelemetry
OTLP (gRPC)	4317	OpenTelemetry
Примечание: NATS на :8222 отдаёт JSON, не Prometheus-формат.
Чтобы собирать метрики — нужен NATS Exporter или встроенный /metrics.

10. DevOps
10.1. Сейчас
Docker Compose — инфраструктура (PostgreSQL, Redis, NATS, Prometheus, Grafana, Jaeger).

Локальные бинарники — сервисы (Auth, Game, Billing, Leaderboard, Gateway).

Makefile — запуск (make start-services).

10.2. План (Ansible + k3s)
Ansible:

Автоматизация деплоя бинарников на VM.

Плейбуки для копирования бинарников и перезапуска systemd-сервисов.

k3s:

Лёгкий Kubernetes (можно поднять на одной VM).

Helm-чарты для микросервисов.

Инфраструктура — либо в Docker Compose, либо в Helm (StatefulSet).

CI/CD:

GitHub Actions → сборка Docker-образов.

Ansible → деплой на VM.

Нагрузочное тестирование (k6) после каждого деплоя.

10.3. Почему не 50 серверов?
Для старта — 1-3 сервера достаточно.
50 серверов — это уровень 1M+ пользователей в месяц.
Пока проект в разработке — Docker Compose + 1 VM — идеально.

11. Будущие сервисы
Сервис	Назначение	БД / Хранилище
Notification	Push-уведомления, email, SMS, Telegram	Firebase FCM, Redis
Analytics	DAU, MAU, Retention, события	ClickHouse (или PostgreSQL)
Payment	Реальные деньги, Boosty, вебхуки	PostgreSQL, Redis
Все новые сервисы:

Общаются через NATS.

Имеют свой Redis и БД.

Запускаются в 2 экземплярах (основной + резервный).

12. Ответы на частые вопросы
Вопрос	Ответ
Gateway обращается к NATS?	Нет, только к сервисам через gRPC.
Как сервисы общаются между собой?	Только через NATS (событийная шина).
Как информация возвращается из кеша/БД?	Тем же маршрутом обратно (сервис → Gateway → клиент).
Когда включать Rate Limiter?	На проде, при >100 RPS.
Нужно ли 50 серверов?	Нет, для старта 1-3, потом k3s.
Мастер + слейвы для БД?	Да, при >1000 RPS.
Ретраи с джиттером уже есть?	В Billing — да, в других — проверить.
Добавлять Consul?	Можно, но пока не критично.
Как реплицируется NATS?	Через кластеризацию (-cluster, -routes).
13. Итог (для себя и для собеса)
Event Horizon — это:

Микросервисная архитектура с чёткими границами.

Событийная шина (NATS) для асинхронного общения.

Самописный балансировщик (Least Connections).

JWT-аутентификация.

Кеширование (Redis) + персистентность (PostgreSQL).

Мониторинг (Prometheus, Grafana, Jaeger, OpenTelemetry).

DevOps: Docker Compose → Ansible + k3s (план).

Готовность к масштабированию (но с осознанием, что 50 серверов — для крупного бизнеса).

Что важно для собеса:

Я знаю, почему выбрал такой стек.

Я понимаю компромиссы (почему не K8s сейчас, почему не 50 серверов).

Я умею объяснять каждый компонент.

У меня есть работающий код, который можно показать.