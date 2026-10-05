#!/usr/bin/env bash
# Fail if any gated service's internal/service coverage is below MIN_COVERAGE (default 70).
# Wave 4 gates the service layer (business logic), not handlers/repos/workers.
# Gateway is intentionally excluded (thin HTTP↔gRPC adapters).
#
# Usage: MIN_COVERAGE=70 bash scripts/coverage-gate.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

MIN="${MIN_COVERAGE:-70}"
SERVICES=(auth billing game inventory leaderboard profile shop payment authors history analytics)

failed=0
for svc in "${SERVICES[@]}"; do
  svc_dir="services/${svc}/internal/service"
  if [[ ! -d "$svc_dir" ]]; then
    echo "===== coverage gate: $svc — skip (no internal/service) ====="
    continue
  fi
  echo "===== coverage gate: $svc/internal/service (min ${MIN}%) ====="
  pct="$(
    cd "services/${svc}"
    GOWORK=off go test ./internal/service/ -coverprofile=coverage.out -covermode=atomic -count=1 >/dev/null 2>&1
    go tool cover -func=coverage.out | awk '/total:/ {gsub(/%/,"",$3); print $3}'
  )"
  if [[ -z "${pct}" ]]; then
    echo "  FAIL: could not compute coverage"
    failed=1
    continue
  fi
  echo "  service total: ${pct}%"
  if awk -v p="$pct" -v m="$MIN" 'BEGIN { exit !(p+0 < m+0) }'; then
    echo "  FAIL: below ${MIN}%"
    failed=1
  fi
done

if [[ $failed -ne 0 ]]; then
  echo "Coverage gate failed — add service-layer tests or lower MIN_COVERAGE for local runs."
  exit 1
fi
echo "Coverage gate OK (>= ${MIN}% on internal/service for all gated services)."
