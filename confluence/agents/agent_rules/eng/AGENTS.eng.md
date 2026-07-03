# EventHorizon — Rules for AI Agents

## Purpose

This file is the primary operational guide for AI agents (including me, Hoho) working in the EventHorizon repository.

The agent must:

- understand the microservices architecture;
- follow established conventions;
- know where to place code;
- be able to start and test the system;
- avoid unsafe or destructive changes;
- ask when uncertain.

---

## Project Summary

| Field | Value |
|-------|-------|
| Project Name | EventHorizon |
| Project Type | Microservices platform (5+ services) |
| One-line Description | Gaming platform with real-time leaderboard, fair validation, and NATS event bus |
| Primary Users | Players (hexagonal puzzle "Nikusia — Pancake Maker") |
| Business Context | Games without FOMO, fair game design, ethical business |
| Lifecycle Stage | MVP + load testing, frontend in development |
| Owners | Captain (lead developer) + Hoho (AI architect) |
| Main Branch | main |
| Repository State | Active development, stable backend |

---

## Agent Operating Principles

The agent should:

- prefer the minimal safe change that solves the problem;
- preserve existing architecture (microservices, NATS, gRPC);
- update documentation when APIs or configs change;
- verify the result before finishing;
- avoid refactoring without request.

### What We Optimize For

1. **Correctness** — the game must not deceive players (see manifesto).
2. **Maintainability** — code should be understandable a year from now.
3. **Speed** — but not at the expense of the first two.

### Forbidden by Default

- Adding FOMO mechanics (artificial scarcity, progress reset).
- Changing architecture without discussion.
- Adding new dependencies unnecessarily.
- Ignoring failing tests.

---

## Sources of Truth

| Source | Path | When to Use |
|--------|------|-------------|
| Ethical Manifesto | `confluence/ethics/MANIFEST.md` | Any game design changes |
| Technical Debt | `confluence/tech_debt/DEBT_LIST.md` | Planning improvements |
| API Documentation | `docs/openapi.yaml` | Gateway endpoint changes |
| Project Structure | `README.md` | Understanding the big picture |

**Priority:** code > documentation. If they conflict, ask the human.

---

## Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Backend | Go | 1.25.6 |
| Frontend | React + TypeScript | Vite 8 |
| gRPC | protobuf + grpc-go | latest |
| Event Bus | NATS JetStream | 2.10 |
| Database | PostgreSQL | 16 |
| Cache | Redis | 7 |
| Authentication | JWT (HS256) | — |
| Orchestration | Docker Compose | — |

### Key Libraries

| Area | Library | Purpose |
|------|---------|---------|
| gRPC | google.golang.org/grpc | Internal calls |
| NATS | nats.go | Events |
| PostgreSQL | jackc/pgx | DB driver |
| Redis | redis/go-redis | Cache & leaderboard |
| HTTP (gateway) | gin-gonic/gin | API Gateway |
| WebSocket | gorilla/websocket | Real-time updates |
| Frontend state | zustand | Game state management |

---

## Architecture

- **Style:** Microservices with event bus (NATS)
- **Services:**
  - `auth` (50051) — JWT, registration, login
  - `game` (50052) — Validation, score calculation
  - `billing` (50053) — Lamps, tickets
  - `leaderboard` (50054) — Top-10, Redis
  - `gateway` (8080) — HTTP + WebSocket

### Data Flow
Frontend → Gateway → Game → NATS → Leaderboard → Redis
↓
Billing → PostgreSQL

text

### Architecture Rules

- Services communicate via gRPC (sync) or NATS (async).
- Database per service.
- Frontend does NOT calculate scores — only server (fair validation).
- No FOMO mechanics (see manifesto).

---

## Repository Structure

```text
event_horizon/
├── services/              # Microservices (auth, game, billing, leaderboard, gateway)
├── frontend/              # React + TypeScript
├── confluence/            # Documentation, manifesto, tech debt
├── deployments/           # Docker Compose configs
├── scripts/               # Helper scripts
├── docs/                  # OpenAPI specification
├── load_test_results/     # Load test results
├── Makefile               # make all, make restart, make status
└── README.md
Placement Rules

Path	Responsibility
services/<name>/cmd/	Entry point (main.go)
services/<name>/internal/	Internal logic (config, handler, service, repository)
services/<name>/proto/	gRPC contracts
frontend/src/components/	React components
frontend/src/store/	Zustand state (gameStore.ts)
confluence/ethics/	Ethical manifesto
confluence/tech_debt/	Technical debt
Environment Setup

Required Tools

Tool	Version	Installation
Go	1.25.6	pacman -S go
Node.js	20+	pacman -S nodejs npm
Docker	latest	pacman -S docker
Docker Compose	v2	included with Docker
protoc	25.1	Download binary
protoc-gen-go	latest	go install
Quick Start Commands

bash
# Everything at once
cd ~/event_horizon
make all          # Starts Docker + Go services

# Check status
make status

# Stop everything
make stop

# Restart
make restart

# Logs
tail -f /tmp/gateway.log
tail -f /tmp/game.log
Development Commands

Task	Command
Start Docker	make up
Stop Docker	make down
Start all services	make all
Stop services	make stop
Regenerate proto	cd services/<name> && protoc ...
Start frontend	cd frontend && npm run dev
Build frontend	cd frontend && npm run build
Testing Guide

Load Testing (bombardier)

bash
# POST /api/game/submit
bombardier -c 1000 -n 10000 -m POST \
  -H "Content-Type: application/json" \
  -b '{"user_id":"<UUID>","game_id":"hexagon","level":3,"seed":"test","moves":[]}' \
  http://localhost:8080/api/game/submit
WebSocket Testing

bash
# Manual check
npx wscat -c ws://localhost:8080/ws/leaderboard

# Go script
cd scripts && go run websocket_load.go -n 1000 -d 30
Code Style

Entity	Preferred	Example
Go files	snake_case	user_repo.go
React components	PascalCase	HexGrid.tsx
Go functions	camelCase	getUserByID
Go constants	PascalCase / UPPER	MaxRetries
TS variables	camelCase	userScore
TS types	PascalCase	HexTile
Comments

Go: Comment exported functions.
TS: Comments for complex logic.
Manifest: Russian (players and team).
Security Boundaries

Hard Rules

No secrets in code (JWT secret in env).
Validate all user inputs on server.
Player cannot cheat scores (server calculates).
Sensitive Areas

Auth (JWT, passwords) — services/auth/
Billing (lamps, tickets) — services/billing/
Game validation — services/game/games/hexagons/
Escalation

Ask the human before:

changing JWT logic;
changing score calculation rules;
deleting data;
changing deployment configs.
When the Agent Should Stop and Ask

Task violates the ethical manifesto (FOMO, hidden reset).
Architecture change needed (e.g., adding a new service).
Documentation and code diverge.
Tests fail for unknown reasons.
Known Pitfalls

user_id in localStorage — sometimes not saved, need token fallback.
WebSocket — closes during component mount/unmount.
NATS — may not resubscribe after restart (needs durable consumer).
Leaderboard — scores should accumulate, not replace.
Pre-Commit Checklist

make all passes without errors?
make status shows all services?
Frontend starts (npm run dev)?
Login/registration work?
Game (drag-n-drop) works?
Score saves to leaderboard?
Maintenance Checklist

Update this file when architecture changes.
Add new commands to Makefile.
Document bugs in tech_debt/DEBT_LIST.md.