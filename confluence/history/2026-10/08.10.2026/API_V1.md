# `/api/v1` — где живёт версия

## Короткий ответ

**Нет**, мы **не** добавляли `/api/v1` внутрь каждого микросервиса.  
Канон только на **HTTP Gateway** (+ FE + OpenAPI + k6).

| Слой | Префикс |
|------|---------|
| Gateway Gin routes | `/api/v1/...` |
| Legacy rewrite | `/api/foo` → `/api/v1/foo` |
| FE `baseURL` | `/api/v1` |
| `docs/openapi.yaml` + embed | `/api/v1/...` |
| k6 scripts | `/api/v1/...` |
| Auth / Game / Shop / … | **gRPC** (`:5005x`) — без HTTP `/api` |

Микросервисы не «имеют REST v1» — у них protobuf RPC. Клиент всегда ходит:  
`React → :8079 → Gateway /api/v1/* → gRPC`.

## Сверка с документами (09.10)

- OpenAPI / gateway embed — **в sync**  
- FE / k6 — **в sync**  
- `SECURITY_*`, `V1_1_0_FINAL_STATUS`, `MIRO_STICKERS_PASTE` — **v1**  
- Старые упоминания `/api/api/profile` в Q&A — история бага, не текущий контракт  

Paste для Miro: `miro/MIRO_STICKERS_PASTE.md`.
