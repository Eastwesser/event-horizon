# Review result — 03.10.2026 batch

## 1. Summary

This batch turned the Berserk CCG catalog into a shoppable, filterable product surface with correct money/stock ownership and richer card detail. We shipped attribute filters and catalog UX polish, fixed element data (`леса` → `woods`), slimmed the inventory list DTO (~−26%), added end-to-end purchase cancel/refund (at purchase price, idempotent, with inventory stock sync + RBAC fix), and attached prewritten Noiz reviews to 127 set-8 card rows with a dedicated detail UI block.

## 2. Waves covered

| Wave | Result |
|------|--------|
| Attribute filters | Client-side filters on catalog (element, rarity, cost tier, flags, query) via URL state |
| 1→3 polish | Dead icon chips removed; attribute groups collapsed; detail sanity |
| Element fix | Seed + 56 cards: `леса` → `woods` so Лес filter works |
| Bottleneck | `InventoryItemsCatalog` omits bulky attrs (`card_text`, `flavor_text`, later `noiz_review`) |
| Refund / cancel | `POST /api/shop/purchase/:id/cancel` — refund `purchases.price`, `refunded_at` + `REFUNDED`, inventory row removed, shop + catalog stock +1, idempotent `already_refunded` |
| RBAC (bonus) | Shop purchase/cancel passes `withUserRole`; `user` allowed on `ReserveItem` / `ReleaseItem` so stock actually moves |
| Noiz reviews | `attributes.noiz_review` `{text, rating, verdict, author}` on 127 rows; detail-only «Мнение Noiz» block |

## 3. Files changed (by area)

### Frontend
- Shop: `Shop.tsx`, `ShopItemCard.tsx`, `ShopItemDetail.tsx`, `PurchaseModal.tsx`, `CancelPurchaseModal.tsx`, `CatalogFiltersPanel.tsx`, `NoizReviewBlock.tsx`, `cardAttributes.tsx`, `ShopWithInfiniteScroll.tsx`
- Inventory: `InventoryPage.tsx`, `InventoryItemCard.tsx`, `InventoryItemDetail.tsx`, `InventoryCreateModal.tsx`
- Authors: `AuthorsPage.tsx`, `CardArtistPage.tsx`
- UI/lib: `CardImage.tsx`, `CatalogPager.tsx`, `cardArtists.ts`, `catalogNav.ts`, `catalogQuery.ts`, `pluralize.ts`, `shopItemMap.ts`
- Store/API/App: `shopStore.ts`, `api.ts`, `inventoryApi.ts`, `App.tsx`, `Profile.tsx`, `vite.config.ts`

### Gateway
- `internal/app/gateway.go` — purchase cancel route, reserve/release with role metadata, catalog list DTO
- `internal/dto/inventory.go` (+ test) — list omit including `noiz_review`
- `docs/openapi.yaml` — cancel endpoint

### Inventory service
- Proto: `ReleaseItem` RPC + generated pb / validate
- Repo/service/handler: `ReleaseItem` (postgres, mongo, cached)
- `internal/app/di.go` — RBAC: user/author/admin on Reserve/Release
- Seed tree: `services/inventory/internal/seed/berserk_cards/...` + `seed/README.md`

### Shop service
- Proto: `CancelPurchase`, `purchase_price` / `purchase_id` on Item
- Migration: `20261003050000_add_purchase_refunded_at.sql`
- Repo/service/handler: refund TX, inventory join for purchase price
- Model: `ErrPurchaseNotFound`

### Scripts
- `scripts/seed-berserk-cards.py`
- `scripts/backfill-noiz-reviews.py`

### Docs / history
- `confluence/history/2026-09/30.09.2026/*` (BERSERK_*, CARDS_PRICING, WHAT_DO_WE_HAVE, …)
- `confluence/history/2026-10/03.10.2026/*` (TODO lists, NOIZ_*, CRIT_*, REVIEW_RESULT)
- `Makefile` (minor)

## 4. Patterns introduced

| Pattern | Role |
|---------|------|
| `CardImage` | Shared image container (contain/cover, fallback) |
| `CatalogPager` | Page controls for catalog |
| `catalogQuery` | URL-driven filters, sort, facets, pagination |
| `catalogNav` | Prev/next neighbors on detail |
| `pluralize` | RU plural helpers (tickets, etc.) |
| `cardArtists` / `CardArtistPage` | Artist index + per-artist catalog |
| `cardAttributes` | Structured combat/meta rows + flag badges |
| `NoizReviewBlock` | Data-driven quote block iff `noiz_review` |
| `InventoryItemsCatalog` | Slim list DTO; full attrs on detail GET |
| Backfill scripts | Idempotent data loads (Noiz; seed `--update`) |

## 5. Known follow-ups (not in this batch)

- `/shop/items` thin DTO (still deferred)
- **C. Author registration** — next large wave
- Emoji → SVG for icon chips
- Game polish
- Optional: investigate one-off PUT 403 on Пращник during Noiz backfill (admin path; SQL backfill succeeded; not a blocker)

## 6. Data fixes made (pre-commit)

| Fix | Count | Notes |
|-----|-------|--------|
| `Анная Игнатьева` → `Анна Игнатьева` | **1** inventory row (`artist` / `artist_display` / `artist_id` `annaya_ignateva` → `anna_ignateva`) | Seed source `darkness_8th` cards_info updated |
| `Индарри` → `Индарра` | **1** inventory + **1** shop item | Seed `swamps_8th` list name fixed; image file keeps `indarri` slug; stock alias + Noiz alias removed |

No full reseed — in-place updates only.
