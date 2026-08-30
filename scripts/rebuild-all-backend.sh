#!/usr/bin/env bash
# Regenerate protos, build all linux/amd64 binaries, build all Docker images (no push).
# Run from repo root: bash scripts/rebuild-all-backend.sh
# Does not require go-task (`task`) — uses bash + make only.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== 1/3 proto =="
bash scripts/rebuild-proto.sh

echo "== 2/3 go binaries (linux/amd64) =="
SERVICES=(
  auth billing game leaderboard profile shop gateway balancer nats-hub
  inventory payment authors history analytics fulfillment notification
)
for svc in "${SERVICES[@]}"; do
  echo "===== BUILD $svc ====="
  (
    cd "services/$svc"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${svc}-service" ./cmd/main.go
  )
done

echo "== 3/3 docker images =="
make docker-build-all

echo "✅ All backend images built locally."
echo "Push to Docker Hub:"
echo "  make docker-push-all"
echo "  # or: bash scripts/docker-push-images.sh --all"
