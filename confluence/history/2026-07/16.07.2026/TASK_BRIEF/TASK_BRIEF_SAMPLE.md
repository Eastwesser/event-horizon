TASK_BRIEF.md

Task

Evaluate MCP (Model Context Protocol) server integration for Event Horizon - determine if it's needed now, in the near future, or not at all, with clear reasoning and a practical alternative strategy.

Context

Event Horizon is a high-load gaming platform (10k RPS target) with microservices (Go + NATS + PostgreSQL + Redis). Current priorities are backend stability and service development (Shop, Notification, Analytics planned). The user asked specifically about MCP server needs during active development phase.

Scope

In scope

· Analysis of MCP server benefits/costs for current architecture
· Evaluation against existing infrastructure (NATS, Prometheus, Grafana, HTTP endpoints)
· Recommendation with timeline (now / v2.0 / never)
· Lightweight alternative approach using current stack

Out of scope

· Implementation code
· Full API documentation (OpenAPI - separate task)
· MCP server development
· Third-party AI agent integration

Files / modules

· README.md - for roadmap context
· Architecture docs (if any)
· Existing API endpoints: /api/auth/*, /api/game/*, /api/leaderboard, /api/profile, /api/billing/*, WebSocket /ws/leaderboard

Acceptance criteria

· Clear yes/no decision with justification
· If yes: timeline (when to implement) provided
· If no: alternative approach specified using existing tools
· Analysis considers: team size (solo), development phase, planned services

Constraints

· Do not change: Current microservices architecture
· Do not break: Existing API contracts, NATS streams, database schemas
· Do not use: External AI services or commercial MCP implementations
· Must fit: Solo developer workflow, active development phase

Verification

· Decision is documented in this brief
· Reasoning is consistent with repo's roadmap (v1.0.5 → future services → k3s)
· Alternative approach can be tested with curl or nats CLI

Plan for the agent

1. Review the existing stack (already done from README).
2. Analyze MCP relevance based on:
   · Current state (still in active development)
   · Team size (solo dev - overhead matters)
   · Existing infrastructure (NATS, Prometheus, HTTP APIs)
   · Planned features (OpenAPI, logging, monitoring)
3. Propose clear recommendation.
4. If "no now": outline a lightweight "manual AI integration" approach using:
   · Prometheus API for metrics queries
   · NATS CLI/pub-sub for admin actions
   · HTTP endpoints for user data
5. Output brief and wait for approval (no code).