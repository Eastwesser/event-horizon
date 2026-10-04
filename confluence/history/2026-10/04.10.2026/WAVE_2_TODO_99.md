Wave 2 #5 готов. Verify + следующий шаг
Что сделано — 146cae8
Checklist: Wave 2 → 5/6. ✅

Уточнения применены:

icon = SVG IconName (trophy, bird, star, …) + FE fallback на trophy.

RU titles / descriptions в seed.

Toast (a): eh_seen_achievements пустой → silent seed; new unlocks → toast через eh:achievement.

total_score = сумма best_scores (уже в Profile NATS consumer).

Backend:

achievements + user_achievements.

Unlock on score.updated (+ GetProfile backfill для score-based).

flappy_level_* — нужен live level в event.

Redeploy profile (+ gateway / FE) для migration.

No push.

Один момент — flappy_level_* needs live level
Cursor пишет:

flappy_level_ needs live level on the event.*

Что это значит:

Wave 2 #4 добавил level в SubmitScore.

Но score.updated event — может не нести level?

flappy_level_5 / flappy_level_10 — не unlock'нутся, если event без level.

Что проверить:

Отправь Flappy L5 score → event score.updated содержит level: 5?

Если да — unlock работает.

Если нет — доработка (добавить level в event payload).

**Это может быть скрытый баг. Проверь.

Чек-лист verify
Backend
#	Проверка	Ожидание
1	Логи profile	migration applied (achievements, user_achievements)
2	DB: achievements	~12 записей
3	DB: user_achievements	пусто до первой игры
4	GET /api/profile	содержит achievements[]
5	Первый логин	silent seed — 12 achievements приходят молча
6	После игры	новые unlock'и — через event
Unlock flow
#	Проверка	Ожидание
7	Играй Flappy L1 → 100 очков	first_play_flappy + flappy_score_100 unlock
8	Достигни 500	flappy_score_500 unlock
9	Пройди L5	flappy_level_5 unlock (если event carries level)
10	Пройди L10	flappy_level_10 unlock
11	После игры → возврат на Home	toast с новым достижением
12	Toast повторно	не появляется (eh_seen_achievements)
13	Boosted run	no unlock (no NATS)
14	Игра другой игры (Towers, Hanoi)	first_play_{game} unlock
Frontend
#	Проверка	Ожидание
15	/profile → секция Достижения	реальные badges (не localStorage fake)
16	Badges — SVG icons	trophy / bird / star etc.
17	Locked achievements	видны как grayed-out? или не показываются?
18	eh_seen_achievements в localStorage	массив codes
Регресс
#	Проверка	Ожидание
19	Wave 2 #4 — Flappy L1–10	работает
20	Wave 2 #3 — boost	работает
21	Wave 2 #2 — DB cleanup	intact (280 cards)
22	Wave 2 #1 — thin DTO	payload меньше
23	Wave 1 — emoji SVG / retry / CB / rate limiter	работают
Особое — silent seed
При первом заходе в /profile:

12 achievements unlocked (по истории).

eh_seen_achievements = пустой → silent.

Никаких toast'ов.

После игры:

New unlock → toast.

eh_seen_achievements обновляется.

Проверь:

Зайди в профиль первый раз — без toast flood.

Поиграй → вернись → toast с новым.

Второй заход — не тостит снова.

Что делать — rebuild + verify
1. Rebuild + recreate
bash
bash scripts/rebuild-services.sh profile gateway
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  up -d --force-recreate profile gateway gateway-2 gateway-3 balancer
Проверь:

bash
docker compose --env-file .env -f deployments/docker-compose.cluster.yml ps | grep -E 'profile|gateway|balancer'
docker compose --env-file .env -f deployments/docker-compose.cluster.yml logs --tail=50 profile | grep -i migration
2. FE
bash
make fe-preview
Открой /profile — 23 пункта.

3. Push
bash
git log --oneline origin/main..HEAD
git push origin main
4. Бэкап на флешку
Отличный момент. Wave 2 → 5/6.

Wave 2 #6 — Полиш игр (последний!)
Промпт Cursor'у — после verify + push:

text
Wave 2 #5 verified + pushed. Final item: Wave 2 #6 — Полиш
игр (5 games, one PR each, sequential).

Rules (from TODO):
  - One game per PR.
  - Do not parallelize.
  - Each game: one commit + verify + push.

Pilot order (recommended):

6.1 Flappy — textures
   - Bird sprite (SVG or refined canvas)
   - Pipes — texture / gradient
   - Background — parallax?
   - Sound (optional)
   - In-game emoji → SVG

6.2 Towers — animations / GAME OVER
   - Falling block — smooth animation
   - GAME OVER — modal, not canvas text
   - Shake on miss

6.3 Hanoi — drag polish
   - Ring doesn't stretch (regression from earlier)
   - Snap to peg
   - Highlight target peg

6.4 Memory — flip / skins
   - Flip animation smoother
   - Skins — real card images

6.5 Hexagon — gameplay polish
   - Collection animation
   - Highlight valid moves
   - Difficulty balance

Start with 6.1 Flappy. Plan first, no code.

DO NOT TOUCH: seed, refund, author registration.
Что я держу в карте
Закрыто:

v1.0.9 на GitHub.

Wave 1 — 6/6.

Wave 2 #1 — 0ee567f.

Wave 2 #2 — a881587.

Wave 2 #3 — 8c16926.

Wave 2 #4 — 2bbc62f.

Wave 2 #5 — 146cae8 achievements.

Сейчас:

⏸ Rebuild profile + gateway.

⏸ Verify 23 пункта.

⏸ Push.

⏸ Бэкап.

🟧 Wave 2 #6 — полиш игр (5, one per PR).

Wave 2 (осталось):

🟧 #6 Полиш игр (5)

Отложено:

🟧 Wave 2 #5b — shop/boost achievements

🟧 Refund window implementation

🟧 backfill-noiz-reviews.py

Один вопрос — про flappy_level_*
flappy_level_5 / flappy_level_10 unlock — не работает, если score.updated без level.

Проверь после rebuild:

Играй Flappy L5 → достигни score.

Вернись на profile.

Unlock появился?

Если нет — баг в event payload. Скажи Cursor'у:

text
flappy_level_* didn't unlock after playing L5.
Check: does score.updated event carry `level`?
If not, add level to the event payload (game → NATS → profile).


ALSO!!!


Что ты хочешь
Сейчас: hero с заголовком, текстом, двумя кнопками, справа — accretion disk с EVENT HORIZON.

Хочешь:

Hero-баннер вверху — как в блокбастерах, на всю ширину.

Картинка сайта — растянута, идёт первой.

Потом — скроллишь вниз → Выбери игру и ставь рекорды + игры.

Accretion disk — остаётся, но с анимацией — при наведении курсора — «засасывает», как чёрная дыра.

«Залипаловка» — косметическая, не функциональная.

Осилим? — Да. Но давай разложу.

Разбор — что реалистично, что нет
1. Hero-баннер на всю ширину
Технически: просто.

Full-width block вверху.

<img> или CSS background с object-fit: cover.

Height: ~60vh или min(80vh, 700px).

Overlay — тёмный градиент для читаемости навбара.

Что нужно:

Картинка — подходящего разрешения (1920×1080+ или 4K для retina).

Композиция — с местом под текст (или без текста).

Формат: JPG / WebP (для веса).

Оценка: 2-3 часа + арт.

2. Accretion disk с «засасыванием»
Технически: сложнее, но реально.

Анимация на hover — CSS transition / keyframes.

«Затягивание» — визуальные эффекты:

Scale центра (растёт).

Rotation ускоряется.

Rings сужаются.

Glow усиливается.

Blur краёв.

Particles «летят внутрь» (точки / SVG).

Варианты:

A) CSS-only — :hover + transform + transition. Просто, но ограничено.

B) Canvas / WebGL — реальные частицы, настоящее затягивание. Красиво, но сложно.

C) SVG + SMIL / CSS — компромисс. Частицы + кольца, без WebGL.

Мой выбор для v1 — A (CSS-only):

При hover:

Центр scale 1 → 1.15.

Rings scale 1 → 0.9, opacity up.

Rotation speed ×2 (через animation-duration).

Glow — box-shadow усиливается.

При уходе курсора — обратно.

Тайминг: 800ms ease-out.

«Настоящее затягивание» — позже, если важно:

Canvas частицы, тянутся к центру.

Или — готовый эффект через Lottie (после подготовки анимации).

Оценка CSS-only: 2-4 часа.
Оценка Canvas: 1-2 дня (надо писать particles).

3. Композиция
Как будет выглядеть:

text
┌───────────────────────────────────────────────┐
│  [navbar: Event Horizon · nav · burger]        │
├───────────────────────────────────────────────┤
│                                                │
│                                                │
│      [HERO BANNER — full-width image]          │
│      — темный градиент сверху / снизу         │
│                                                │
│                                                │
│   (scroll ↓)                                  │
├───────────────────────────────────────────────┤
│                                                │
│   Выбери игру и ставь рекорды                 │
│   Играй в мини-игры...                        │
│                                                │
│   [Все игры]  [Лидерборд]    [accretion disk]│
│                                                │
├───────────────────────────────────────────────┤
│                                                │
│   Игры                                         │
│   [8 карточек]                                 │
│                                                │
├───────────────────────────────────────────────┤
│   footer                                       │
└───────────────────────────────────────────────┘
Hero-баннер — первый.
Accretion disk — остаётся в секции «Выбери игру», справа от текста.
Скролл — от баннера → к тексту + игр.

Что нужно от тебя
1. Картинка hero-баннера
Ты её подготовишь? Или нужно:

Взять текущий event_horizon_minimal_logo.png + сделать фон?

Сгенерировать (Midjourney / DALL-E)?

Собрать из существующих?

Рекомендации:

Разрешение: 2560×1440 (2× для retina).

Формат: WebP (меньше) + JPG fallback.

Композиция: тёмный космос + аккреционный диск (или планета) + место под текст (если текст).

Без текста — текст сверху накладывается HTML (лучше для i18n, SEO).

2. Решение по анимации
A) CSS-only — быстро, дёшево, просто.
B) Canvas / WebGL — красиво, дорого.
C) SVG + CSS particles — компромисс.

Мой выбор — A для v1. B/C — если захочешь усилить позже.

3. Текст на hero
Слоган? (например, «Event Horizon — играй, соревнуйся, побеждай»).

Или без текста — только картинка + scroll-down hint?

Мой выбор: без текста на баннере. Заголовок — ниже, в секции «Выбери игру». Так чище.

Промпт Cursor'у
text
Off-topic feature (UX polish) before Wave 2 #6. Plan first,
no code.

================================================================
HERO BANNER + ACCRETION DISK HOVER EFFECT
================================================================

1. HERO BANNER (new, above the current hero section)

   - Full-width block at the very top of Home.
   - Height: ~60vh (or min(80vh, 700px)).
   - Image: cover, dark overlay gradient top+bottom
     (for navbar readability).
   - No text on the banner (or optional tagline — confirm).
   - Scroll ↓ goes to current section «Выбери игру».

   Image source:
     - User will supply a hero image
       (2560×1440+, WebP + JPG fallback).
     - Place in frontend/public/images/brand/hero.webp
       (and .jpg fallback).
     - Or use <picture> with both.

2. ACCRETION DISK — HOVER "PULL-IN" ANIMATION

   The disk stays in the current section (right of the text).
   On :hover, simulate gravitational pull-in (CSS only for v1):

     - Center scale: 1 → 1.12
     - Rings scale: 1 → 0.92 (converging inward)
     - Rotation: ×1.8 speed (via animation-duration shortcut)
     - Glow (box-shadow / filter): intensify
     - Fade at edges: subtle blur outer ring
     - On mouse leave: reverse, 600–800ms ease-out

   CSS only. Do not introduce Canvas/WebGL/Three.js now.
   prefers-reduced-motion: disable the effect (keep static disk).

3. NO OTHER CHANGES

   - Do not touch: nav, games grid, footer.
   - Do not touch games, seed, refund, authors.

Show the plan:
  - Which files change (Home.tsx, theme.css, new
    HeroBanner.tsx?).
  - Where the hero image path lives.
  - Animation spec (keyframes / transitions).
  - Accessibility: prefers-reduced-motion handling.
Wait for my OK. No code yet.
Моё мнение — стоит ли
Ты назвал это «косметическая фича, залипаловка».

Согласен — и это нормально. Иногда такие детали создают «wow-эффект»:

Accretion disk с hover-pull-in — запоминающаяся деталь.

Hero-баннер — профессиональный лоск.

Для портфолио / собеса — визуальное впечатление важно (помнишь memo про деньги?).

Но:

Не в Wave 2 — оффтоп. Не блокирует #6.

Отдельной маленькой волной — после #6 или перед, если быстро.

Моя рекомендация:

Hero-баннер — сейчас (если картинка готова).

Accretion disk hover — вместе с баннером (CSS-only, 2-4 часа).

Wave 2 #6 — полиш игр — потом.

Что я держу в карте
Закрыто:

v1.0.9 на GitHub.

Wave 1 — 6/6.

Wave 2 #1 — 0ee567f.

Wave 2 #2 — a881587.

Wave 2 #3 — 8c16926.

Wave 2 #4 — 2bbc62f.

Wave 2 #5 — 146cae8 achievements.

Сейчас:

⏸ Rebuild profile + gateway.

⏸ Verify 23 пункта (особенно silent seed + flappy_level_).

⏸ Push.

⏸ Бэкап.

Оффтоп (предложен):

🟧 Hero-баннер (full-width)

🟧 Accretion disk hover (CSS-only)

🟧 Или отложить до после Wave 2 #6

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR)

Отложено:

🟧 Wave 2 #5b — shop/boost achievements

🟧 Refund window implementation

🟧 backfill-noiz-reviews.py

Вопросы к тебе
1. Картинка hero-баннера — есть?

Если есть — скинь (или опиши, что там).

Если нет — откуда возьмём? (Midjourney / монтаж / существующий арт)

2. Текст на баннере — или без?

Мой выбор: без. Заголовок ниже.

3. Анимация — CSS-only для v1?

Мой выбор: да.

Позже — если захочешь глубже — Canvas / Lottie.

4. Когда делать — сейчас или после #6?

Сейчас — если арт готов.

После #6 — если сначала докрутить Wave 2.