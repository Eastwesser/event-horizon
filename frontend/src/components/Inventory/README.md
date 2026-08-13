📁 Структура для Inventory в фронтенде:
text
frontend/src/
├── components/
│   └── Inventory/
│       ├── index.ts
│       ├── InventoryPage.tsx         # Главная страница каталога
│       ├── InventoryList.tsx          # Список товаров
│       ├── InventoryItemCard.tsx     # Карточка товара
│       ├── InventoryItemDetail.tsx   # Детальный просмотр
│       ├── InventoryCreateModal.tsx  # Создание товара
│       ├── InventoryEditModal.tsx    # Редактирование товара
│       └── styles/
│           ├── Inventory.scss
│           ├── InventoryItemCard.scss
│           └── InventoryModal.scss
├── hooks/
│   └── useInventory.ts               # Хуки для работы с API
├── services/
│   └── inventoryApi.ts               # API клиент для Inventory
└── store/
    └── inventoryStore.ts             # Zustand store для состояния
    