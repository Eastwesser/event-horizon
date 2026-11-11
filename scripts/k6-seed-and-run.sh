#!/usr/bin/env bash
# scripts/k6-seed-and-run.sh — засев k6-юзеров, баланс, запуск 10k-probe.
#
# Usage:
#   bash scripts/k6-seed-and-run.sh seed          # зарегать N юзеров
#   bash scripts/k6-seed-and-run.sh balance       # начислить 100к
#   bash scripts/k6-seed-and-run.sh json          # собрать k6_users.json
#   bash scripts/k6-seed-and-run.sh run 50        # P=50 с пулом
#   bash scripts/k6-seed-and-run.sh all 50        # seed + balance + json + run
#   bash scripts/k6-seed-and-run.sh clean         # удалить k6_u_* юзеров
#
# Env:
#   EH_SEED_COUNT      — сколько юзеров (default 10)
#   EH_SEED_PASSWORD   — пароль (default k6_secret_2026)
#   EH_BALANCE         — баланс (default 100000)

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

SEED_COUNT="${EH_SEED_COUNT:-10}"
SEED_PASSWORD="${EH_SEED_PASSWORD:-k6_secret_2026}"
BALANCE="${EH_BALANCE:-100000}"
USERS_JSON="$ROOT/load_test_results/k6_users.json"
AUTH_PG="event-horizon-postgres"
BILLING_PG="event-horizon-postgres-billing"
AUTH_DB="eventhorizon"
BILLING_DB="eventhorizon_billing"

log() { echo "==> $*"; }

do_seed() {
  log "seed: $SEED_COUNT users (password=$SEED_PASSWORD)"
  env -u K6_VUS -u K6_DURATION -u K6_RAMP -u K6_SCENARIOS \
    EH_SEED_COUNT="$SEED_COUNT" EH_SEED_PASSWORD="$SEED_PASSWORD" \
    EH_SEED_PAUSE_MS="${EH_SEED_PAUSE_MS:-250}" \
    k6 run deployments/k6/highload/seed-users.js
}

do_balance() {
  log "balance: $BALANCE tickets+lamps for k6_u_* (auth → billing)"
  # 1) user_id из auth, без \r
  docker exec "$AUTH_PG" psql -U eventhorizon -d "$AUTH_DB" \
    -c "SELECT id FROM users WHERE email LIKE 'k6_u_%';" -t -A \
    | tr -d '\r' > /tmp/k6_ids.txt

  local n
  n=$(wc -l < /tmp/k6_ids.txt | tr -d ' ')
  log "found $n user_ids"
  if [ "$n" -eq 0 ]; then
    echo "ERROR: no k6_u_* users. Run 'seed' first." >&2
    exit 1
  fi

  # 2) вставляем в billing одним SQL-файлом, без TTY-цикла
  {
    echo "BEGIN;"
    while read -r uid; do
      [ -z "$uid" ] && continue
      echo "INSERT INTO user_currencies (user_id, currency_type, balance)
            VALUES ('$uid', 'tickets', $BALANCE), ('$uid', 'lamps', $BALANCE)
            ON CONFLICT (user_id, currency_type) DO UPDATE SET balance = $BALANCE;"
    done < /tmp/k6_ids.txt
    echo "COMMIT;"
  } > /tmp/k6_balance.sql

  docker exec -i "$BILLING_PG" psql -U eventhorizon -d "$BILLING_DB" \
    < /tmp/k6_balance.sql
  log "balance done"
}

do_json() {
  log "json: $USERS_JSON"
  mkdir -p "$(dirname "$USERS_JSON")"
  docker exec "$AUTH_PG" psql -U eventhorizon -d "$AUTH_DB" \
    -c "SELECT email FROM users WHERE email LIKE 'k6_u_%' ORDER BY created_at DESC LIMIT $SEED_COUNT;" \
    -t -A \
    | tr -d '\r' \
    | grep -v '^$' \
    | jq -Rn --arg pwd "$SEED_PASSWORD" \
        '[inputs | {email: ., password: $pwd}]' \
    > "$USERS_JSON"
  log "json: $(jq length "$USERS_JSON") users"
}

do_run() {
  local profile="${1:-50}"
  if [ ! -f "$USERS_JSON" ]; then
    echo "ERROR: $USERS_JSON not found. Run 'json' first." >&2
    exit 1
  fi
  log "run: EH_PROFILE=$profile with pool $(jq length "$USERS_JSON")"
  EH_K6_USERS_FILE="$USERS_JSON" \
  EH_PROFILE="$profile" \
  EH_DURATION="${EH_DURATION:-60s}" \
  EH_RAMP="${EH_RAMP:-30s}" \
    env -u K6_VUS -u K6_DURATION -u K6_RAMP -u K6_SCENARIOS \
    k6 run deployments/k6/highload/10k-probe.js
}

do_clean() {
  log "clean: delete k6_u_* users + balance"
  docker exec "$AUTH_PG" psql -U eventhorizon -d "$AUTH_DB" -c "
    DELETE FROM users WHERE email LIKE 'k6_u_%';
  " || true

  # billing: удалить балансы по сохранённым id
  if [ -f /tmp/k6_ids.txt ]; then
    {
      echo "BEGIN;"
      while read -r uid; do
        [ -z "$uid" ] && continue
        echo "DELETE FROM user_currencies WHERE user_id = '$uid';"
      done < /tmp/k6_ids.txt
      echo "COMMIT;"
    } > /tmp/k6_clean.sql
    docker exec -i "$BILLING_PG" psql -U eventhorizon -d "$BILLING_DB" \
      < /tmp/k6_clean.sql || true
  fi
  log "clean done"
}

cmd="${1:-all}"
case "$cmd" in
  seed)    do_seed ;;
  balance) do_balance ;;
  json)    do_json ;;
  run)     do_run "${2:-50}" ;;
  all)     do_seed; do_balance; do_json; do_run "${2:-50}" ;;
  clean)   do_clean ;;
  *)
    echo "Usage: $0 {seed|balance|json|run [PROFILE]|all [PROFILE]|clean}" >&2
    exit 1
    ;;
esac