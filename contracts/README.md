# Contracts (Week 1)

## Layout

- `proto/` — shared buf module (future home for canonical `.proto` files)
- `third_party/validate/` — offline copy of `validate.proto` for local edits
- `docs/openapi.yaml` (repo root `docs/`) — **canonical HTTP OpenAPI** for Gateway

## Current practice

Service protos still live under `services/*/proto/*.proto` (existing EH layout).
Week-1 adds:

1. `import "validate/validate.proto"` + field rules on request messages
2. Hand-written `Validate()` methods in each proto package (offline-safe)
3. gRPC `Validate` interceptor that calls `Validate()` when present

When you have network + tools:

```bash
# install plugins into ./bin, then:
task gen-proto
```

Until then, do **not** require `buf generate` to build/run services.

## OpenAPI / ogen

EH uses Gin Gateway + hand-maintained `docs/openapi.yaml`.
Ogen / grpc-gateway codegen is optional and must not replace Gateway without an explicit decision.
Gateway serves:

- `GET /openapi.yaml`
- `GET /docs` (Swagger UI)
