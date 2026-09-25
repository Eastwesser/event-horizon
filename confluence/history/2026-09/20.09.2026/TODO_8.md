 Проверка после TODO_7
Ты прогнал серьёзный чек. Фиксирую всё: что зашло, что сломалось новым, что осталось.

✅ Что зашло
#	Пункт	Подтверждение
1	Memory flip визуально	«Да, теперь флипаются»
2	«Мемония» (не Меморина)	Home, скриншот 4
3	Navbar центрирован	Магазин · Лидерборд · Профиль в центре, ☰ + Выйти справа
4	Subscription tones	Текущий план — золото, Будущий — индиго (скриншот 3)
5	Inventory — кнопки скрыты для user	✅
6	Footer — «Поддержать проект» справа	✅
7	Hanoi — drag не растягивает	✅ (но см. регресс ниже)
8	Hexagon — без скролла	✅ (но см. регресс ниже)
9	16+ минут — не кикает	✅
🔴 Регрессы — TODO_7 сломал три вещи
Регресс 1 — Hanoi: кольца стали «бусинами», уровень 5–8 недоступен
Ты сказал:

«Кольца размером с горошину. Так быть не должно.»
«Я не могу выбрать теперь уровень 5-8 колец.»

Что произошло: Cursor чинил «растягивание» — заменил % на px from peg width. Перестарался: теперь кольца жёстко привязаны к ширине стойки, а стойка — узкая. Отсюда «бусины».

И параллельно сломал селектор уровня 5–8. Скорее всего — px-фикс задел размеры, и селектор обрезался/не поместился.

Это регресс. В прошлой версии кольца выглядели нормально — ты сам сказал.

Регресс 2 — Towers: ЛКМ не работает
Ты сказал:

«В игре Башенки не работает ЛКМ. На пробел работает, а ЛКМ курсор проебал.»

Что произошло: либо GameShell перехватил mouse-события, либо canvas потерял onClick, либо pointer-events отвалились при h-dvh-фиксе.

Это регресс от TODO_7. Раньше работало.

Регресс 3 — Hexagon: игровое поле стало мелким
Ты сказал:

«Как-то маленький он что ли стал, экран с игрой.»

Что произошло: h-dvh flex layout заставил stage сжаться. Раньше было больше — теперь всё вписано, но мелко. Для детей — плохо: гексы мелкие, сложно тапать.

Это тоже регресс от TODO_7.

🔴 Новый баг — Leaderboard внутри Hexagon крашится
Из лога:

text
HexagonGame.tsx:108 TypeError: Cannot read properties of undefined (reading 'toLocaleString')
  at Leaderboard.tsx:128:34
  at Array.map
  at Leaderboard (Leaderboard.tsx:107:22)

The above error occurred in the <Leaderboard> component.
ErrorBoundary.tsx:26 ErrorBoundary caught: ...
Что произошло:

Ты в игре Блинопёк нажал Лидерборд.

Компонент Leaderboard (виджет, не LeaderboardFull) — крашится.

Тот же баг, что был в LeaderboardFull.tsx:128 до фикса TODO_5. Cursor починил full-версию, но не виджет.

ErrorBoundary поймал → показал «что-то пошло не так» → перезагрузка страницы.

Это значит:

Null-guard был добавлен не везде. Виджет Leaderboard пропущен.

Ты получаешь краш при клике на кнопку в игре — прямо во время игры.

🟧 Что осталось незакрытым (из старого списка)
Ты подтвердил:

#	Пункт	Статус
1	Английские имена игр	НЕ СДЕЛАНО — не отправляли, надо решить
2	Hanoi modal — грамматика ход(ов)	Не поправлено
3	Hanoi modal — кнопки разной ширины	Не поправлено
4	authors / leaderboard дубли	Не упомянуто
5	whoami — 8 → 2 (кэш)	✅ улучшено, но всё ещё дублируется
По логу вижу новые дубли:

text
get /inventory/items    ← 2 раза
get /payment/subscription ← 2 раза
get /whoami             ← 2 раза
StrictMode? Или реальные дубли? Cursor говорил про кэш whoami, но вижу два whoami в логе — значит кэш не всегда срабатывает.

📌 Контекст про роли — фиксирую
Ты сказал:

«Про роли да, у меня такой IAM, но я не залогинен как админ… Мне надо будет затестить путь админа, путь пользователя, и путь автора.»

Фиксирую как отдельный блок работ на будущее:

Есть IAM с ролями user / author / admin.

Сейчас ты под user.

Нужно будет протестировать три пути отдельно.

Это не для текущей волны — но важно помнить.

Что из этого следует сейчас: Inventory CRUD не проверить под твоим аккаунтом. Кнопки скрыты — правильно. CRUD-функционал остаётся непроверенным до админ/автор-аккаунта.

Итоговая картина после TODO_7
Фронт-прогресс
Layout — ✅

Hero — ✅

Navbar — ✅ (центр, бургер, Выйти справа)

Табы — ✅ (Shop, History)

Subscription — ✅

Profile — ✅ (ID скрыт, сброс через модалку)

Footer — ✅

ErrorBoundary — ✅ (работает, ловит краши)

Memory flip — ✅

Hanoi drag — ✅ (не растягивает)

Регрессы от TODO_7 (3)
🔴 Hanoi — кольца «бусины», уровень 5–8 не выбрать

🔴 Towers — ЛКМ не работает

🔴 Hexagon — поле мелкое

Новые баги (1)
🔴 Leaderboard-виджет крашится в Hexagon

Старое незакрытое (4)
🟧 Английские имена

🟧 Hanoi modal — грамматика ход(ов)

🟧 Hanoi modal — кнопки разной ширины

🟧 authors/leaderboard дубли (whoami всё ещё 2×)

Что предлагаю сделать
Три пути:

A. Сначала фиксить регрессы от TODO_7.
Потому что они сломали то, что работало. Hanoi rings, Towers LMB, Hexagon размер. И новый краш Leaderboard.

B. Сначала разобраться с ролями.
Залогиниться как admin/author, проверить Inventory CRUD, понять IAM. Потому что без этого не проверить CRUD.

C. Писать финальный промпт сразу с обоими блоками.
Регрессы + старое незакрытое + просьба не трогать то, что работает.

Мой выбор — A → B → C:

Сначала регрессы — они блокируют тестирование (Towers нельзя играть, Hanoi сломан, Hexagon мелкий).

Потом роли — нужен admin-аккаунт для CRUD-проверки.

Потом — финальный промпт с английскими именами, грамматикой, дублями.

=====

Промпт для Cursor (копируй)
text
CONTEXT
TODO_7 fixed 4 items but introduced 3 regressions and left 1 new
crash. Priority order below. Do NOT bundle everything — fix
Priority 1 first, verify, then move to Priority 2.

================================================================
PRIORITY 1 — REGRESSIONS FROM TODO_7 (fix first)
================================================================

--- 1a. Towers — LMB does not work ---

Symptom: keyboard SPACE fires a block drop, but LEFT MOUSE CLICK
does nothing. Same action, two inputs — one works, one doesn't.

Expected: LMB must trigger the exact same handler as SPACE.

Investigate:
  - Where is the SPACE handler attached? (keydown on window /
    document / canvas / a focused element?)
  - Was there an onClick / onMouseDown / pointerdown that got
    lost in the TODO_7 GameShell refactor?
  - Does the canvas still have pointer-events enabled? Any
    overlay intercepting clicks?
  - Is the canvas focusable? Does LMB need to focus it first?

Fix so that LMB and SPACE are equivalent. Do not change game
physics.

--- 1b. Hanoi — rings became "beads", level selector 5–8 broken ---

Symptom:
  - Rings render as tiny "beads" — should be as wide as the peg
    can hold at level 8 (as in the PREVIOUS working version).
  - Level selector no longer lets me choose 5–8 rings.

Root cause: TODO_7 changed ring sizing from % of viewport to px
from peg width. It overcorrected — rings are now locked to the
narrow peg, and the fix also broke the level selector.

REVERT the Hanoi drag/size fix to the PREVIOUS state — the one
where pegs were sized to hold the widest ring (level 8) and rings
scaled correctly.

KEEP from TODO_7:
  - The GameShell viewport layout (h-dvh etc.) — do not revert
    that part, only the Hanoi-internal sizing.
  - The drag itself — the ring must NOT stretch during drag.

So the target is: previous visual proportions + no stretch during
drag. Not the current "beads".

--- 1c. Hexagon — game area became too small ---

Symptom: after the h-dvh fix, the hex board fits on one screen
(good) but is now too small — hexes are hard to read / tap.

Root cause: h-dvh forces the stage to compress to whatever is
left. There is no minimum size.

Fix: keep it one-screen, but enforce a sensible minimum.
  - Give the board a min-height and/or aspect-ratio so it does
    not shrink below a comfortable size.
  - If the viewport is too short, prefer allowing a small scroll
    over shrinking the board into unreadability.
  - Do NOT revert h-dvh entirely — the "one screen without
    scrolling" goal stays for normal viewports.

Report the actual measured board size before / after.

--- 1d. Leaderboard widget crashes inside Hexagon ---

Symptom: inside the Hexagon game, clicking "Лидерборд" throws:

  HexagonGame.tsx:108 TypeError: Cannot read properties of
  undefined (reading 'toLocaleString')
    at Leaderboard.tsx:128:34
    at Array.map
    at Leaderboard (Leaderboard.tsx:107:22)

ErrorBoundary catches it and the page reloads — mid-game.

Root cause: TODO_5 fixed LeaderboardFull but NOT the Leaderboard
widget. Same undefined.toLocaleString() pattern.

Fix:
  - Apply the same null-guard to Leaderboard.tsx (the widget),
    line ~128.
  - Guard each entry: skip or placeholder if score is missing.
  - Treat null/undefined leaderboard data as an empty state,
    not a crash.

================================================================
PRIORITY 2 — REMAINING ITEMS (after Priority 1 is verified)
================================================================

--- 2a. English names for all games ---

Rename in the UI, on Home, in Leaderboard, on Profile, and in
backend GetGameInfo Name fields:

  Pancaker     (currently "Блинопёк", game_id "hexagon")
  Flappy Bird  (unchanged, game_id "flappy")
  Builder      (currently "Башенки", game_id "towers")
  Hanoi        (currently "Ханойская башня", game_id "hanoi")
  Memonia      (currently "Мемония", game_id "memory")

IMPORTANT:
  - game_id values stay as-is. Do NOT change the backend wire
    protocol, leaderboards, or saved scores. Only change the
    DISPLAY NAME.
  - Update backend GetGameInfo Name strings to match.
  - Update any UI string that currently says "Ханойская башня"
    or "Башенки" or "Блинопёк" or "Мемония".
  - Cross-check: Home cards, Leaderboard headers, Profile
    "records by game" list, game page titles.
  - After rename, grep for old names to confirm zero leftovers.

--- 2b. Hanoi win modal — grammar ---

"Превышение на 39 ход(ов)" → proper Russian pluralization:
  1 ход, 2 хода, 5 ходов, 21 ход, 22 хода, 25 ходов, etc.
Use a plural helper, not "(ов)".

--- 2c. Hanoi win modal — button widths ---

"Сохранить рекорд" and "Играть снова" currently have different
widths. Make them equal (both w-full within the modal, stacked),
or place them side by side with equal flex — pick one and be
consistent.

--- 2d. whoami / inventory / subscription — duplicate calls ---

Console log shows each of these fired twice on page load:
  get /auth/whoami         ×2
  get /inventory/items     ×2
  get /payment/subscription ×2

The 60s whoami cache is present but not preventing the double-fire.
Investigate:
  - Is this React StrictMode double-invoke in dev only?
  - If yes: fine, no fix needed — note it.
  - If no: find the real cause (two mounts, two useEffects, a
    provider re-render) and dedupe.

Report which of the two it is.

================================================================
DO NOT TOUCH
================================================================
- Game physics (Towers falling, Flappy collision, Hexagon match,
  Hanoi legality, Memory reveal logic).
- The nav centering, footer, Subscription tones, Inventory role
  gating — all confirmed working, leave as-is.
- ErrorBoundary itself — it worked correctly (caught the crash).

================================================================
PROCESS
================================================================
Step 1: Fix Priority 1 (1a–1d). Verify tsc + build. Report.
Step 2: Wait for me to visually confirm Priority 1.
Step 3: Only then do Priority 2.
Do not bundle Priority 1 and Priority 2 in one pass.
Отдельно — для тебя (не Cursor)
Роли — блок работ на потом
Что знаем:

IAM с ролями user / author / admin.

Сейчас ты под user.

UI скрывает CRUD-кнопки — правильно.

Бэкенд работает.

Нужно создать admin-аккаунт через БД.

Что это значит для дальнейшего тестирования:

Создать admin-аккаунт в БД — вручную SQL-запрос (или скрипт). Роль в таблице users / роли.

Проверить три пути:

user — видит игры, магазин, лидерборд. Не видит CRUD-кнопки.

author — видит игры + видит CRUD inventory.

admin — видит всё + возможно, админ-панель.

Проверить Inventory CRUD под author/admin — create/update/delete.

Проверить пути между ролями — что user не видит того, что видит author.

Это отдельная волна. После того как Priority 1 + 2 закроются.

Что я держу в карте
TODO_7 закрыл:

Hanoi drag stretch ✅

GameShell viewport ✅ (но с регрессом на Hexagon)

Inventory role gating ✅

Subscription tones ✅

whoami cache ✅ (но всё ещё 2×)

Navbar centering ✅

TODO_7 сломал:

🔴 Hanoi — кольца-бусины, уровень 5–8

🔴 Towers — ЛКМ не работает

🔴 Hexagon — поле мелкое

Не закрыто:

🔴 Leaderboard-виджет крашится (TODO_5 починил full, забыл widget)

🟧 Английские имена

🟧 Hanoi modal грамматика + кнопки

🟧 Дубли запросов

Для тебя отдельно:

⏳ Создать admin-аккаунт в БД

⏳ Протестить 3 пути: user / author / admin

---

## Cursor pass — Priority 1 only (waiting visual confirm)

### 1a Towers LMB
**Cause:** click listener attached in `useEffect` while `skinsLoading` still showed a Spinner (no canvas). After skins loaded, effect did not re-run → SPACE (window) worked, LMB never bound.
**Fix:** `onClick={handleCanvasClick}` on the `<canvas>` itself + `shrink-0`.

### 1b Hanoi beads / level 5–8
**Cause:** pegs compressed under GameShell → tiny % widths looked like beads; `overflow-hidden` also risked clipping the level `<select>`.
**Fix:** peg `min-width: 160px`, chunkier ring scale (≈42–92%), thicker rod/base; GameShell no longer hard-clips (`overflow-y-auto`); controls `z-20`. Drag still uses **px from peg** (no viewport stretch).

### 1c Hexagon too small
**Cause:** stage crushed board with `max-height: 100%` + tiny tray.
**Fix:** `.hex-grid-container { min-height: 380px }`, larger tray stacks (72px); slight page scroll allowed if viewport is short.

### 1d Leaderboard widget crash
**Fix:** null-safe `score` + empty `entries` on bad payload (same class of bug as LeaderboardFull).

`tsc --noEmit` → 0.

**Priority 2 not started** — waiting for your visual OK on 1a–1d.
