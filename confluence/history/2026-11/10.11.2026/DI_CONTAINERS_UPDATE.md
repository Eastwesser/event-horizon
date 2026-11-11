3. Остальные сервисы — потом, но в этой сессии
Согласен. Патчи для остальных:

game/di.go — highscores:

go
if _, err := db.ExecContext(ctx, `SELECT pg_prewarm('highscores')`); err != nil {
    a.log.Warn("pg_prewarm failed (non-fatal)", "err", err)
}
profile/di.go — user_profiles, achievements:

go
for _, tbl := range []string{"user_profiles", "achievements"} {
    if _, err := db.ExecContext(ctx, fmt.Sprintf(`SELECT pg_prewarm('%s')`, tbl)); err != nil {
        a.log.Warn("pg_prewarm failed (non-fatal)", "table", tbl, "err", err)
    }
}
Но — profile на pgxpool, не sql.DB. ExecContext есть у *pgxpool.Pool? Да — pool.Exec(ctx, sql). Патч:

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
CREATE EXTENSION в каждой PG (5 БД):

bash
for db in game profile billing leaderboard; do
  docker exec event-horizon-postgres-$db psql -U eventhorizon -d eventhorizon_$db -c "CREATE EXTENSION IF NOT EXISTS pg_prewarm;"
done
Плюс auth (users)? Auth на read path — users маленькая (10 строк), pg_prewarm не нужен. Но — можно добавить для полноты. Не критично.