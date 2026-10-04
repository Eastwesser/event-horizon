Понял. Планета должна быть ПОД надписью
Что вижу
Хорошо:

✅ EVENT HORIZON — одна связка, центрирована.

✅ Буквы — вплотную.

✅ Hero — 80vh, вертикально центрировано.

✅ Chevron — внизу.

✅ Mobile — стек, центр.

Проблема:

🟧 Планета СЗАДИ надписи — надпись перекрывает планету, накладывается на горизонт.

Планета должна быть ПОД надписью (ниже по вертикали), не за ней.

Что ты хочешь
Порядок сверху вниз:

Navbar.

EVENT HORIZON — центр.

Планета — под надписью, свой слой.

Chevron — внизу hero.

Сейчас: планета — absolute за надписью, пересекаются.

Промпт Cursor'у
text
Screenshot: planet is BEHIND the wordmark — they overlap.
The planet should sit BELOW «EVENT HORIZON», not behind it.

Target layout (top to bottom):
  1. navbar
  2. EVENT HORIZON (centered wordmark)
  3. planet PNG (below the wordmark, its own block)
  4. chevron at the bottom of hero

Fix:
  - Remove the planet from being absolutely positioned
    behind the wordmark.
  - Structure .eh-hero as a flex column:
      <div class="eh-hero-inner">
        <h1 class="eh-hero-wordmark">...</h1>
        <img class="eh-hero-planet" src="/...planet-hero.png">
      </div>
  - .eh-hero-inner { display: flex; flex-direction: column;
    align-items: center; gap: ~40-60px; }
  - .eh-hero-planet { width: 100%; max-width: 1000px;
    height: auto; display: block; }
  - Keep top/bottom fades and radial mask on the planet.
  - Chevron absolutely at the bottom of .eh-hero.

Verify 1920×1080: wordmark centered on top, planet cleanly
below, no overlap.

Mobile 375×812: stacked wordmark, planet below, still no
overlap.

No push until I OK.
Что проверить **после
#	Проверка	Ожидание
1	1920 — EVENT HORIZON	сверху
2	1920 — планета	под надписью, без overlap
3	Пробел между	~40-60px
4	Планета целиком	видна
5	Fades	плавные
6	Chevron	внизу
7	375 — EVENT HORIZON	сверху
8	375 — планета	под надписью
9	Mobile	без overlap
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

VOID core fix — 5a18d99.

VOID v3 — 2abb45d / ebf6a92.

Hero baseline — 8bfa3db.

Локально (не запушено):

1694cc0 — MAINPAGE_TODO_1.

dc6869e — Hero Plan A.

6a1e574 — Hero (mask + wordmark + air).

7674c51 — Hero (центр + mobile).

Сейчас:

🟧 Планета — под надписью (не за ней).

🟧 Verify.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

Действия
Отправь Cursor'у промпт — планета под надписью.

Проверь — 9 пунктов.

Скинь скрин 1920.

Push после OK.

Wave 2 #6.