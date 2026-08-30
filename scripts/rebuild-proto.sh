#!/usr/bin/env bash
# Regenerate Go + gRPC stubs from services/*/proto/*.proto
# Run from repo root: bash scripts/rebuild-proto.sh [service ...]
#
# Note: Gateway is HTTP-only — there is no services/gateway/proto/gateway.proto.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
PROTO_INCLUDE="${ROOT}/contracts/third_party"

if ! command -v protoc >/dev/null 2>&1; then
  echo "protoc not found. Install protobuf + protoc-gen-go + protoc-gen-go-grpc."
  exit 1
fi

DEFAULT_SERVICES=(
  auth game analytics billing shop leaderboard profile payment authors history inventory
)

if [[ $# -gt 0 ]]; then
  SERVICES=("$@")
else
  SERVICES=("${DEFAULT_SERVICES[@]}")
fi

for svc in "${SERVICES[@]}"; do
  proto="services/${svc}/proto/${svc}.proto"
  if [[ ! -f "$proto" ]]; then
    echo "SKIP $svc — no $proto"
    continue
  fi
  echo "===== protoc $svc ====="
  (
    cd "services/${svc}"
    protoc -I . -I "${PROTO_INCLUDE}" --go_out=. --go_opt=paths=source_relative \
      --go-grpc_out=. --go-grpc_opt=paths=source_relative \
      "proto/${svc}.proto"
  )
done

echo "✅ Proto regeneration done."
