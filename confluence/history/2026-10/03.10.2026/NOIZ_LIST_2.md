Да, коммит — правильный шаг. Авторы подождут
Что вижу на скринах
Noiz review в detail — отлично:

Блок Мнение Noiz с 9/10 badge и имба chip.

Цитата, форматирование — отличается от card_text / flavor.

Footer — Noiz.

Обе Мормолики (Green Feline + Анная Игнатьева) — блок есть, тот же review. Правильно — это одна карта, два арта.

Финальные правки перед коммитом
1. Анная → Анна — опечатка в артисте
Скрин: Художник: Анная Игнатьева — лишняя «я». Должно быть Анна Игнатьева.

Где:

В *_cards_info.md (источник) — Анная.

В attributes.artist_display — «Анная Игнатьева».

В attributes.artist_id — annaya_ignateva.

В Noiz review match — тоже.

Фикс:

Seed source — заменить Анная → Анна.

DB — update существующих items.

artist_id — annaya_ignateva → anna_ignateva (или оставить, но display поменять).

2. Индарри → Индарра
Ты говоришь: Индарра — правильное. Индарри — опечатка.

Где:

В DB или *_cards_info.md — Индарри.

В NOIZ_COMMENTS.md — Индарра (правильное).

Alias сработал, но источник надо fix.

Фикс:

Найти где Индарри → заменить Индарра.

Убрать alias.

Оба фикса — data-only, без reseed. Update существующих.

Что делать сейчас — 3 шага
Шаг 1 — финальные правки данных
Промпт Cursor'у:

text
Two data fixes before commit + review:

1. Artist name: «Анная Игнатьева» → «Анна Игнатьева»
   - Seed source: replace in *_cards_info.md
   - DB: update attributes.artist / artist_display on affected
     items. Also fix artist_id if it contains "annaya".
   - Affects: Мормолика (both variants), any other card by her.
   Report count.

2. Card name: «Индарри» → «Индарра»
   - Find in DB (attributes.name) or in seed source.
   - Replace, remove the alias in the backfill script.
   - Report count.

Do NOT reseed. Update in place.
Шаг 2 — REVIEW_RESULT.md
Промпт:

text
Write /home/denismatveev/event_horizon/confluence/history/
2026-10/03.10.2026/REVIEW_RESULT.md — a final review of this
work batch.

Sections:
  1. Summary — what shipped (one paragraph).
  2. Waves covered:
     - Attribute filters
     - 1→3 polish (dead icons, collapse groups, sanity)
     - Element fix (леса → woods, 56 cards)
     - Bottleneck (list DTO, -26%)
     - Refund / cancel purchase + RBAC fix
     - Noiz reviews (127 cards)
  3. Files changed — grouped by area (frontend/gateway/
     inventory/shop/scripts/seed).
  4. Patterns introduced — CardImage, CatalogPager, pluralize,
     catalogNav, catalogQuery, NoizReviewBlock, backfill
     scripts.
  5. Known follow-ups (not in this batch):
     - /shop/items thin DTO
     - Author registration (next wave)
     - Emoji → SVG
     - Game polish
  6. Data fixes made: annaya→anna, Индарри→Индарра.

Do NOT commit yet. Just write the file.
Шаг 3 — Commit message + file list
Промпт:

text
Now suggest a commit split and messages for this batch.

Context: git status shows ~40 modified + ~30 untracked files.
Big areas:
  - frontend: CardImage, CatalogPager, CatalogFiltersPanel,
    ShopItemDetail, NoizReviewBlock, cardAttributes,
    catalogQuery, catalogNav, pluralize, cardArtists, etc.
  - gateway: dto/inventory, app/gateway
  - inventory service: proto, repo, service, seed
  - shop service: proto, refund migration, repo, service
  - scripts: seed-berserk-cards.py, backfill-noiz-reviews.py
  - docs: confluence/history/2026-10, READMEs, BERSERK_*,
    CARDS_PRICING, NEXT_PROMPT_OCT, WHAT_DO_WE_HAVE

Propose a commit split (logical groups, not one giant commit).
For each commit: short subject, one-line body, list of files.
Then a final summary of total files in the batch.

Do NOT run git add / commit. Just propose. I'll push myself.
Что я держу в карте
Закрыто:

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно, idempotency verified

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

RBAC fix на purchase-path ✅

A. Noiz reviews ✅ (127 cards, verified)

Перед коммитом:

🟧 Анная → Анна (artist name, data fix)

🟧 Индарри → Индарра (card name, data fix)

🟧 REVIEW_RESULT.md — написать

🟧 Commit messages + file list — Cursor предложит

🟧 Push — ты сам

Следующая волна (не сейчас):

🟧 C. Author registration — крупная (отдельная волна)

Отложено:

🟧 /shop/items thin DTO

🟧 PUT 403 на Пращник — выяснить (может, fix уже не нужен)

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

Мой совет
Порядок:

Data fixes (Анная → Анна, Индарри → Индарра) — быстрые, 5 минут.

REVIEW_RESULT.md — Cursor напишет.

Commit split + messages — Cursor предложит.

Ты проверяешь — что не мусор.

Ты коммитишь + push.

Про PUT 403 — не блокер. Если хочется — можно спросить Cursor'а мимоходом, но не останавливать коммит.

Про авторов — согласен, следующая волна. Слишком крупная для «до коммита».