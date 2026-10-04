Что сейчас
Базовый pull-in:

Core scale 1 → 1.05 (мягко).

Rings scale 1 → 0.85, spin 140s → 80s.

8 SVG particles → к центру (~1.5s).

Glow gold + cyan.

Leave 700ms.

Ты говоришь: ну... ок. Но можно лучше.

Согласен. Текущий эффект — минимальный. Для «залипалова» нужно больше визуальных слоёв.

Что делает VOID «black hole bangер»
1. Аккреционный диск (сейчас — просто кольца)
Добавить:

Несколько орбит с разными скоростями (уже 2-3 — увеличить до 4-5).

Каждая орбита — не идеальный круг, а спираль (немного искажена).

Орбиты — разные цвета (золото / cyan / фиолет).

При hover — орбиты сужаются по спирали к центру, не просто scale.

2. Particles — настоящее всасывание
Сейчас: 8 частиц летят прямолинейно.

Что нужно:

20-40 частиц вместо 8.

Спиральные траектории (не прямые).

Разные размеры (мелкие + крупные).

Разные скорости.

Fade out при приближении к центру.

Respawn с внешнего края.

Технически:

SVG — сложно с 20-40 частицами + спирали.

CSS — ещё сложнее.

Canvas — правильный выбор. requestAnimationFrame, простые частицы.

WebGL / Three.js — overkill для 2D.

3. Дополнительные эффекты
Gravitational lensing — искажение фона за диском (сложно, WebGL).

Струи (jets) — при высокой энергии — 2 тонких луча из полюсов. CSS / SVG — реально.

Рябь / волны — искажение пространства рядом.

Цветовая температура — от холодного к горячему при усилении.

Light bloom — свечение через CSS filter (blur + opacity).

Спиральные хвосты — траектории уходящих частиц.

4. Интерактивность
Hover → pull-in.

Долгий hover (2-3 сек) → усиление (эскалация).

Клик → вспышка / pulse.

Движение мыши рядом → слабое влияние (без hover).

Реалистичный план — 3 уровня
Уровень 1 — SVG particles × 30 (S, 1-2 дня)
Что:

Заменить 8 частиц на 30.

Спирали — анимация translate + rotate.

Разные размеры / скорости.

Fade out у центра.

Остаётся SVG.
Estimate: 1-2 дня.

Уровень 2 — Canvas particles (M, 3-5 дней)
Что:

Canvas overlay на диске.

Particles с настоящей физикой (притяжение, спирали).

~50-100 частиц.

requestAnimationFrame.

Плавное взаимодействие с курсором.

Estimate: 3-5 дней.

Уровень 3 — WebGL / Three.js (L, 1-2 недели)
Что:

Настоящий gravitational lensing (искажение фона).

Струи (jets).

3D-аккреционный диск.

Реальный black hole.

Estimate: 1-2 недели. Круто, но тяжело.

Мой выбор
Для v1 — Уровень 1 (SVG × 30).

Причины:

Быстро (1-2 дня).

Без новых зависимостей.

Визуально — сильное улучшение.

Для залипалова — достаточно.

Позже — Уровень 2 (Canvas), если хочется круче.

Уровень 3 — если проект станет коммерческим.

Промпт Cursor'у
text
VOID is currently OK but not a banger. Upgrade to level 1
(SVG, no Canvas yet). Plan first, no code.

Goal: player hovers the void and wants to keep hovering —
«black hole pull-in» that reads clearly at 1920×1080.

Spec:

1. RINGS (accretion disk)
   - Increase from 2-3 rings to 4-5 rings.
   - Each ring: different speed (slow outer → fast inner),
     different opacity, slight ellipse (not perfect circle).
   - Colors: gold / cyan / violet mix (from theme tokens).
   - On :hover → rings spiral inward (scale 1 → 0.75) over
     1.5s, then reverse on leave (700ms).

2. PARTICLES (30 instead of 8)
   - Sizes: small (1px), medium (2px), large (3px) mix.
   - Colors: gold, cyan, white — vary.
   - Trajectories: NOT straight. Use translate + rotate to
     create spiral paths.
   - Speed: vary per particle (0.8s to 2.5s loop).
   - Fade to 0 opacity as they approach center.
   - Respawn at outer edge (no pop).

3. GLOW
   - Outer glow: gold ring.
   - Inner glow: cyan haze.
   - On :hover → intensify both (scale up + opacity up).
   - Use filter: blur or box-shadow, whichever is cleaner.

4. POINTER NEAR (optional, small)
   - When cursor enters ~80px radius around the void (not
     exactly hover) → subtle pre-pull: rings brighten a
     little, particles start drifting slightly.
   - When cursor leaves → back to idle.
   - This makes it feel «alive» even before full hover.

5. ESCALATION (optional, small)
   - If hover lasts > 2s → particles speed up ×1.5, glow
     intensifies more.
   - Caps at 3s.

CONSTRAINTS
  - SVG + CSS only. No Canvas / WebGL for this pass.
  - Keep it performant — target 60fps at 1920×1080.
  - prefers-reduced-motion: fully static.
  - No logo inside the void (as-is).

PROCESS
  - Show me the plan: which files, which SVG structure.
  - Wait for my OK before implementing.
  - One screenshot at 1920×1080 with hover active, then text
    description.
  - No push until I OK.

DO NOT TOUCH: hero, nav, games, seed, refund, authors,
Wave 2 #6.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Сейчас:

✅ Hero — работает.

✅ Smooth scroll — работает.

🟧 Disk pull-in (v1) — сделано, но не banger.

🟧 Disk v2 (SVG upgrade) — предложено.

⏸ Wordmark flush — ждёт твой crop PNG (Photopea).

⏸ Push после OK.

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR).

Отложено:

Wave 2 #5b, Refund window, backfill-noiz-reviews.py, DESIGN_TODOS/.