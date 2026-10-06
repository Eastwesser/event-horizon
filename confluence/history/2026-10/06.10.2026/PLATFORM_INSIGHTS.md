# Platform insights map — 06.10.2026+

Не-игровой бэклог (магазин, chrome, infra, deferred product).  
Игры живут отдельно: [`GAME_INSIGHTS.md`](../../history/2026-10/06.10.2026/GAME_INSIGHTS.md) — **не смешивать**.  
Живая карта треков: [`FINAL_TRACKS.md`](../../architecture/FINAL_DETAILS.md/FINAL_TRACKS.md).  
Волны: [`TODO_FINAL_PRE_PROD_DETAILS.md`](../../architecture/FINAL_DETAILS.md/TODO_FINAL_PRE_PROD_DETAILS.md).

**Правило:** тикай `- [ ]` → `- [x]`. Берсерк CCG (~281) не трогать без явного OK.

**Порядок (платформа):**  
1. Shop examples / spinner → 2. Inventory/chrome polish → 3. Perf (/shop load) → 4. Track C k3s → 5. C4/deferred → games

---

## Progress

| Срез | Done / Total |
|------|--------------|
| P0 Already OK (voice) | 5 / 5 |
| P1 Shop examples & art | 6 / 6 |
| P2 Shop UX / load | 3 / 3 |
| P3 Chrome / nav | 2 / 2 |
| P4 Track C infra | 1 / 1 |
| P5 Deferred product | 5 / 8 |

---

## P0 — Already OK (voice confirmed) ✅

- [x] Уведомления admin (заявка автора) — кликабельны
- [x] Бургер: инвентарь / история / автор / подписка
- [x] Refund cancel в инвентаре («Нет» / диалог) — работает
- [x] Фильтры и сортировки магазина — не ломать
- [x] Берсерк-карты в выдаче первыми / не тронуты cleanup’ом

---

## P1 — Shop examples & art ✅

Source: voice 06.10 + shop cleanup leftovers.

- [x] Удалить или скрыть example **«Картина «Туманность Horizon»»** (`a1111111-…102`) — soft-delete inventory + `available=false` shop
- [x] **Значок Event Horizon** — `/images/shop/badge-horizon.jpg`
- [x] **Космический брелок** — `/images/shop/brelok-cosmic.jpg`
- [x] **Фенечка «Орбита»** — `/images/shop/fenechka-orbit.jpg`
- [x] C3 Smoke Card — hide in shop (`available=false`; inventory карточка row untouched)
- [x] Cards count guard: 281 → 281; script `scripts/platform-p1-shop-examples.sql`

**Не трогать:** inventory/shop rows type=`карточка` Берсерк (author seed).

---

## P2 — Shop UX / load ✅

- [x] Спиннер загрузки `/shop` — центр экрана (`LoadingSpinner fullscreen`; detail + infiniteshop тоже)
- [x] Профиль load path (code): cold `/shop` = `searchAllItems` (inventory pages ≤100) **после** `fetchInventory` + лишний `fetchItems` (`/shop/items` + ещё inventory); каждый апдейт inventory **перекачивал** весь каталог
- [x] Thin follow-up: catalog once on mount ∥ balance/inventory; owned patch local; drop `fetchItems` on Shop; `searchAllItems` remaining pages `Promise.all` (не Wave 5 server page/filter)

---

## P3 — Chrome / nav ✅

- [x] Glossary RU на currency chips: `LAMP_HINT` / `TICKET_HINT` — Balance, Profile, Shop
- [x] Balance header + Notification toast: emoji → SVG

**Parked:** мелкий UX «кнопок платформы» — нужен список Emma (остаётся в open).

---

## P4 — Track C — Infra ✅

- [x] k3s data plane: NATS ×3 + Postgres StatefulSets in Helm (`dataPlane.enabled`)
  - `make deploy-k3s-dataplane` / `helm-template-k3s-dataplane`

---

## P5 — Deferred product ✅ (actionable slice)

- [ ] Мелкий UX «кнопок платформы» (список Emma) — **open**
- [x] C4 monetization lock = **D defer** — [`TRACK_C4_MONETIZATION_LOCK.md`](../../architecture/FINAL_DETAILS.md/TRACK_C4_MONETIZATION_LOCK.md)
- [x] Wave 2 #5b: `first_purchase` + `first_boost` (migration + NATS; FE sync toast)
- [x] Multi-VU purchase EXPLAIN — **skip** (нет real slow report)
- [x] Catalog bottleneck server page/filter — **skip** until ~500+ cards (today ~281)
- [x] Cursor pull-in — **skip v1** (VOID stretch; disk particle pull already shipped)
- [x] `backfill-noiz-reviews.py` — script OK; `--dry-run` verified (126 reviews). Live write = ops when gateway+admin up
- [x] Authors marketplace beyond Berserk — **skip content** (platform already supports; no new authors now)

---

## Closed (context, не тикать заново)

- [x] Track A: refund · notifications · JWT refresh
- [x] Shop cleanup v2 (placeholders hidden; 281 cards safe)
- [x] Wave 1–4 / Wave 4.5
- [x] Track B games parked in GAME_INSIGHTS (hotfix → boost UX → effects → levels)

---

## Next action

```text
Platform P0–P5 actionable DONE (button list still needs Emma).

NEXT = Manual QA (games smoke / Flappy LB) · or Emma chrome list when ready.
Tracks A–D pushed (CONTINUE through ef6843e+).
```
