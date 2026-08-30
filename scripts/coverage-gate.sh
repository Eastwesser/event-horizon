#!/usr/bin/env bash
# Fail if any gRPC service package coverage is below MIN_COVERAGE (default 70).
# Usage: MIN_COVERAGE=70 bash scripts/coverage-gate.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

MIN="${MIN_COVERAGE:-70}"
SERVICES=(auth billing game inventory leaderboard profile shop payment authors history analytics)

failed=0
for svc in "${SERVICES[@]}"; do
  if [[ ! -d "services/${svc}" ]]; then
    continue
  fi
  echo "===== coverage gate: $svc (min ${MIN}%) ====="
  pct="$(
    cd "services/${svc}"
    GOWORK=off go test ./... -coverprofile=coverage.out -covermode=atomic >/dev/null 2>&1
    go tool cover -func=coverage.out | awk '/total:/ {gsub(/%/,"",$3); print $3}'
  )"
  echo "  total: ${pct}%"
  if awk -v p="$pct" -v m="$MIN" 'BEGIN { exit !(p+0 < m+0) }'; then
    echo "  FAIL: below ${MIN}%"
    failed=1
  fi
done

if [[ $failed -ne 0 ]]; then
  echo "Coverage gate failed — add tests or lower MIN_COVERAGE for local runs."
  exit 1
fi
echo "Coverage gate OK (>= ${MIN}% all services)."
