#!/bin/bash

# Сохраняем все рекорды из Redis в PostgreSQL
GAME_DB="postgres://eventhorizon:eventhorizon@localhost:5461/eventhorizon_game"

# Получаем все ключи leaderboard
KEYS=$(docker exec -it event-horizon-redis-leaderboard redis-cli KEYS "leaderboard:*" | tr -d '\r')

for KEY in $KEYS; do
    # Извлекаем game_id из ключа (leaderboard:hexagon → hexagon)
    GAME_ID=$(echo "$KEY" | sed 's/leaderboard://')
    
    echo "📊 Processing $GAME_ID..."
    
    # Получаем все записи из Redis
    docker exec -it event-horizon-redis-leaderboard redis-cli ZREVRANGE "$KEY" 0 -1 WITHSCORES | while read -r USER_ID; do
        read -r SCORE
        
        # Вставляем в PostgreSQL
        docker exec -it event-horizon-postgres-game psql -U eventhorizon -d eventhorizon_game -c "
            INSERT INTO scores (user_id, game_id, score, created_at)
            VALUES ('$USER_ID', '$GAME_ID', $SCORE, NOW())
            ON CONFLICT (user_id, game_id) DO UPDATE
            SET score = EXCLUDED.score, created_at = NOW();
        " 2>/dev/null
        
        echo "  ✅ $USER_ID → $SCORE"
    done
done

echo "✅ Done!"