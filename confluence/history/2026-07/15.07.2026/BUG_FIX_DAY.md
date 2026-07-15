# BUGFIXING

🐛 Баг 1: Дата покупки - показывает сегодняшнюю, а не реальную
Проблема в том, что в shopStore.ts при создании PurchasedItem мы используем new Date().toISOString() вместо реальной даты из ответа API.

Исправление - в shopStore.ts:

typescript
// Было:
purchased_at: item.purchased_at || item.PurchasedAt || new Date().toISOString(),

// Стало:
purchased_at: item.purchased_at || item.PurchasedAt || item.created_at || new Date().toISOString(),
Но лучше - получать дату из БД. Проверь, что API возвращает purchased_at.

🐛 Баг 2: Инвентарь показывает 0 пока не кликнешь
В Shop.tsx вкладка "Мой инвентарь" показывает ({inventory.length}), но inventory обновляется только при клике.

Исправление - в Shop.tsx:

tsx
// При первом рендере тоже загружаем инвентарь
useEffect(() => {
  const token = localStorage.getItem('accessToken');
  if (token) {
    fetchItems();
    fetchBalance();
    fetchInventory(); // ← добавить это
  }
}, [fetchItems, fetchBalance, fetchInventory]);
🐛 Баг 3: Кнопки применения скинов не видны
В FlappyGame.tsx кнопки появляются только если скины куплены, но они могут не отображаться из-за useSkins().

Проверь, что skins.flappy.hasRainbowPipes и skins.flappy.hasGoldenBird возвращают true.

🐛 Баг 4: В профиле не отображается баланс
В Profile.tsx баланс загружается, но не отображается. Проверь:

tsx
// Profile.tsx - должно быть так:
<div className="profile-balance">
  <span>💡 Лампочки: {balance.lamps}</span>
  <span>🎫 Билетики: {balance.tickets}</span>
</div>

Судя по логам, баланс приходит с бэкенда (lamps: 10045, tickets: 10028). Значит проблема в отрисовке.

🔍 Проверяем через curl

bash
# Проверяем баланс
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login -H "Content-Type: application/json" -d '{"email":"tuzer@example.com","password":"tuzer1"}' | jq -r '.access_token')
curl -s -X GET "http://localhost:8079/api/billing/balance/all" -H "Authorization: Bearer $TOKEN" | jq '.'

# Проверяем инвентарь
curl -s -X GET http://localhost:8079/api/shop/inventory -H "Authorization: Bearer $TOKEN" | jq '.'

# Проверяем профиль
curl -s -X GET http://localhost:8079/api/profile -H "Authorization: Bearer $TOKEN" | jq '.'
```

📋 Чек-лист исправлений
Баг	Статус	Нужно
Дата покупки	❌	Исправить в shopStore.ts
Инвентарь (0)	❌	Загружать при монтировании
Кнопки скинов	❌	Проверить useSkins
Баланс в профиле	❌	Проверить рендеринг
WebSocket	❌	Не подключается к 8080


[denismatveev@c0der event_horizon]$ TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login -H "Content-Type: application/json" -d '{"email":"tuzer@example.com","password":"tuzer1"}' | jq -r '.access_token')
curl -s -X GET "http://localhost:8079/api/billing/balance/all" -H "Authorization: Bearer $TOKEN" | jq '.'
{
  "lamps": 10045,
  "tickets": 10028
}
[denismatveev@c0der event_horizon]$ curl -s -X GET http://localhost:8079/api/shop/inventory -H "Authorization: Bearer $TOKEN" | jq '.'
[
  {
    "id": "d4734224-ab7b-46a6-b503-6886bfbf4bd2",
    "name": "Радужные блоки",
    "description": "Разноцветные блоки для башни",
    "price": 100,
    "category": "game_skin",
    "game_id": "towers",
    "image_url": "/images/rainbow_blocks.png",
    "available": true
  },
  {
    "id": "7bbca382-403c-400c-b1ef-ee03af49e4de",
    "name": "Космические блины",
    "description": "Блины в стиле космос!",
    "price": 100,
    "category": "game_skin",
    "game_id": "hexagon",
    "image_url": "/images/space_pancakes.png",
    "available": true
  },
  {
    "id": "521ada79-8821-4304-83d8-d903f3faf987",
    "name": "Карточки со зверями",
    "description": "Животные вместо фруктов в Меморине",
    "price": 150,
    "category": "game_skin",
    "game_id": "memory",
    "image_url": "/images/animal_cards.png",
    "available": true
  },
  {
    "id": "82be50db-670b-48c6-beb9-7e00d584f6de",
    "name": "Золотая птичка",
    "description": "Птичка становится золотой!",
    "price": 200,
    "category": "game_skin",
    "game_id": "flappy",
    "image_url": "/images/golden_bird.png",
    "available": true
  },
  {
    "id": "b5b332fd-146d-464c-b5f3-389a8fa48b83",
    "name": "Блинный мерч",
    "description": "Футболка с блином",
    "price": 50,
    "category": "merch",
    "image_url": "/images/pancake_tshirt.png",
    "available": true
  },
  {
    "id": "6a1de8dd-9457-4aa4-99a7-78267aee731d",
    "name": "Радужные трубы",
    "description": "Сделайте трубы в Flappy радужными!",
    "price": 100,
    "category": "game_skin",
    "game_id": "flappy",
    "image_url": "/images/rainbow_pipes.png",
    "available": true
  }
]
[denismatveev@c0der event_horizon]$ curl -s -X GET http://localhost:8079/api/profile -H "Authorization: Bearer $TOKEN" | jq '.'
{
  "user_id": "7fc8a659-1bb2-4d7c-b60e-c140239d5c62",
  "total_score": 725,
  "best_scores": {
    "flappy": 30,
    "hexagon": 325,
    "memory": 180,
    "towers": 190
  },
  "lamps": 85,
  "tickets": 45
}
[denismatveev@c0der event_horizon]$ 

