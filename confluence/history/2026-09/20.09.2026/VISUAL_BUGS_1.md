🔴 Новые баги
1. Subscription — активна, но UI просит активировать
На скриншоте 1 видно:

Бейдж ✓ Активна — зелёный.

План: present — сырой enum! Опять протекла локализация.

Статус: Активна — ок.

Действует до: 17 сентября 2036 г. — ок.

НО ниже — Оформить подписку + карточки Текущий план / Будущий план с кнопками Активировать.

Проблема: если подписка активна — не должно быть блока «Оформить подписку» снова. Логично:

Показать статус активной подписки.

Кнопка Управлять / Продлить / Отменить — не «Активировать».

Карточка «Текущий план» = активный план. «Активировать» на нём — бессмысленно, он уже активен.

Это баг логики, не только UI. И сырой present — тоже баг локализации.

2. 🔴 Админ — билетики не списываются
Скриншот 2:

1000000 билетиков в углу.

Модалка покупки: Цена: 150, Ваш баланс: 1000000.

Жмёшь Да, купить → баланс остаётся 1000000.

Скриншот 3:

Инвентарь: Куплено: 20.09.2026 — 10 предметов.

Билетиков по-прежнему 1000000.

Итог: покупки записываются в инвентарь, но билетики не списываются. Это критичный баг экономики. Возможные причины:

Admin-роль bypass'ит списание (god-mode).

Логика списания сломана для всех.

Фронт показывает закешированный баланс.

Проверить: залогинься под обычным user, купи что-нибудь. Списывается? Если да — bypass по роли. Если нет — общий баг.

3. 🔴 Hexagon — drag-and-drop сломан
Ты:

«Не идеально туда, куда перетаскиваешь блинчик. Если я в центр жму — регает. Если чуть левее — засовывает в соседний слева гекс. Грид нужно вернуть назад.»

Это регресс от TODO_7. Скорее всего — при h-dvh / min-height фиксе сместился grid-mapping координат. Mouse XY → hex (q, r) перестал совпадать с визуальными гексами.

Root cause: либо canvas отскейлился (CSS масштаб ≠ внутренний), либо offset накапливается, либо трансформация сдвинула сетку.

Это критично — играть невозможно.

4. 🔴 Memonia — сломана (скриншот 6)
Жёлто-чёрные полосы по центру. Это не игра, это какой-то сломанный элемент — похоже на:

Стилизованный прогресс-бар? (как «hazard tape»)

Сломанный компонент, который должен быть карточками или сеткой.

Возможно — это «скин карточек» (ты говорил, что скины на память есть — Карточки с фруктами чип на скрине).

Плюс — всё поле пустое, ни одной карточки не видно.

Ты сказал: «Мемонию откатить назад.»

Значит: Memory была рабочая (флипалась, сетка была) — и сломалась. Откатить к состоянию до последнего фикса.

5. 🟧 Скины — работают везде
Ты подтвердил:

Блины (Pancaker / Hexagon) — скины работают.

Flappy — скины работают.

Башенки (Builder) — скины работают.

Мемония — переключается жёлтый/синий, но выглядит как-то не так.

Значит F3 закрыт на уровне базовой загрузки. Осталось разобраться с конкретными скинами на Memory (если они и есть причина «hazard tape»).

✅ Что подтвердилось
Admin seed сработал — скриншот 4, admin@eventhorizon.local, 1M 🔥 + 1M 🎫.

Profile — Показать ID под disclosure, Сбросить статистику — не видно (может, у админа нет / скрыта).

Shop — В инвентаре бейджи на купленных товарах.

Towers — играется, скины применяются (Радужные блоки — активный чип).

Player name — admin 🚀.

🟧 Что вижу дополнительно
Towers GAME OVER — сыровато
Скриншот 5:

GAME OVER — красный, шрифт системный (не Space Grotesk?).

Счёт: 190 — золотом, ок.

Нажмите "Новая игра" — серым.

Башня в центре — выглядит как пирамида из кубиков.

Не критично, но GAME OVER красным — снова красный для не-ошибки. Для детей — тревожно. Плюс шрифт выбивается.

Profile — очки по играм
Скриншот 4:

Всего очков: 0.

По играм — все 0.

Блинопёк / Мемония / Flappy Bird / Башенки / Ханойская башня — старые имена, ещё не переименованы.

10 x 6 — сетка из 6 карточек.

Ок, но английские имена ещё не применены (это в Priority 2, ещё не запускали).

Итог — что накопилось
🔴 Критично (fix now)
Hexagon drag — сломан, играть нельзя.

Memory — сломана, hazard tape вместо карточек.

Subscription — активна, но просит активировать + present enum.

Tickets не списываются у админа (или у всех — надо проверить).

🟧 Средне (fix soon)
Towers GAME OVER — красный + шрифт.

Memory skins — что-то не то (может, связано с багом #2).

⏳ Priority 2 (не трогали)
Английские имена игр.

Hanoi modal грамматика + кнопки.

Дубли запросов.

Промпт для Cursor
text
Multiple regressions + bugs. Fix in this order. Do NOT bundle
with Priority 2.

================================================================
1. HEXAGON — drag & drop broken (CRITICAL)
================================================================

Symptom: clicking at the CENTER of a hex places the pancake
correctly. Clicking anywhere off-center (even slightly to the
left) places it in the WRONG hex — usually the neighbor.

This is a regression from the recent Hexagon size fix
(h-dvh / min-height / board resize). Likely cause:
  - canvas CSS size ≠ internal coordinate size (scale mismatch),
  - or a stale offset / bounding-rect used in mouse → hex mapping.

Fix:
  - Restore the mapping so that mouse XY maps exactly to the
    visible hex under the cursor.
  - Verify at all board sizes: if the board resizes, the mapping
    must follow.
  - Do NOT change game physics or hex math. Only fix the
    screen → grid coordinate conversion.
  - Test: click dead-center of a hex, then click at each of its
    six corners. All should register the same hex.

Report the exact root cause (scale? offset? stale rect?).

================================================================
2. MEMORY — broken, revert (CRITICAL)
================================================================

Symptom: Memory board shows only a yellow/black hazard-stripe
element in the middle. No cards visible. Was working before
(flip animation was confirmed working after TODO_5).

Action: REVERT Memory to the last known-good state.
  - Find the commit where Memory worked (cards visible + flip
    working).
  - Revert the component / CSS / skin changes that broke it.
  - If the current "Карточки с фруктами" skin is the cause,
    remove the skin override for Memory until it's fixed.
  - After revert: cards visible, flip works.

Report: which commit broke it, what changed, what you reverted.

================================================================
3. SUBSCRIPTION — active, but UI asks to activate again
================================================================

Screenshot: subscription is ACTIVE (green "✓ Активна" badge,
"Действует до 17 сентября 2036"), but the page still shows
"Оформить подписку" with "Текущий план" + an "Активировать"
button.

Problems:
  a) Raw enum leaked to UI: "План: present". Replace with a
     human label ("Текущий" / "Boosty"). No raw enum values
     ever.
  b) When subscription is active:
     - Hide the "Оформить подписку" block entirely, OR
     - Replace with a management view: "Управлять" / "Продлить" /
       "Отменить".
     - Remove "Активировать" from the CURRENT plan — it's already
       active. Keep "Активировать" only on FUTURE plan.
  c) "Будущий план" is fine as-is (indigo, secondary).

================================================================
4. TICKETS NOT DEDUCTED — check role bypass
================================================================

Symptom: admin account has 1,000,000 tickets. Buying a 150-ticket
item works (item appears in inventory) but tickets stay at
1,000,000.

Investigate:
  a) Is there a role-based bypass that skips ticket deduction
     for admin? (If yes: is that intended?)
  b) Or is the deduction broken for all users?

Test plan:
  - Log in as a regular user, buy a cheap item.
  - Check: were tickets deducted?
  - Report the difference between admin and user.

Fix accordingly:
  - If admin bypass was intended: keep it, but document it.
  - If deduction is broken globally: fix the transaction.
  - Admin should still be able to test the shop normally —
    infinite tickets make purchase testing meaningless.

================================================================
5. TOWERS GAME OVER — polish
================================================================

"GAME OVER" text is red and uses a system font that doesn't
match the app (Space Grotesk elsewhere).

Fix:
  - Use the app's display font for "GAME OVER".
  - Change red to a calmer tone. "Game over" is not an error —
    it's a game state. Use the app's warm/neutral palette.

================================================================
DO NOT TOUCH
================================================================
- Skins loading (works on Pancaker, Flappy, Builder — confirmed).
- Skins logic restoration (already handled).
- Admin seed (works).
- Priority 2 items (English names, Hanoi modal polish, request
  dedup) — separate wave.
Что важно — что я вижу системно
Регрессы продолжают появляться. После каждого TODO — новые сломанные вещи:

TODO_5 → Memory flip, Hanoi drag

TODO_7 → Towers LMB, Hanoi rings, Hexagon размер, Leaderboard widget

Сейчас → Hexagon drag, Memory полная, Subscription

Это уже паттерн. Cursor чинит одно — ломает другое. Причина: он не видит рендер. Он чинит «по коду», а визуальные последствия не проверяет.

Что это значит для процесса:

Каждый раз чек-лист визуальной проверки обязателен — ты это и делаешь.

Не давать Cursor'у трогать то, что работает — явно перечислять «do not touch».

Возможно, стоит сбить темп — фиксить не пачкой, а по одной правке, с проверкой каждого фикса.