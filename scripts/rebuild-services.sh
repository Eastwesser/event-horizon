#!/usr/bin/env bash
# Build linux/amd64 binaries + docker images for selected services.
# Usage: bash scripts/rebuild-services.sh game analytics [gateway ...]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ $# -eq 0 ]]; then
  echo "Usage: $0 <service> [service ...]"
  echo "Example: $0 game analytics gateway"
  exit 1
fi

for svc in "$@"; do
  echo "===== build $svc binary ====="
  (
    cd "services/${svc}"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${svc}-service" ./cmd/main.go
  )
  echo "===== docker $svc ====="
  docker build -f "Dockerfile.${svc}.bin" -t "eastwesser/${svc}:latest" .
done

echo
echo "✅ Done. Push with: bash scripts/docker-push-images.sh $*"
echo "   Recreate: docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d $*"
