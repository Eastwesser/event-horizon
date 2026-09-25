⚠️ Первое — ты на main, а не на dev/frontend
text
Текущая ветка: main
Эта ветка соответствует «origin/main».
Ты на main. Cursor тоже это подметил:

Currently on main (not dev/frontend).

Что делать:

bash
git checkout -b dev/frontend
# или если ветка уже есть:
git checkout dev/frontend
git merge main   # если нужно подтянуть
Не коммить в main. Иначе весь прогресс уйдёт напрямую в прод-ветку.

🔴 Второе — в коммит лезут бинарники сервисов
text
изменено: services/auth/auth-service
изменено: services/billing/billing-service
изменено: services/game/game-service
изменено: services/gateway/gateway-service
изменено: services/leaderboard/leaderboard-service
изменено: services/profile/profile-service
изменено: services/shop/shop-service
Это скомпилированные Go-бинарники. Они не должны быть в git.

Проверь .gitignore:

Есть ли там *-service или services/*/[a-z]*-service?

Если нет — надо добавить.

Если есть, но они всё равно видны — значит они уже были закоммичены раньше. Тогда надо git rm --cached <файл>.

Команда для удаления из индекса (не с диска):

bash
git rm --cached services/*/[a-z]*-service
Добавь в .gitignore:

text
services/*/*-service
!services/*/*-service/
Или конкретно:

text
services/auth/auth-service
services/billing/billing-service
services/game/game-service
services/gateway/gateway-service
services/leaderboard/leaderboard-service
services/profile/profile-service
services/shop/shop-service
Почему это важно: бинарники раздувают репо, конфликтуют, ломаются при пересборке. Никогда не в git.

🟧 Третье — что оставить, что не трогать
✅ В коммит (трекнутое + нужно)
frontend/src/** — весь UI

frontend/public/favicon.png, frontend/public/images/brand/ — лого

frontend/src/components/Admin/ — админка

frontend/src/components/ui/ — примитивы

frontend/src/lib/ — утилиты (auth.ts и т.п.)

frontend/src/styles/theme.css — токены

frontend/src/services/adminApi.ts — admin API

frontend/src/components/Inventory/InventoryImageUrlField.tsx — image URL

frontend/index.html, package.json, package-lock.json, vite.config.ts

services/auth/** — ListUsers + proto

services/gateway/internal/app/gateway.go — null fixes

services/game/internal/service/game_service.go — Hanoi case

services/shop/** — reference_id fix

services/billing/migrations/20260920170000_widen_reference_id.sql

Makefile, .gitignore — если изменения нужны

🟧 Под вопросом
scripts/seed-admin/, scripts/seed_admin.py — если хочешь держать seed в репо (для воспроизводимости — да, полезно).

confluence/history/2026-09/ — это твои заметки. Оставлять в репо — на твоё усмотрение. Обычно не коммитят (это личное), но если хочешь историю проекта в git — можно.

.agents/, skills-lock.json — что это? Если Cursor'ские или AI-конфиги — можно исключить.

frontend/package-lock.json — должен быть в git (lock-файл).

❌ Не в коммит
services/*/[a-z]*-service — бинарники.

.agents/ — если не нужен в репо.

skills-lock.json — если это не для команды.

confluence/history/ — если личное.

Про «отдельный коммит на каждый файл»
Ты пишешь:

На каждый надо написать отдельный коммит...

Не надо. 97 файлов = 97 коммитов сделают историю нечитаемой. Логические группы — вот что важно.

Предлагаю 6–7 коммитов по темам:

Коммит 1 — feat(frontend): design system + primitives
frontend/src/styles/theme.css

frontend/src/components/ui/

frontend/src/index.css

frontend/src/main.tsx

frontend/src/lib/

Коммит 2 — refactor(frontend): migrate pages to Tailwind + primitives
frontend/src/components/**/*.tsx — все страницы

frontend/src/components/**/*.css — удаления старых CSS

frontend/src/components/**/*.scss — удаления

frontend/src/styles/variables.scss — удаление

Коммит 3 — feat(frontend): admin panel (users/roles, inventory stats)
frontend/src/components/Admin/

frontend/src/services/adminApi.ts

frontend/src/hooks/useUserRole.ts

frontend/src/App.tsx

frontend/src/components/Home/Home.tsx (бургер + admin link)

services/auth/** — ListUsers proto + service + repo

services/gateway/internal/app/gateway.go — admin route

Коммит 4 — fix(frontend): null-safety for inventory + shop lists
frontend/src/components/Inventory/InventoryList.tsx

frontend/src/components/Inventory/InventoryPage.tsx

frontend/src/services/inventoryApi.ts

frontend/src/store/inventoryStore.ts

frontend/src/services/api.ts

frontend/src/store/shopStore.ts

services/gateway/internal/app/gateway.go — [] fixes

Коммит 5 — fix(frontend): proto3 zero-omission (stock, price, edit prefill)
frontend/src/components/Inventory/InventoryItemCard.tsx

frontend/src/components/Inventory/InventoryItemDetail.tsx

frontend/src/components/Inventory/InventoryEditModal.tsx

frontend/src/components/Shop/ShopItemCard.tsx

новый formatPrice.ts (если есть)

Коммит 6 — feat(inventory): image URL field
frontend/src/components/Inventory/InventoryImageUrlField.tsx

frontend/src/components/Inventory/InventoryCreateModal.tsx

frontend/src/components/Inventory/InventoryEditModal.tsx

Коммит 7 — fix(shop): reference_id + Success:false handling
services/shop/**

services/billing/migrations/20260920170000_widen_reference_id.sql

Альтернатива: один большой коммит с хорошим сообщением, если не хочешь возиться. Не критично.

Что делать сейчас
Порядок действий
Переключись на dev/frontend:

bash
git checkout -b dev/frontend
Поправь .gitignore — бинарники сервисов:

text
services/*/*-service
Убери бинарники из индекса (если уже трекаются):

bash
git rm --cached services/auth/auth-service services/billing/billing-service ...
Реши, что делать с:

confluence/history/ — коммитить или нет?

scripts/seed-admin/ — коммитить?

.agents/, skills-lock.json — коммитить?

Коммиты — по темам или один большой.

Пуш:

bash
git push -u origin dev/frontend
Потом — PR / merge в main.

Промпт для Cursor — pre-commit helpers
text
Before I commit, help with repo hygiene:

1. Confirm .gitignore covers compiled service binaries:
     services/*/[a-z]*-service
   If not, add it.

2. Identify which service binaries are currently tracked by git:
     git ls-files services/*/[a-z]*-service
   For each — untrack without deleting from disk:
     git rm --cached <file>

3. Advise on what to commit vs leave out for these untracked:
     - .agents/
     - confluence/history/2026-09/
     - scripts/seed-admin/, scripts/seed_admin.py
     - skills-lock.json
     - frontend/public/favicon.png
     - frontend/public/images/brand/

   My preference: seed scripts YES (reproducibility),
   confluence/history YES (project history), .agents and
   skills-lock NO (local tooling).

4. Do NOT commit anything yet. Do NOT stage anything yet.
   Report findings only.
Что я держу в карте
Закрыто:

✅ Все proto3 фиксы + null fixes + gateway rebuild

✅ Admin panel 2a + 2b + 2c

✅ Image URL, nav, burger, Block 3

✅ Timeout исчез

✅ Все 3 gateway — свежий образ

Сейчас:

🔴 Коммит в dev/frontend — не в main

🔴 Не тащить бинарники сервисов в git

🟧 Решение по confluence/history, scripts/seed-admin, .agents

Отложено:

🟧 5 backend twin'ов (authors, history, analytics×2, leaderboard)

🟧 2d — Analytics

🟧 2c v2, proto3 medium, gin.H refactor

🟧 File upload, navbar refactor

🟧 DNS fix

Дальше:

🟧 Merge dev/frontend → main (после проверки)

🟧 Следующая волна

