Нашли. hide=core убирает квадрат
Твой тест
Тест	Квадрат
hide=pull	ЕСТЬ
hide=rings	ЕСТЬ
hide=core	НЕТ! ← вот он
hide=particles	ЕСТЬ
hide=glow&hide=pull	ЕСТЬ
Источник — .eh-disk-core.

Что это значит
.eh-disk-core — чёрный кружок в центре. Computed был:

black radial + large gold box-shadow

Вот оно — box-shadow. box-shadow имеет прямоугольный spread вокруг элемента. Даже если элемент круглый (border-radius: 50%), box-shadow с spread может рисовать скруглённый прямоугольник, а не круг.

Или — .eh-disk-core имеет большой box-shadow с blur, который обрезается границей контейнера → прямоугольник на краю.

Cursor писал про core:

black radial + large gold box-shadow | circular

«Circular» — но box-shadow — прямоугольный по определению. Blur сглаживает, но при большом spread + clip на родителе — углы видны.

Промпт Cursor'у
text
Found the square: it's .eh-disk-core.

Test results:
  hide=pull        → square still there
  hide=rings       → square still there
  hide=core        → SQUARE GONE
  hide=particles   → square still there
  hide=glow+pull   → square still there (you thought it
                     might be pull/rings — it's not)

.hide=core computed was: black radial-gradient + large
gold box-shadow.

The box-shadow is the culprit. Even though the element has
border-radius: 50%, a box-shadow with a large spread/blur
creates a soft-edged rectangle around the circular element.
If any parent has overflow: hidden or clip-path, the blur
is cut off at the rectangular boundary → visible square.

Fix:
  1. Replace the box-shadow on .eh-disk-core with a
     radial-gradient (background) so the glow is a true
     circle, not a box shadow.
  2. Or: reduce the box-shadow spread and increase blur so
     the edges fade to fully transparent well before the
     parent's clip boundary.
  3. Or: put the glow on a separate absolutely-positioned
     <div> with its own radial-gradient and no clip issues.

Recommended: (1) — radial-gradient on .eh-disk-core:
  background:
    radial-gradient(circle at center,
      rgba(0,0,0,1) 0%,
      rgba(0,0,0,1) 40%,
      rgba(232,184,74,0.35) 55%,
      rgba(232,184,74,0.10) 70%,
      rgba(232,184,74,0.00) 100%);
  box-shadow: none;

Then verify: hover → no square boundary anywhere.

Report the exact change to .eh-disk-core. No push.

Also: after this is clean, please proceed with the
"rings thicker" tweak (+1–2px). Same commit or separate,
your call.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

🎯 Square — найден источник: .eh-disk-core box-shadow.

Fix — radial-gradient вместо box-shadow.

🟧 Rings thicker (+1–2px) — после фикса.

🟧 PNG crop — Photopea.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

Действия
Отправь Cursor'у промпт — fix .eh-disk-core.

Verify — квадрат исчез.

Попроси — rings thicker в тот же заход.

PNG crop — Photopea.

Push.

Wave 2 #6.