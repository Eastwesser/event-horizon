#!/usr/bin/env bash
# Apply scripts/cleanup-shop-content-v2.sql with backup + Berserk card guard.
# Usage (after Emma OK):
#   bash scripts/apply-shop-cleanup-v2.sh
# Dry report only:
#   bash scripts/apply-shop-cleanup-v2.sh --dry-run
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SQL="$ROOT/scripts/cleanup-shop-content-v2.sql"
STAMP="$(date +%Y%m%d_%H%M%S)"
BACKUP_DIR="${BACKUP_DIR:-$HOME/backups/shop-cleanup-v2-$STAMP}"
DRY=0
[[ "${1:-}" == "--dry-run" ]] && DRY=1

INV_CTR=event-horizon-postgres-inventory
SHOP_CTR=event-horizon-postgres-shop
PGUSER=eventhorizon
INV_DB=eventhorizon_inventory
SHOP_DB=eventhorizon_shop

cards_before() {
  docker exec "$INV_CTR" psql -U "$PGUSER" -d "$INV_DB" -tAc \
    "SELECT COUNT(*) FROM inventory_items WHERE type='карточка' AND deleted_at IS NULL;"
}

echo "=== Shop cleanup v2 ==="
echo "SQL: $SQL"
CARDS_BEFORE="$(cards_before)"
echo "Berserk cards (inventory type=карточка, active) BEFORE: $CARDS_BEFORE"
if [[ "$CARDS_BEFORE" -lt 270 || "$CARDS_BEFORE" -gt 290 ]]; then
  echo "ABORT: unexpected card count $CARDS_BEFORE (want ~280). Refusing to continue."
  exit 1
fi

if [[ "$DRY" -eq 1 ]]; then
  echo "DRY-RUN: no backup, no writes. Review $SQL"
  exit 0
fi

mkdir -p "$BACKUP_DIR"
echo "Backup → $BACKUP_DIR"
docker exec "$INV_CTR" pg_dump -U "$PGUSER" "$INV_DB" >"$BACKUP_DIR/inventory.sql"
docker exec "$SHOP_CTR" pg_dump -U "$PGUSER" "$SHOP_DB" >"$BACKUP_DIR/shop.sql"

# Split SQL by section markers into temp files
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
awk '
  /^-- SECTION A/ { out="'"$TMP"'/a.sql"; next }
  /^-- SECTION B/ { out="'"$TMP"'/b.sql"; next }
  out != "" { print > out }
' "$SQL"

echo "Applying SECTION A → inventory…"
docker exec -i "$INV_CTR" psql -U "$PGUSER" -d "$INV_DB" -v ON_ERROR_STOP=1 <"$TMP/a.sql"

echo "Applying SECTION B → shop…"
docker exec -i "$SHOP_CTR" psql -U "$PGUSER" -d "$SHOP_DB" -v ON_ERROR_STOP=1 <"$TMP/b.sql"

CARDS_AFTER="$(cards_before)"
echo "Berserk cards AFTER: $CARDS_AFTER"
if [[ "$CARDS_AFTER" != "$CARDS_BEFORE" ]]; then
  echo "ABORT GUARD FAILED: card count changed ($CARDS_BEFORE → $CARDS_AFTER)."
  echo "Restore from: $BACKUP_DIR"
  echo "  docker exec -i $INV_CTR psql -U $PGUSER -d $INV_DB < $BACKUP_DIR/inventory.sql"
  echo "  docker exec -i $SHOP_CTR psql -U $PGUSER -d $SHOP_DB < $BACKUP_DIR/shop.sql"
  exit 2
fi

echo "=== Verify counts ==="
docker exec "$SHOP_CTR" psql -U "$PGUSER" -d "$SHOP_DB" -c "
SELECT category, COUNT(*) FILTER (WHERE available) AS available, COUNT(*) AS total
FROM items GROUP BY category ORDER BY category;
SELECT name, available FROM items
WHERE id IN (
  'a1111111-1111-4111-8111-111111111101',
  'a1111111-1111-4111-8111-111111111102',
  'a1111111-1111-4111-8111-111111111103',
  'a1111111-1111-4111-8111-111111111104',
  '82be50db-670b-48c6-beb9-7e00d584f6de',
  '6a1de8dd-9457-4aa4-99a7-78267aee731d'
) ORDER BY name;
"
echo "✅ Done. Cards untouched ($CARDS_AFTER). Backup: $BACKUP_DIR"
