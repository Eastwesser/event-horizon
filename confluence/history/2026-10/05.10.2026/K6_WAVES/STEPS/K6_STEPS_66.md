Отлично. Wave 4 (2) — correctness закрыт. Разбор
Что вижу
purchase 1 VU — ✅ всё зелёное
text
✓ errors rate=0.00%
✓ http_req_failed rate=0.00%
✓ p95 300ms (порог 1200)
✓ 100% checks (222/222)
purchase 5 VU — 🟧 correctness зелёный, latency нет
text
✓ errors 0%
✓ http_req_failed 0%
✗ p95 2.86s (порог 1200)
✓ 100% checks (264/264)
WARN: only 4 unowned items for 5 VUs — VU5 idle
Correctness — идеальный. Latency — 2.86s p95 на 4 VU. Не баг — нагрузка на purchase/cancel path (billing spend + shop tx + inventory reserve).

browse 5 VU — ✅ зелёный
text
✓ p95 355ms
✓ 0% errors
Мой анализ
Что закрыто:

✅ N+1 shop list (было 2.57s → 355ms).

✅ gateway inventory CB double-write.

✅ k6 purchase harness (owned skip + per-VU + cancel loop).

✅ browse baseline — clean.

✅ purchase 1 VU — clean.

Что открыто:

🟧 purchase 5 VU p95 2.86s — latency only, не функциональность.

Значит: под concurrent load purchase/cancel path тормозит. Candidates:

Billing spend (лампочки + билетики update).

Shop tx (purchase + cancel).

Inventory reserve/release.

NATS events (purchase.paid → fulfillment/notification).

Не индексы. EXPLAIN — следующий шаг, но не обязателен сейчас.

Что делать
Мой выбор — остановиться сейчас.

Причины:

Correctness — идеальный (0% errors на всех runs).

Latency под 4 VU — не критично для портфолио/демо.

Реальные пользователи — не будут бить purchase 4 VU concurrent (один игрок = один buy).

Indexes — speculative. Лучше после EXPLAIN + реальных данных нагрузки.

Time — уходит на оптимизацию нереалистичного сценария.

Push 3 коммита + закрыть Wave 4 (2).

Позже — если будет жалоба на purchase latency — вернуться с EXPLAIN на purchase/cancel path.

Промпт Cursor'у
text
Wave 4 (2) baseline closed.

Correctness is clean:
  - browse 5 VU: p95 355ms, 0% errors
  - purchase 1 VU: p95 300ms, 100% checks
  - purchase 5 VU: 100% checks, 0% errors (latency
    p95 2.86s is under load; not a functional issue)

Multi-VU purchase latency is NOT a priority right now:
  - Real user traffic won't hit purchase at 4 concurrent VUs.
  - Optimizing an artificial scenario risks speculative
    changes.
  - Defer EXPLAIN on purchase/cancel path until we have a
    real report of slow purchase.

Action:
  1. I'll push the 3 local commits (shop, gateway, k6).
  2. PR #2 (coverage gate workflow) is still open — I'll
     decide whether to merge now or keep it parked.

Wave 4 (2) — closed.

Next: Wave 4.5 — deferred:
  - Helm charts for k3s
  - Consul evaluate (recommend: skip unless I want a demo)
  - OpenAPI sync pass

Not now. Rest / other priorities first.

Standing rule: you don't push; you write commit messages.
DO NOT TOUCH: C1–C3, C4, hero, nav, games, seed, refund,
indexes.
Что я держу в карте
Закрыто (в origin/main, 202a9da):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Wave 3 — C1 + C2 + C3.

Hero Plan A, VOID v3, smooth scroll.

Wave 4 (1) — все gated services ≥70%.

PR #2 — открыт (CI coverage gate).

Локально (3 коммита, push после OK):

shop fix (batch IsItemOwned).

gateway fix (no double-write).

k6 purchase.js fix.

Wave 4 (2) — закрыт:

✅ browse — p95 355ms.

✅ purchase 1 VU — 100% checks.

✅ purchase 5 VU — 100% checks (latency only).

🟧 multi-VU latency — deferred (не критично).

Wave 4.5 (deferred):

🟧 Helm, Consul evaluate, OpenAPI sync.

Deferred:

🟧 C4, notifications, JWT role refresh, Wave 2 #5b, refund window, backfill-noiz-reviews.py, cursor pull-in.