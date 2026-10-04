План
Шаг 1 — Commit — зафиксировать текущий VOID v3 (работает, квадрат есть, но всё остальное ок).

Шаг 2 — Layer-by-layer toggle — Cursor добавляет URL-параметры для отключения каждого слоя:

?no-glow=1 — без .eh-disk-glow.

?no-pull=1 — без .eh-disk-pull.

?no-rings=1 — без .eh-disk-rings.

?no-core=1 — без .eh-disk-core.

?no-particles=1 — без svg.particles.

?no-parent-bg=1 — убрать bg-void с родителей.

И так далее — до <body>.

Шаг 3 — Ты открываешь по одному URL, смотришь, пропал ли квадрат. Так находим точный источник.

Шаг 4 — Fix — точечно.

Промпт Cursor'у
text
Before layer-by-layer elimination, commit the current
state. Then add dev-only toggles.

================================================================
1. COMMIT
================================================================

Commit everything VOID-related currently uncommitted on
top of 6722de1.

Subject: feat(home): void v3 — equator pull, stars, debug
overlays, reduced-motion respect

Body:
  - Rings flatten to equator on hover (--ry 0.4 → 0.06).
  - rotateX(52deg) removed from .eh-disk-pull (was creating
    a visible square boundary).
  - Stars: 4-point for colored particles, tiny gold = dots.
  - prefers-reduced-motion: static (verified).
  - ?motion=force — dev override to ignore reduced-motion.
  - ?debug-void=1 — dev outlines on .eh-disk-* layers.

Do NOT push. Just commit so we have a rollback point.
Report the new HEAD hash.

================================================================
2. LAYER-TOGGLE URL PARAMS (dev only)
================================================================

Add URL params that HIDE individual layers, so I can find
which one shows the square:

  ?hide=glow        → .eh-disk-glow display: none
  ?hide=pull        → .eh-disk-pull display: none
  ?hide=rings       → .eh-disk-rings display: none
  ?hide=core        → .eh-disk-core display: none
  ?hide=particles   → svg.particles display: none
  ?hide=disk        → whole .eh-disk display: none
  ?hide=bg-nebula   → remove any bg-nebula / bg-void from
                      wrappers around the hero + disk
  ?hide=hero        → .eh-hero-banner display: none
  ?hide=glow-filter → remove filter / backdrop-filter from
                      .eh-disk-glow

Rules:
  - Only active in import.meta.env.DEV.
  - Multiple ?hide= can be combined.
  - The point: find which single removal makes the square
    disappear.

Keep ?debug-void=1 as-is — it still helps.

================================================================
PROCESS
================================================================
  - Commit first. Report HEAD hash.
  - Then add the toggles.
  - Report which URL to open for each layer.
  - No push.

DO NOT TOUCH: hero layout, nav, games, seed, refund,
authors.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

🟧 Commit VOID v3 (local, точка отката).

🟧 Layer-toggle — ?hide=glow|pull|rings|core|particles|....

🟧 Найти источник квадрата.

🟧 Rings thicker — после square.

🟧 PNG crop — Photopea.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

Действия
Отправь Cursor'у промпт — commit → layer toggles.

После toggle'ов — открываешь по URL — смотришь.

Как найдёшь слой — fix.

Rings thicker + PNG crop.

Push.

Wave 2 #6.

Хорошая идея. Метод исключения — быстрее всего найдёт источник.