# Recap перед фронтом / собесом (13.08.2026)

## MCP — реально ли done?
**Да, для Stage 2.** Есть `services/mcp/mcp-server` (stdio), инструменты NATS/PG SELECT/Redis/`search_prydwen`, `.cursor/mcp.json`, тесты validate+RAG.  
**Не сделано (и не блокирует):** Qdrant/эмбеддинги, Docker-образ MCP (не нужен для Cursor), live-проверка consumers в вашем кластере глазами.

## Что осталось по продукту (честный backlog)

| Приоритет | Что | Статус |
|-----------|-----|--------|
| Frontend lag | Payment/Authors/Analytics UI, shop merch UX, игры | ❌ дальше |
| Hanoi Towers | `FRONTEND_KHANOY_TOWERS.md` | ❌ следующий крупный FE |
| OpenAPI sync | новые routes в `docs/openapi.yaml` | 🟡 |
| Coverage ≥70% / testcontainers | частично | 🟡 |
| Notification harden e2e | частично | 🟡 |
| k3s/Ansible prod | tech_debt | 🟡 позже |
| MCP/RAG Stage2 | done | ✅ |
| Stage1 services push | Payment→Analytics pushed вами | ✅ |

## Собес red_mad_robot
Читать `confluence/agents/prydwen_knowledge/README.md` → затем `8.legend_projects/03_PROJECT_EVENT_HORIZON.md` → брокеры/Outbox/Postgres/Go concurrency.
