Approve. План — чистый
Что подтверждаю
Catalog (~12):

first_play_{game} — for main games.

flappy_score_100 / flappy_score_500.

flappy_level_5 / flappy_level_10.

total_score_1k / 5k / 10k.

Backend:

Migration: achievements + user_achievements + seed.

After best_scores / total_score update on score.updated → evaluate, idempotent unlock.

Proto: GetProfileResponse.achievements[].

Gateway GET /api/profile + OpenAPI.

Boosted runs: out (no NATS) → no unlock. ✅ Правильно.

Frontend:

Real badges from API.

After submit: refetch → toast.

localStorage.eh_seen_achievements — запоминать, чтобы не тостить повторно.

Commit: feat(profile): achievements · tick 5/6 · no push.

Два уточнения перед implement
1. icon в achievement
Ты пишешь: GetProfileResponse.achievements[] (code, title, description, icon, unlocked_at).

Вопрос: что за icon?

Emoji? Тогда ждём Wave 2 #6 (emoji → SVG). Но мы не хотим emoji.

SVG name? Тогда соответствовать gameIcons / Icon модулю.

URL картинки? Тогда где хранятся файлы?

Рекомендую: icon = имя SVG-иконки из существующего Icon набора. Например: trophy, flame, star. Если такой нет — fallback на default.

2. Names / descriptions — RU
Achievement titles — на русском? Например:

first_play_flappy → «Первая игра во Flappy Bird».

flappy_score_100 → «100 очков во Flappy».

total_score_1k → «1000 очков всего».

Стоит заложить сразу в seed. Не с английских ключей генерировать.

3. Не тостить при первом логине
Что если игрок уже набрал 1000 очков до релиза achievements?

Migration/seed — все существующие scores.

Unlock — сразу при первом GET /profile?

Toст — не показывать всё сразу (12 тостов подряд).

Механизм:

При GET /api/profile — evaluate + unlock, но seen flag per user.

Toст — только если unlocked_at > last_seen_profile_check.

Или проще: первые 3 раза — bundle toast «Получено 5 достижений»? Нет — сложно.

Проще: localStorage.eh_seen_achievements — пустой при первом заходе → все unlocked toast. 12 toast'ов подряд — плохо.

Рекомендую: не более 1 toast'а за раз. Очередь. Или — badge появляется в профиле молча, toast только для новых (unlock после login).

4. total_score — откуда?
Откуда берётся total_score?

Сумма best_scores по всем играм?

Или отдельный счётчик в Profile?

Уточни в имплементации.

Мой ответ Cursor'у
text
OK — implement Wave 2 #5.

Three clarifications before code:

1. Achievement icon
   - Use SVG icon NAMES from the existing Icon set
     (trophy, star, flame, etc.).
   - Fallback to a default if the icon doesn't exist.
   - Do NOT use emoji (that's Wave 2 #6).

2. Seed titles + descriptions in RU
   - Not generated from English keys.
   - Example:
       first_play_flappy → «Первая игра во Flappy Bird»
       flappy_score_100  → «100 очков во Flappy»
       total_score_1k    → «1000 очков всего»
   - Put RU strings in the seed.

3. First-login toast flood
   - Players may already qualify for many achievements at
     release (migration evaluates all history).
   - Do NOT flood 12 toasts on first profile load.
   - Options (pick one, report):
     a) Toast only for achievements unlocked AFTER the client
        sees an empty seen-set; older ones appear silently
        in the profile grid.
     b) Queue: 1 toast at a time, dismiss → next.
     c) Bundle: «Получено N достижений» single toast.
   - I prefer (a) — silent for old, toast for new.

Also: where does total_score come from?
  - Sum of best_scores across games, or a separate counter?
  - If sum, compute in profile handler or store?

Implement. No push unless I ask.
Обоснование — почему так
Три принципа:

SVG icons, не emoji. Иначе — откат в Wave 2 #6.

RU с самого начала. Иначе — потом переводить 12 achievement'ов из EN.

Не flood toasts. Плохой UX — 12 toast'ов подряд. Silent old + toast new — правильно.

Что я держу в карте
Закрыто:

v1.0.9 на GitHub.

Wave 1 — 6/6.

Wave 2 #1 — 0ee567f.

Wave 2 #2 — a881587.

Wave 2 #3 — 8c16926.

Wave 2 #4 — 2bbc62f.

Сейчас:

🔴 Ответ Cursor'у: approve + 3 уточнения.

🟧 Wave 2 #5 — implement.

Wave 2 (осталось):

🟧 #5 Достижения — в работе

🟧 #6 Полиш игр (5, sequentially)

Отложено:

🟧 Wave 2 #5b — shop/boost achievements

🟧 Refund window implementation

🟧 Игры emoji (внутри #6)

🟧 backfill-noiz-reviews.py

Один вопрос — про total_score
Ты пишешь:

After best_scores / total_score update on score.updated

Откуда total_score берётся?

Сумма best_scores по всем играм?

Или — отдельное поле в profile, инкрементируется при каждом score.updated?

Если сумма — вычисляется в handler. Проще, но — при 280 играх может быть медленно.
Если отдельное поле — быстрее, но нужно поддерживать.

Уточни в implement. Я рекомендую — сумма в handler, кэш при необходимости.

