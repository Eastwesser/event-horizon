#!/bin/bash

echo "========== USER CURRENCIES =========="
docker exec -it event-horizon-postgres-billing psql -U eventhorizon -d eventhorizon_billing -c "SELECT user_id, currency_type, balance FROM user_currencies;"

echo ""
echo "========== RECENT TRANSACTIONS =========="
docker exec -it event-horizon-postgres-billing psql -U eventhorizon -d eventhorizon_billing -c "SELECT user_id, currency_type, amount, balance_after, reason FROM transactions ORDER BY created_at DESC LIMIT 10;"

echo ""
echo "========== GAME HIGHSCORES =========="
docker exec -it event-horizon-postgres-game psql -U eventhorizon -d eventhorizon_game -c "SELECT user_id, game_id, score FROM highscores;"

echo ""
echo "========== LEADERBOARD BACKUP =========="
docker exec -it event-horizon-postgres-leaderboard psql -U eventhorizon -d eventhorizon_leaderboard -c "SELECT user_id, game_id, score, user_email FROM leaderboard_backup ORDER BY score DESC LIMIT 10;"

echo ""
echo "========== USERS =========="
docker exec -it event-horizon-postgres psql -U eventhorizon -d eventhorizon -c "SELECT id, email, created_at FROM users ORDER BY created_at DESC LIMIT 10;"