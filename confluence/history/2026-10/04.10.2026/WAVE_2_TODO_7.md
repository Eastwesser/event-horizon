Вопрос 1 — Leaderboard storage для per-level
A) Composite game_id like flappy:5.

Плюс: минимальное изменение. Leaderboard service не трогаем — просто передаём другой ключ.

Минус: хак. game_id перестаёт быть чистым ID игры — теперь «game_id + level». Может путать в других местах (аналитика, rewards, профиль).

B) Реальное поле level + PG/Redis schema.

Плюс: чисто. game_id остаётся, level — отдельный параметр. Аналитика, фильтры, rewards — всё явное.

Минус: больше работы — миграция, обновление leaderboard service, Redis-ключи, gateway, DTO.

Мой выбор — B.

Причины:

Ты уже накопил техдолг из-за быстрых решений (reference_id varchar 100 → 200 — помнишь?). Не повторяй паттерн.

Per-level boards — это долгосрочная фича. Профиль, достижения, rewards — все захотят level как отдельное поле.

Аналитика — game_id:flappy:5 в ClickHouse — уродливо. game_id=flappy, level=5 — чисто.

Composite hack — сработает сейчас, сломает позже. Правильное решение — B.

Оценка: B — +1–2 дня работы. Стоит.

Но: если хочешь быстрее — A. Тоже сработает. Только пометь в TODO, что позже заменить на B.

Вопрос 2 — Scope
A) Flappy 1–10 only.

Плюс: пилот. Меньше работы, проще откатить.

Минус: только Flappy — остальные игры без уровней.

B) Flappy 1–20 now.

Плюс: сразу всё.

Минус: 20 уровней — сложнее дизайн. Надо 20 раз продумать баланс (скорость, gap, density). Дольше.

C) Flappy 1–10 + stub picker on other games.

Плюс: единый UI сразу.

Минус: stub — обманка. Игрок выбирает уровень в Towers, а механика не меняется. Плохо для kids — разочарование.

Мой выбор — A.

Причины:

Пилот — правильно. Flappy 1–10 — достаточно для проверки модели.

20 уровней — проще добавить после пилота. Сначала 10, убедиться, что работает.

Stub picker — нет. Не обманывай игрока. Если игра не поддерживает — не показывай picker.

После успеха A → B (расширить до 20).

Полный ответ Cursor'у
text
Answers:

1. Leaderboard storage = B.
   Real `level` field, not composite game_id.
   - Leaderboard service: add `level` to entry key.
   - Redis: leaderboard:<game_id>:<level>.
   - PG: add `level` column (default 1).
   - Gateway: accept `level` query param.
   - SubmitScore: level comes from the request (already exists,
     Flappy just sent 1 always).
   - Rationale: proper long-term, avoids a hack like the
     reference_id incident. Cost ~1-2 days extra, worth it.

2. Scope = A.
   Flappy 1–10 only.
   - Pilot to validate the model.
   - Levels 1–20 later once A works.
   - No stub picker on other games (do not show a picker that
     does nothing — kids get confused).

Locked recommendations from your doc:
  - Pre-game level picker in GameShell.
  - Rewards scale with level (multiplier).
  - Boost + level allowed; boosted runs still unranked.
  - Persist last selected level: localStorage first, profile
    later.

Draft the implementation plan. No code yet.

DO NOT TOUCH: seed, refund, author registration.
Обоснование — почему так
Два принципа:

Долгосрочное > быстрое. Ты уже обжигался на reference_id. Не повторяй.

Kids-safe > stub. Не обманывай игрока. Picker только там, где реально работает.

Плюс:

B — +2 дня, но чисто.

A scope — пилот, откат лёгкий.

20 уровней — позже, когда модель проверена.

Что я держу в карте
Закрыто:

v1.0.9 на GitHub.

Wave 1 — 6/6.

Wave 2 #1 — 0ee567f.

Wave 2 #2 — a881587.

Wave 2 #3 — 8c16926.

Push — сделано.

Сейчас:

🔴 Ответ Cursor'у: 1B + 2A.

🟧 Wave 2 #4 — design → implement.

Wave 2 (осталось):

🟧 #4 Уровни сложности (Flappy 1–10, real level field)

🟧 #5 Достижения

🟧 #6 Полиш игр (5, sequentially)

Отложено:

🟧 Refund window implementation

🟧 Игры emoji

🟧 backfill-noiz-reviews.py