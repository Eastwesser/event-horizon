2. Что осталось — полный список
Из наших обсуждений (не сделано):

#	Что	Где	Приоритет	Время
1	pg_prewarm для game/profile/billing/leaderboard	4 di.go	P2	5 мин
2	CREATE EXTENSION pg_prewarm в 4 PG	bash	P2	30 сек
3	make ride-on	Makefile	P2	2 мин
4	Косметика — комментарии в 8 di.go	8 файлов	P3	5 мин
5	10k-probe-big.js для 32 GiB	новый файл	P3	15 мин
6	applyOwnedFlags — убрать старый ListOwnedItemIDs из интерфейса/репо (мёртвый код)	shop	P3	5 мин
7	Фикс теста mockShopStore	shop_service_test.go	P0	2 мин
8	Постмортем — финальный отчёт	confluence	P0	30 мин
9	pg_prewarm для auth/payment/authors/history/notification	5 di.go	P3	5 мин
10	JAEGER_ENDPOINT — проверить все 15 сервисов	compose	P3	2 мин
Забыли напрочь:

balancer_active_connections — метрика сломана (всегда 0). Техдолг, не блокер.

applyOwnedFlags — старый ListOwnedItemIDs закомментирован, но не удалён. Мёртвый код.

shop_service.go — комментарий «P1 batch — 2 SQL per shop/items» устарел (теперь 1 SQL). Поправить.

10k-probe-big.js — не создан (решили не надо, 10k-probe.js принимает PROFILE).

pg_prewarm для остальных 5 сервисов — не сделано.


4. Остальные тесты — по образу shop/inventory
Что ещё фиксить по образу (по твоему **списку):

game — pool 100/50 (сделал), pg_prewarm — нет (добавить, 5 минут).

profile — pool 100/50 (сделал), pg_prewarm — нет.

billing — pool 64/32 (сделал), pg_prewarm — нет.

leaderboard — pool 64/32 (сделал), pg_prewarm — нет.

auth/payment/authors/history/notification — pool есть, pg_prewarm — нет (для них не критично, но можно добавить).

Патчи pg_prewarm для game/profile/billing/leaderboard:

game/di.go — highscores:

go
if _, err := db.ExecContext(ctx, `SELECT pg_prewarm('highscores')`); err != nil {
    a.log.Warn("pg_prewarm failed (non-fatal)", "err", err)
}
profile/di.go — user_profiles, achievements (pgxpool):

go
for _, tbl := range []string{"user_profiles", "achievements"} {
    if _, err := dbpool.Exec(ctx, fmt.Sprintf(`SELECT pg_prewarm('%s')`, tbl)); err != nil {
        a.log.Warn("pg_prewarm failed (non-fatal)", "table", tbl, "err", err)
    }
}
billing/di.go — user_currencies, transactions (pgxpool):

go
for _, tbl := range []string{"user_currencies", "transactions"} {
    if _, err := dbpool.Exec(ctx, fmt.Sprintf(`SELECT pg_prewarm('%s')`, tbl)); err != nil {
        a.log.Warn("pg_prewarm failed (non-fatal)", "table", tbl, "err", err)
    }
}
leaderboard/di.go — leaderboard_backup (pgxpool):

go
if _, err := dbpool.Exec(ctx, `SELECT pg_prewarm('leaderboard_backup')`); err != nil {
    a.log.Warn("pg_prewarm failed (non-fatal)", "err", err)
}
CREATE EXTENSION pg_prewarm в 4 PG:

bash
for db in game profile billing leaderboard; do
  docker exec event-horizon-postgres-$db psql -U eventhorizon -d eventhorizon_$db -c "CREATE EXTENSION IF NOT EXISTS pg_prewarm;"
done
Плюс — game на sql.DB, не pgxpool. Смотри game/di.go — там db.SetMaxOpenConns (значит sql.DB). Патч для game — как у shop.