#!/bin/sh
# Substitute Telegram credentials into alertmanager.yml, or fall back to noop receiver.
set -eu

CFG_DIR="${ALERTMANAGER_CONFIG_DIR:-/etc/alertmanager}"
OUT="/tmp/alertmanager.yml"

TOKEN="${TELEGRAM_BOT_TOKEN:-}"
CHAT="${TELEGRAM_CHAT_ID:-}"

if [ -n "$TOKEN" ] && [ -n "$CHAT" ]; then
  # chat_id must be a bare integer in YAML
  sed \
    -e "s|__TELEGRAM_BOT_TOKEN__|${TOKEN}|g" \
    -e "s|__TELEGRAM_CHAT_ID__|${CHAT}|g" \
    "${CFG_DIR}/alertmanager.yml" > "$OUT"
  echo "alertmanager: telegram receiver enabled (chat_id=${CHAT})"
else
  cp "${CFG_DIR}/alertmanager.noop.yml" "$OUT"
  echo "alertmanager: TELEGRAM_* unset — using null receiver (noop)"
fi

exec /bin/alertmanager \
  --config.file="$OUT" \
  --storage.path=/alertmanager \
  "$@"
