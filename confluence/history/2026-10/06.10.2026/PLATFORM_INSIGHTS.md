# Platform insights map — 06.10.2026+

Не-игровой бэклог (магазин, chrome, infra, deferred product).  
Игры живут отдельно: [`GAME_INSIGHTS.md`](../../history/2026-10/06.10.2026/GAME_INSIGHTS.md) — **не смешивать**.  
Живая карта треков: [`FINAL_TRACKS.md`](../../architecture/FINAL_DETAILS.md/FINAL_TRACKS.md).  
Волны: [`TODO_FINAL_PRE_PROD_DETAILS.md`](../../architecture/FINAL_DETAILS.md/TODO_FINAL_PRE_PROD_DETAILS.md).

**Правило:** тикай `- [ ]` → `- [x]`. Берсерк CCG (~281) не трогать без явного OK.

**Порядок (платформа):**  
1. Shop examples / spinner → 2. Inventory/chrome polish → 3. Perf (/shop load) → 4. Track C k3s (если нужен) → 5. C4/deferred

---

## Progress

| Срез | Done / Total |
|------|--------------|
| P0 Already OK (voice) | 5 / 5 |
| P1 Shop examples & art | 6 / 6 |
| P2 Shop UX / load | 0 / 3 |
| P3 Chrome / nav | 0 / 2 |
| P4 Track C infra | 0 / 1 |
| P5 Deferred product | 0 / 7 |

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

## P2 — Shop UX / load 🟧 NEXT

- [ ] Спиннер загрузки `/shop` — центр экрана (горизонталь **и** вертикаль), не у хедера
- [ ] Профилировать долгую загрузку магазина (gateway / shop list / FE waterfall)
- [ ] Если p95 плохой после профиля — тонкий follow-up (не полный Wave 5 bottleneck)

---

## P3 — Chrome / nav (final polishing, non-game)

- [ ] Мелкий UX «кнопок платформы» вне игр (после P1–P2) — список уточнит Emma
- [ ] Glossary RU в UI chrome: лампы / билеты (иконки уже в shop; игры — в GAME_INSIGHTS)

---

## P4 — Track C — Infra (later)

- [ ] k3s data plane: NATS + Postgres StatefulSets в Helm  
  - Нужно только для prod-grade k3s demo; compose достаточно сейчас

---

## P5 — Deferred product (Emma OK only)

- [ ] C4 payouts / monetization model lock
- [ ] Wave 2 #5b shop/boost achievements
- [ ] Multi-VU purchase EXPLAIN (только при реальном slow report)
- [ ] Catalog bottleneck server page/filter (~500+ cards)
- [ ] Cursor pull-in
- [ ] `backfill-noiz-reviews.py` one-shot
- [ ] Authors marketplace beyond Berserk (другие авторы — уже заложено; контент не сейчас)

---

## Closed (context, не тикать заново)

- [x] Track A: refund · notifications · JWT refresh
- [x] Shop cleanup v2 (placeholders hidden; 281 cards safe)
- [x] Wave 1–4 / Wave 4.5
- [x] Track B games parked in GAME_INSIGHTS (hotfix → boost UX → effects → levels)

---

## Next action

```text
NEXT = P2 Shop spinner + load profile.

P1 done: example arts + painting/C3 smoke hidden; Berserk 281 untouched.
```
