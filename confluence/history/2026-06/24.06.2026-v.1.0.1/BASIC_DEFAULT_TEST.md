# 1. Регистрация
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","nickname":"TestUser"}'

# 2. Логин (получить токен)
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

echo "Токен: $TOKEN"

# 3. Проверить баланс
curl -X GET http://localhost:8079/api/billing/balance/all \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 4. Отправить рекорд
curl -X POST http://localhost:8079/api/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"user_id":"test-user","game_id":"hexagon","level":1,"score":150,"seed":"test_seed_123","moves":[]}' | jq '.'

# 5. Проверить лидерборд
curl -s "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10" | jq '.'

# RESULTS:

```curl
[denismatveev@c0der event_horizon]$ curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","nickname":"TestUser"}'
[denismatveev@c0der event_horizon]$ TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \ge":"user registered successfully"}[denismatveev@c0der event_horizon]$ TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')
[denismatveev@c0der event_horizon]$ echo "Токен: $TOKEN"
Токен: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJleHAiOjE3ODI0MDQ0NzcsInVzZXJfaWQiOiJhMTFkMzc2Ny1hMjE4LTQyODEtYjBjNy1iOWNlNGM4MTEwODAifQ.Mh07Q2LmZxUOgWzcJRiOOzAFlHQw6AV0swTzNfUwArA
[denismatveev@c0der event_horizon]$ curl -X GET http://localhost:8079/api/billing/balance/all \
  -H "Authorization: Bearer $TOKEN" | jq '.'
  % Total    % Received % Xferd  Average Speed  Time    Time    Time   Current
                                 Dload  Upload  Total   Spent   Left   Speed
100     23 100     23   0      0     66      0                              0
{
  "lamps": 0,
  "tickets": 0
}
[denismatveev@c0der event_horizon]$ curl -X POST http://localhost:8079/api/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"user_id":"test-user","game_id":"hexagon","level":1,"score":150,"seed":"test_seed_123","moves":[]}' | jq '.'
  % Total    % Received % Xferd  Average Speed  Time    Time    Time   Current
                                 Dload  Upload  Total   Spent   Left   Speed
100    213 100    114 100     99   1034    898                              0
{
  "success": true,
  "new_highscore": 150,
  "message": "score submitted successfully",
  "lamps_earned": 10,
  "tickets_earned": 1
}
[denismatveev@c0der event_horizon]$ curl -s "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10" | jq '.'
{
  "entries": [
    {
      "rank": 1,
      "user_id": "test-user",
      "score": 300,
      "updated_at": 1782318105
    }
  ]
}
```