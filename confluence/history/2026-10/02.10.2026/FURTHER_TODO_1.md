Wave: card-aware UI. 10 items. Confirm plan, then implement
in order. Do NOT add attribute filters yet — next wave.

================================================================
1. IMAGE RENDERING IN SHOP
================================================================

ShopItemCard shows emoji only. It maps image_url but does not
render it. Fix: render <img src={images[0]}> like
InventoryItemCard, with the emoji as fallback when images is
empty. Same treatment everywhere shop items are shown.

================================================================
2. CARD DETAIL — CLICKABLE
================================================================

Shop cards are currently not clickable. Make the whole card a
link to a detail page (or open a detail modal — pick whichever
fits the existing pattern):
  - Route: /shop/item/:id (or /shop/:id)
  - Detail view shows: full image, name, description, price,
    stock, all attributes (icons, class, cost, hp, move, dice,
    artist, set, foil/noir/flying/companion badges), flavor text.
  - Reuse InventoryItemDetail pattern if it exists.

================================================================
3. STOCK VISIBILITY
================================================================

Card and detail view must show stock clearly:
  - stock > 0  → "В наличии: N"
  - stock = 0  → "Нет в наличии" (muted)
  - stock null → hide the line
Same treatment as InventoryItemCard (already fixed earlier).

================================================================
4. ICONS IN ATTRIBUTES — REAL COMPONENTS
================================================================

Currently attributes.icons is dumped as String(value) →
"[object Object]". Replace with a small icon component:
  - armor:2      → shield icon + "2"
  - zoal         → arrow-wings icon
  - regen:1      → crab + "1"
  - zov / zoz / zot / zor / zoo / zom  → respective icons
  - direct / uchr / ova / ovz / ovs / stamina  → respective
If no asset exists yet, render a text chip with the type name
(still better than "[object Object]"). Also handle structured
fields: class (array), cost_tier.

================================================================
5. BADGES FOR FLAGS
================================================================

Add small badges next to the name or in the card header:
  - foil      → "✦ ФОЙЛ"   (gold)
  - noir      → "◐ НУАР"   (silver)
  - flying    → "🕊"
  - companion → paw icon or "C"
Pulled from attributes.foil / noir / flying / companion.

================================================================
6. TICKET PRICE DISPLAY
================================================================

Prices are in tickets, not rubles. Fix labels:
  - InventoryItemCard / Detail: N 🎫 instead of N ₽
  - ShopItemCard / PurchaseModal: same
Remove formatRubPrice usage for these items (or rename to
formatTickets).

================================================================
7. CREATE MODAL ×2000 → ×1000
================================================================

The create modal still suggests market_rub × 2000. Seed uses
×1000. Fix helper text / formula: 1 ₽ = 1000 tickets, foil ×2.

================================================================
8. PAGINATION
================================================================

Shop catalog has 304 items. Add pagination:
  - 100 items per page → 4 pages (304 / 100)
  - URL param: ?page=1 (default), ?page=2, ...
  - "Назад" / "Далее" buttons + "Стр. N из M"
  - Page size configurable via a constant; default 100.
  - Preserve existing filters (type=карточка, etc.) across
    pages.
Reuse the pagination pattern already used in /admin Users tab.

================================================================
9. SORTING
================================================================

Add a sort control above the grid:
  - Name A–Z / Z–A
  - Rarity (common → ultra, then reverse)
  - Price (low → high, high → low)
  - Artist A–Z
  - Set number (asc / desc)
  - Newest first (by created_at) — default
UI: a small dropdown next to the filter chips.
Default: newest first (created_at desc) so seeded order
matches.

================================================================
10. ARTIST PAGE + LIST
================================================================

Add a page /authors/:artist_id (or /shop?artist=...):
  - Header: artist display name.
  - Grid of that artist's cards (same ShopItemCard).
  - Count: "N карт".
Also a page /authors (list all artists):
  - Each row: artist name, card count.
  - Click → artist page.
Long-term: authors will self-register and upload their own
cards. For now, seeding is done by me (admin). No author
registration UI yet — just the read side.

================================================================
PROCESS
================================================================

1. Show a short plan (files touched, order).
2. Implement 1–3 first (images, click, stock). Report.
3. I verify.
4. Then 4–5 (icons, badges).
5. Then 6–7 (price display, create modal).
6. Then 8 (pagination).
7. Then 9 (sorting).
8. Then 10 (artist pages).

DO NOT TOUCH:
  - Seed script.
  - Non-card inventory types (брелок / картина / фенечка).
  - Game mechanics.
  - Nav / footer.
  - Admin panel.
  - Attribute filters (element / rarity / stats) — those come
    AFTER this wave.