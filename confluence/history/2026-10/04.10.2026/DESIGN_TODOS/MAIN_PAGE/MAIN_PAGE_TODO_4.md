Разбираю. 4 пункта
Что вижу на скрине
✅ Буквы внутри слов — вплотную, ок.

🟧 EVENT и HORIZON — разъехались по краям, гигантская дырка между ними.

🟧 EVENT HORIZON — прижато к верху (сразу под navbar).

✅ PNG — лучше, прямоугольник почти не виден.

🟧 Ниже — пустота (много).

Что ты хочешь
1. EVENT и HORIZON — ближе друг к другу
Сейчас: space-between раскидывает слова по краям.
Хочешь: центрировать связку EVENT HORIZON по экрану, с одним пробелом между словами.

Fix:

justify-content: center вместо space-between.

Пробел между словами — фиксированный margin или <span> с padding.

2. Мобильный (375×812)
Хочешь:

EVENT — сверху.

HORIZON — снизу.

Оба центрированы.

Fix:

На mobile — flex-direction: column, оба слова с центром.

3. Больше отступа от хедера
Сейчас: надпись сразу под navbar.
Хочешь: отступ больше, чтобы надпись была по центру экрана (по высоте).

Fix:

padding-top на .eh-hero-wordmark — больше.

Или — hero позиционируется flex align-items: center — надпись вертикально центрируется.

4. EVENT HORIZON — по центру экрана
Сейчас: сверху.
Хочешь: в центре экрана по горизонтали и вертикали.

Промпт Cursor'у
text
Four fixes based on screenshot.

================================================================
1. BRING EVENT AND HORIZON CLOSER
================================================================

Current: two-word flex with justify-content: space-between
pushes EVENT to the far left and HORIZON to the far right.
Huge gap in the middle.

Change:
  - .eh-hero-wordmark { justify-content: center; gap: 0.35em; }
  - Both words centered as a unit.
  - Single fixed gap between EVENT and HORIZON.
  - Do NOT let them fly to the edges.

================================================================
2. MOBILE — STACK WORDS, CENTER EACH
================================================================

@media (max-width: 640px):
  - .eh-hero-wordmark { flex-direction: column; align-items: center; gap: 0.2em; }
  - EVENT on top, HORIZON below.
  - Both horizontally centered.
  - Reduce font-size if needed so each word fits its line.

================================================================
3. MORE TOP PADDING — CENTER WORDMARK VERTICALLY
================================================================

Current: wordmark sits right under the navbar.

Goal: hero section min-height 80vh; wordmark centered
vertically (or at ~40% from top).

Approach:
  - .eh-hero { min-height: 80vh; display: flex;
    flex-direction: column; justify-content: center;
    align-items: center; padding-top: <navbar height + 40px>;
    padding-bottom: 40px; }
  - Planet as background/absolutely-positioned behind.
  - Chevron absolute at bottom.

================================================================
4. WORDMARK CENTERED ON DESKTOP (not stretched)
================================================================

Since you now want EVENT HORIZON centered and close
together (not stretched to shell edges), remove the flex
space-between approach entirely.

New target:
  - EVENT HORIZON centered horizontally in the viewport.
  - Single gap between words.
  - Font-size: clamp(3rem, 8vw, 7rem) — large, but centered.
  - E/N flush is no longer the goal; centering is.

Confirm this is what I want — if I misunderstood, say so.

================================================================
PROCESS
================================================================
  - Implement 1+2+3+4 in one pass.
  - Screenshot 1920×1080 and 375×812.
  - No push until I OK.

DO NOT TOUCH: nav, disk, games, seed, refund, authors.
Что проверить после
#	Проверка	Ожидание
1	1920×1080 — EVENT HORIZON	центр, слова близко
2	Пробел между EVENT / HORIZON	один, равномерный
3	Отступ сверху	есть (не липнет к navbar)
4	Вертикальный центр	надпись по центру экрана
5	PNG	без прямоугольника
6	Chevron	внизу
7	Air до Выбери	есть
8	375×812 — EVENT	сверху, центр
9	375×812 — HORIZON	снизу, центр
10	Zoom Ctrl+	резко
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

Сейчас:

🟧 Wordmark — центр связки, убрать space-between.

🟧 Mobile — EVENT / HORIZON стек.

🟧 Padding сверху — больше.

🟧 PNG — ок (можно отшлифовать в Photopea позже).

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

Действия
Отправь Cursor'у промпт — 4 задачи.

Проверь — центр, близко, отступ, mobile стек.

Скинь — скрин 1920 + 375.

Push после OK.

Wave 2 #6.