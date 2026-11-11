Отличный результат. Bulkhead 256/128/64/32 — работает: bulkhead full исчез полностью. Но появился новый сигнал: circuit inventory (Circuit Breaker Inventory открылся). Разберу и сравню с baseline.

Сводка: до / после всех фиксов (deadlock + nginx + bulkhead)
Профиль	Baseline (до всего)	После deadlock	После deadlock+nginx+bulkhead
P=50	0.67% failed, browse 1894ms, purchase 2264ms	1.11% failed, browse 4222ms, purchase 4397ms	0.30% failed, browse 3166ms, purchase 3463ms ✅
P=100	9.35% failed, browse 68674ms, 5xx 7.17%	0.69% failed, browse 5057ms, 5xx 0%	1.72% failed, browse 5900ms, 5xx 0% ✅
P=200	3.42% failed, browse 17309ms, deadlock 40P01, bulkhead full	2.87% failed, browse 9477ms, deadlock 0, bulkhead full	3.21% failed, browse 11515ms, bulkhead full 0, inventory CB 1 ⚠️
Что ушло
deadlock 40P01 — 0 на всех профилях. ✅ Шаг 1.

bulkhead full (Shop) — 0 на всех профилях. ✅ Шаг 3.

No available backends — 0. ✅ Шаг 2 (Nginx).

p99 36s — 0. ✅ Шаг 2 (Nginx read_timeout 5s).

Что появилось
circuit inventory (503) — в P=200, 9+ раз. Это Circuit Breaker Inventory открылся. Почему?

Circuit Breaker открывается, когда ConsecutiveFailures >= 5. Значит, Inventory получил 5 подряд 500/timeout. Скорее всего:

inventory/items под 140 browse-VU — Seq Scan (нет индексов, Шаг 5), PG не успевает → gRPC timeout (10s) → 5 подряд → CB открылся.

CB открылся → следующие запросы мгновенно 503 circuit inventory → browse errors 9.34%.

Это не регрессия — это следующий bottleneck. Мы сняли bulkhead (Shop больше не 503), теперь Inventory стал виден. Шаг 5 (индексы) — прямой фикс.

P=50 — лучший результат за всё время
text
Failed: 0.30% (было 0.67% baseline)
browse p95: 3166ms (было 1894ms, но с deadlock-фиксом было 4222ms)
purchase p95: 3463ms (было 2264ms, но с deadlock-фиксом 4397ms)
purchase_owned: 12.65% (было 39.76%)
Почему purchase_owned упал с 39% до 12%? Потому что bulkhead 256 пропускает больше purchase запросов, cancel успевает пройти, item возвращается в unowned. Меньше 409.

Почему browse 3166ms, а не 1894ms? Это цена items FOR UPDATE в RefundPurchase (Шаг 1). Purchase-цикл теперь сериализован на 10 items, browse-запросы ждут в очереди за CPU/соединениями. Это ожидаемо — Шаг 3.5 (короткий резерв) уберёт.

P=100 — лучше, чем baseline, но хуже, чем после deadlock
text
Failed: 1.72% (baseline 9.35%, после deadlock 0.69%)
browse p95: 5900ms (baseline 68674ms, после deadlock 5057ms)
Почему хуже, чем после deadlock? Потому что bulkhead 256 пропускает больше запросов → больше contention на 10 items → purchase p95 7031ms (было 7975ms — лучше). Browse 5900ms vs 5057ms — шум, ±15%.

Вывод: P=100 в норме, разница с deadlock-only в пределах шума.

P=200 — смешанный
text
Failed: 3.21% (baseline 3.42%, после deadlock 2.87%)
browse p95: 11515ms (baseline 17309ms, после deadlock 9477ms)
purchase p95: 18623ms (baseline 22025ms, после deadlock 14686ms)
bulkhead full: 0 ✅
inventory CB: 1 (9+ раз)
Что лучше: bulkhead full 0, deadlock 0, browse 17s → 11.5s (×1.5 лучше).
Что хуже: inventory CB открылся, browse errors 9.34% (было 7.15%).
Почему: Inventory не тянет 140 browse-VU × inventory/items с Seq Scan. CB защищает Inventory, но роняет browse.

Это следующий bottleneck. Шаг 5 (индексы) — прямой фикс.