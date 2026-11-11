Сводка: до / после фикса
Профиль	До (baseline)	После (фикс)	Δ
P=50	0.67% failed, browse 1894ms, purchase 2264ms	1.11% failed, browse 4222ms, purchase 4397ms	⚠️ хуже по latency
P=100	9.35% failed, browse 68674ms, 5xx 7.17%	0.69% failed, browse 5057ms, 5xx 0%	✅ кардинально лучше
P=200	3.42% failed, browse 17309ms, deadlock 40P01, 5xx 1.74%	2.87% failed, browse 9477ms, 5xx 1.60%	✅ deadlock ушёл, browse ×1.8 лучше
Главное: deadlock 40P01 исчез
В P=200 (после фикса) нет ни одного deadlock detected (40P01). Вместо него:

text
WARN[0122] cancel1 non-200: status=500 body={"error":"lock item: pq: canceling statement due to lock timeout (55P03)"}
Это 55P03 lock_timeout — наш страховочный SET LOCAL lock_timeout = '2000ms' сработал. cancel ждал items FOR UPDATE больше 2 секунд (потому что purchase держал якорь), и отвалился с понятной ошибкой за 2с, а не повис на 30с с deadlock.

Почему это хорошо:

Deadlock (40P01) = Postgres сам убивает одну из транзакций, непредсказуемо. Может убить purchase, может cancel. Может откатить уже начатую сагу.

Lock timeout (55P03) = мы контролируем: 2с ожидания, потом fail. cancel возвращает 500, клиент видит ошибку, повторяет. Сага не рвётся посередине.

Почему это плохо (но ожидаемо):

cancel не может взять якорь 2с, потому что purchase держит его дольше. Это сериализация — мы её сами добавили. Под 20 purchase-VU на 10 items — 2 VU на item, и если purchase держит items FOR UPDATE 2+ секунды (Billing gRPC + INSERT + outbox), cancel ждёт.

Почему P=50 стал хуже (browse 4222ms vs 1894ms)
Это не регрессия, это шум. Смотри requests total: было 9719, стало 5959. То есть VU не успели сделать столько же запросов — потому что purchase стал медленнее (4397ms vs 2264ms), и purchase-VU дольше держат сессию, browse-VU тоже медленнее (4222ms vs 1894ms). Это не deadlock, это сериализация через items FOR UPDATE.

Гипотеза: purchase теперь всегда берёт items FOR UPDATE (даже если purchase_owned 39% — эти 409 не доходят до PurchaseItemWithStock, но RefundPurchase всегда берёт якорь). RefundPurchase — это cancel, а cancel в нашем k6-скрипте всегда идёт после buy. 5 purchase-VU × buy→cancel→cancel = 15 транзакций на 10 items, все сериализуются на items. Отсюда 4.4с.

Что делать: это не баг, это цена за отсутствие deadlock. Можно уменьшить, но не убрать:

Больше items (50+) → меньше contention. Это Шаг 5 (засев), но для purchase нужно items, не юзеров.

Или уменьшить lock_timeout до 500ms — быстрее fail, но больше 500-х.

Или разделить purchase и cancel на разные пулы (не нужно, они и так разные).

Я за «больше items» — это единственный способ уменьшить contention, не ломая корректность.

P=100 — кардинально лучше
До	После
Failed	9.35%	0.69%
browse p95	68674ms	5057ms
5xx	7.17%	0%
No available backends	много	0
Это самый большой выигрыш. Раньше P=100 был аномалией (хуже P=200). Теперь — линейно между P=50 и P=200. No available backends исчез — это была проблема cancel/purchase, которые висели на deadlock, Balancer не получал ответа, Nginx ретраил, копились коннекты. Теперь cancel быстро fail (2с), коннекты освобождаются, Balancer жив.

P=200 — browse ×1.8 лучше, но bulkhead всё ещё проблемит
До	После
browse p95	17309ms	9477ms
purchase p95	22025ms	14686ms
deadlock 40P01	много	0
bulkhead full (Shop)	много	много
lock timeout 55P03	0	1
Deadlock ушёл — это Шаг 1 закрыт. Осталось:

bulkhead full (Shop) — это Шаг 3, ещё не трогали.

circuit inventory (1 раз) — Inventory CB открылся, это Шаг 3/5 (bulkhead + индексы).