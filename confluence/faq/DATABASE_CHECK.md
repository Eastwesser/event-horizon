# Команды для проверки БД 

# Billing — балансы игроков
docker exec -it event-horizon-postgres-billing psql -U eventhorizon -d eventhorizon_billing -c "SELECT user_id, currency_type, balance FROM user_currencies;"

# Billing — транзакции
docker exec -it event-horizon-postgres-billing psql -U eventhorizon -d eventhorizon_billing -c "SELECT user_id, currency_type, amount, balance_after, reason FROM transactions ORDER BY created_at DESC LIMIT 10;"

# Game — рекорды игроков (если сохраняем)
docker exec -it event-horizon-postgres-game psql -U eventhorizon -d eventhorizon_game -c "SELECT user_id, game_id, score FROM highscores;"

# Leaderboard — бэкап топа (если сохраняем)
docker exec -it event-horizon-postgres-leaderboard psql -U eventhorizon -d eventhorizon_leaderboard -c "SELECT user_id, game_id, score, user_email FROM leaderboard_backup ORDER BY score DESC LIMIT 10;"

# Auth — список пользователей
docker exec -it event-horizon-postgres psql -U eventhorizon -d eventhorizon -c "SELECT id, email, created_at FROM users;"

# Auth — список пользователей с никами
docker exec -it event-horizon-postgres psql -U eventhorizon -d eventhorizon -c "SELECT id, email, nickname, created_at FROM users;"

# Profile — агрегированный профиль
docker exec -it event-horizon-postgres-profile psql -U eventhorizon -d eventhorizon_profile -c "SELECT user_id, email, nickname, total_score, lamps, tickets FROM user_profiles;"
