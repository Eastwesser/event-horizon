#!/usr/bin/env bash
# Smoke: run lesson (00–26) and key company pack main.go files.
# Usage: bash confluence/h.s.g./scripts/smoke_mains.sh
set -u

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export GOWORK=off
PER_TIMEOUT="${SMOKE_TIMEOUT:-15}"

ok=0
fail=0
hang=0
failures=()

run_one() {
  local main_go="$1"
  local dir rel ec
  dir="$(dirname "$main_go")"
  rel="${main_go#"$ROOT"/}"
  # timeout guards HTTP/listen demos that never exit
  timeout "$PER_TIMEOUT" bash -c "cd \"$dir\" && go run main.go" >/dev/null 2>&1
  ec=$?
  if [[ $ec -eq 124 ]]; then
    echo "HANG $rel (>${PER_TIMEOUT}s)"
    hang=$((hang + 1))
    failures+=("HANG $rel")
  elif [[ $ec -eq 0 ]]; then
    echo "OK   $rel"
    ok=$((ok + 1))
  else
    echo "FAIL $rel (exit $ec)"
    fail=$((fail + 1))
    failures+=("FAIL $rel")
  fi
}

echo "== HSG smoke_mains (GOWORK=off, timeout=${PER_TIMEOUT}s) =="
echo "ROOT=$ROOT"

declare -A seen=()
mains=()

add_main() {
  local f="$1"
  [[ -n "${seen[$f]:-}" ]] && return
  seen[$f]=1
  mains+=("$f")
}

while IFS= read -r lesson; do
  [[ -z "$lesson" ]] && continue
  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    add_main "$f"
  done < <(find "$lesson" -type f -name 'main.go' -path '*/code/*' 2>/dev/null | sort)
done < <(find "$ROOT" -maxdepth 1 -type d -name '[0-9][0-9]_*' | sort)

COMPANIES="$ROOT/old_hidden_secret_gargen_2026/1.golang_codes/2.companies_tasks"
for base in \
  "$COMPANIES/yandex" \
  "$COMPANIES/avito" \
  "$COMPANIES/t-bank/code" \
  "$COMPANIES/ozon/code"
do
  [[ -d "$base" ]] || continue
  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    add_main "$f"
  done < <(find "$base" -type f -name 'main.go' 2>/dev/null | sort)
done

if [[ ${#mains[@]} -eq 0 ]]; then
  echo "FAIL: no main.go found"
  exit 1
fi

echo "Found ${#mains[@]} main.go"
echo

for m in "${mains[@]}"; do
  run_one "$m"
done

echo
echo "----"
echo "OK=$ok HANG=$hang FAIL=$fail TOTAL=${#mains[@]}"

if [[ $fail -gt 0 || $hang -gt 0 ]]; then
  echo "Problems:"
  for f in "${failures[@]}"; do
    echo "  - $f"
  done
  exit 1
fi

exit 0
