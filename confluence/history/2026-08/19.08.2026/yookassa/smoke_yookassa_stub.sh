#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8079}"
PLAN="${PLAN:-present}"
PASS="${PASS:-secret123}"
STAMP="$(date +%s)"
EMAIL="${EMAIL:-smoke${STAMP}@example.com}"
NICK="${NICK:-Smoke${STAMP}}"
WEBHOOK_SECRET="${PAYMENT_WEBHOOK_SECRET:-${WEBHOOK_SECRET:-}}"

echo "== EH YooKassa stub smoke =="
echo "BASE_URL=$BASE_URL"
echo "EMAIL=$EMAIL"
echo "PLAN=$PLAN"

echo
echo "[1/6] Register user"
curl -sS -X POST "${BASE_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASS}\",\"nickname\":\"${NICK}\"}" | tee /tmp/eh_register.json

echo
echo "[2/6] Login and extract JWT"
LOGIN_JSON="$(curl -sS -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASS}\"}")"
TOKEN="$(echo "${LOGIN_JSON}" | jq -r '.access_token // empty')"
if [[ -z "${TOKEN}" ]]; then
  echo "ERROR: empty JWT token from login"
  echo "Login response: ${LOGIN_JSON}"
  exit 1
fi
echo "TOKEN_LEN=${#TOKEN}"

echo
echo "[3/6] Create checkout"
CHECKOUT_JSON="$(curl -sS -X POST "${BASE_URL}/api/payment/checkout" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d "{\"plan\":\"${PLAN}\"}")"
echo "${CHECKOUT_JSON}" | tee /tmp/eh_checkout.json
PAYMENT_ID="$(echo "${CHECKOUT_JSON}" | jq -r '.payment_id // empty')"
if [[ -z "${PAYMENT_ID}" ]]; then
  echo "ERROR: checkout did not return payment_id"
  exit 1
fi
echo "PAYMENT_ID=${PAYMENT_ID}"

echo
echo "[4/6] Simulate YooKassa webhook (1st delivery)"
WEBHOOK_BODY="{\"event\":\"payment.succeeded\",\"object\":{\"id\":\"${PAYMENT_ID}\",\"status\":\"succeeded\"}}"
if [[ -n "${WEBHOOK_SECRET}" ]]; then
  WEBHOOK1="$(curl -sS -X POST "${BASE_URL}/api/payment/yookassa/webhook" \
    -H "Content-Type: application/json" \
    -H "X-Webhook-Secret: ${WEBHOOK_SECRET}" \
    -d "${WEBHOOK_BODY}")"
else
  WEBHOOK1="$(curl -sS -X POST "${BASE_URL}/api/payment/yookassa/webhook" \
    -H "Content-Type: application/json" \
    -d "${WEBHOOK_BODY}")"
fi
echo "${WEBHOOK1}" | tee /tmp/eh_webhook1.json

echo
echo "[5/6] Verify subscription + merch gate"
SUB_JSON="$(curl -sS -X GET "${BASE_URL}/api/payment/subscription" \
  -H "Authorization: Bearer ${TOKEN}")"
echo "${SUB_JSON}" | tee /tmp/eh_subscription.json
CAN_JSON="$(curl -sS -X GET "${BASE_URL}/api/payment/can-purchase-merch" \
  -H "Authorization: Bearer ${TOKEN}")"
echo "${CAN_JSON}" | tee /tmp/eh_can_purchase_merch.json

echo
echo "[6/6] Simulate duplicate webhook (idempotency)"
if [[ -n "${WEBHOOK_SECRET}" ]]; then
  WEBHOOK2="$(curl -sS -X POST "${BASE_URL}/api/payment/yookassa/webhook" \
    -H "Content-Type: application/json" \
    -H "X-Webhook-Secret: ${WEBHOOK_SECRET}" \
    -d "${WEBHOOK_BODY}")"
else
  WEBHOOK2="$(curl -sS -X POST "${BASE_URL}/api/payment/yookassa/webhook" \
    -H "Content-Type: application/json" \
    -d "${WEBHOOK_BODY}")"
fi
echo "${WEBHOOK2}" | tee /tmp/eh_webhook2.json

echo
echo "== Smoke complete =="
echo "Check files:"
echo "  /tmp/eh_register.json"
echo "  /tmp/eh_checkout.json"
echo "  /tmp/eh_webhook1.json"
echo "  /tmp/eh_subscription.json"
echo "  /tmp/eh_can_purchase_merch.json"
echo "  /tmp/eh_webhook2.json"
