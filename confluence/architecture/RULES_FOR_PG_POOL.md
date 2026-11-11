Сводка: у кого что
Сервис	Тип пула	MaxConns	MinConns	Статус
analytics	ClickHouse (ch.Client)	—	—	не pgxpool, не трогаем
auth	pgxpool	25	10	⚠️ фикс
authors	pgxpool	25	10	⚠️ фикс
billing	pgxpool	25	10	⚠️ фикс
history	pgxpool	25	10	⚠️ фикс
leaderboard	pgxpool	25	10	⚠️ фикс
notification	pgxpool	25	10	⚠️ фикс
payment	pgxpool	25	10	⚠️ фикс
profile	pgxpool	25	10	⚠️ фикс
Все 8 (кроме analytics) — 25/10. Тот же баг. PG max_connections=100 на каждый сервис — 8 сервисов × 25 = 200 соединений потенциально (плюс inventory/shop/game по 100). Под P=200 не исчерпали (повезло — auth быстрый, billing на read path), но это мина.

Что делать — правило для всех
Не всем 100. Read-heavy — 100/50, write-heavy — 64/32, read-rare — 32/16.

Сервис	MaxConns	MinConns	Почему
auth	100	50	ValidateToken (пока без кэша — был bottleneck), GetUser, Login. Read-heavy. 100 — с запасом
billing	64	32	SpendCurrency/AddCurrency — write, деньги. 64 — норм для 20 purchase-VU. 100 — лишнее (contention на user_currencies)
leaderboard	64	32	Redis Sorted Set — основной, PG — бэкап. Read идёт в Redis, PG редко. 64 — с запасом
profile	100	50	Агрегатор (Auth + Game + Billing + PG). Read-heavy. 100 — с запасом
payment	32	16	Merch gate — редко, write (подписка). 32 — норм
authors	32	16	Applications — редко (admin), write. 32 — норм
history	32	16	Retention + read (редко). 32 — норм
notification	32	16	In-app inbox — редко. 32 — норм
analytics	—	—	ClickHouse, не pgxpool. Не трогаем

Логика: 100 — только auth/profile/shop/inventory/game (read-heavy, высокий RPS). 64 — billing/leaderboard (write/Redis). 32 — остальные (редко).