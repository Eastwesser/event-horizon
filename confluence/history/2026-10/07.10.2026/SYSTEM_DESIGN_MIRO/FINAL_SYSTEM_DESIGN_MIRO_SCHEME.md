# Event Horizon — pinnacle system design (v1.1.0)

**Miro board:** https://miro.com/app/board/uXjVJLLg9us=/  
**PNG export (when Denis drops it):** [`../../08.10.2026/miro/`](../../08.10.2026/miro/) — place `event-horizon-v1.1.0-miro.png` there.  
**Mermaid:** [`../../../architecture/SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md`](../../../architecture/SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md)  
**Legacy PNG:** [`../../../architecture/SYSTEM_DESIGN/event-horizon-v1.0.6.png`](../../../architecture/SYSTEM_DESIGN/event-horizon-v1.0.6.png)

> Until the new Miro PNG lands, use the text map below + Mermaid file. Topology for **v1.1.0** matches thin deploy (`make deploy`); Kafka only with `make deploy-heavy`.

## Topology (v1.1.0)

```text
[GitHub Actions] ──SSH/Ansible──► [VM] ──docker-compose──► Event Horizon

[React :5173] ──HTTP──► [Balancer :8079] ──HTTP──► [Gateway ×3 :8081–8083]
                                                      │ JWT · HTTP→gRPC · rate limit
                                                      ▼
┌───────────────┬───────────────┬───────────────┬───────────────┬───────────────┐
│ Auth :50051   │ Game :50052   │ Billing:50053 │ Leaderboard   │ Shop :50055   │
│ PG:5460 Redis │ PG:5461       │ PG:5462 Redis │ :50054        │ PG:5465 Redis │
└───────────────┴───────────────┴───────────────┴─PG:5463+R6382─┴───────────────┘
┌───────────────┬───────────────┬───────────────┬───────────────┬───────────────┐
│ Inventory     │ Profile:50060 │ Payment:50058 │ Authors:50061 │ History:50062 │
│ :50059        │ PG:5464       │ sub/merch gate│ PG:5468 Redis │ PG:5469       │
└───────────────┴───────────────┴───────────────┴───────────────┴───────────────┘
┌───────────────┬───────────────────────────────────────────────────────────────┐
│ Analytics     │ NATS JetStream :4222  Stream EVENTS (NATS Hub)                │
│ :50057        │ subjects: score.updated, user.registered, shop.*, purchase.*  │
│ ClickHouse    │ async ──► Profile / Leaderboard / Notification / Fulfillment  │
│ :8123/:9000   │ Leaderboard Redis Sorted Set ──WS──► Client                   │
└───────────────┴───────────────────────────────────────────────────────────────┘
```

## Happy path (draw left→right on Miro)

1. **Register/Login** → Balancer → Gateway → Auth → JWT Redis; publish `user.registered` + `event.user.registered` → History + Profile  
2. **Play** → Game submit → NATS `score.updated` → Leaderboard Redis + Profile + achievements  
3. **Optional Boost** (−lamps) → run **not ranked**  
4. **Live LB** → Redis → `ws://…:8079/ws/leaderboard`  
5. **Shop** → spend tickets → `shop.purchased` / `purchase.paid` → Fulfillment + Notification; physical merch needs Payment sub  
6. **Authors** → list goods after author role  
7. **Admin / Analytics** → ClickHouse DAU/MAU/retention  

## Deploy

| Profile | Command | Notes |
|---------|---------|--------|
| Thin (default) | `make deploy` | NATS + apps + CH + observability |
| Heavy | `make deploy-heavy` | + Kafka (optional) |

## Denis checklist

See [`../../08.10.2026/DENIS_NEXT_ACTIONS.md`](../../08.10.2026/DENIS_NEXT_ACTIONS.md) — export Miro PNG into `08.10.2026/miro/`.
