Ok. Can we do a loadtest now? Or run the tests. I need metrics, timings, business metrics, results, all the stuff.

Like here (rather old data):

Ключевые метрики и цифры:1. Mau: 100k2. Dau: 10k (10% mau) 3. N users per year: ~1mln new players4. Content per 1 user a day: 10-20 gaming sessions per user5. RPS (r, w, search):
 Для 10к DAU реальная нагрузка — 30-50 RPS.
 средняя 50 (пиковая 100)
 Read: 30-50 (<60),
 Write: 10-50 (<100),
 Search: 10-20 (<100)6. N simultaneous connections (net, ws):  Net/HTTP: 5000 (до 10к net соединений),  WebSockets: 2500 (до 5к ws соединений)7. Load: общая <70к По CPU: 40-60% (пиковая <70%) Ram: 50-60% (пиковая <70%) I/o: 30-50% Net: 100-200 Мбит/с8. Db weight + other 'storage instances' PostgreSQL: <1Tb Redis: 5-32Gb MongoDb: 100-500Gb (authors) Clickhouse: 1-5Tb (events)9. $ price, traffic and storage: 430$ в месяц на 5 серверов Selectel10. Latency (p50, p90, p95, p99):
 p50: < 50 ms
 p90: <150 ms
 p95: <200 ms
 p99: <300 ms11. Read/write ratio: 3 к 1 или 2 к 1 (чтения больше)
 (индексы не должны тормозить вставку, партиционирование по дням)12. Avg weight (size) of 1 query/object in bytes: <10kb13. Retention policy (n days of storaging data) сырые ивенты: 30 дней14. Peak load vs average: пик х2 от среднего


Golang metrics
 -goroutines: 444 (<1000 на одном сервере) 
 -gc cycles: <10 раз в минуту
 -heap: <512Мб   (min 3.7 Мб на сервере сервиса Game)
 -stack: - (autoflush?
)
 -mutex waits: если >1 сек, то это проблема

DB metrics:
 -connections: <100
 -slow queries (>200ms): <5%
 -deadlocks: 0
 -replication lag: <1сек

Графики в 3 часа ночи:  - P99 Latency на Auth/Gateway  - HTTP error rate (500 статусы кода)  - Free memory (RAM)
  - NATS JetStream Consumer Lag
    (лаг >1000, проблема с подпиской)

DevOps:
Delivery: Ansible + GitHub ActionsDeploy: docker compose + k3s

Базы данных (04.08.2026):

БД                — Объём    — Назначение                 — Инстансы

1) PostgreSQL (Auth)       — 10-50Gb     — Пользователи, сессии              — 1 мастер, 1 реплика
2) PostgreSQL (Game)     — 50-200Gb   — Рекорды, статистика             — 1 мастер, 1 реплика
3) PostgreSQL (Billing)     — 50-200Gb     — Балансы, транзакции (внутриигровые)  — 1 мастер, 1 реплика
4) PostgreSQL (Shop)       — 50-200Gb  — Товары, инвентарь, покупки          — 1 мастер, 1 реплика
5) PostgreSQL (Inventory)     — 10-50Gb    — Товары авторов                — 1 мастер, 1 реплика
6) PostgreSQL (Profile)        — 50-200Gb     — Агрегированные профили           — 1 мастер, 1 реплика
7) Redis (кэши)            — 5-32Gb    — Сессии, лидерборд, кэш            — (1 на сервис)
8) MongoDB (для Inventory)  — 100-500Gb    — Товары авторов (динамические поля)   — (1 кластер 3 ноды)
9) ClickHouse         — 1-5Tb       — Raw data/events, аналитика         — 1 кластер (3 ноды)

Business mertics:- Популярность игр/авторов: n (score_updated) по полю game_id. А авторы - кого больше покупают.- Активность: DAU, user_id за день.- Retention: возвращаются ли игроки?

Selectel Server pricing (04.08.2026)
Стоимость/месяц - Назначение - CPU - RAM - Диск

1) $50 - Balancer, Gateway, Auth - 4 vCPU - 8Gb - 50Gb SSD
2) $70 - Game, Billing, Shop - 4 vCPU - 8Gb - 100Gb SSD
3) $150 - PostgreSQL (все БД) - 8 vCPU - 32Gb - 500Gb NVMe
4) $90 - Redis, MongoDB - 4 vCPU - 16Gb - 200Gb SSD
5) $70 - ClickHouse, Analytics, Monitoring - 4 vCPU - 16Gb - 500Gb SSD

~$430/месяц в итоге за 5 серверов  (в будущем добавить 6 и 7 облачный сервер)

Правильный расчет RPS: - У тебя 10k DAU. - Каждый игрок делает 10-20 сессий в день. Возьмем 15 сессий. - Итого действий в день: 10 000 * 15 = 150 000 событий. - Средний RPS: 150 000 / 86 400 ≈ 1.7 RPS. - Пиковый (х2): ~3.4 RPS.Но! Content per user 10-20 gaming sessions. Если это игровые раунды, внутри которых десятки запросов
(ходы, синхронизация, запись прогресса), то RPS вырастает.Допустим, 1 сессия = 10 API вызовов. Тогда 150 000 * 10 = 1.5M в день.Средний RPS = ~17 RPS. Пик = ~35 RPS.