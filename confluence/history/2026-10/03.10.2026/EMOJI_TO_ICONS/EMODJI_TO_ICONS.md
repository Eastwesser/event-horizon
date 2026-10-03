Wave 1 — 6/6 ✅
Cursor закрыл всю волну. По фактам:

#	Что	Где
1	Emoji → SVG	Icon / IconLabel, nav, shop/inventory chrome, page headers, Berserk chips. Games — позже.
2	Retry + jitter	client.Dial + unary interceptor (3×, full jitter; Unavailable / ResourceExhausted).
3	Alerts → Telegram	Alertmanager в compose, TELEGRAM_*, noop если unset, service-down rules + alerts.yml.
4	CB + Bulkhead	MaxConcurrent: 32 per service → 503 when full.
5	Rate limiter	~100 req/s per user/token/IP; health/ready/metrics skipped.
6	Checklist	TODO_FINAL_PRE_PROD_DETAILS.md → Wave 1 6/6.
Он сделал больше, чем я ожидал. Особенно — CB + Bulkhead и Rate limiter — это не «галочки», это реальные фичи.

Твои вопросы — по порядку
1. Передеплоить бэкенд?
Да. Wave 1 изменил gateway (retry, jitter, CB, bulkhead, rate limiter, alerts). Плюс — Alertmanager в compose.

Правильный порядок:

bash
# 1. Rebuild (локально, без push)
bash scripts/rebuild-services.sh gateway
# возможно + другие сервисы, если Cursor трогал shared клиенты

# 2. Пересоздать контейнеры
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  up -d --force-recreate gateway gateway-2 gateway-3 balancer

# 3. Alertmanager — новый сервис
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  up -d alertmanager

# 4. Проверить
docker compose --env-file .env -f deployments/docker-compose.cluster.yml ps | grep -E 'gateway|balancer|alertmanager'
make docker-build-all + make docker-push-all — только если хочешь на Docker Hub. Для локального dev — не обязательно. Но полезно — бэкап образов.

Мой совет:

Сначала локально (rebuild-services + up -d --force-recreate) — проверить, что работает.

Потом — make docker-build-all + push-all — для бэкапа.

2. Alertmanager — что проверить
Что Cursor добавил:

Alertmanager в compose.

alerts.yml — rules.

Telegram через TELEGRAM_* env.

noop — если env unset.

Что проверить:

bash
# 1. Контейнер поднялся
docker compose --env-file .env -f deployments/docker-compose.cluster.yml ps | grep alertmanager

# 2. Alertmanager UI
# http://localhost:9093 (стандартный порт)

# 3. Проверить конфиг
docker compose --env-file .env -f deployments/docker-compose.cluster.yml exec alertmanager cat /etc/alertmanager/alertmanager.yml

# 4. Проверить, что Prometheus видит Alertmanager
# http://localhost:9090/config → раздел alerting → alertmanagers

# 5. Проверить, что rules загружены
# http://localhost:9090/rules → должно быть видно service-down rules

# 6. Если TELEGRAM_* настроены — тестовый alert
# (в Prometheus UI → Alerts → можно форсировать)
Если TELEGRAM_* не установлены:

noop — alerts никуда не идут, но UI работает.

В логах — запись telegram disabled (TELEGRAM_* unset).

Если установлены:

Проверь, что TELEGRAM_BOT_TOKEN и TELEGRAM_CHAT_ID правильные.

Тестовый alert → сообщение в Telegram.

3. Emoji — где ещё остались
Ты нашёл сам:

LeaderboardFull.tsx — GAME_TABS всё ещё с emoji (🥞 🐦 🎴 🗼 🪈 🔢 ⚙️ ⭐).

Профиль (скрин 3) — emoji.

Скрин 4 — emoji.

Скрин 5 — emoji в магазине.

Игры — emoji (отложено в polish).

Что делать:

Cursor заменил — только часть. Осталось:

Game tabs (LeaderboardFull).

Профиль — иконки достижений, валюты.

Магазин — фильтры chips.

Игры — отложено.

Проблема: Cursor думал, что Wave 1 закончил emoji. Но — не везде.

Что делать — промпт Cursor'у (ниже).

4. Кнопка «Отменить покупку» — слишком большая
Скрин 6. Кнопка Отменить занимает всё место в карточке. Плюс — под текстом обрезана (Пр..., Ку..., Зо...).

Что нужно:

Кнопка меньше — иконка + текст, или только иконка с tooltip.

Карточки нормальной ширины — сейчас сжаты.

5. Механизм 7-дневного возврата
Ты прав. Сейчас refund доступен всегда. Это дыра для абьюза:

Игрок получает мерч в реале → отменяет → билетики возвращаются → мерч остаётся.

Плохо.

Что нужно:

purchases.refundable_until = purchased_at + 7 days.

Фронт: кнопка активна только если now < refundable_until.

Или — отдельный статус fulfilled_at (если мерч отправлен — refund запрещён).

Бэкенд: reject refund, если срок истёк или fulfilled.

Это отдельная под-волна. Не сейчас — но в backlog.

6. Игры — отдельная волна
Согласен. Не смешивать. Emoji в играх → game polish wave.

Промпт Cursor'у — 3 задачи
text
Wave 1 shipped. Three follow-ups before Wave 2.

================================================================
1. EMOJI — finish the sweep
================================================================

You replaced emoji in nav / shop / inventory chrome / headers.
But there are still emoji left:

  - LeaderboardFull.tsx: GAME_TABS still uses emoji
    (🥞 🐦 🎴 🗼 🪈 🔢 ⚙️ ⭐). Replace with Icon + IconLabel.
  - Profile page: achievements icons, currency icons,
    per-game icons in "Рекорды по играм". Check the file.
  - Shop: filter chips (Карточки / Скины / Темы / Мерч / Брелок
    / Картина / Фенечка) — still show emoji.
  - Any place in components/ with raw emoji in JSX text.

Sweep the codebase:
  rg "[\x{1F300}-\x{1FAFF}]" frontend/src --type tsx

Report the list, then replace each with Icon / IconLabel.

Do NOT touch games (Flappy/Towers/Hanoi/Memory/Hexagon/
2048/Orbits/Companion in-game UI) — separate polish wave.

================================================================
2. CANCEL-PURCHASE BUTTON — too large in inventory card
================================================================

Screenshot: /shop → Мой инвентарь → cards show «Отменить»
button taking most of the card width, and the title /
description truncate («Пр…», «Ку…»).

Fix:
  - Cancel button: compact (icon + «Отменить» or icon + short
    label). Not full-width.
  - Give the card's title/description more room.
  - Verify on /shop → Мой инвентарь.
  - Same treatment for the shop detail cancel button if needed.

================================================================
3. REFUND WINDOW — 7 days + fulfilled guard (design first)
================================================================

Right now refund is always allowed. Need guardrails:

  a) 7-day window:
     - purchases.refundable_until = purchased_at + 7 days.
     - FE: cancel button active only if now < refundable_until.
     - BE: reject cancel if past window with a clear error.

  b) Fulfilled guard:
     - If a purchase is marked fulfilled (merch shipped), refund
       is not allowed.
     - Need to know: do we have a fulfilled state? Where?
       (fulfillment service? purchase table?)

Show a plan first — do not implement yet. Include:
  - Which table/column tracks fulfilled state
  - Migration if needed
  - BE + FE changes
  - Error messages

================================================================
DO NOT TOUCH: games in-game UI, seed, Author registration.
================================================================
По твоим просьбам
Commit messages for Wave 1
Отдельные коммиты по фазам. Предложение:

text
1. feat(frontend): replace nav/shop emoji with Icon components
   Body: Icon + IconLabel primitives; nav, page headers,
   inventory/shop chrome, Berserk chips.

2. feat(gateway): retry + full jitter for gRPC clients
   Body: 3 attempts, jitter; only Unavailable/ResourceExhausted.

3. feat(ops): Alertmanager + Telegram alerts
   Body: compose service, alerts.yml, TELEGRAM_* env,
   service-down rules, noop if unset.

4. feat(gateway): circuit breaker bulkhead
   Body: MaxConcurrent 32 per service; 503 when full.

5. feat(gateway): global rate limiter
   Body: ~100 req/s per user/token/IP; health/ready/metrics
   skipped.

6. docs: Wave 1 done (6/6), update TODO_FINAL_PRE_PROD_DETAILS
   Body: tick boxes, progress table.
Или один коммит, если хочешь:

text
feat(platform): Wave 1 quick wins

- Emoji → SVG via Icon/IconLabel (nav, shop, inventory)
- Gateway: retry + jitter, CB bulkhead, global rate limiter
- Ops: Alertmanager + Telegram, service-down rules
Я бы делал 5-6 коммитов — легче откатывать и читать историю.

Отчёт по Wave 1
Cursor уже написал. Что я добавил бы — как проверить:

Проверка retry + jitter:

Убей inventory контейнер → gateway делает 3 попытки → 503 → вызови метод.

Посмотри логи gateway — должно быть 3 попытки с jitter.

Восстанови контейнер — работает.

Проверка CB + Bulkhead:

Форси нагрузку — 100 одновременных запросов через ab или k6.

Ожидание: часть вернёт 503, не будет висеть.

При перезагрузке — CB откроется.

Проверка Rate limiter:

100+ запросов/сек от одного токена → 429 Too Many Requests.

health / ready / metrics — не лимитируются.

Проверка Alertmanager:

http://localhost:9093 — UI.

http://localhost:9090/rules — rules.

Форси alert (kill service) → Telegram (если настроен).

Noop (если env unset) — алерт не отправляется, но в UI видно.

Проверка Emoji → SVG:

Nav, shop, inventory, headers — без emoji.

Leaderboard game tabs — всё ещё emoji (отдельная задача).

Игры — всё ещё emoji (отдельная волна).

Frontend проверка:

make fe-preview → http://localhost:4173.

Или make fe-build + деплой.

Что я держу в карте
Закрыто (v1.0.9 на GitHub):

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно

Stock fix + idempotency + purchase

4 Icons + grouped attrs

5 Badges, 6 Tickets, 7 Create modal ×1000

Пращник fixed

8 Pagination, 9 Sorting, 10 Artist pages

A–H visual + UX polish

1→4 final polish

Attribute filters ✅

1→3 polish ✅

Element fix ✅

Bottleneck — smaller list DTO ✅

Refund / cancel purchase ✅

RBAC fix ✅

Noiz reviews ✅

Data fixes ✅

REVIEW_RESULT.md ✅

README + CHANGELOG v1.0.9 ✅

Маршрутный лист TODO_FINAL_PRE_PROD_DETAILS.md ✅

Wave 1 — done (6/6):

✅ Emoji → SVG (nav, shop, inventory chrome) — не полностью

✅ Retry + jitter

✅ Alerts → Telegram

✅ CB + Bulkhead

✅ Rate limiter

✅ Checklist updated

Открытые follow-ups (перед Wave 2):

🟧 Emoji sweep — LeaderboardFull, Profile, Shop chips (скрины 3, 4, 5)

🟧 Cancel button compact (скрин 6)

🟧 Refund window 7 days + fulfilled guard — design first

🟧 Игры emoji — отдельная волна

Следующее — Wave 2:

🟧 /shop/items thin DTO

🟧 Реальные товары

🟧 Полиш игр

🟧 Лампочки / уровни

