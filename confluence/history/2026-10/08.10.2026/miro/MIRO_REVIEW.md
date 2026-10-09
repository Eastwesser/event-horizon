# Miro review vs METRICS / code (09.10.2026)

Screenshots: `miro_pics/` · **Paste stickers:** `MIRO_STICKERS_PASTE.md`  
Metrics: `../07.10.2026/LOAD_TESTS/METRICS.md`

**Happy path:** видео Дениса (reg → play → shop/sub → pay → card), не цепочка боксов на Miro.

---

## `/api/v1` — где прописано?

| Слой | `/api/v1`? |
|------|------------|
| Gateway HTTP (Gin) | **да** — все публичные REST |
| FE axios | **да** — `baseURL: '/api/v1'` |
| OpenAPI / k6 | **да** |
| Auth/Game/… gRPC | **нет** — только protobuf RPC |

Legacy `/api/*` на gateway один раз переписывается в `/api/v1/*`.

---

## METRICS panel — вердикт

Правая колонка Miro **=** capacity planning из `METRICS.md` (не local CORE k6). OK.

---

## Drift — ответы (после сверки compose)

| # | Тема | Вердикт |
|---|------|---------|
| 1 | MCP | **Сделан** (`services/mcp`, stdio, RAG+tools). На схеме: Cursor→MCP, **не** player→GW→MCP. |
| 2 | api routes | Список для копипаста → `MIRO_STICKERS_PASTE.md` |
| 3 | Нумерация | Оставь свою; визуально выстрой edge→core→async (см. paste sheet) |
| 4 | Authors + Mongo | Authors = **PG 5468 + Redis 6387**. Mongo **только Inventory** (`27017` в inventory compose). |
| 5 | Inventory ports | PG **5466**, Redis **6384** (стикеры 5446 / 6364 — **ошибка**) |
| 6 | Notification gRPC | **50063** (не 50056); PG host **5470** |
| 7 | PROMO | ignore (crop) |
| 8 | Outbox | Есть: Game, Billing, Shop, Inventory, Payment, Authors. **Нет:** Auth, Profile, Leaderboard, History, Analytics, Notification, Fulfillment |

---

## ASCII map (актуально)

```text
[CLIENT React :5173]
        |  HTTP /api/v1/*          WS /ws/leaderboard
        v
[LOAD BALANCER :8079] -----> [API GW ×3 :8081-8083]
                                    |
     +---------+----------+---------+----------+----------+
     v         v          v         v          v          v
  [AUTH]    [GAME]    [BILLING] [LEADERBOARD] [PROFILE] [SHOP]
  :50051    :50052     :50053     :50054       :50060    :50055
  PG5460    PG5461     PG5462     PG5463       PG5464    PG5465
  R6379     R6380      R6381      R6382        R6385     R6383
  (no OB)   OUTBOX     OUTBOX     (no OB)      (no OB)   OUTBOX
     |         |          |         |            |          |
     +---------+----+ NATS JETSTREAM HUB (4222/3/4) +-------+
                        |         |         |         |
                        v         v         v         v
                 [INVENTORY] [AUTHORS] [PAYMENT] [HISTORY]
                   :50059     :50061    :50058    :50062
                   PG5466     PG5468    PG5467    PG5469
                   R6384      R6387     R6386     (consumer)
                   Mongo27017 OUTBOX    OUTBOX    (no OB)
                   OUTBOX     (no Mongo)
                        |         |
                        v         v
                 [ANALYTICS] [NOTIFICATION] [FULFILLMENT]
                   :50057      :50063         (worker)
                   ClickHouse  PG5470         opt Kafka
                   (no OB)     (no OB)        (no OB table)

[Cursor/agent] --stdio--> [MCP SERVER] --> NATS / PG(ro) / Redis / Prydwen RAG
                          (NOT on player HTTP path)

[Observability] Prom:9090 Grafana:3000 Jaeger:16686 NATS Explorer:7777 Alertmanager
```

---

## Что осталось после v1.1.0 wave (честно)

**IRL (ты):**  
- Вставить стикеры из `MIRO_STICKERS_PASTE.md`  
- Поправить порты / убрать лишние Outbox / MCP стрелку  
- Опционально: Boosty URL в `BOOSTY_DONE`, `git tag v1.1.0`, ranked smoke для нулей на профиле  

**Код / продукт (не блокер тега):** avatar, 108 authors, LB seed, Tamagotchi tickets gift, deep-link «рекорд побит», AuthZ matrix  

**Parked:** C4 payouts, flower/3D/Dodo/Sims, per-game Boosty 1.1.1…  

**MCP:** код есть — для Cursor; допиливать tools можно позже, не блокирует схему.
