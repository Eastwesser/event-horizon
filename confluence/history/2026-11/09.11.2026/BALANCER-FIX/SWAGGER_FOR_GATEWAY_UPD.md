Что в нём будет:

password.minLength: 8 (не 6)

enum games: [hexagon, flappy, memory, towers, hanoi, 2048, gears, tamagotchi] (8 игр)

include_deleted в GET /api/v1/inventory/items

version в PUT /api/v1/inventory/items/{id}

idempotency_key в attributes для POST /api/v1/inventory/items/bulk

nickname в GET /api/v1/leaderboard → entries[]

days в GET /api/v1/analytics/dau и mau

cohort_days_ago, window_days в GET /api/v1/analytics/retention

id=all в POST /api/v1/notifications/{id}/read

GetMyAuthor описание для GET /api/v1/authors/me (soft-deleted 200)

GET /api/v1/authors/{user_id} → 404 для soft-deleted

POST /api/v1/admin/shop/purchases/cancel

POST /api/v1/admin/payments/grant-subscription

POST /api/v1/admin/payments/revoke-subscription

Плюс: в services/gateway/api/swagger.go — api.OpenAPIYAML через //go:embed. Файл openapi.yaml в той же папке — не трогаем, он и так embed'ится. Просто заменяем содержимое.

