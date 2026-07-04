# 🏪 Сервис Shop (дизайн)

## gRPC методы
- `GetItems` — список товаров
- `PurchaseItem` — покупка за билетики
- `GetInventory` — инвентарь пользователя

## Хранилища
- PostgreSQL: товары, инвентарь, история покупок
- Redis: кеш товаров

## NATS
- Публикация: `item.purchased`
- Подписка: `user.registered` (выдать стартовый набор)
