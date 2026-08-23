# Event Horizon — Still Tech Debt (as of 2026-08-19)

This file enumerates the “audit exclusions” from `TODO_FINAL_LIST.md` and marks which ones are:
- already satisfied by current implementation/design choice, or
- still technically outstanding (i.e., real remaining work).

## Still outstanding (real work)

### 0) Game Outbox for `score.updated`
**Status:** implemented (23.08.2026). Rebuild/push `eastwesser/game:latest` to run in cluster.
See `confluence/architecture/GAME_OUTBOX.md`.

### 1) Live Boosty signature verification
**Status:** partial (shared secret equality check), not full Boosty signature verification. Boosty treated as secondary/manual path.

**What exists now**
- `POST /api/payment/webhook` ultimately calls `PaymentService.ConfirmPayment(..., webhookSecret)` and compares it to `PAYMENT_WEBHOOK_SECRET`.
- Evidence:
  - `services/payment/internal/service/payment_service.go` uses:
    - `if s.webhookSecret != "" && webhookSecret != s.webhookSecret { return nil, model.ErrUnauthorized }`
  - gRPC handler passes the request field `webhook_secret`:
    - `services/payment/internal/handler/grpc_handler.go` `ConfirmPayment(... req.GetWebhookSecret())`
  - Config field:
    - `services/payment/internal/config/config.go` `WebhookSecret`

**What’s missing**
- No HMAC/signature verification flow (e.g. validating Boosty-provided signature header(s), timestamp/replay protection, canonical payload).

**Target outcome**
- Implement Boosty’s official signature verification:
  - verify signature using Boosty scheme (HMAC/secret + payload),
  - validate timestamp/nonce if required,
  - reject replayed/incorrect requests.

### 2) ≥70% coverage across whole services (enforcement + remaining gaps)
**Status:** not fully achieved / not enforced by CI.

**What exists now**
- Many unit tests exist (and some critical areas were improved), and there is a built-in coverage task:
  - `Taskfile.yml` `test-coverage` runs `go test -coverprofile=coverage.out` and prints `go tool cover -func`.

**What’s missing**
- There is no current “coverage gate” guaranteeing:
  - every service (or every critical package) meets ≥70%,
  - and coverage deltas remain stable after changes.

**Target outcome**
- Run coverage measurement per gRPC service (`task test-coverage`) and add tests in the packages that are below threshold.
- (Optional but recommended) add a CI gate/threshold so this doesn’t regress.

### 3) MCP/Qdrant polish (embeddings + vector DB upgrade path)
**Status:** MCP is done with offline TF-IDF; Qdrant embeddings are still not implemented.

**What exists now**
- MCP server is implemented with:
  - `search_prydwen` tool using **offline TF-IDF**.
- Evidence:
  - `services/mcp/README.md` documents TF-IDF `search_prydwen`.
  - `confluence/done/2026-08-13/MCP_RAG_STATUS.md` explicitly lists “Not in this pass: Qdrant / paid embeddings”.

**What’s missing**
- No Qdrant (or similar) running integration for embeddings.
- No embedding/ANN storage + retrieval upgrade.

**Target outcome**
- Upgrade Stage 2 RAG:
  - generate embeddings (OpenAI/Ollama or other plan),
  - store in Qdrant,
  - keep MCP tool contract stable (`search_prydwen` signature stays the same).

Resolved audit exclusions (squirrel/swaggo/zap/Envoy) were moved to:
`confluence/tech_debt/DONE/AUDIT_EXCLUSIONS_RESOLVED.md`.

