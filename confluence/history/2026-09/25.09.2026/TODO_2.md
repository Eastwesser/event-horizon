Шаг 3 — Промпт Cursor'у на следующую волну
После merge выбирай, что дальше. Мой приоритет:

🔴 Приоритет 1 — Backend twin'ы (5 штук)
Тот же proto3-паттерн, но пока защищён только на фронте. Бэкенд всё ещё отдаёт null:

authors → if nil { Authors = []*authorsPb.Author{} }

history/events

analytics/dau

analytics/retention

leaderboard/entries

Промпт:

text
Post-merge cleanup — apply the same nil→[] fix to the remaining
gateway handlers that return slices.

Endpoints (from your own audit):
  GET /api/authors              → {"authors": resp.Authors}
  GET /api/history/events       → {"events": resp.Events}
  GET /api/analytics/dau        → {"days": resp.Days}
  GET /api/analytics/retention  → {"points": resp.Points}
  GET /api/leaderboard/...      → {"entries": resp.Entries}

For each:
  - Coerce nil slice to empty typed slice before c.JSON.
  - Same pattern as inventory/items (Ln 1380–1390) and
    shop/items (Ln 799–805).
  - Do NOT change response shape (keep gin.H wrapper, keep the
    existing key names).
  - Do NOT touch the frontend (already guarded).

Verify with tsc/build after.

Report: file:line for each fix, total files changed.

DO NOT TOUCH:
  - Game mechanics, nav, footer, role gating.
  - Anything from the merged batch.
🟧 Приоритет 2 — 2d Analytics в админке
После twin'ов. Промпт отдельно.

🟧 Приоритет 3 — 2c v2, proto3 medium, мелочи
Отложено.

Что я держу в карте
Закрыто (в dev/frontend, готово к merge):

✅ Дизайн-система, все страницы, игры, admin panel

✅ Proto3 фиксы + null-safety

✅ Image URL, role gating, nav/burger

✅ Timeout исчез, gateway cluster свежий

Следующее (после merge):

🔴 Backend twin'ы — 5 эндпоинтов

🟧 2d — Analytics в админке

🟧 2c v2 — top-5, total stock, author names

🟧 Proto3 medium / gin.H refactor

🟧 File upload, navbar refactor, DNS fix

Ждёт:

🟧 Merge dev/frontend → main