.

📊 Текущее состояние Shop
✅ Что уже работает:
Backend Shop Service - работает, API отвечает

Товары в БД - есть 6 товаров (радужные трубы, золотая птичка и т.д.)

Покупка - работает через /api/shop/purchase

Инвентарь - работает через /api/shop/inventory

Баланс - работает через Billing

Frontend Shop - есть компоненты, но падают из-за ошибки парсинга ответа

❌ Что не работает:
Frontend не может отобразить товары - ошибка Cannot read properties of undefined (reading 'map')

Нет картинок для товаров - только пути /images/...

Нет применения кастомизации в играх

🎯 План действий:
1. Сначала починим фронтенд
Ошибка в shopStore.ts - ответ от API приходит не в том формате. Давайте исправим:

typescript
// frontend/src/store/shopStore.ts
// В fetchItems нужно обработать ответ правильно

fetchItems: async () => {
  const { lastFetch, items } = get();
  const now = Date.now();
  
  if (items.length > 0 && (now - lastFetch) < CACHE_DURATION) {
    console.log('📦 Используем кеш товаров');
    return;
  }

  set({ loading: true, error: null });
  try {
    const userId = localStorage.getItem('userId');
    if (!userId) throw new Error('Пользователь не авторизован');
    
    const response = await getShopItems();
    console.log('📦 Ответ от API:', response.data);
    
    // Проверяем структуру ответа
    let itemsData = response.data;
    // Если ответ - объект с полем items, берем items
    if (response.data && response.data.items) {
      itemsData = response.data.items;
    }
    // Если ответ - массив, используем его
    if (!Array.isArray(itemsData)) {
      console.error('❌ Неверный формат данных:', itemsData);
      throw new Error('Неверный формат данных от сервера');
    }
    
    const shopItems: ShopItem[] = itemsData.map((item: any) => ({
      id: item.id || item.Id,
      name: item.name || item.Name,
      description: item.description || item.Description || '',
      price_tickets: item.price || item.Price || 0,
      icon_url: item.image_url || item.ImageUrl || '',
      type: item.category || item.Category || 'merch',
      category: item.category || item.Category || 'merch',
      game_id: item.game_id || item.GameId,
      image_url: item.image_url || item.ImageUrl,
      available: item.available !== undefined ? item.available : item.Available !== undefined ? item.Available : true,
      owned: item.owned !== undefined ? item.owned : item.Owned !== undefined ? item.Owned : false,
    }));
    
    set({ 
      items: shopItems, 
      loading: false,
      lastFetch: now 
    });
  } catch (error: any) {
    console.error('❌ Ошибка загрузки товаров:', error);
    set({ 
      error: error.response?.data?.message || 'Ошибка загрузки товаров',
      loading: false 
    });
    throw error;
  }
},
2. Добавляем картинки для товаров
Создайте папку с картинками:

bash
mkdir -p /home/denismatveev/event_horizon/frontend/public/images/shop
Добавьте туда картинки:

rainbow_pipes.png - радужные трубы

golden_bird.png - золотая птичка

space_pancakes.png - космические блины

rainbow_blocks.png - радужные блоки

animal_cards.png - карточки со зверями

pancake_tshirt.png - блинный мерч

3. Архитектура Shop (ваш вопрос)
Вы правы, что магазин можно разделить. Вот как я вижу:

text
┌─────────────────────────────────────────────────────────────┐
│                     SHOP ECOSYSTEM                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Shop API   │    │   Catalog    │    │  Inventory   │  │
│  │   Gateway    │◄───│   Service    │    │   Service    │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│         │                   │                   │           │
│         ▼                   ▼                   ▼           │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Purchase   │    │   Supplier   │    │  Analytics   │  │
│  │   Service    │    │   Service    │    │   Service    │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
Но для MVP лучше оставить один Shop Service, а разделить позже, когда будет реальная нагрузка.

4. Применение кастомизации в играх
Нужно добавить логику применения скинов в каждой игре. Например, для Flappy:

typescript
// frontend/src/components/Games/Flappy/FlappyGame.tsx
// Проверяем инвентарь игрока
const checkInventory = async () => {
  const response = await getInventory();
  const items = response.data;
  
  // Проверяем, есть ли у игрока радужные трубы
  const hasRainbowPipes = items.some(item => 
    item.game_id === 'flappy' && 
    item.name.includes('Радужные трубы')
  );
  
  // Проверяем, есть ли золотая птичка
  const hasGoldenBird = items.some(item => 
    item.game_id === 'flappy' && 
    item.name.includes('Золотая птичка')
  );
  
  if (hasRainbowPipes) {
    // Используем радужные трубы
    setPipeColor('rainbow');
  }
  
  if (hasGoldenBird) {
    // Используем золотую птичку
    setBirdColor('golden');
  }
};
5. План действий на сегодня
🔥 Срочно:
Починить фронтенд (ошибка map)

Добавить картинки для товаров

Настроить инвалидацию кеша баланса

⚡ Важно:
Добавить применение скинов в Flappy Bird

Добавить применение скинов в Hexagon

Добавить применение скинов в Towers

Добавить применение скинов в Memory

💎 Дополнительно:
Добавить мерч (футболки, кружки)

Настроить WebSocket для обновления баланса в реальном времени

Добавить аналитику покупок

