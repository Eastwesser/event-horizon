cd /home/denismatveev/event_horizon

# Auth
[denismatveev@c0der auth]$ goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5460/eventhorizon?sslmode=disable" up
2026/06/23 00:37:48 OK   20260530005336_init_users.sql (100.24ms)
2026/06/23 00:37:48 goose: successfully migrated database to version: 20260530005336
[denismatveev@c0der auth]$ 

# Billing
cd services/billing
goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5462/eventhorizon_billing?sslmode=disable" up
cd ../..

# Game
cd services/game
goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5461/eventhorizon_game?sslmode=disable" up
cd ../..

# Leaderboard
cd services/leaderboard
goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5463/eventhorizon_leaderboard?sslmode=disable" up
cd ../..

--

[denismatveev@c0der event_horizon]$ cd services/billing
goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5462/eventhorizon_billing?sslmode=disable" up
cd ../..
2026/06/23 00:39:40 OK   20260530020818_init_billing_schema.sql (150.67ms)
2026/06/23 00:39:40 goose: successfully migrated database to version: 20260530020818
[denismatveev@c0der event_horizon]$ cd services/game
goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5461/eventhorizon_game?sslmode=disable" up
cd ../..
2026/06/23 00:39:46 OK   20260530020905_init_game_schema.sql (43.08ms)
2026/06/23 00:39:46 goose: successfully migrated database to version: 20260530020905
[denismatveev@c0der event_horizon]$ cd services/leaderboard
goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5463/eventhorizon_leaderboard?sslmode=disable" up
cd ../..
2026/06/23 00:39:51 OK   20260530020936_init_leaderboard_schema.sql (40.71ms)
2026/06/23 00:39:51 goose: successfully migrated database to version: 20260530020936
[denismatveev@c0der event_horizon]$ curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'
{"user_id":"6de3aa05-61af-4144-8682-ae501218b805","email":"test@example.com","success":true,"message":"user registered successfully"}[denismatveev@c0der event_horizon]$ curl -X POST http://localhost:8079/api/curl -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456"}'
{"access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJleHAiOjE3ODIyNTA4MTksInVzZXJfaWQiOiI2ZGUzYWEwNS02MWFmLTQxNDQtODY4Mi1hZTUwMTIxOGI4MDUifQ.Ln2LnzPbdk8IzTYAYRCcaHTrW3p2HOWXxSjlQvn9tJQ","expires_in":86400,"token_type":"Bearer","user_id":"6de3aa05-61af-4144-8682-ae501218b805"}[denismatveev@c0der event_horizon]$ 