📦 Что будет: Магазин (Shop Service)
Ты хочешь микросервис, где игроки тратят билетики на ништяки.

Ништяки могут быть:

Категория	Примеры
Косметика	Скины для игр, аватарки, рамки
Бонусы	Ускорители, дополнительные жизни
Контент	Стихи, арты, мемы (твой Home Made браслетик)
Статус	Значки, титулы («Легенда», «Мастер»)
NFT-like	Уникальные предметы с историей

Архитектура:

text
[Shop :5055] — gRPC
    │
    ├── PostgreSQL (товары, инвентарь)
    ├── Redis (кеш)
    └── NATS (покупки, обновления инвентаря)
События в NATS:

shop.purchase — игрок купил предмет

shop.inventory_updated — изменился инвентарь

shop.new_item — появился новый товар

Не забудь добавить :

    "net/http"
    _ "net/http/pprof"

В сервис


## UPD INFO 3rd of July, 2026:


cd ~/event_horizon/services/store
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o store-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.store.bin -t eastwesser/store:latest .
docker-compose -f deployments/docker-compose.cluster.yml up -d store


git add .
git commit -m "feat: add Shop Service (v1.0.5)

- Shop Service: gRPC методы GetItems, PurchaseItem, GetInventory
- PostgreSQL: items, inventory, purchases
- Redis: кеш товаров (TTL 5 мин)
- NATS: shop.purchased событие
- Интеграция с Billing (списание билетиков)
- Обновлён Gateway: /api/shop/* эндпоинты
- Обновлён NATS Hub: добавлен shop.purchased
- Добавлены миграции для Shop БД"

git push origin main