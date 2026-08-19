# Event Horizon — DONE: Audit exclusions resolved (2026-08-19)

This file documents the items from the audit exclusion list that are **not owed as tech debt** because the current implementation/design already satisfies them (or they are intentionally excluded).

## Not owed / already satisfied by design

### Squirrel
- **Resolution:** No production code usage of `github.com/Masterminds/squirrel`.
- Rationale: audit explicitly excluded it; SQL stays raw / typed.

### swaggo annotations
- **Resolution:** No `swaggo` annotations added.
- Rationale: gateway uses hand-maintained OpenAPI in `docs/openapi.yaml` and serves Swagger UI at `/docs`.

### zap
- **Resolution:** Project uses `log/slog`, not `uber-go/zap`.

### Envoy
- **Resolution:** Envoy proxy/config is not part of the deployed runtime.
- Note: `envoyproxy/protoc-gen-validate` exists only as a protobuf validation generator dependency.

