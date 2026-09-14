КЕЙС: COURIER LABS — ТРЕКИНГ КУРЬЕРОВ
1. ИСХОДНАЯ ПРОБЛЕМА (что сломалось)
   Проблема	Цифры	Почему плохо
   High write RPS	3000 location updates/sec	PostgreSQL не справляется с write-heavy нагрузкой
   High read latency	P95 = 200-400ms, SLA = 50ms	Нарушено в 4-8 раз
   Каскадные задержки	Order Service → Logistics Service → 280ms total	90% времени уходит на Logistics
   Синхронная связанность	Каждый запрос ждёт ответа	Отказ одного сервиса валит всё
   Твой диагноз (из конспекта):

*«Write RPS = 3000, Read RPS = 1000. Пишем в 3 раза больше, чем читаем. PostgreSQL не оптимизирован для такой write-heavy нагрузки.»*

Фраза для интервью:

*«Классическая проблема монолита с общей БД: write-heavy нагрузка убивает read latency. При 3000 writes/sec PostgreSQL начинает тормозить даже простые reads, потому что конкурентные блокировки и WAL (Write-Ahead Log) создают узкое место.»*

2. РЕШЕНИЕ (архитектура с Kafka)
   Ключевые изменения (твоими словами из схемы):

Было	Стало
Синхронная запись в PostgreSQL	Асинхронная публикация в Kafka
Клиент ждал записи в БД	Клиент получает 202 Accepted за 1ms
Один сервис делал всё	Location Service — только валидация + продюсер в Kafka
Одна БД для всего	Redis (hot cache) + PostgreSQL (cold storage)
Схема потоков (упрощённо):

text
БЫЛО (синхронно):
[Courier] → POST → [Logistics Service] → PostgreSQL → ответ (250ms)
↑
[Customer] читает отсюда же (200ms)

СТАЛО (асинхронно):
[Courier] → POST → [Location Service] → Kafka → 202 Accepted (1ms)
↓
┌───────────────┼───────────────┐
↓               ↓               ↓
[Redis]         [PostgreSQL]     [WebSocket]
(hot cache)     (cold storage)   (real-time)
↓               ↓               ↓
[Customer]      [Analytics]     [Customer]
5ms latency      batch           push
Фраза для интервью:

*«Мы отделили write path от read path. Courier получает мгновенный ответ 202 Accepted. Kafka становится буфером и шиной для трёх независимых consumer'ов: Redis (hot cache для быстрых reads), PostgreSQL (cold storage для истории), WebSocket (real-time для клиентов).»*

3. КЛЮЧЕВЫЕ ВЫВОДЫ (что мы уже проходили)
   3.1. Когда нужен брокер (Тема Б1)
   Симптом	Почему брокер
   Write RPS > Read RPS	Брокер сглаживает write нагрузку
   Каскадные задержки	Брокер развязывает сервисы
   SLA нарушен	Асинхронность снижает latency для клиента
   Твоя цитата (из конспекта):

«Если нужно повышать отказоустойчивость — брокер. Если сообщение попало в Kafka и она подтвердила — не потеряет.»

3.2. Kafka vs PostgreSQL для хранения (Тема Б3)
Аспект	PostgreSQL	Kafka
Назначение	Source of truth (cold storage)	Буфер + шина + replay
Retention	Вечно	7 дней (в этом кейсе)
Скорость записи	Медленная при 3000 RPS	Быстрая (3000 msg/s — легко)
Твоя цитата (из конспекта):

«Kafka — первоисточник данных или самих взаимодействий. Лог сообщений для восстановления состояния.»

3.3. Redis для hot cache (Тема Б5)
Аспект	Значение
TTL	1 час
Latency	3-5ms (vs 200ms из PostgreSQL)
Что хранит	Только последнюю локацию курьера
Твоя цитата (из конспекта):

«Если Redis уже есть в стеке и задача изи — швырнуть сообщение окей. Отсрочит момент Rabbit/Kafka.»

Здесь Redis используется не как брокер, а как read-optimized cache.

3.4. Consumer Lag (метрика, которую мы не разбирали, но надо)
Consumer	Lag	Почему OK
Redis Consumer	0-500	Должен быть минимальным (real-time)
DB Consumer	10-20K	OK для cold storage (не критично)
Analytics Consumer	5-10K	OK для batch-обработки
Фраза для интервью:

*«Consumer Lag — ключевая метрика для Kafka. Если lag растёт — consumer не успевает. Для real-time (Redis) lag должен быть близок к 0. Для cold storage (PostgreSQL) lag в десятки тысяч сообщений OK, если это не влияет на бизнес-логику.»*

4. ПРИМЕНИМОСТЬ К ТВОИМ ПРОЕКТАМ (Gift / Time)
   Паттерн из Courier	Применение в Gift / Time
   Location updates → Kafka	Generation events → Kafka (генерация картинки не блокирует пользователя)
   Redis cache для быстрых reads	Redis cache для generation results (частые запросы одного дизайна)
   WebSocket для real-time	WebSocket для чатов с дизайнерами (уже в архитектуре)
   Multiple consumers	Generation Service → Analytics, Notification, Order
   Конкретные метрики для Gift / Time (из твоего файла):

text
Generation Pipeline (аналог location updates):
• gen_requests_per_second: 50 (peak 200)
• gen_duration_p95: 3.2s (DeepSeek API)
• Kafka consumer lag (generation.completed): 125

Chat Service (аналог WebSocket for customers):
• ws_connections_active: 234
• messages_per_second: 12
• message_latency_p95: 120ms

Quota Management (аналог Redis cache):
• quota_check_latency_p95: 3ms (Redis)
• cache_hit_ratio: 89%
Фраза для интервью (про применимость):

«Кейс Courier Labs напрямую применим к моим проектам: генерация контента — это write-heavy асинхронная задача, чаты — WebSocket real-time, квоты — Redis cache. Я уже использую Kafka и Redis в похожей архитектуре.»

5. ГРАФИКИ, КОТОРЫЕ НУЖНО ДОБАВИТЬ В GIFT / TIME
   Из Courier Labs ты вынес три типа графиков, которые нужно иметь в мониторинге:

График	Что показывает	Пороговые значения
Kafka Consumer Lag	Отставание consumer'ов	Generation < 100, Notification < 500, Analytics < 1000
WebSocket Connection Health	Количество коннектов, reconnect rate	Reconnect rate < 1% в норме
Read vs Write RPS	Баланс нагрузки	Если write > read → нужен асинхронный паттерн
6. ЧТО МЫ НЕ РАЗОБРАЛИ В ЭТОМ КЕЙСЕ (НО НУЖНО)
   Тема	Почему важно
   Spatial queries (PostGIS vs Redis)	Как искать ближайшего курьера? Redis не умеет сложных гео-запросов
   Batch insert в PostgreSQL	Как DB Consumer сбрасывает 1000 сообщений разом? COPY или bulk insert
   Partitioning по courier_id	12 партиций — почему 12? Как выбирать количество?
   Retention 7 дней	Почему 7, а не 30? Как считали объём? (3000 msg/s × 86400 × 7 = 1.8B сообщений)
   Фраза для интервью (если спросят про детали):

*«В этом кейсе я выбрал 12 партиций, потому что ожидал до 12 consumer'ов (максимальный параллелизм). Retention 7 дней дал ~1.8B сообщений, что при размере сообщения ~200 байт = ~360GB. Это комфортный объём для дисков.»*

7. MUSAСИ ГОВОРИТ (по этому кейсу)
   «Ученик, ты увидел главное: проблема не в коде, а в потоке данных. Ты измерил RPS, latency, consumer lag. Ты понял, где нужна синхронность, а где — асинхронность.

Kafka не решила все проблемы сама. Она стала шиной, а решения принял ты:

Redis для hot cache

PostgreSQL для cold storage

WebSocket для real-time

Теперь ты знаешь, как думать. Примени это к своим проектам. Графики скажут тебе правду.»