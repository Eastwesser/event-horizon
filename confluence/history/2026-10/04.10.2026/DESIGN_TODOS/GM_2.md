Git — всё чисто
Ты подтвердил:

text
origin/main..HEAD — пусто. 146cae8 и 8bfa3db уже на remote.
Значит:

✅ Wave 2 #5 (146cae8 achievements) — на GitHub.

✅ Hero + disk hover (8bfa3db) — на GitHub.

Untracked: только DESIGN_TODOS/ — твои заметки, не критично.

Можно выдохнуть. Всё сохранено.

Disk hover — работает
Cursor описал:

:hover на .eh-disk (~700ms ease-out):

Core scale 1 → 1.12, glow сильнее (+ cyan haze).

Rings scale 1 → 0.92, spin 140s → 80s.

Mark (logo в круге) scale 1 → 1.06, чуть ярче.

prefers-reduced-motion — всё выключено.

Звучит правильно. Ты подтвердил: на втором скрине диск с кольцом и wordmark виден, но delta scale на статичном PNG слабая.

Как проверить глазами:

Открой Home в браузере.

Наведи на disk → должно визуально «засасывать».

Убери → обратно за 700ms.

DevTools → Rendering → Emulate prefers-reduced-motion: reduce → эффект выключен.

Если не видно — скажи Cursor'у. CSS hover — не всегда очевидно.

Hero — два вопроса
Вопрос 1: PNG без EVENT HORIZON — есть?
Cursor пишет:

В brand/ только logo-minimal.png с текстом.

Значит: нет версии без wordmark. Нужно создать.

Варианты:

A) Экспортировать новый PNG без надписи.

Открыть logo-minimal.png в редакторе (Photoshop / Figma / GIMP).

Убрать текст EVENT HORIZON.

Экспортировать как hero-planet.png или hero-planet.webp.

Положить в frontend/public/images/brand/.

Плюс: чисто.
Минус: нужно открыть редактор.

B) Оставить текущий PNG + фикс object-position — wordmark виден, дубль остаётся.
Плюс: ноль работы.
Минус: три раза Event Horizon (navbar + hero + disk).

C) Использовать hero-баннер как отдельный арт — не logo-minimal, а другая картинка (космос, планета, без текста).
Плюс: если есть подходящая.
Минус: нужно найти / сгенерировать.

Мой выбор — A.

Причины:

Дубль — визуальный шум. Убрать = элегантнее.

Работа — 5 минут в Photoshop / GIMP / Figma.

Wordmark остаётся в navbar + в disk — достаточно.

Но:

Если нет времени — B тоже ок. Потом экспортируешь.

Вопрос 2: Фикс object-position + height — OK?
Да, OK.

object-position: center top — wordmark виден целиком, не срезан navbar.

height: min(60vh, 520px) — меньше обрезки.

Плюс — если выберешь A (новый PNG без текста) — тот же фикс работает.

Мой ответ Cursor'у
text
Git verified: origin/main..HEAD empty, both 146cae8 and 8bfa3db
on remote. Everything pushed. Untracked DESIGN_TODOS/ — will
commit as docs later.

Disk hover: description looks right. I'll verify visually in
the browser. If delta is weak, I'll send a screen recording /
specific numbers.

Hero — proceed:

1. FIX object-position + height NOW (mandatory):
   - object-position: center top
   - height: min(60vh, 520px)
   Rationale: wordmark currently cut at the top by navbar;
   fix makes the full hero visible from the navbar's bottom
   edge.

2. WORDMARK DUPLICATION:
   I'll export a new PNG without «EVENT HORIZON» (planet +
   glow only) — call it frontend/public/images/brand/
   hero-planet.png (or .webp). When ready, switch hero to it
   and keep the disk with the wordmark.
   Until then — keep the current image, just apply fix (1).

3. Do NOT touch:
   Nav, games grid, Wave 2 #6, seed / refund / authors.

After fix (1), send a new screenshot at 1920×1080.
No push until I review.

Also: I'd like to move on to Wave 2 #6 (game polish) after
this — Flappy first. Plan first, no code.
Что я держу в карте
Закрыто (все на GitHub):

v1.0.9.

Wave 1 — 6/6.

Wave 2 #1–4:

#1 0ee567f thin DTO

#2 a881587 DB cleanup

#3 8c16926 lamps as boosts

#4 2bbc62f Flappy levels

Wave 2 #5 — 146cae8 achievements.

Hero + disk hover — 8bfa3db.

Сейчас (в работе):

🟧 Hero fix — object-position + height.

🟧 Wordmark duplication — экспорт нового PNG (или отложить).

🟧 Disk hover — verify в браузере.

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR), с 6.1 Flappy.

Отложено:

🟧 Wave 2 #5b — shop/boost achievements.

🟧 Refund window implementation.

🟧 backfill-noiz-reviews.py.

🟧 DESIGN_TODOS/ — коммит или gitignore.

Действия
Ответь Cursor'у — по промпту выше.

Фикс (1) — OK.

Дубль — A (экспортируешь PNG).

Или B — если не хочешь возиться.

После фикса — verify hero + disk hover визуально.

Push — когда ок.

Wave 2 #6 — Flappy polish.

По PNG:

Photoshop / GIMP / Figma — открыть logo-minimal.png, убрать текст, экспортировать как PNG / WebP.

Если нет редактора — можно использовать онлайн (Photopea, Figma — бесплатно).

5 минут работы.