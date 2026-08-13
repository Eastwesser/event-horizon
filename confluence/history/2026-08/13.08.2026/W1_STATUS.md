# Week 1 status — Event Horizon (13.08.2026)

## Done

- [x] Cursor: `.cursor/rules/event-horizon.mdc`, `.cursor/agents/event-horizon-agent.md`, `.cursor/settings.json`
- [x] `go.work` + root `Taskfile.yml` (`build`, `test`, `w1-check`, `gen-proto` note)
- [x] `contracts/` (buf config + vendored `validate.proto` + README)
- [x] gRPC interceptors on all live gRPC services: Recovery → Logger (slog) → Validate → (otel)
- [x] `validate.rules` on all service `.proto` request messages
- [x] Hand-written `Validate()` in each `services/*/proto/validate.go` (offline-safe)
- [x] `internal/model` present for auth/billing/game/leaderboard/profile/shop (+ inventory already)
- [x] Gateway Swagger: `GET /openapi.yaml`, `GET /docs`
- [x] Auth `User` moved to `internal/model`

## Explicitly deferred (not W1 blockers)

- Full `buf generate` / `protoc-gen-validate` codegen → run `task gen-proto` when you have network + plugins
- Migrating every proto into `contracts/proto/` (layout ready; services still own current protos)
- Replacing Gin Gateway with ogen/grpc-gateway (rejected unless you ask)
- Method-level Prometheus metrics / Zap migration → Week 7–8

## Your rebuild

Binaries already built under `services/*/*-service`. Docker push is yours:

```bash
for svc in auth gateway billing inventory profile shop leaderboard game; do
  docker build -f Dockerfile.${svc}.bin -t eastwesser/${svc}:latest .
done
```

If `go build` under `go.work` acts up: `GOWORK=off go build ...`
