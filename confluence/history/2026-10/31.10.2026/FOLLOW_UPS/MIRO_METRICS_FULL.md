Метрики для Miro — что писать на доске
Ты делаешь схему в Miro и хочешь указать приблизительные метрики. На собесе это не «точные цифры из прода», а capacity planning — порядок величин, который ты можешь защитить.

Вот что писать на доске, сгруппированное по блокам. Плюс — что не писать.

1. Бизнес-метрики (верхний блок)
Метрика	Значение	Откуда
MAU	~100k	Модель EH (00_METRICS.md)
DAU	~10k (10% MAU)	Модель EH
Сессий на юзера / день	10–20	Модель EH
RPS avg	~17–50	10k × 15 / 86400
RPS peak	~35–100	×2 от avg
Read:Write	~2:1 … 3:1	Больше чтений
Что писать на доске:

text
DAU ~10k
RPS avg ~30, peak ~100
Read:Write ~2:1
2. Латентность (верхний/средний блок)
Метрика	Значение	Откуда
p50	< 50ms	Цель EH
p90	< 150ms	Цель EH
p95	< 200ms	Цель EH
p99	< 300ms	Цель EH
Важно: это цель прода, не факт локального k6. На 8 GiB VM локально p95 ~700ms — это ok для laptop + docker.

Что писать на доске:

text
p50 < 50ms
p95 < 200ms
p99 < 300ms
(цель, не факт)
3. Соединения (верхний/средний блок)
Метрика	Значение	Откуда
HTTP concurrent	~5k	Модель EH
WebSocket concurrent	~2.5k (до 5k)	Модель EH
NATS connections	~20	Сервисы
PG connections	~50–100 на БД	Pool 25
Redis connections	~10–50 на инстанс	—
Что писать на доске:

text
HTTP ~5k
WS ~2.5k
PG ~50/БД
4. Хранилища (нижний блок)
Метрика	Значение	Откуда
PG (на сервис)	~10–100 GB	Модель EH (Auth, Game, Billing…)
PG (всего)	~1 TB	Модель EH
Redis (на инстанс)	~1–10 GB	Сессии, кэш, Sorted Set
Redis (всего)	~10–50 GB	8 инстансов
Mongo	~100 GB (experimental)	Inventory
ClickHouse	~1–5 TB	Analytics
S3	~TB	Картинки, мерч
NATS stream	~1 GB (MaxBytes)	Stream EVENTS
Retention (NATS)	7 дней	Stream EVENTS
Retention (History)	30 дней	History PG
Что писать на доске:

text
PG ~1 TB
Redis ~10–50 GB
ClickHouse ~1–5 TB
NATS stream 1 GB / 7d
History 30d
5. Железо (нижний блок)
Метрика	Значение	Откуда
CPU	< 70% на пике	Модель EH
RAM	< 70% на пике	Модель EH
Сеть	100–200 Mbit/s	Selectel
Диск	SSD / NVMe	—
Бюджет	~$430/мес (5 VM)	Selectel
Для 10k DAU	8c/32G × 1	SUITABLE_SERVER.md
Для 100k DAU	8c/32G × 3–5	Follow-up
Что писать на доске:

text
CPU/RAM < 70%
Сеть 100–200 Mbit/s
$430/мес (5 VM)
8c/32G для 10k DAU
6. Go runtime (средний блок, если есть место)
Метрика	Значение	Откуда
Goroutines	< 1000 на сервис	Модель EH
Heap	< 512 MB на сервис	Модель EH
GC cycles	< 10/мин	Модель EH
Что писать на доске:

text
Goroutines < 1000
Heap < 512 MB
GC < 10/мин
7. NATS / брокеры (средний блок)
Метрика	Значение	Откуда
NATS stream	EVENTS, 10 subjects	HOW_DOES_EH_WORK.md
NATS MaxAge	7 дней	Stream Config
NATS MaxBytes	1 GB	Stream Config
NATS consumer lag	< 1000 (алерт)	SOS-дашборд
Kafka topics	purchase.paid, purchase.fulfilled	Heavy
Kafka partitions	3 (пример)	Heavy
Что писать на доске:

text
NATS: EVENTS, 10 subjects
MaxAge 7d, MaxBytes 1 GB
Lag < 1000
8. Метрики по конкретным сервисам (если рисуешь детально)
Сервис	Ключевая метрика	Значение
Auth	login p95	< 500ms
Game	submit p95	< 500ms
Billing	spend p95	< 200ms
Leaderboard	top p95	< 50ms
Shop	purchase p95	< 500ms
Inventory	create p95	< 300ms
Profile	get p95	< 100ms
Payment	webhook p95	< 500ms
Authors	approve p95	< 500ms
History	get p95	< 100ms
Analytics	dau p95	< 500ms
Notification	list p95	< 100ms
Fulfillment	assembly p95	< 12s
Gateway	p95	< 200ms
Balancer	p95	< 50ms
NATS Hub	connected	1
MCP	tool p95	< 1s
Что писать на доске: только ключевые (Auth, Billing, Shop, Gateway, Balancer), не все 17.

9. Как это выглядит на доске (пример)
text
┌─────────────────────────────────────────────────────┐
│  Clients                                            │
│  React :5173                                        │
└──────────────────────┬──────────────────────────────┘
                       │ HTTP/WS
                       ▼
┌─────────────────────────────────────────────────────┐
│  Edge                                               │
│  Balancer :8079 (Least Connections)                 │
│  Gateway ×3 :8081–8083 (CB, Rate Limit, WS Ticket) │
│  p95 < 200ms · RPS ~30 avg / ~100 peak              │
└──────────────────────┬──────────────────────────────┘
                       │ gRPC
                       ▼
┌─────────────────────────────────────────────────────┐
│  Services (17)                                      │
│  Auth · Game · Billing · Leaderboard · Shop · …     │
│  p95 < 500ms (Auth, Game, Shop)                     │
│  p95 < 200ms (Billing)                              │
│  p95 < 50ms (Leaderboard)                           │
│  DAU ~10k · HTTP ~5k · WS ~2.5k                     │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────┐
│  Data                                               │
│  PG ~1 TB · Redis ~10–50 GB · CH ~1–5 TB           │
│  NATS stream 1 GB / 7d · History 30d                │
│  CPU/RAM < 70% · 8c/32G × 1 (для 10k DAU)          │
└─────────────────────────────────────────────────────┘
10. Что НЕ писать на доске
«10k RPS» — это 10k DAU, десятки RPS. Не путай.

«100k RPS» — не твоя модель.

«p95 429ms» — это локальный k6 на 8 GiB VM, не цель прода. На доске — цель (< 200ms).

«1M DAU» — не твоя модель (это follow-up, не основа).

«Redis 6379» на Gateway — это внутренний порт, не публичный.

«Kafka 9092» — на хосте 19092. Внутри 9092.

«Mongo 27017» — в текущем compose нет. Experimental.

11. Три уровня детализации (выбери)
Минимум (для быстрой доски)
text
DAU ~10k
RPS ~30 avg / ~100 peak
p95 < 200ms
PG ~1 TB, Redis ~10–50 GB, CH ~1–5 TB
8c/32G × 1
Средний (для нормальной доски)
text
DAU ~10k, MAU ~100k
RPS ~30 avg / ~100 peak
Read:Write ~2:1
HTTP ~5k, WS ~2.5k
p50 < 50ms, p95 < 200ms, p99 < 300ms
PG ~1 TB, Redis ~10–50 GB, Mongo ~100 GB, CH ~1–5 TB
NATS stream 1 GB / 7d, History 30d
CPU/RAM < 70%, 8c/32G × 1
$430/мес (5 VM)
Максимум (для детальной доски)
Всё из среднего + метрики по сервисам (Auth p95 < 500ms, Billing p95 < 200ms, Leaderboard p95 < 50ms, …) + Go runtime (goroutines < 1000, heap < 512 MB) + NATS lag < 1000.

12. Что сказать на собесе
«На доске — capacity planning, не факт прода. Модель: 10k DAU, ~30 RPS avg, пик ~100, p95 < 200ms. Read:Write ~2:1. HTTP ~5k, WS ~2.5k. PG ~1 TB, Redis ~10–50 GB, ClickHouse ~1–5 TB. NATS stream 1 GB / 7 дней. CPU/RAM < 70%. Железо: 8c/32G для 10k DAU, ~$430/мес на 5 VM. Локально k6: 20 VU browse PASS, ~50 RPS, p95 429ms на 8 GiB VM — это факт стенда, не цель прода.»

Итого: что писать
Блок	Метрики
Бизнес	DAU ~10k, MAU ~100k, RPS ~30 avg / ~100 peak
Латентность	p50 < 50ms, p95 < 200ms, p99 < 300ms
Соединения	HTTP ~5k, WS ~2.5k, PG ~50/БД
Хранилища	PG ~1 TB, Redis ~10–50 GB, CH ~1–5 TB, NATS 1 GB/7d
Железо	CPU/RAM < 70%, 8c/32G × 1, $430/мес
Go	goroutines < 1000, heap < 512 MB, GC < 10/мин
NATS	EVENTS, 10 subjects, MaxAge 7d, MaxBytes 1 GB, lag < 1000
Сервисы	Auth p95 < 500ms, Billing p95 < 200ms, Leaderboard p95 < 50ms, Gateway p95 < 200ms
Не писать: «10k RPS», «100k RPS», «p95 429ms» (это локальный k6), «1M DAU», «Mongo 27017 в compose».