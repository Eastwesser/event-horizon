# MCP + RAG status — Event Horizon (13.08.2026)

## Done (Stage 2 / Priority 5)

- [x] `services/mcp` — mark3labs/mcp-go stdio server
- [x] Tools: `nats_list_streams`, `nats_list_consumers`, `postgres_query` (SELECT-only), `redis_get`, `redis_keys`, `search_prydwen`
- [x] Offline TF-IDF RAG over Prydwen
- [x] `.cursor/mcp.json` example + Prydwen AI docs filled
- [x] Unit tests: pg validate + rag Outbox query

## Not in this pass

- Qdrant / paid embeddings (upgrade path documented)
- Docker image for MCP (stdio is meant to run on the host for Cursor)

## Build

```bash
(cd services/mcp && CGO_ENABLED=0 go build -ldflags="-s -w" -o mcp-server ./cmd/main.go)
# point Cursor MCP at that binary — see services/mcp/README.md
```
