#!/usr/bin/env bash
# Rebuild Phase-1 changed services (same pattern as services/auth/README.md).
# Run from repo root as your normal user (not the Cursor sandbox):
#   bash scripts/rebuild-phase1.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SERVICES=(
  auth
  gateway
  billing
  inventory
  profile
  shop
  leaderboard
  game
)

echo "📦 go mod tidy for auth + profile (new redis deps)..."
(cd services/auth && go mod tidy)
(cd services/profile && go mod tidy)

echo
echo "🔨 Building linux/amd64 binaries..."
for svc in "${SERVICES[@]}"; do
  echo "----- $svc -----"
  (
    cd "services/$svc"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${svc}-service" ./cmd/main.go
  )
  ls -lh "services/$svc/${svc}-service"
done

echo
echo "🐳 Docker images (local tags only — no push)..."
for svc in "${SERVICES[@]}"; do
  echo "----- docker $svc -----"
  docker build -f "Dockerfile.${svc}.bin" -t "eastwesser/${svc}:latest" .
done

echo
echo "✅ Done. Restart only the rebuilt services, e.g.:"
echo "   docker-compose -f deployments/docker-compose.cluster.yml up -d auth gateway billing inventory profile shop leaderboard game"
echo "   make migrate-all   # applies new outbox/role/status migrations"
