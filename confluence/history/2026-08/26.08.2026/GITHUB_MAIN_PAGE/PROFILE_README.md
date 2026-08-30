<div align="center">

# Denis Matveev · Eastwesser

**Backend engineer** — Go microservices, event-driven systems, async Python, production DevOps

![Typing SVG](https://readme-typing-svg.demolab.com?font=Fira+Code&size=22&pause=1000&color=6366F1&width=500&lines=Go+%7C+Python+%7C+gRPC+%7C+NATS;Building+Event+Horizon+%F0%9F%8E%AE;Microservices+%E2%86%92+Production)

[![Event Horizon](https://img.shields.io/badge/Flagship-Event%20Horizon-6366f1?style=for-the-badge)](https://github.com/Eastwesser/event-horizon)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![Python](https://img.shields.io/badge/Python-3.11+-3776AB?style=flat-square&logo=python&logoColor=white)](https://python.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose%20%7C%20k3s-2496ED?style=flat-square&logo=docker&logoColor=white)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=flat-square&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![NATS](https://img.shields.io/badge/NATS-JetStream-27AAE1?style=flat-square&logo=nats.io&logoColor=white)](https://nats.io/)

<p align="center">
  <img src="https://github-readme-stats.vercel.app/api?username=Eastwesser&show_icons=true&theme=radical" alt="GitHub stats" height="165" />
  <img src="https://github-readme-stats.vercel.app/api/top-langs/?username=Eastwesser&layout=compact&theme=radical" alt="Top languages" height="165" />
</p>
<p align="center">
  <img src="https://github-readme-streak-stats.herokuapp.com/?user=Eastwesser&theme=radical" alt="GitHub streak" height="165" />
</p>

</div>

---

## About me

I'm a **Golang & Python developer** focused on backend systems that stay understandable under load: clear service boundaries, explicit config, observability from day one, and deploy paths you can actually run locally.

I care about routine craft — migrations, health checks, OpenAPI contracts, tests that protect real behavior — not just greenfield demos. Current flagship: **[Event Horizon](https://github.com/Eastwesser/event-horizon)**, a microservices game platform I use as both product and reference architecture.

---

## Featured project — Event Horizon

[![GitHub stars](https://img.shields.io/github/stars/Eastwesser/event-horizon?style=social&label=Star)](https://github.com/Eastwesser/event-horizon/stargazers)
[![License MIT](https://img.shields.io/badge/License-MIT-green?style=flat-square)](https://github.com/Eastwesser/event-horizon/blob/main/LICENSE)

High-load-friendly **gaming platform** (v1.0.7): real-time leaderboard via NATS + WebSockets, shop economy, JWT auth, target design ~10k RPS.

```text
React → Balancer → Gateway (×3) → gRPC mesh → Postgres / Redis / NATS / ClickHouse
```

| Layer | What's inside |
|-------|----------------|
| **Services** | 14+ Go microservices · Clean Architecture (handler → service → repository) |
| **API** | HTTP Gateway `/api/` → gRPC · hand-maintained OpenAPI + Swagger UI |
| **Auth** | JWT roles (`user` \| `author` \| `admin`) · sessions in Redis · bcrypt cost 12 |
| **Events** | NATS JetStream · transactional **outbox** (Inventory, Billing, Shop, Payment, Game) |
| **Data** | PostgreSQL per domain · Redis · MongoDB (inventory) · ClickHouse (analytics) |
| **Ops** | Prometheus, Grafana, Jaeger · `/health` + `/ready` · GitHub Actions → Docker Hub · Ansible · k3s & Patroni stubs |

```bash
git clone https://github.com/Eastwesser/event-horizon.git
cd event-horizon
# create .env with JWT_SECRET locally — never commit .env
make deploy && make status
```

Architecture docs: [`confluence/architecture/`](https://github.com/Eastwesser/event-horizon/tree/main/confluence/architecture) in the repo.

---

## Other projects

| Project | Description | Link |
|---------|-------------|------|
| **Car-sharing LUXURY** | Fullstack luxury car rental app (Go + TypeScript) | [car-rental](https://github.com/Eastwesser/car-rental) |
| **CloudMiu** | Telegram bot (aiogram3) with YandexGPT & Kandinsky API integration | [CloudMiu](https://github.com/Eastwesser/CloudMiu) |

---

## Tech stack

### Languages & frameworks

**Go (primary)** — Gin gateway, gRPC, `testify`, Goose migrations, structured logging (`slog`). Event Horizon: circuit breakers, rate limits, interceptor chains (recovery, validate, logging).

**Python** — FastAPI, Django, Flask · async (`asyncio`, `aiohttp`) · Aiogram3 for Telegram bots · `pytest`.

**Frontend touchpoints** — TypeScript, React (Event Horizon SPA).

### API & communication

REST · gRPC · WebSocket · JSON-RPC · OpenAPI / Swagger

### Data & storage

| Type | Tools |
|------|--------|
| **SQL** | PostgreSQL, SQLite · sqlx · SQLAlchemy 2.0 · asyncpg |
| **NoSQL / cache** | MongoDB · Redis (sessions, cache, pub/sub, leaderboard sorted sets) |
| **OLAP** | ClickHouse (analytics, time-series) |
| **ORM / queries** | GORM · SQLAlchemy (Python side projects) |

### Messaging & patterns

NATS JetStream · RabbitMQ · Apache Kafka · event-driven architecture · **outbox pattern** · CQRS (basics) · DLQ · retries with jitter

### Observability

Prometheus · Grafana · Jaeger · Loki + Promtail · ELK (legacy exposure)

### DevOps & infra

Docker · Docker Compose · multi-stage builds · Kubernetes (k3s) · Helm · HPA · liveness/readiness probes · GitHub Actions · GitLab CI · Ansible · Consul (basics) · Patroni HA pilots

### Security & quality

JWT · OAuth2 · env-based secrets (no committed `.env`) · Go table-driven tests · gomock · k6 load testing · golangci-lint

### Tooling

Git · Postman · Markdown · Confluence · Celery (Python jobs)

---

## Certifications

| Course | Certificate |
|--------|-------------|
| Go (Golang) — first acquaintance | [Stepik](https://stepik.org/cert/2525739?lang=en) |
| Good, Good Python (Sergey Balakirev) | [Stepik](https://stepik.org/cert/2165774) |
| Quick Start in FastAPI Python | [Stepik](https://stepik.org/cert/2363817) |
| SQL Introduction | [Stepik](https://stepik.org/cert/2336687) |

---

## Connect

- **GitHub:** [@Eastwesser](https://github.com/Eastwesser)
- **Email:** eastwesser@gmail.com

Questions, code review, or collaboration ideas — feel free to reach out.

---

<div align="center">

*If Event Horizon helped you learn microservices in Go, a ⭐ on the repo means a lot.*

**Ad Victoriam!**

</div>
