Разбор — три задачи. Идём по шагам
Что вижу на скрине 2 (близко)
VOID v2 — красиво:

5 колец — золото, cyan, фиолет. Эллипсы, не идеальные круги. Аккреционный диск.

Центр — чёрный, без логотипа.

~30 частиц — точки (разного цвета и размера).

Glow — золото + cyan.

Ты говоришь: я вижу круг в разрезе с диаметром, а хочу сферу с экватором.

Понял. Разница:

Сейчас: плоский круг с кольцами по обе стороны.

Хочешь: сфера — диск виден как эллипс (сплюснутая орбита) — «экватор» = горизонтальная линия через центр.

Дополнительно:

Частицы → звёздочки (не точки).

Голубые — 4 зубца.

Остальные — 5 зубцов.

Мельчайшие золотые — оставить как есть.

При hover:

Кольца плавно выравниваются по горизонтали (формируется экватор).

Останавливаются на экваторе, дают свет.

Затем частицы всасываются.

Курсор — визуально затягивается и искажается к центру.

Hover сейчас не работает — ты говоришь, только подсветка.

По задачам
1. Commit — зафиксировать v2
Просто. Пока v2 не идеален — но лучше, чем ничего. Коммит = точка отката.

Промпт Cursor'у:

text
Commit VOID v2 as-is before further experiments.

Subject: feat(home): void disk v2 — 5 rings, 30 particles,
near/hot states

Body:
  - Rings: 5 ellipses, gold/cyan/violet.
  - Particles: 30 SVG circles, spiral drift.
  - States: eh-disk--near (~80px), eh-disk--hot (>2s hover).
  - prefers-reduced-motion: static.

Do NOT push. Just commit locally so we can iterate on top.
После — git log — увидим hash. Не push пока.

2. Частицы → звёздочки
Промпт Cursor'у:

text
Particles → stars, not dots.

Change each .eh-disk-particle from <circle> to a small <path>
star shape:
  - Cyan particles: 4-point star.
  - Gold + white (medium): 5-point star.
  - Tiny gold particles (< 1.5px): keep as circle (dots).

Approach:
  - Define an SVG <defs><symbol id="star4"> and id="star5">.
  - Use <use href="#star4" /> inside <svg class="eh-disk-particles">.
  - Size via font-size / width, color via fill.
  - Rotation for star5 maybe 36° on some to vary.

No extra dependencies. Pure SVG.
3. Hover-анимация — выравнивание по экватору
Ты хочешь:

При hover — кольца плавно выравниваются по горизонтали (становятся эллипсами с очень маленькой вертикальной полуосью — визуально плоский диск на экваторе).

Останавливаются в этом положении.

Свет усиливается.

Затем частицы всасываются.

Дополнительно: сам hover сейчас не работает — только brighten.

Промпт Cursor'у:

text
Two hover fixes:

A) The pull-in animation currently reads as «just brighter».
   Debug why rings don't visibly scale/spin on hover.
   Possible causes:
     - .eh-disk--pull class not applied on pointer enter
     - animation-duration overridden
     - transform overridden by another rule
     - prefers-reduced-motion is active (check DevTools)
   Confirm with a console log on pointer enter: which class
   is applied, which animation runs.

B) Change the hover effect itself to «equator alignment»:
   Instead of scale 0.75, do:
     1. Rings rotate around the central axis so their
        vertical semi-axis collapses toward the equator:
        from ellipse (ry = rx * 0.35) → (ry = rx * 0.05).
        In CSS this is tricky on <ellipse>; easier:
        rotate3d(1,0,0, 75deg) on the ring group → makes
        them look like they lie flat on an equatorial plane.
     2. Once flat, stop rotation (hold for ~200ms).
     3. Glow ramps up.
     4. Then particles accelerate into the core.
   Duration: ~1.8s to full state. Reverse 700ms on leave.

   Visual target: you should see a flat disc lying on the
   horizontal equator of a sphere, with a bright ring at
   the centre.

C) Cursor pull-in visual
   On :hover, the mouse cursor visually gets «pulled» toward
   the center. Options (pick one, report):
     a) CSS cursor: none + custom ::after dot that animates
        toward center. Fiddly.
     b) A small SVG overlay at the cursor position, dragged
        toward the center by JS on pointer move. 1–2 days.
     c) Skip for v1; note as a stretch goal.
   My recommendation: (c) for now. The ring alignment +
   particle pull-in is already a strong effect.

prefers-reduced-motion: static (all of the above off).

Do the star change AND debug the hover first. Report findings.

Then decide on (C) after I see the hover work.
Мой комментарий
«Затягивание курсора» (C) — реально сложно:

cursor: none — скрывает системный курсор.

Кастомный — рисуем свой.

Притягивание — JS на pointermove, интерполяция к центру.

2 дня минимум.

Для v1 — пропустить. Сфокусироваться на кольцах + частицах. Курсор — позже.

Если хочешь — можно сделать без cursor: none:

Только «магнитное» поведение: частицы реагируют на курсор рядом.

Курсор остаётся обычным.

Меньше работы, безопаснее UX.

Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

🟧 VOID v2 — commit → звёздочки → hover-фикс.

🟧 Wordmark flush — ждёт crop PNG (Photopea).

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in — stretch.

Wave 2 #5b, Refund window, backfill-noiz-reviews.py.