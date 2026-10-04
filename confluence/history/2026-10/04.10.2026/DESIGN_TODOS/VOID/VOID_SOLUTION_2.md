Что вижу
Анимация работает — кольца, частицы, звёздочки.

Проблема: виден квадратный контур — как будто золотой квадрат с границами за диском.

Причина: .eh-disk-glow — элемент с box-shadow или filter: blur. Blur вырезает по границе контейнера → мягкий край становится резким на краю → виден квадрат.

Технически:

Glow рисуется внутри прямоугольного div.

filter: blur(...) — размывает внутри границ.

За границей — ничего. Резкий разрыв.

Результат — квадратный контур.

Fix — 3 варианта
A. radial-gradient вместо box-shadow
Проще всего:

css
.eh-disk-glow {
  background: radial-gradient(
    circle at center,
    rgba(232, 184, 74, 0.35) 0%,
    rgba(232, 184, 74, 0.15) 30%,
    transparent 70%
  );
  /* убрать filter: blur и box-shadow */
}
radial-gradient — круглый по определению. Не даёт квадрата.

B. Расширить контейнер + mask-image
Или:

css
.eh-disk-glow {
  /* оставить blur, но добавить плавную маску по краям */
  mask-image: radial-gradient(
    circle at center,
    black 0%,
    black 60%,
    transparent 100%
  );
}
Тогда blur плавно уходит в прозрачность до края div'а.

C. Увеличить padding контейнера
Если blur(40px) — нужно padding ≥ 2× blur радиус.

css
.eh-disk-glow {
  inset: -80px; /* вместо inset: 0 */
  filter: blur(40px);
}
Тогда размытие закончится внутри контейнера.

Мой выбор — A + C (комбо)
A — radial-gradient для основного glow.
C — inset: -80px если нужен blur.

Промпт Cursor'у:

text
Animation works now. Remaining issue: a square outline
around the disk glow is visible during hover.

Cause: .eh-disk-glow uses box-shadow / filter: blur on a
rectangular container. The blur is clipped at the container
edge → hard square boundary.

Fix:
  1. Replace box-shadow with radial-gradient:
     background: radial-gradient(circle at center,
       rgba(232, 184, 74, 0.35) 0%,
       rgba(232, 184, 74, 0.15) 30%,
       transparent 70%);
     Round shape, no square.
  2. If filter: blur is still used anywhere on the glow,
     add mask-image: radial-gradient(circle at center,
       black 0%, black 60%, transparent 100%);
     to fade the blur to transparent before the container
     edge.
  3. Alternatively, expand the glow container: inset: -80px
     so blur has room to fade inside.

Verify in browser: hover → no visible square boundary, only
soft circular glow.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

✅ VOID v3 — работает (после prefers-reduced-motion fix).

🔴 Glow square — fix (radial-gradient / mask).

🟧 Wordmark flush — Photopea crop.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

Действия