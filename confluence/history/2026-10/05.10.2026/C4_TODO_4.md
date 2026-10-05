Что сделано — 7 коммитов локально
Coverage delta — отлично:

authors 0% → 93.2% (service)

shop ~2% → 83.9% (service)

Shipped:

COVERAGE.md

browse.js + purchase.js (k6)

loadtest.js → legacy

Флаги (важные):

merch_gate tests ожидали неправильные error strings → fixed to ErrSubscriptionRequired. «Test bug, not prod» — но это тревожно: если тест ожидал другое сообщение, значит когда-то поведение изменялось без обновления тестов. Проверить, что prod-поведение действительно ErrSubscriptionRequired.

shop gRPC dial + inventory NATS sync — moved service → app (Clean Architecture). Поведение unchanged — но это рефакторинг prod-кода в тестовом PR. Должно быть отдельным коммитом с объяснением, не в тестах. Проверь commit hash.

k6:

Готовы, но stack down — smoke blocked. Разблокируешь сам после make deploy.

Моя тревога — 2 вещи
1. merch_gate test fix — где prod?
Возможно:

Тесты были устаревшими → prod давно ErrSubscriptionRequired → fix тестов — правильно.

Или — prod меняли без обновления тестов → тогда тест был прав, а prod сломан → fix тестов — маскирует баг.

Проверь сам:

Открой prod путь merch_gate в коде.

Какой error возвращается для subscription-required?

Соответствует ли он какому-то известному error коду в OpenAPI/доках?

2. Clean Architecture refactor — вне тестового PR
shop gRPC dial + inventory NATS sync — moved. Это не тесты, это рефакторинг архитектуры. Должно быть:

Отдельный commit.

Commit message: что и зачем.

Отдельно от test-only PR.

Если Cursor запихнул это в test-commit — rebase/отделить.

Мой ответ Cursor'у
text
Wave 4 batch reviewed. Coverage deltas are excellent.

Before pushing — two flags:

1. merch_gate test fix — CONFIRM prod behavior first.
   The tests were expecting wrong strings — were the tests
   outdated (prod correct) OR was prod changed without
   updating tests (tests correct, prod broken)?

   Show:
     - the current merch_gate prod path
     - the error returned for subscription-required
     - how it maps to OpenAPI / docs (if documented)
   If prod is correct → tests fix is fine, note it.
   If prod is broken → DO NOT paper over it. Fix prod
   separately.

2. Clean Architecture refactor (shop gRPC dial + inventory
   NATS sync service → app) is NOT a test-only change.
   Split it into its own commit with a clear message.
   Rebase the batch so test commits stay test-only.

After both are handled:

  - Push the batch (7 commits, or 8 with the refactor split).
  - Then continue with inventory → auth tests.
  - k6 smoke: I'll run it myself after `make deploy`.

Also note:
  - k6 was not executed (stack down). That's fine —
    mark it in the map as "ready, needs live stack".
  - Coverage gate is not wired in CI yet — remember to
    wire after inventory + auth hit ≥70%.

DO NOT TOUCH: C1–C3, C4, hero, nav, games, seed, refund.
Что я держу в карте
Закрыто (в origin/main, 6f75867):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Wave 3 — C1 + C2 + C3.

Hero Plan A, VOID v3, smooth scroll.

Локально (7 коммитов, не запушено):

9587569 authors tests.

2ff6583 COVERAGE.md.

fba74c6 browse.js.

6014e2c purchase.js.

03afe83 shop tests.

2 docs commits.

⚠️ Refactor в test-commit — требует split.

Wave 4 — в работе:

✅ authors — 93.2% service.

✅ shop — 83.9% service.

🟧 inventory (7%) — следующее.

🟧 auth (22%) — следующее.

🟧 k6 — готовы, smoke отложен.

🟧 COVERAGE.md — есть.

🟧 coverage-gate.sh — не в CI (после inv+auth).

Wave 4.5 (deferred):

🟧 Helm, Consul, OpenAPI sync.

Deferred (не сейчас):

🟧 C4 (payouts), author notifications, JWT role refresh, Wave 2 #5b, refund window, backfill-noiz-reviews.py, cursor pull-in.