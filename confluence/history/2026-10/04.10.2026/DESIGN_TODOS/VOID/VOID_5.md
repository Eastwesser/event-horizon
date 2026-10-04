1. Анимация повторяется — только 1 раз
Ты диагностировал: hover-эффект проигрывается, но один раз, не зацикливается, пока курсор наведён.

Причина: animation: ... forwards (или без iteration-count: infinite) — играет 1 раз и останавливается на последнем кадре.

Фикс: animation-iteration-count: infinite, пока :hover активен. Или — перезапуск на pointerenter, сброс на pointerleave.

Промпт Cursor'у:

text
Hover animations currently play once. They must loop while
the pointer is on the void.

Change:
  - .eh-disk--pull animations: animation-iteration-count:
    infinite (or a sufficiently large count).
  - On pointerleave: animations reset (class removed).
  - On pointerenter again: restart from frame 0.

If some animation should play once AND THEN loop (e.g. equator
alignment once, then continuous glow pulse) — split into two:
  a) align keyframes: 1 iteration
  b) pulse/particle-drift: infinite
And chain via animation-delay.
2. Звёздочки — все 4-конечные
Ты сказал: 4-конечные приятнее. 5-конечные → 4-конечные.

Промпт Cursor'у:

text
Replace all 5-point stars with 4-point stars.
Keep: tiny gold particles as plain dots.
Result: only two particle types — 4-point stars (colored)
and small gold dots.
3. VOID v3 — проверить
Cursor заявил:

Стars — cyan 4-point, gold/white 5-point (→ заменить на 4).

Equator pull — rings flatten (--ry → 0.06, tilt → 0) + rotateX(52°), spin pauses, glow up, particles spiral after ~0.95s.

Leave 700ms.

Cursor pull-in — skipped (c).

Коммит 6722de1 — local, не запушен.

Что проверить в браузере:

Наведи на VOID → rings flatten в эллипс (экватор)?

Spin pauses?

Glow усиливается?

Particles spiral в центр после ~1s?

Всё это — циклично, пока курсор наведён?

Уход → обратно 700ms?

Промпт Cursor'у — общий
text
Three fixes on top of 6722de1:

1. LOOP THE HOVER ANIMATIONS
   Currently play once. Make them loop while :hover / --pull
   is active.
   - Infinite iteration-count OR restart on pointerenter.
   - If equator-alignment should play once then hold: split
     into two animation blocks (align 1× + pulse ∞).
   - Reset on pointerleave.

2. STARS — ALL 4-POINT
   Remove 5-point stars. All colored particles = 4-point stars
   (cyan, gold, white). Tiny gold = dots.

3. VERIFY IN BROWSER (not screenshot)
   Tell me what you see on hover:
     - rings flatten to an equatorial ellipse?
     - spin pauses?
     - glow up?
     - particles spiral to core after ~1s?
     - all of the above loop while hovering?
   If not — report which is missing.

DO NOT TOUCH: hero, nav, games, seed, refund, authors.
No push until I OK.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

🟧 VOID v3 — 6722de1 (local).

🔴 Loop hover animations (сейчас 1×).

🟧 Все звёздочки — 4-конечные.

🟧 Verify в браузере.

🟧 Wordmark flush — ждёт crop PNG (Photopea).

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in — stretch.

Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Действия
Отправь Cursor'у промпт — loop + 4-point stars + verify.

Проверь в браузере (не по скрину).

Параллельно — crop PNG в Photopea.

Push после OK.

Wave 2 #6.

Про 4-конечные звёздочки — согласен. 4-point — более «космично», менее «ёлочно».