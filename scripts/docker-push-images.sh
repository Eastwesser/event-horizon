#!/usr/bin/env bash
# Push Docker images one ref per call (docker push does not accept multiple tags).
# Usage:
#   bash scripts/docker-push-images.sh game analytics gateway
#   bash scripts/docker-push-images.sh --all
set -euo pipefail

ALL_IMAGES=(
  auth billing game leaderboard profile shop gateway balancer nats-hub
  inventory payment authors history analytics fulfillment notification
)

if [[ $# -eq 0 ]]; then
  echo "Usage: $0 [--all | image ...]"
  echo "  image names without tag, e.g. game gateway (pushes eastwesser/<name>:latest)"
  exit 1
fi

if [[ "${1:-}" == "--all" ]]; then
  NAMES=("${ALL_IMAGES[@]}")
else
  NAMES=("$@")
fi

for name in "${NAMES[@]}"; do
  img="eastwesser/${name}:latest"
  echo "===== PUSH $img ====="
  docker push "$img"
done

echo "✅ Push complete."
