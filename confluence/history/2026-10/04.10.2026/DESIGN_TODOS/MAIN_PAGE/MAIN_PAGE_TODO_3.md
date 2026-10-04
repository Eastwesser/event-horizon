Разбор + твои задачи + Paint
Что вижу на скринах
Скрин 1 (1920×1080)
EVENT HORIZON — растянуто, E flush лево, N flush право. ✅

Планета — видна.

Не видно прямоугольник PNG по бокам (радиальный маск работает).

Тесно — между буквами и планетой, между планетой и Выбери — мало воздуха.

Диск — справа, ок.

Скрин 2 (375×812 — мобильный)
EVENT HORIZON — маленькое, не растянуто.

Планета — маленькая, в середине.

Много пустого места сверху.

Скрин 3
Выбери игру + диск — ок.

Что ты хочешь
1. PNG планеты — видно, что вставлена. Можно вырезать и посадить на сайт?
Про «вырезать и посадить на сайт»:

Что уже сделано:

PNG вырезан из logo-minimal — planet-hero.png.

HTML — <h1> + <img>.

Что ты видишь как «вставлено»:

Фоновый оттенок PNG ≠ точно page-void.

Разница ~4% (Cursor отрепортил: (0,1,19) vs (11,14,26)).

Из-за этого планета «сияет» на чуть другом фоне — видно край.

Можно ли исправить на уровне CSS?

Да — усилить mask.

Или — заменить фон PNG на точный #0b0e1a (Photopea).

Или — mix-blend-mode: screen / lighten на <img> — сливается с фоном.

2. Буквы — разное расстояние
E V E — далеко друг от друга.
R I Z — I визуально далеко от соседей.
HORIZON — сжать вправо по правому бордеру.

Причина: flex space-between даёт равные промежутки между <span>-буквами**, не учитывая визуальную ширину буквы. I — узкая, R, Z — широкие. Промежуток одинаковый, но визуально неравный.

Fix — сгруппировать буквы с одинаковыми промежутками:

Между E, V, E — letter-spacing: 0 (буквы вплотную).

Между словами — spacer.

HORIZON — сжать справа (не space-between, а фиксированная ширина).

Другой вариант: font-stretch или letter-spacing с оптимизацией под глифы.

3. Тесно — нужен воздух по расположению из Paint
Твой Paint — скрин 2 из второго сообщения. Ты хочешь:

EVENT HORIZON — верх.

Стрелка ˅ под ними.

Выбери игру + диск — ниже.

Юзер скроллит → видит hero.

Клик на стрелку → Выбери игру (smooth scroll).

Или скролл вниз → то же.

Проблема сейчас: hero слишком низкий (min(60vh, 520px)), планета упирается в Выбери, нет воздуха.

Промпт Cursor'у — 3 задачи
text
Three fixes based on the current hero.

================================================================
1. PLANET PNG — reduce visible «insertion»
================================================================

Corner colors: PNG ~(0,1,19) vs page-void #0b0e1a (11,14,26).
Delta ~4% per channel — enough to show the PNG boundary.

Options (try in order):
  a) Stronger radial mask on .eh-hero-planet — expand
     transparent stop further:
       mask-image: radial-gradient(
         ellipse 65% 80% at center,
         black 45%,
         transparent 95%
       );
  b) Add mix-blend-mode: screen or lighten on the planet img
     so its dark background blends into the page void.
  c) If (a) and (b) still leave the edge visible, I'll
     re-export planet-hero.png in Photopea with the exact
     page-void color as its background — tell me the exact
     hex of page-void.

Report which you pick and why. I'll verify visually.

================================================================
2. WORDMARK — letter spacing is visually uneven
================================================================

Current: flex space-between on individual letter spans →
equal flex gaps → I in HORIZON looks isolated, RIZ too loose,
E V E too spread.

Goal: tight, poster-like wordmark. E flush left, N flush
right, but letters within each word tight, with one clear
gap between EVENT and HORIZON.

Approach:
  - Drop space-between on letters.
  - Wrap each word in a .eh-hero-wordmark-word:
      <span class="word">EVENT</span>
      <span class="gap" aria-hidden="true"></span>
      <span class="word">HORIZON</span>
  - .eh-hero-wordmark { display: flex; justify-content:
    space-between; } — the two WORD spans are the direct
    children, so space-between distributes between words,
    not letters.
  - Letter spans inside each word: no gap, natural kerning.
  - If HORIZON needs to shrink toward right, adjust
    letter-spacing or font-stretch — report if needed.

Keep aria-label="EVENT HORIZON".

================================================================
3. HERO LAYOUT — more air (per my Paint)
================================================================

Goal (from my sketch):
  Row 1: EVENT HORIZON (stretched wordmark)
  Row 2: big gap
  Row 3: chevron ˅
  Row 4: another gap
  Row 5: «Выбери игру…» + disk below the fold

Fix:
  - Give .eh-hero-wordmark margin-bottom (~8-10vh).
  - Move the chevron OUT of hero — or push it to the very
    bottom of .eh-hero, separated from the wordmark.
  - Add margin/padding between .eh-hero (with planet) and
    the .eh-choose section (which has «Выбери игру» + disk).
  - The planet PNG should sit under the wordmark, with its
    own space above «Выбери игру» — no touching.
  - Target: hero fills ~70vh, chevron at bottom edge,
    then a full ~10-15vh breathing room, then «Выбери».

No push. Verify at 1920×1080 and 375×812.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

VOID core fix — 5a18d99.

VOID v3 — 2abb45d / ebf6a92.

Hero baseline — 8bfa3db.

Локально (не запушено):

1694cc0 — MAINPAGE_TODO_1.

dc6869e — Hero Plan A refactor.

Сейчас:

🟧 Planet PNG — уменьшить «вставленность».

🟧 Wordmark — буквы вплотную, пробел только между словами.

🟧 Hero layout — воздух, chevron, отступы (по Paint).

🟧 Verify + commit.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

Действия
Отправь Cursor'у промпт — 3 задачи.

Проверь — PNG сливается? буквы вплотную? воздух есть?

Скинь скрин.

Push после OK.

Wave 2 #6.