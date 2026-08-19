# Event Horizon — System Design (v1.0.7)

> Text + Mermaid successor to Miro [`event-horizon-v1.0.6.png`](./event-horizon-v1.0.6.png).  
> Canonical ports: `deployments/docker-compose.cluster.yml`, `confluence/architecture/EH_SCHEMAS.md`.  
> **Legend:** solid arrows = sync (HTTP/gRPC) · dashed arrows = async (NATS/Kafka/Outbox/WS)

---

## 1. Context — who talks to whom

```mermaid
flowchart TB
  subgraph clients["Clients"]
    FE["React SPA<br/>dev :5173"]
    MCP["MCP Server RAG<br/>local stdio · Cursor"]
  end

  subgraph edge["Edge"]
    LB["Balancer<br/>HTTP :8079<br/>metrics :9098"]
    GW["Gateway ×3<br/>HTTP :8081–8083<br/>OpenAPI /docs · WS /ws/leaderboard"]
  end

  subgraph sync["Sync domain gRPC"]
    AUTH["Auth :50051"]
    GAME["Game :50052"]
    BILL["Billing :50053"]
    LEAD["Leaderboard :50054"]
    SHOP["Shop :50055"]
    ANAL["Analytics :50057"]
    PAY["Payment :50058"]
    INV["Inventory :50059"]
    PROF["Profile :50060"]
    AUTHR["Authors :50061"]
    HIST["History :50062"]
  end

  subgraph async["Async consumers"]
    FUL["Fulfillment<br/>metrics :9101"]
    NOTIF["Notification<br/>metrics :9102"]
  end

  subgraph buses["Message buses"]
    NATS["NATS JetStream ×3<br/>:4222 / :4223 / :4224"]
    KAFKA["Kafka KRaft<br/>host :19092 → in-net :9092"]
    HUB["NATS Hub<br/>Stream EVENTS · metrics :9097"]
  end

  subgraph obs["Observability"]
    PROM["Prometheus :9090"]
    GRAF["Grafana :3000"]
    JAEG["Jaeger UI :16686 · OTLP :4317"]
    NEXP["NATS Exporter :7777"]
  end

  FE -->|HTTP JSON /api| LB
  LB -->|Least Connections| GW
  MCP -.->|tools docs| FE

  GW -->|gRPC + JWT CB| AUTH & GAME & BILL & LEAD & SHOP & ANAL & PAY & INV & PROF & AUTHR & HIST
  GW -.->|WS score.updated| FE

  SHOP -->|SpendCurrency| BILL
  SHOP -->|CanPurchaseMerch| PAY

  GAME & BILL & SHOP & INV & PAY & AUTHR -.->|Outbox / publish| NATS
  NATS --> HUB
  NATS -.-> LEAD & PROF & HIST & ANAL & SHOP

  SHOP -->|purchase.paid| KAFKA
  KAFKA -.-> FUL & NOTIF
  FUL -->|purchase.fulfilled| KAFKA

  GW & sync & async -.->|/metrics scrape| PROM
  PROM --> GRAF
  sync -.->|OTLP traces| JAEG
  NATS -.-> NEXP --> PROM
```

---

## 2. Entry path (request lifecycle)

```mermaid
sequenceDiagram
  autonumber
  participant C as React :5173
  participant B as Balancer :8079
  participant G as Gateway :808x
  participant A as Auth :50051
  participant R as Redis sessions :6379
  participant S as Domain service gRPC

  C->>B: GET/POST /api/...
  B->>G: HTTP (pick least-conn backend)
  G->>A: ValidateToken (protected routes)
  A->>R: session lookup
  R-->>A: ok
  A-->>G: user_id + role
  G->>S: gRPC unary (metadata x-user-role)
  S-->>G: response / gRPC error
  G-->>C: JSON + HTTP status (CB → 503 if open)
```

Gateway extras: `GET /openapi.yaml`, `GET /docs`, `WS /ws/leaderboard` (NATS `score.updated` fan-out).

---

## 3. Services, stores, outboxes (full port map)

```mermaid
flowchart LR
  subgraph core["Core"]
    direction TB
    AUTH2["Auth<br/>gRPC 50051 · m 9091"]
    PG_A[("PG :5460")]
    RD_A[("Redis :6379")]

    GAME2["Game<br/>gRPC 50052 · m 9092"]
    PG_G[("PG :5461")]

    BILL2["Billing<br/>gRPC 50053 · m 9093"]
    PG_B[("PG :5462")]
    RD_B[("Redis :6381")]
    OB_B[["Outbox<br/>balance.updated"]]

    LEAD2["Leaderboard<br/>gRPC 50054 · m 9094"]
    PG_L[("PG :5463")]
    RD_L[("Redis :6382<br/>Sorted Set")]

    PROF2["Profile<br/>gRPC 50060 · m 9099"]
    PG_P[("PG :5464")]
    RD_P[("Redis :6385")]

    SHOP2["Shop<br/>gRPC 50055 · m 9095"]
    PG_S[("PG :5465")]
    RD_S[("Redis :6383")]
    OB_S[["Outbox<br/>shop.purchased"]]
  end

  subgraph commerce["Commerce & content"]
    direction TB
    PAY2["Payment<br/>gRPC 50058 · m 9103"]
    PG_PAY[("PG :5467")]
    RD_PAY[("Redis :6386")]
    OB_PAY[["Outbox<br/>payment.completed"]]

    INV2["Inventory<br/>gRPC 50059 · m 9096"]
    PG_INV[("PG :5466")]
    RD_INV[("Redis :6384")]
    OB_INV[["Outbox<br/>inventory.item.*"]]

    AUTHR2["Authors<br/>gRPC 50061 · m 9104"]
    PG_AUTHR[("PG :5468")]
    RD_AUTHR[("Redis :6387")]
    OB_AUTHR[["Outbox<br/>author.upserted"]]

    HIST2["History<br/>gRPC 50062 · m 9105"]
    PG_HIST[("PG :5469<br/>retention 30d")]

    ANAL2["Analytics<br/>gRPC 50057 · m 9106"]
    CH[("ClickHouse<br/>:8123 / :9000")]
  end

  AUTH2 --- PG_A & RD_A
  GAME2 --- PG_G
  BILL2 --- PG_B & RD_B & OB_B
  LEAD2 --- PG_L & RD_L
  PROF2 --- PG_P & RD_P
  SHOP2 --- PG_S & RD_S & OB_S

  PAY2 --- PG_PAY & RD_PAY & OB_PAY
  INV2 --- PG_INV & RD_INV & OB_INV
  AUTHR2 --- PG_AUTHR & RD_AUTHR & OB_AUTHR
  HIST2 --- PG_HIST
  ANAL2 --- CH
```

| Service | gRPC | Metrics | PostgreSQL | Redis | Outbox subjects |
|---------|------|---------|------------|-------|-----------------|
| Auth | 50051 | 9091 | 5460 | 6379 | — |
| Game | 50052 | 9092 | 5461 | — | — (direct NATS) |
| Billing | 50053 | 9093 | 5462 | 6381 | `balance.updated` |
| Leaderboard | 50054 | 9094 | 5463 | 6382 | — |
| Shop | 50055 | 9095 | 5465 | 6383 | `shop.purchased` |
| Analytics | 50057 | 9106 | — | — | — (NATS ingest) |
| Payment | 50058 | 9103 | 5467 | 6386 | `payment.completed` |
| Inventory | 50059 | 9096 | 5466 | 6384 | `inventory.item.created` (+ updated/deleted in stream) |
| Profile | 50060 | 9099 | 5464 | 6385 | — |
| Authors | 50061 | 9104 | 5468 | 6387 | `author.upserted` |
| History | 50062 | 9105 | 5469 | — | — (NATS ingest) |
| Gateway ×3 | HTTP 8081–8083 | per instance | — | 6379 (shared) | — |
| Balancer | HTTP 8079 | 9098 | — | — | — |
| Fulfillment | — | 9101 | — | — | — |
| Notification | — | 9102 | — | — | — |
| NATS Hub | — | 9097 | — | — | — |

> **Note:** Auth has **no** outbox in code (unlike v1.0.6 Miro). Inventory compose uses `INVENTORY_DRIVER=postgres`; Mongo driver remains in repo for legacy/experiment.

Every gRPC service exposes **`/health`** + **`/ready`** on its metrics HTTP port.

---

## 4. NATS JetStream — stream EVENTS

```mermaid
flowchart TB
  subgraph producers["Producers"]
    G2["Gateway<br/>event.user.registered"]
    GM["Game<br/>score.updated direct"]
    OB_ALL["Outbox workers<br/>Billing · Shop · Payment · Authors · Inventory"]
    SHOP_FAIL["Shop<br/>shop.purchase.failed direct"]
  end

  subgraph nats["NATS cluster"]
    N1["nats-1 :4222"]
    N2["nats-2 :4223"]
    N3["nats-3 :4224"]
    STREAM["Stream EVENTS<br/>subjects: event.&gt; score.updated user.registered<br/>shop.purchased payment.completed<br/>inventory.item.*"]
  end

  subgraph hub["NATS Hub"]
    HUB2["Ensures stream · logs event.&gt;"]
  end

  subgraph consumers["Durable consumers"]
    LB2["Leaderboard<br/>score.updated"]
    PR2["Profile<br/>score.updated · event.user.registered"]
    SH2["Shop<br/>inventory.item.created"]
    BI2["Billing<br/>score.updated"]
    HI2["History ingest<br/>payment.completed author.upserted<br/>shop.purchased score.updated user.registered"]
    AN2["Analytics ingest<br/>same 5 subjects"]
    GW2["Gateway WS hub<br/>score.updated"]
  end

  G2 & GM & OB_ALL & SHOP_FAIL --> STREAM
  STREAM --- N1 & N2 & N3
  N1 --> HUB2
  STREAM -.-> LB2 & PR2 & SH2 & BI2 & HI2 & AN2 & GW2
```

### Outbox pattern (reference: Inventory / Billing)

```mermaid
flowchart LR
  SVC["Service handler"] --> TX["DB transaction<br/>business row + INSERT outbox"]
  TX --> COMMIT["COMMIT"]
  COMMIT --> WORKER["Outbox worker<br/>poll every 1–2s"]
  WORKER --> PUB["js.Publish event_type"]
  PUB --> NATS2["NATS JetStream"]
  WORKER --> MARK["UPDATE outbox processed=true"]
```

Services with outbox table + worker today: **Billing, Shop, Payment, Authors, Inventory**.

---

## 5. Kafka purchase path (Week 5 / Order + Assembly)

```mermaid
sequenceDiagram
  autonumber
  participant U as User
  participant GW as Gateway
  participant SH as Shop
  participant BI as Billing
  participant PG as Shop PG + outbox
  participant OW as Shop outbox worker
  participant NATS as NATS
  participant K as Kafka :9092
  participant F as Fulfillment
  participant N as Notification

  U->>GW: POST /api/shop/purchase
  GW->>SH: PurchaseItem gRPC
  SH->>BI: SpendCurrency gRPC
  BI-->>SH: ok
  SH->>PG: TX stock↓ inventory purchases outbox shop.purchased
  PG-->>SH: commit
  SH->>K: purchase.paid post-commit
  OW->>PG: poll outbox
  OW->>NATS: shop.purchased
  NATS-.->History & Analytics ingest

  K->>F: purchase.paid
  F->>F: assemble delay ~10s
  F->>K: purchase.fulfilled
  K->>N: notify Telegram optional

  Note over SH,NATS: shop.purchase.failed = direct NATS on refund failure
```

| Kafka topic | Producer | Consumers |
|-------------|----------|-----------|
| `purchase.paid` | Shop (after commit) | Fulfillment, Notification |
| `purchase.fulfilled` | Fulfillment | Notification (optional Shop) |

Env: `KAFKA_BROKERS=kafka:9092` · host maps **`19092:9092`** (Game metrics owns host `:9092`).

---

## 6. Key sync business flows

```mermaid
flowchart TB
  subgraph shop_ticket["Shop purchase tickets"]
    ST1["Gateway"] --> ST2["Shop gRPC"]
    ST2 --> ST3["Billing SpendCurrency"]
    ST2 --> ST4["Shop PG TX + outbox"]
    ST4 --> ST5["Kafka purchase.paid"]
  end

  subgraph shop_merch["Shop merch gate"]
    SM1["Gateway"] --> SM2["Shop gRPC"]
    SM2 --> SM3["Payment CanPurchaseMerch"]
    SM3 --> SM4["fulfill if subscribed"]
  end

  subgraph game_lb["Game → leaderboard real-time"]
    GL1["Game SaveScore"] --> GL2["NATS score.updated"]
    GL2 --> GL3["Leaderboard PG + Redis ZSET"]
    GL2 --> GL4["Gateway WS broadcast"]
    GL4 --> GL5["React client"]
  end

  subgraph inv_shop["Inventory → shop catalog"]
    IS1["Inventory CreateItem + outbox"] --> IS2["NATS inventory.item.created"]
    IS2 --> IS3["Shop creates shop.items row"]
  end
```

---

## 7. Gateway resilience

```mermaid
flowchart LR
  GW3["Gateway HTTP"] --> CB["Circuit breaker<br/>all gRPC clients"]
  CB -->|closed| GRPC["Auth Billing Shop Payment<br/>Inventory Authors History Analytics<br/>Game Leaderboard Profile"]
  CB -->|open| E503["HTTP 503 Service Unavailable"]
  GW3 --> RL["Rate limit Redis"]
  GW3 --> JWT["JWT + Auth ValidateToken"]
  GW3 --> RBAC["RequireRole user author admin"]
```

---

## 8. Observability stack

```mermaid
flowchart TB
  subgraph services["All services"]
    M["/metrics /health /ready"]
  end

  PROM2["Prometheus :9090"]
  GRAF2["Grafana :3000"]
  JAEG2["Jaeger :16686"]
  OTLP["OTLP gRPC :4317"]
  PGX["postgres-exporter :9187"]
  RDX["redis-exporter :9121"]
  NEX2["nats-exporter :7777"]

  M --> PROM2
  PGX & RDX & NEX2 --> PROM2
  PROM2 --> GRAF2
  services -.-> OTLP --> JAEG2
```

---

## 9. Frontend surface (v1.0.7)

| Route / area | Backend |
|--------------|---------|
| Auth, games, billing, shop, inventory, leaderboard, profile | Gateway REST → gRPC |
| `/history` | `GET /api/history` → History :50062 |
| Authors, analytics admin, payment checkout UI | API clients exist · partial UI |
| Hanoi `/game/hanoi` | Game service |

---

## 10. Miro legend mapping (for board updates)

| Miro shape | v1.0.7 element |
|------------|----------------|
| Hexagon client | React :5173 |
| Blue LB / GW | Balancer :8079 · Gateway :8081–8083 |
| Yellow service box | gRPC microservice (table §3) |
| Cylinder PG | Per-service Postgres :546x |
| Diamond Redis | Per-service Redis :638x (Auth :6379) |
| Purple outbox | Billing · Shop · Payment · Authors · Inventory |
| Purple NATS hub | JetStream ×3 + NATS Hub stream EVENTS |
| Kafka bus | purchase.paid / purchase.fulfilled |
| Green obs block | Prometheus · Grafana · Jaeger · NATS exporter |

**Delta vs v1.0.6 PNG:** Shop outbox added (19.08) · Authors/History/Analytics/Fulfillment/Notification explicit · Kafka host port 19092 · Gateway CB on all gRPC · History FE · OpenAPI auth routes v1.0.7.

---

## Related docs

- [`../EH_SCHEMAS.md`](../EH_SCHEMAS.md) — compact ASCII schemas
- [`HOW_DOES_EH_WORK.md`](./HOW_DOES_EH_WORK.md) — interview narrative
- [`../../contracts/events/PURCHASE_KAFKA.md`](../../../contracts/events/PURCHASE_KAFKA.md) — Kafka payloads
- [`../../history/2026-08/14.08.2026/TODO_FINAL_LIST.md`](../../history/2026-08/14.08.2026/TODO_FINAL_LIST.md) — P0–P2 status
