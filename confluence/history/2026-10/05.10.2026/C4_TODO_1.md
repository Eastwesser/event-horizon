Push. Дальше C4 — но сначала решаем monetization
Что принято — 986736f
✅ /author/dashboard — Мои карты / Продажи / Профиль.

✅ Burger «Автор» (author|admin).

✅ GET /api/authors/me, PUT .../me + portfolio.

✅ GET /api/authors/me/sales (masked buyer email; aggregates ignore refunds).

✅ Inventory soft-delete list + restore с ownership check.

✅ Soft-deleted visible с «Удалено» + «Восстановить» через ?include_deleted=1 — моё уточнение выполнено.

✅ Smoke — 6 пунктов, всё ок.

Хорошо. Ownership check на restore — правильно (нельзя восстановить чужой item).

Про refunded purchases в sales
aggregates ignore refunds

Правильно. Refund → не считается продажей для автора.

Проверь, что при refund:

Sales count — уменьшается.

Total tickets — уменьшается.

Buyer email — не виден больше (или помечен refunded).

Если так — ок. Если нет — Cursor доделает.

C4 — нужен monetization decision
Прежде чем стартовать C4, нужно ответить на вопросы:

Q1. Какую долю получает автор?
Варианты:

A) 0% — автор просто публикует для удовольствия. Payouts не нужны.

B) Фиксированный % (например 30%) от цены покупки в билетиках.

C) Фиксированная сумма ₽ за карту (не за продажу) — выплата за факт публикации.

D) Модель позже. C4 — отложить, делать C5 или Wave 4.

Q2. Как выплачивать?
Если B:

В билетиках автору на аккаунт?

Или в ₽ на внешний кошелёк (нужна интеграция с платежами)?

Когда — по факту продажи? Раз в месяц? По запросу?

Q3. Что делать с refund'ами?
Если игрок купил → вернул — автор получает 0 (уже исключено из aggregates).

Если уже выплачено — вычитать из следующей выплаты?

Q4. Kids-safe нюанс
Платформа для детей. Автор получает деньги за продажи детям — это нормально? Юридически/этически?

Или — модель «волонтёрская» — автор публикует ради удовольствия?

Мои рекомендации
Для v1 — не спешить с C4.

Причины:

Payouts — сложная фича (деньги, юридические, refund'ы, налоги).

У тебя нет пользователей пока — payouts не срочны.

Wave 4 (юнит-тесты, k6, Helm) — важнее для качества/деплоя.

Проще сделать C4 позже, когда будет monetization модель.

Мой выбор — отложить C4, перейти на Wave 4 или незакрытые follow-up'ы.

Что ещё открыто в Wave 3
C4 — Payouts (L) — отложить до monetization decision.

Wave 3 — можно считать закрытой на C1–C3 (3/4). C4 — deferred.

Ответ Cursor'у
text
C3 accepted. Push 986736f → origin/main. Report new HEAD.

Before C4 — we need a monetization decision:

  Q1. Author share of a sale?
      A) 0% (author publishes for fun, no payouts)
      B) Fixed % of ticket price (e.g. 30%)
      C) Fixed ₽ per card (published, not sold)
      D) Decide later → defer C4, move to other work

  Q2. If B — pay in tickets to author account, or ₽ to
      external wallet? Cadence: per-sale, monthly, on-request?

  Q3. Refund handling — already excluded from aggregates;
      what if already paid out? Deduct from next payout?

  Q4. Kids-safe: platform is for children. Should authors
      earn money from sales to kids? Or volunteer-only model?

My recommendation: DEFER C4 (payouts) until monetization is
locked. Move to one of:

  Option 1 — Wave 4 infra:
    - Helm charts for k3s.
    - Service Discovery (Consul).
    - k6 load + DB indexes.
    - Unit tests ≥ 70%.
    - OpenAPI docs (gateway-only — already decided).

  Option 2 — Wave 2 follow-ups still open:
    - author.application.submitted → admins (notification).
    - author.application.approved → author (notification).
    - JWT role refresh (re-login UX gap).

  Option 3 — Housekeeping:
    - Wave 2 #5b — shop/boost achievements (bonus).
    - Refund window 7d + fulfilled guard (design ready).
    - backfill-noiz-reviews.py (PUT 403 fix).

Tell me which one to start after push.
Что я держу в карте
Закрыто (в origin/main, 021f7a3):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Hero Plan A, VOID v3, smooth scroll.

Wave 3 C1 — 7110d1b.

Wave 3 C2 — 021f7a3.

Локально (пуш сейчас):

986736f — C3 Author dashboard.

Wave 3 — 3/4:

✅ C1 — Author application.

✅ C2 — Admin approval.

✅ C3 — Author dashboard (986736f).

🟧 C4 — Payouts (deferred до monetization).

Follow-up (отложено):

🟧 author.application.submitted → admins (notification).

🟧 author.application.approved → author (notification).

🟧 JWT role refresh — UX-gap.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

