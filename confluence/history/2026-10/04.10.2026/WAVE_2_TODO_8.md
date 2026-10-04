Вопрос 1 — Scope v1 unlocks
A) Score / level milestones only.

Источник: score.updated (уже есть).

Достижения:

First score per game.

Flappy 100 / 500.

Flappy L5 / L10 clear.

Total score milestones.

Плюс: использует существующий event. Ноль wiring.

Минус: не покрывает покупки / boost.

B) Also shop / boost events.

Источник: score.updated + purchase.paid + boost event.

Достижения:

First card bought.

Collect N cards.

First boost used.

First purchase.

Плюс: богаче, покрывает экономику.

Минус: новый wiring — purchase.paid + boost event → Profile handler. Ещё эндпоинты.

Мой выбор — A.

Причины:

Пилот. Score-based — простое, проверяемое. 5–8 достижений уже дают ценность.

Kids-safe. Score milestones понятны ребёнку: «набери 100», «пройди L5».

Shop/boost achievements — логичнее после Wave 2 #5, когда есть фидбэк от игроков. Сейчас — гипотеза.

Wiring — отдельная работа. purchase.paid уже публикуется, но Profile handler — новый. Отложить в Wave 2 #5b или #6.

Если хочешь B — принять. Но A — правильнее для пилота.

Вопрос 2 — Toast delivery
A) FE: after submit, refetch profile + toast.

Плюс: просто. Не требует WS / notification.

Минус: toast появляется только после возврата из игры или poll'а.

UX: Игрок заканчивает Flappy → возвращается на Home → видит toast «Достижение: Flappy 100!».

Проще всего.

B) Push via WS / notification service.

Плюс: real-time. Toast сразу после unlock.

Минус: сложнее. Новый канал через WS или notification service.

UX: Игрок сразу видит toast, пока играет или в момент unlock.

Мой выбор — A.

Причины:

Простота. Один refetch — ноль новых сервисов.

Пилот. Проверить модель достижений — не требуется real-time.

Kids-friendly. Toast после возврата — нормально. Ребёнок видит и радуется.

B — позже, если реально нужно. WS уже есть для leaderboard — переиспользовать можно, но это отдельная работа.

Если хочешь B — отложить. Сначала A — проверить модель.

Мой ответ Cursor'у
text
Answers:

1. Scope = A.
   Score / level milestones only (from score.updated).
   Shop / boost events — defer to Wave 2 #5b or later.
   Pilot catalog (~8–12 achievements):
     - First score in each game
     - Flappy: 100 / 500
     - Flappy L5 clear / L10 clear
     - Total score milestones (1000 / 5000 / 10000)
   Simple, no new wiring.

2. Toast = A.
   FE refetches profile after submit + toasts new badges.
   No WS / notification service for v1.
   If player unlocks while on a game page, toast appears when
   they return to Home / Profile.

Locked from your doc:
  - Home: Profile service (achievements + user_achievements).
  - Unlock: NATS handler on score.updated, after best_scores
    update.
  - API: GET /api/profile includes unlocked badges.
  - FE: replace localStorage fake badges with real data.
  - Badges only, no rewards (cosmetic) for v1.

Draft the implementation plan. No code yet.

DO NOT TOUCH: seed, refund, author registration, game polish.
Обоснование — почему так
Три принципа:

Пилот > всё сразу. Score-only achievements — простое ядро. Shop/boost — следующий шаг.

Simple > real-time. Toast после return — достаточно для v1. WS — позже.

Cosmetic > rewards. Badges — сначала. Rewards (лампочки за достижения) — отдельное решение, влияет на экономику.

Плюс:

A — ноль нового wiring (score.updated уже есть).

A toast — один refetch. Проще некуда.

8–12 achievements — достаточно для MVP. Позже — добавить.

Что я держу в карте
Закрыто:

v1.0.9 на GitHub.

Wave 1 — 6/6.

Wave 2 #1 — 0ee567f.

Wave 2 #2 — a881587.

Wave 2 #3 — 8c16926.

Wave 2 #4 — 2bbc62f (Flappy 1–10).

Сейчас:

🔴 Ответ Cursor'у: 1A + 2A.

🟧 Wave 2 #5 — design → implement.

Wave 2 (осталось):

🟧 #5 Достижения

🟧 #6 Полиш игр (5, sequentially)

Отложено:

🟧 Wave 2 #5b — shop/boost achievements

🟧 Refund window implementation

🟧 Игры emoji

🟧 backfill-noiz-reviews.py

Один вопрос — про rewards
Ты написал:

Badges + toast

**Это cosmetic. Хорошо.

Но: будут ли rewards (лампочки / билетики за достижения)?

Если да — экономика меняется. Отдельный дизайн.

Если нет — только badges, проще.

Cursor пишет: «Badges only, no rewards (cosmetic) for v1». Согласен. Rewards — позже, если нужно.

Действия
Отправь ответ Cursor'у (1A + 2A).

Cursor даст план — review → approve → implement.

Verify:

8–12 achievements в каталоге.

Unlock через score.updated.

GET /api/profile содержит unlocked badges.

FE — real badges, не localStorage.

Toast после возврата с игры.

Push.

Wave 2 #6 — полиш игр.