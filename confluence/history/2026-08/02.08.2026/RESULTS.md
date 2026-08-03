[denismatveev@c0der event_horizon]$ TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tuzer@example.com","password":"tuzer1"}' \
  | jq -r '.access_token')

echo "Token: $TOKEN"
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InR1emVyQGV4YW1wbGUuY29tIiwiZXhwIjoxNzg1ODAzOTE3LCJ1c2VyX2lkIjoiN2ZjOGE2NTktMWJiMi00ZDdjLWI2MGUtYzE0MDIzOWQ1YzYyIn0.b8Xs2m_01dveGepd6xD088K6sgmsq9_37YSxDz98sKY
[denismatveev@c0der event_horizon]$ curl -X POST http://localhost:8079/api/inventory/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "брелок",
    "name": "Тестовый брелок",
    "description": "Красивый брелок из серебра",
    "price": 100.50,
    "stock": 10,
    "attributes": {"material": "серебро", "weight": 15},
    "images": ["/images/keychain1.png"]
  }' | jq '.'
  % Total    % Received % Xferd  Average Speed  Time    Time    Time   Current
                                 Dload  Upload  Total   Spent   Left   Speed
100    783 100    491 100    292   1020    607                              0
{
  "item": {
    "id": "45a3b9d7-0713-455c-9acf-454e296f2d20",
    "author_id": "7fc8a659-1bb2-4d7c-b60e-c140239d5c62",
    "type": "брелок",
    "name": "Тестовый брелок",
    "description": "Красивый брелок из серебра",
    "price": 100.5,
    "stock": 10,
    "attributes": {
      "material": "серебро",
      "weight": 15
    },
    "images": [
      "/images/keychain1.png"
    ],
    "created_at": "2026-08-03 00:38:51.279879845 +0000 UTC m=+192.289047092",
    "updated_at": "2026-08-03 00:38:51.279879961 +0000 UTC m=+192.289047205"
  }
}
[denismatveev@c0der event_horizon]$ curl -X GET "http://localhost:8079/api/inventory/items?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
  % Total    % Received % Xferd  Average Speed  Time    Time    Time   Current
                                 Dload  Upload  Total   Spent   Left   Speed
100    466 100    466   0      0   5141      0                              0
{
  "items": [
    {
      "id": "45a3b9d7-0713-455c-9acf-454e296f2d20",
      "author_id": "7fc8a659-1bb2-4d7c-b60e-c140239d5c62",
      "type": "брелок",
      "name": "Тестовый брелок",
      "description": "Красивый брелок из серебра",
      "price": 100.5,
      "stock": 10,
      "attributes": {
        "material": "серебро",
        "weight": 15
      },
      "images": [
        "/images/keychain1.png"
      ],
      "created_at": "2026-08-03 00:38:51.27988 +0000 +0000",
      "updated_at": "2026-08-03 00:38:51.27988 +0000 +0000"
    }
  ],
  "total": 1
}

json
{
  "item": {
    "id": "45a3b9d7-0713-455c-9acf-454e296f2d20",
    "author_id": "7fc8a659-1bb2-4d7c-b60e-c140239d5c62",
    "type": "брелок",
    "name": "Тестовый брелок",
    "price": 100.5,
    "stock": 10,
    "attributes": {"material": "серебро", "weight": 15}
  }
}
✅ Товар создан!
✅ Список получен!
✅ Gateway работает!
✅ Inventory работает!
✅ Outbox работает!

🎯 ЧТО МЫ СДЕЛАЛИ ЗА ЭТУ СЕССИЮ:

Компонент	Статус	Что сделано
Inventory Service	✅	Полностью написан, скомпилирован, запущен
Outbox + NATS	✅	Реализован transactional outbox
PostgreSQL	✅	Своя БД, таблицы, индексы
MongoDB	✅	Адаптер для тренировки
Redis	✅	Кеширование товаров
Gateway	✅	Добавлены все роуты для Inventory
OpenAPI	✅	Обновлена документация
Docker	✅	Собраны образы, запущены контейнеры
Prometheus	✅	Метрики собираются
Health check	✅	Работает