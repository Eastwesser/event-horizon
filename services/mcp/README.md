# Event Horizon MCP server

stdio MCP server for Cursor / Claude Desktop. Tools:

| Tool | Purpose |
|------|---------|
| `nats_list_streams` | JetStream streams |
| `nats_list_consumers` | Durable consumers on a stream |
| `postgres_query` | Read-only `SELECT` / `WITH` (mutations rejected) |
| `redis_get` / `redis_keys` | Cache inspection |
| `search_prydwen` | Offline TF-IDF RAG over `prydwen_knowledge` |

## Build

```bash
cd services/mcp
CGO_ENABLED=0 go build -ldflags="-s -w" -o mcp-server ./cmd/main.go
```

## Env

| Var | Default | Notes |
|-----|---------|-------|
| `NATS_URL` | `nats://localhost:4222` | Cluster URL list OK |
| `MCP_POSTGRES_DSN` | empty | Prefer read-only role |
| `REDIS_ADDR` | `localhost:6379` | Auth redis or any |
| `PRYDWEN_ROOT` | `confluence/agents/prydwen_knowledge` | Absolute path recommended |
| `RAG_INDEX_PATH` | empty | Optional JSON cache for TF-IDF index |

## Cursor config

Add to Cursor MCP settings (or project `.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "event-horizon": {
      "command": "/home/denismatveev/event_horizon/services/mcp/mcp-server",
      "args": [],
      "env": {
        "NATS_URL": "nats://localhost:4222",
        "REDIS_ADDR": "localhost:6379",
        "MCP_POSTGRES_DSN": "postgres://eventhorizon:eventhorizon@localhost:5460/eventhorizon?sslmode=disable",
        "PRYDWEN_ROOT": "/home/denismatveev/event_horizon/confluence/agents/prydwen_knowledge"
      }
    }
  }
}
```

## Smoke

```bash
# with compose up (nats):
PRYDWEN_ROOT=$PWD/confluence/agents/prydwen_knowledge \
  NATS_URL=nats://localhost:4222 \
  ./services/mcp/mcp-server
# then from Cursor: call search_prydwen query="Outbox Shop"
```
