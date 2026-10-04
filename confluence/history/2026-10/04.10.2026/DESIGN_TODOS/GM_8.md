1. Стрелка вниз — сделать плавный переход
Ты прав. Сейчас клик на ˅ — резкий jump к #choose.

Фикс — CSS scroll-behavior: smooth:

css
html {
  scroll-behavior: smooth;
}
Плюс — уточнить #choose:

Не scrollTop мгновенно, а плавно.

Учесть navbar — scroll-margin-top: 80px на #choose, чтобы заголовок не уезжал под navbar.

Или — JS на клик (если CSS недостаточно):

js
el.scrollIntoView({ behavior: 'smooth', block: 'start' });
Промпт Cursor'у:

text
Smooth scroll for the down-chevron.

Option 1 (CSS):
  html { scroll-behavior: smooth; }
  #choose { scroll-margin-top: 80px; } /* avoid navbar overlap */

Option 2 (JS):
  onClick → el.scrollIntoView({ behavior: 'smooth', block: 'start' })

Either is fine. Test: click chevron → smooth scroll to «Выбери игру».
prefers-reduced-motion: instant jump (no smooth).
2. Анимации у VOID — до сих пор нет
Cursor пишет:

No disk.

Т.е. — он ждёт твоего OK на (3).

Промпт Cursor'у — сейчас:

text
Now (3) — disk pull-in on hover.

Current: disk right of «Выбери игру» is a static void — rings,
glow, particles, no logo.

On :hover:
  - Core scale 1 → 1.05
  - 2–3 rings: scale 1 → 0.85, opacity up, spin 140s → 80s
  - 6–10 SVG particle circles drift inward toward the core
    over 1.5s, loop
  - Glow intensifies (cyan/gold haze)
On mouse leave: reverse 700ms ease-out.

prefers-reduced-motion: disabled entirely (static disk).

CSS + SVG only — no Canvas/WebGL for v1.
No logo inside.

DO NOT TOUCH: hero, nav, games, seed, refund, authors.
Если Cursor сделает — проверить:

Наведи на диск → видно?

Rings сужаются?

Particles летят к центру?

Glow усиливается?

Уход → обратно 700ms?

prefers-reduced-motion → статика?

3. Wordmark ↔ text flush
Cursor спрашивает:

Wordmark↔text flush still needs your call (a) crop PNG or (b) HTML <h1>.

Ты сказал ранее: буквы должны совпадать по вертикальным границам с Выбери игру.

Сейчас: PNG 1024×993 — буквы внутри с padding. При width 100% буквы уже, чем shell.

Опции:

A. Crop PNG (Photopea):

Обрежь PNG по границам букв (слева/справа).

Тогда при width=100% буквы = shell edges.

5 минут.

B. HTML <h1> + PNG без букв:

PNG — только планета.

HTML EVENT HORIZON — контролируется CSS.

Нужен новый PNG (без букв).

Мой выбор — A. Быстро, без нового экспорта.

Промпт Cursor'у:

text
Wordmark↔text flush: choose (a) crop PNG.

I'll crop logo-minimal.png in Photopea so the wordmark is
flush with the left and right edges of the image. Then at
width: 100% of the shell, wordmark edges match the «Выбери
игру» text block edges.

Wait for my updated PNG, then just leave width: 100% as is.
No CSS change needed.
Тебе: открой logo-minimal.png в Photopea → crop слева/справа по границам букв → экспорт → replace файл. 5 минут.

Что делать сейчас — 3 задачи
Порядок:

Smooth scroll — промпт выше. Быстро.

Disk pull-in — промпт выше. Основная задача.

Crop PNG (Photopea) — 5 минут → Cursor подхватит.

После:

Push — Wave 2 почти закрыт.

Wave 2 #6 — полиш игр.

Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Сейчас:

✅ Hero — работает (картинка целиком).

🟧 Smooth scroll — стрелка вниз.

🟧 Disk pull-in — анимация VOID.

🟧 Wordmark flush — crop PNG.

⏸ Push после OK.

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR).

Отложено:

Wave 2 #5b, Refund window, backfill-noiz-reviews.py, DESIGN_TODOS/.