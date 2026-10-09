# Где лежат Mermaid / схемы (для сверки с Miro)

Когда кидаешь скрины в `08.10.2026/miro/`, сверяем с этими файлами.

| Файл | Что внутри |
|------|------------|
| `confluence/architecture/SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md` | **Главный** Mermaid: clients → edge → gRPC → async → data |
| `confluence/architecture/SYSTEM_DESIGN/HOW_DOES_EH_WORK.md` | Пояснения + диаграммы |
| `confluence/architecture/SYSTEM_DESIGN.md` | Обзор / доп. mermaid |
| `confluence/architecture/EH_SCHEMAS.md` | Текстовые схемы / порты |
| `confluence/architecture/SYSTEM_DESIGN/event-horizon-v1.0.6.png` | Старый Miro export |
| `…/SYSTEM_DESIGN_MIRO/FINAL_SYSTEM_DESIGN_MIRO_SCHEME.md` | Текст v1.1.0 + happy path + куда класть новый PNG |

**Что проверить на твоих скринах Miro (чеклист для меня после дропа):**
- [ ] Balancer `:8079`, Gateway `:8081–8083`
- [ ] Все gRPC сервисы с портами 50051–50062 (как в таблице)
- [ ] NATS JetStream / EVENTS (не только Kafka)
- [ ] Kafka помечен как optional (`deploy-heavy`)
- [ ] WS leaderboard с edge
- [ ] Outbox хотя бы на inventory/shop path
- [ ] ClickHouse у Analytics
- [ ] Fulfillment + Notification как consumers
