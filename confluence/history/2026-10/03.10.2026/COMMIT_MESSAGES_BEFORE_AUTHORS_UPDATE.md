Промпт Cursor'у
text
Two doc updates after the v1.0.9 push.

================================================================
1. README — bump to v1.0.9, add section
================================================================

Update README.md:
  - Version line → v1.0.9 (currently v1.0.8 in the file).
  - Add a new section at the top of the "Что нового" list:

### Что нового в v1.0.9

- **Berserk CCG catalog:** 280 карт засеяно (5 сетов, 4–8),
  attribute filters (element / rarity / class / flags / stats /
  icons / artist), URL-shareable state, sort (name / rarity /
  price / artist / set / element), pagination
- **Purchase + refund:** cancel purchase (`POST /api/shop/
  purchase/:id/cancel`), refund по цене покупки, idempotent,
  restore stock, RBAC fix (user allowed on Reserve/Release)
- **Noiz reviews:** «Мнение Noiz» блок на картах 8-го сета
  (127 rows), attributes.noiz_review {text, rating, verdict,
  author}
- **UI polish:** CardImage (5/7, cover/contain), CatalogPager,
  artist pages + per-artist grid, prev/next navigation,
  attribute grouping in detail, cancel modal
- **Data:** element normalization (леса → woods, 56 cards),
  slim list DTO InventoryItemsCatalog (−26% payload)
- **Full review:** [REVIEW_RESULT.md](confluence/history/
  2026-10/03.10.2026/REVIEW_RESULT.md)

Remove the placeholder «Текущая: v1.0.8» line or update it.

================================================================
2. CHANGELOG.md — add v1.0.9 entry
================================================================

If CHANGELOG.md exists at repo root, prepend a v1.0.9 entry.
If it doesn't exist, create one at the root with this format:

## [v1.0.9] — 2026-10-03

### Added
- Berserk CCG card seed: 280 cards across sets 4–8
- Attribute filters panel (element, rarity, class, flags,
  stats with operators, icons multi-select, artist autocomplete)
- Sort controls (name, rarity, price, artist, set, element)
- Pagination with URL state (100 per page)
- Artist pages (/authors + /authors/:id)
- Prev/next navigation on card detail
- Noiz reviews UI block on 8th-set cards (127 rows)
- Cancel purchase (refund at purchase price, idempotent)
- Slim inventory list DTO (InventoryItemsCatalog)

### Fixed
- Element normalization: леса → woods (56 cards)
- Purchase-path RBAC: user allowed on ReserveItem/ReleaseItem
- Data typos: Анная → Анна, Индарри → Индарра
- Detail hero white corners
- Shop card layout (drop description)

### Docs
- REVIEW_RESULT.md (full batch review)
- Berserk history: rules, icons, how-to-read, pricing
- NOIZ source comments

### Files
- See REVIEW_RESULT.md for full file list.

Keep the previous CHANGELOG entry (whatever it is) below.

================================================================
DO NOT TOUCH: code. Only README.md and CHANGELOG.md.
================================================================

Report: files changed, commit message suggestion.
Что важно
1. Плейсхолдер Текущая: v1.0.8 (30.08.2026) — заменить на v1.0.9 (03.10.2026) или убрать.

2. Если CHANGELOG уже есть — не перезаписать, prepend новую секцию сверху.

3. Не дублировать — в README коротко, в REVIEW_RESULT подробно.

4. После правки — отдельный коммит:

bash
git add README.md CHANGELOG.md
git commit -m "docs: bump version to v1.0.9, add changelog"
git push origin main
Это будет commit 10. Отдельный от batch — правильно.

Что я держу в карте
Закрыто (в origin/main, v1.0.9):

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно

Stock fix + idempotency + purchase

4 Icons + grouped attrs

5 Badges, 6 Tickets, 7 Create modal ×1000

Пращник fixed

8 Pagination, 9 Sorting, 10 Artist pages

A–H visual + UX polish

1→4 final polish

Attribute filters ✅

1→3 polish ✅

Element fix ✅

Bottleneck — smaller list DTO ✅

Refund / cancel purchase ✅

RBAC fix ✅

Noiz reviews ✅

Data fixes ✅

REVIEW_RESULT.md ✅

__pycache__ в .gitignore ✅

Push ✅ (388 files, +15 895, −594)

Сейчас:

🟧 README v1.0.9 — Cursor обновит

🟧 CHANGELOG v1.0.9 — Cursor добавит

🟧 Отдельный commit + push

Следующая волна:

🟧 C. Author registration

Отложено:

🟧 /shop/items thin DTO

🟧 PUT 403 на Пращник

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр