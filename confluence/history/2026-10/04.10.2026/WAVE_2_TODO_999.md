OK. Реализуй
Что подтверждаю
Арт: logo-minimal.png — используем как есть (1238×1200). Никаких Midjourney.

Layout:

Navbar

HERO — full-bleed, min(70vh, 640px), gradient overlays.

Scroll hint (chevron) — опционально. Мой выбор — да, оставь.

Выбери игру + accretion disk — как сейчас.

Games grid

Footer

Hero: object-fit: cover + object-position: center 60% (на горизонт/glow). Void letterboxing — ок на ultra-wide.

No text on banner. Лого уже содержит бренд. Согласен.

Disk hover (CSS only):

Center scale 1 → 1.12

Rings 1 → 0.92, spin 140s → 80s

Glow up

Leave 700ms ease-out

prefers-reduced-motion → no effect. Обязательно.

Files:

Home.tsx — full-bleed section + hover class.

theme.css — .eh-hero-banner + disk hover.

HeroBanner.tsx — опционально. Мой выбор — inline в Home.tsx. Если Home станет грязным — вынести.

Out of scope: games, seed, refund, authors, Wave 2 #6.

Defaults locked:

art = logo-minimal

no banner copy

CSS-only

before #6

Мой ответ Cursor'у
text
OK. Implement.

Confirmations:
  - Height: min(70vh, 640px) — OK.
  - object-position: center 60% — OK.
  - Scroll hint chevron: YES, add it. Subtle. Fades on scroll
    past hero (or just animates indefinitely).
  - HeroBanner: inline in Home.tsx for v1. Extract only if
    Home.tsx gets noisy.
  - Disk hover: use CSS variables so timing is tweakable later
    (--eh-disk-pull-duration).
  - prefers-reduced-motion: hard disable — no scale, no spin
    speedup, no glow change.

After implement:
  - Report files changed.
  - Screenshot Home at 1920×1080 for me to verify.
  - Do NOT push. I'll push after review.

DO NOT TOUCH: games, seed, refund, authors, Wave 2 #6.
Что ожидать после реализации
Verify hero-баннер
#	Проверка	Ожидание
1	Home → top	Full-bleed hero во всю ширину
2	Высота	~640px (или 70vh)
3	object-position center 60%	Горизонт/glow виден, не обрезан
4	Ultra-wide (2560px)	Тёмный letterbox по бокам, ок
5	Navbar поверх hero	Виден, читается (gradient сверху)
6	Scroll вниз	Переход к Выбери игру
7	Scroll hint chevron	Виден, анимирован
8	Мобильный (375px)	Hero адаптируется, без обрезки
Verify accretion disk
#	Проверка	Ожидание
9	Home → Выбери игру	Disk справа
10	Hover на disk	Center растёт, rings сужаются
11	Rotation ускоряется	140s → ~80s
12	Glow усиливается	Заметно
13	Mouse leave	Возврат за 700ms ease-out
14	prefers-reduced-motion	Нет эффекта (эмуляция в DevTools)
Регресс
#	Проверка	Ожидание
15	Navbar	не сломан
16	Games grid	8 карточек
17	Footer	ок
18	Скролл всей страницы	плавный
19	Wave 2 #5 — achievements	не сломаны
Что я держу в карте
Закрыто:

v1.0.9 на GitHub.

Wave 1 — 6/6.

Wave 2 #1–5:

#1 0ee567f

#2 a881587

#3 8c16926

#4 2bbc62f

#5 146cae8

Сейчас (в работе):

🟧 Hero banner + accretion hover (план approve → implement)

⏸ Rebuild profile + gateway (Wave 2 #5 verify)

⏸ Verify 23 пункта

⏸ Push

⏸ Бэкап

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR)

Отложено:

🟧 Wave 2 #5b — shop/boost achievements

🟧 Refund window implementation

🟧 backfill-noiz-reviews.py

Действия
Отправь OK Cursor'у — implement.

Пока Cursor делает — verify Wave 2 #5:

Rebuild profile + gateway.

/profile — silent seed, achievements.

Flappy L5 — flappy_level_5 unlock (проверить баг).

Push, если ок.

После hero — verify баннер + disk.

Push.

Wave 2 #6 — полиш игр.

Скриншот Home (1920×1080) — после hero — сюда.