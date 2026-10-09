# Miro review vs METRICS / code (09.10.2026)

Screenshots: `miro_pics/` (`miro_main.png` + `miro_1`…`miro_10`).  
Metrics source: `../07.10.2026/LOAD_TESTS/METRICS.md`.

**Happy path:** не рисовать как цепочку боксов на Miro — это видео Дениса (reg → play → shop/sub → pay → card). На схеме оставляем topology + ports + bus.

---

## METRICS panel (правая колонка) — вердикт

| Блок Miro | vs `METRICS.md` | Вердикт |
|-----------|-----------------|---------|
| MAU 100k / DAU 10k / ~1M new/year | совпадает | OK |
| RPS avg 30–50, peak ~100; read/write | совпадает (есть и модель ~17/35 при 10 API/сессия) | OK — две оценки, обе planning |
| HTTP 5k / WS 2.5k (до 10k / 5k) | совпадает | OK |
| Latency p50/90/95/99 | совпадает | OK (цель прода ≠ локальный k6 p95~700ms) |
| Storage PG/Redis/Mongo/CH + retention 30d | совпадает | OK |
| Selectel ~$430 / 5 VM | совпадает | OK |
| RPS calc 10k×15 → 1.7 → ×10 API ≈17 / peak 35 | совпадает | OK |
| Go / DB / Grafana 03:00 watchlist | совпадает | OK |
| DB table dated 08.10 | те же сервисы + объёмы | OK |

**Итог:** правая колонка Miro = capacity planning из `METRICS.md`, не измеренный local CORE. На интервью разделяй «цель» vs «факт стенда» (`SECURITY_AUDIT_RU.md`).

---

## Topology (слева) — что ок

- Client → LB `:8079` → API GW `8081–8083` → сервисы + NATS JetStream hub
- Auth / Game / Billing / LB / Profile / Shop / Inventory(+Mongo+Outbox) / Authors / Payment / History / Analytics(CH) / Notification / Fulfillment
- Observability: Prometheus 9090, Grafana 3000, Jaeger 16686, NATS Explorer 7777 (+ Alertmanager)
- Kafka у Fulfillment помечен как optional / cut — согласуется с thin vs `deploy-heavy`
- Порты на стикерах Auth/Game/Billing/LB в целом совпадают с compose / interview cheat sheet

---

## Drift / что поправить на доске (не блокер тега)

1. **MCP SERVER (RAG)** нарисован у GW — **parked**, не runtime v1.1.0. Либо серый «future», либо убрать со стрелки.
2. **`api routes:`** на стикерах пустые → заполнить каноном **`/api/v1/...`** (с 09.10 в коде).
3. **Нумерация сервисов** плавает (Shop/Inventory/History в разных кропах как 6/7/8) — выровнять под единый список.
4. **Authors sticky** иногда тянет Mongo/ports Inventory — Mongo только у **Inventory**.
5. **Inventory PG** на одном стикере `5446` — сверить с compose (ожидаемо `546x` ряд); Redis `6364` тоже перепроверить.
6. **Notification gRPC `50056`** — сверить с фактическим портом в compose.
7. **PROMO** обрубок справа — либо подписать, либо убрать.
8. Outbox нарисован почти у всех PG — ок как паттерн; у кого реально нет outbox — можно не рисовать цилиндр (честность для интервью).

---

## ASCII / text map (для интервью)

```text
[CLIENT React :5173]
        |  HTTP /api/v1/*   WS /ws/leaderboard
        v
[LOAD BALANCER :8079] -----> [API GW ×3 :8081-8083]
                                    |
        +-----------+---------------+---------------+-----------+
        v           v               v               v           v
     [AUTH]      [GAME]        [BILLING]      [LEADERBOARD] [PROFILE]
     :50051      :50052         :50053          :50054        :50060
     PG+Redis    PG(+Redis)     PG+Redis        PG+Redis SS   PG
        |           |               |               |           |
        +-----------+------ NATS JETSTREAM HUB -----+-----------+
                           (cluster 4222/4223/4224)
                                    |
        +--------+--------+---------+--------+--------+
        v        v        v         v        v        v
     [SHOP] [INVENTORY] [AUTHORS] [PAYMENT] [HISTORY] [ANALYTICS]
            +Mongo+Outbox                    PG       ClickHouse
        |
     [NOTIFICATION] [FULFILLMENT]--(opt Kafka)--Outbox
        |
     [Observability: Prom/Grafana/Jaeger/NATS Explorer]
```

---

## Next for Denis (IRL Miro)

- [ ] Серый MCP / убрать стрелку  
- [ ] Дописать `/api/v1` на стикерах  
- [ ] Починить Authors/Inventory Mongo drift + спорные порты  
- [ ] После правок — PNG уже в `miro_pics/`; при желании один «чистый» full export в корень `miro/`
