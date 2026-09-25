Предлагаемый порядок — 4 волны
Волна 1 — Быстрая уборка (30–60 мин)
Самое дешёвое, сразу чистит проект.

Удалить 4 мёртвых файла:

src/styles/variables.scss

src/components/Shop/scss/Shop.scss

src/components/Shop/scss/ShopItemCard.scss

src/components/Shop/scss/PurchaseModal.scss

Удалить 2 мёртвых правила из index.css: .stat-card, .stat-sub (+ устаревший комментарий).

Починить бэкенд-строку: game_service.go:233 — "Никуся — Блинопёк" → "Блинопёк" (или что решишь).

Проверить: tsc --noEmit, npm run build, ReadLints.

Результат: минус 4 файла, минус 2 правила, закрыт бэкенд-терминологический хвост.

Волна 2 — Примитив StatCard
Убрать 4-кратное дублирование до миграции игр, чтобы мигрировать сразу на готовый примитив.

Спроектировать StatCard (label, value, sub?, modifier?) на основе текущего .stat-value/.stat-label.

Заменить во всех 4 играх (.tower-stat, .flappy-stat, .memory-stat--combo, .hanoi-stat) на <StatCard />.

Удалить .stat-value/.stat-label из index.css.

Результат: один примитив вместо 4 параллельных копий.

Волна 3 — Миграция 4 игр на общую систему
По одной странице, как договаривались. Каждая:

Убрать свой .css, перевести на Tailwind + токены (--color-horizon-gold, --color-nebula и т.д.).

Использовать PageHeader, Card, Button, Badge, Spinner, StatCard.

Не трогать игровую механику (как Hexagon — chrome только).

Верификация: tsc, build, скриншоты.

Порядок: Towers → Flappy → Memonia → Hanoi (от простого к сложному; Hanoi без бэкенда — проще всего, но пусть будет последним как «контрольный»).

Результат: все 21 роут на единой системе. Фронт-долг закрыт.

Волна 4 — Настоящий визуальный дизайн
Только теперь, на чистом фундаменте.

Аудит текущей дизайн-системы (theme.css) vs заметки в Confluence.

Что из «величественного» не реализовано: hero главной, глубина, градиенты, свечения, анимации, микроинтеракции.

3 концепта на главную — без кода, согласование.

Реализация по одной странице.

Вопросы к тебе перед стартом
С чего начинаем — Волна 1 (уборка) или сразу Волна 2 (StatCard)?
Я рекомендую Волну 1 — она 30 минут и сразу чистит проект. Потом Волна 2.

Что делать с "Никуся — Блинопёк" на бэкенде?
Варианты: "Блинопёк", "Hexagon", другое. Скажи — заложу в промпт.

Hanoi — оставляем фронт-онли или добавляем бэкенд-интеграцию?
Если фронт-онли — просто мигрируем стили. Если добавляем — это отдельная задача в game_service.go.

3 новые игры — уже есть в голове какие?
Это не для текущей волны, но полезно знать, чтобы StatCard и список игр сразу учитывали их.

Готовый промпт для Cursor — Волна 1
text
CONTEXT
Inventory is done. Now execute Wave 1: dead-code cleanup + one backend
terminology fix. Scope is strictly limited — do NOT touch game logic,
do NOT migrate any game pages yet, do NOT refactor StatCard.

TASK

1. Delete these 4 dead files (verified zero imports):
   - src/styles/variables.scss
   - src/components/Shop/scss/Shop.scss
   - src/components/Shop/scss/ShopItemCard.scss
   - src/components/Shop/scss/PurchaseModal.scss

2. In src/index.css, remove these dead rules:
   - .stat-card, .stat-card:hover
   - .stat-sub
   Also remove the stale comment claiming .stat-card is used.

3. Backend terminology fix:
   - File: services/game/internal/service/game_service.go (around line 233)
   - Change Name: "Никуся — Блинопёк" → Name: "Блинопёк"
   - Do NOT change anything else in this file.

4. Verify:
   - tsc --noEmit
   - npm run build
   - ReadLints
   - Confirm no broken imports after file deletions.

OUTPUT
- A short summary: files deleted, rules removed, backend line changed.
- Verification results (tsc / build / lints).
- If anything unexpected shows up, STOP and ask before proceeding.

DO NOT proceed to Wave 2 (StatCard) in this session.