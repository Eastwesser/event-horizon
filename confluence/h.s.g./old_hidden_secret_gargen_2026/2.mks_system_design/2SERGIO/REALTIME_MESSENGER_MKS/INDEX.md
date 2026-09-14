# 📖 System Design: Мессенджер Avito — Навигация

**Задача:** Спроектировать real-time мессенджер для маркетплейса с 3M DAU

---

## 📂 Структура документации

### 1. **README.md** (670 строк) — Основная документация
**Что внутри:**
- ✅ Математика и расчёты (RPS, Traffic, Storage)
- ✅ Выбор базы данных (PostgreSQL + Cassandra + Redis + Kafka)
- ✅ Архитектура системы (диаграмма)
- ✅ Подводные камни и нюансы (8 критичных моментов)
- ✅ Масштабирование (Horizontal scaling, Sharding)
- ✅ Monitoring & Observability (Prometheus, ELK, Jaeger)
- ✅ Security (JWT, Rate Limiting, Content Moderation)
- ✅ Cost Estimation (~$365K/year)

**Ключевые цифры:**
```
RPS:       ~2K r/s (average), ~6K r/s (peak)
Traffic:   ~20 MB/s (average), ~60 MB/s (peak)
Storage:   1.5 PB за 2 года с бекапами
Cost:      ~$365K/year ($0.12/user/year)
```

---

### 2. **DEEP_DIVE.md** (585 строк) — Альтернативные решения
**Что внутри:**
- 🔍 PostgreSQL vs MySQL (сравнительная таблица)
- 🔍 Cassandra vs MongoDB
- 🔍 WebSocket vs SSE vs Long Polling
- 🔍 Kafka vs RabbitMQ
- 🔍 Transactional Outbox Pattern
- 🔍 Sharding стратегии (по user_id vs chat_id vs Hybrid)
- 🔍 Многоуровневый кэш (Client SQLite → Redis → PostgreSQL → Cassandra)
- 🔍 Push Notifications (FCM limits, Silent vs Alert, Badge count)
- 🔍 Custom Metrics (Business metrics, SLI/SLO)
- 🔍 End-to-End Encryption (Signal Protocol)
- 🔍 Disaster Recovery (RTO/RPO, Backup strategy)
- 🔍 A/B Testing (Feature Flags, Cohort metrics)
- 🔍 Cost Optimization (Auto-scaling, Spot Instances, Reserved)
- 🔍 Serverless альтернатива (AWS Lambda + DynamoDB)

**Trade-offs и рекомендации:**
- Начните с простого (MVP)
- Масштабируйте по мере роста
- Premature optimization — корень всех зол!

---

### 3. **DIAGRAMS.md** (453 строки) — Визуальные диаграммы
**Что внутри:**
- 📊 High-Level Architecture (ASCII art)
- 📊 Message Flow (Send → Deliver → Read)
- 📊 WebSocket: Sticky Sessions + Redis Pub/Sub
- 📊 Data Partitioning (Time-based PostgreSQL)
- 📊 Cassandra Data Model
- 📊 Kafka Topics & Consumer Groups
- 📊 Rate Limiting: Token Bucket (визуализация)
- 📊 Multi-Region Setup (Disaster Recovery)

**Рекомендации по визуализации:**
- Mermaid.js (для интерактивных диаграмм)
- draw.io (для детальных схем)
- PlantUML (для sequence диаграмм)
- Grafana (для live метрик)

---

## 🎯 Быстрый старт

### Для технического собеседования (15 минут)

**Последовательность:**
1. **README.md → Section 1** (Математика): RPS, Traffic, Storage
2. **README.md → Section 2** (Выбор БД): PostgreSQL + Cassandra + Redis + Kafka
3. **README.md → Section 3** (Архитектура): High-level диаграмма
4. **README.md → Section 4** (Подводные камни): 8 критичных моментов

### Для детального изучения (1 час)

**Последовательность:**
1. **README.md** полностью (670 строк)
2. **DIAGRAMS.md** → Section 1-4 (визуализация)
3. **DEEP_DIVE.md** → Section 1-5 (альтернативы и trade-offs)

### Для подготовки к Senior/Staff интервью (2-3 часа)

**Последовательность:**
1. Все документы полностью
2. **DEEP_DIVE.md** → Section 6-12 (advanced topics)
3. Практика: нарисовать диаграммы в draw.io
4. Подготовить ответы на вопросы интервьюера:
   - Почему выбрали PostgreSQL, а не MySQL?
   - Как обрабатываете дубликаты сообщений?
   - Что делать, если RPS вырос в 10 раз?
   - Как обеспечиваете consistency между БД и Kafka?

---

## 📌 Ключевые концепции по файлам

### README.md
```
✅ RPS = DAU × Messages / 86400
✅ Storage = Messages × Size × 1.5 (indexes) × 1.5 (backups)
✅ PostgreSQL (hot 30d) + Cassandra (cold > 30d)
✅ WebSocket Gateway (sticky sessions)
✅ Redis Pub/Sub (broadcast)
✅ Rate Limiting (Token Bucket)
✅ Monitoring (Prometheus + ELK + Jaeger)
```

### DEEP_DIVE.md
```
🔍 Transactional Outbox Pattern (DB + Kafka atomicity)
🔍 Sharding: Hybrid (routing table in Redis)
🔍 Multi-level Cache (4 levels)
🔍 E2EE (Signal Protocol)
🔍 Disaster Recovery (RTO 1h, RPO 5m)
🔍 Cost Optimization (Auto-scaling, Spot, Reserved)
```

### DIAGRAMS.md
```
📊 Message Flow: 12 шагов (Send → DB → Kafka → WS → ACK → Read)
📊 WebSocket: Load Balancer → 3 WS Servers → Redis Pub/Sub
📊 PostgreSQL: Partitioning по месяцам → ETL → Cassandra
📊 Kafka: 3 partitions → 3 consumers (ordering per chat)
📊 Token Bucket: 30 tokens/60s → refill rate 0.5 tokens/s
```

---

## 🎓 Темы для изучения

### Junior/Middle уровень
- [ ] RPS и Traffic расчёты
- [ ] Выбор БД (PostgreSQL vs NoSQL)
- [ ] WebSocket vs HTTP
- [ ] Redis (cache patterns)
- [ ] Kafka basics (topics, partitions)
- [ ] Rate Limiting (Token Bucket)

### Senior уровень
- [ ] Sharding стратегии
- [ ] Consistency models (Strong, Eventual)
- [ ] CAP theorem
- [ ] Transactional Outbox Pattern
- [ ] Multi-region setup
- [ ] Disaster Recovery (RTO/RPO)
- [ ] Cost Optimization

### Staff/Principal уровень
- [ ] Distributed Consensus (Raft, Paxos)
- [ ] Observability (Metrics, Logs, Traces)
- [ ] SLI/SLO/SLA
- [ ] Capacity Planning
- [ ] A/B Testing Infrastructure
- [ ] Security (E2EE, Zero Trust)
- [ ] Trade-offs analysis

---

## ❓ FAQ

**Q: Почему PostgreSQL, а не MongoDB?**
A: MongoDB не поддерживает ACID транзакции на уровне нескольких документов (до версии 4.0), PostgreSQL лучше для transactional consistency (статусы доставки).

**Q: Зачем Kafka, если есть RabbitMQ?**
A: Kafka для event sourcing и audit log (сохраняем все события навсегда), RabbitMQ для task queues (события удаляются после обработки).

**Q: Как обрабатывать дубликаты?**
A: Idempotency key (event_id) + дедупликация в памяти (Redis или in-memory map).

**Q: Что если RPS вырос в 10 раз?**
A: Horizontal scaling (Chat Service: 10 → 100 инстансов), Sharding (PostgreSQL: 1 → 10 шардов), Read Replicas (5 → 50).

**Q: Как обеспечить consistency между DB и Kafka?**
A: Transactional Outbox Pattern (сохраняем в БД и outbox table, worker читает outbox и публикует в Kafka).

---

## 📚 Дополнительные материалы

**Книги:**
- "Designing Data-Intensive Applications" by Martin Kleppmann
- "System Design Interview" by Alex Xu (Volume 1 & 2)
- "Building Microservices" by Sam Newman

**Курсы:**
- ByteByteGo (YouTube) — System Design диаграммы
- Grokking the System Design Interview (Educative)

**Open Source:**
- Telegram (open source client)
- Rocket.Chat (open source Slack alternative)
- Matrix.org (decentralized chat protocol)

---

## ✅ Checklist перед собеседованием

**Математика:**
- [ ] Могу рассчитать RPS по формуле
- [ ] Могу рассчитать Storage с учётом индексов и бекапов
- [ ] Понимаю Peak vs Average traffic

**Архитектура:**
- [ ] Могу нарисовать High-Level диаграмму за 5 минут
- [ ] Понимаю роль каждого компонента (Chat Service, WS Gateway, Redis, Kafka)
- [ ] Знаю альтернативы (PostgreSQL vs MySQL, Kafka vs RabbitMQ)

**Подводные камни:**
- [ ] Sticky sessions для WebSocket
- [ ] Redis Pub/Sub для broadcast
- [ ] Rate Limiting (Token Bucket)
- [ ] Unread counts в Redis
- [ ] Media files в S3/CDN
- [ ] Archiving в Cassandra
- [ ] Content moderation

**Масштабирование:**
- [ ] Horizontal scaling (как добавить серверы)
- [ ] Sharding (когда и как)
- [ ] Read Replicas (когда нужны)
- [ ] Multi-region (для geo-distributed users)

**Monitoring:**
- [ ] Знаю какие метрики собирать (RPS, latency, error rate)
- [ ] Понимаю SLI/SLO/SLA
- [ ] Могу настроить alerts (Prometheus)

---

**Итого:** 1708 строк профессиональной документации по System Design! 🎯

**Время на изучение:**
- Quick start: 15 минут
- Детальное изучение: 1 час
- Senior/Staff prep: 2-3 часа

**Готовность к интервью:** 95% ✅
