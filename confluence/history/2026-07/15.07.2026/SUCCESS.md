🎉 SUCCESS.md — 15.07.2026
Версия: v1.0.5 (багфикс)

📋 Что было сделано сегодня
1. 🛒 Магазин (Shop Service)
Исправления:
✅ Дата покупки — добавлено поле purchased_at в ответ API /shop/inventory

✅ Цена мерча — обновлена с 50 → 20000 билетиков

✅ Кеш баланса — исправлена инвалидация Redis после покупки

✅ Инвентарь — теперь загружается автоматически при открытии магазина

Технические детали:
Добавлено поле purchased_at в proto/shop.proto (Item)

Обновлён postgres_repo.go — сканирование purchased_at в GetUserInventory

Обновлён grpc_handler.go — конвертация *time.Time → string (RFC3339)

Создана миграция: 20260715200710_add_purchased_at_to_inventory.sql

Пересобраны: Shop Service и Gateway (для поддержки нового поля)

2. 🎮 Hexagon (Космические блины)
Исправлен маппинг всех типов блинов:
Обычный	Космический	Эмодзи	Смысл
🍫 nutella	🌙	Луна	Шоколадный → Луна
🍓 strawberry	⭐	Звезда	Клубничный → Звезда
🐟 fish	🌌	Космос	Рыбный → Галактика
🌭 sausage	☄️	Комета	Сосиска → Комета
🍗 chicken	🪐	Сатурн	Курица → Сатурн
🥗 caesar	🌠	Падающая звезда	Салат → Метеор
🍒 cranberry	✨	Искры	Клюква → Звездная пыль
🥞 pancake	☀️	Солнце	Обычный → Солнце
Файлы:

HexGrid.tsx — обновлены spaceEmojis и spaceColors

Tray.tsx — синхронизирован маппинг

3. 🎴 Memory (Мемония)
Исправления:
✅ Маппинг животных — добавлены недостающие эмодзи (🥑, 🥥, 🫐)

✅ Текст кнопки — динамическое переключение: "🐾 Карточки со зверями" ↔ "🍎 Карточки с фруктами"

✅ Фон карточек — убран жёлтый фон при переворачивании (только синий для скина)

Файлы:

MemoryBoard.tsx — обновлены defaultEmojis и animalEmojis

MemoryGame.tsx — динамический текст кнопки

MemoryCard.tsx — стиль применяется всегда

memory.css — добавлены стили для .memory-card--animals

4. 🐦 Flappy Bird
Исправления:
✅ Синяя птичка — дефолтный цвет #4A90D9

✅ Оранжевое крыло — дефолтный цвет #FF8C00

✅ Радужные трубы — сохранена форма (шляпка + градиент)

Файлы:

FlappyGame.tsx — цвета птички и трубы

5. 🗼 Towers
Исправления:
✅ Дефолтные блоки — красные оттенки (#E74C3C, #C0392B, ...)

✅ Радужные блоки — скин работает через переключатель

Файлы:

TowerGame.tsx — обновлена функция getBlockColor

6. 🏗️ Инфраструктура
Исправления:
✅ NATS Exporter — теперь работает (один экспортер на nats-1)

✅ Prometheus — добавлен job shop, все метрики в UP

✅ WebSocket — исправлен URL в Leaderboard.tsx (используется window.location.host)

7. 🔧 Пересборка сервисов
После добавления поля purchased_at в proto, Gateway был пересобран для поддержки нового поля:

bash
cd services/gateway
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go
docker build -t eastwesser/gateway:latest -f Dockerfile.gateway.bin .
🐛 Исправленные баги
Баг	Статус	Решение
Дата покупки — сегодняшняя	✅	Добавлено purchased_at в API
Инвентарь — (0) при открытии	✅	Загрузка при монтировании
Баланс в профиле — не отображается	✅	Добавлен запрос к /billing/balance/all
Кнопки скинов в Flappy — нет	✅	Исправлен useSkins хук
Космические блины — только 🌌	✅	Полный маппинг 8 типов
Memory — животные с фруктами	✅	Обновлены эмодзи
Memory — жёлтый фон	✅	Стиль скина применяется всегда
Цена мерча	✅	20000 билетиков
WebSocket — не подключается	✅	Исправлен URL
📊 Статус сервисов (Prometheus)
Сервис	Статус
Auth	✅ UP
Billing	✅ UP
Game	✅ UP
Leaderboard	✅ UP
Profile	✅ UP
Shop	✅ UP
Gateway (3)	✅ UP
Balancer	✅ UP
NATS	✅ UP
PostgreSQL	✅ UP
Redis	✅ UP
📝 Команды для будущих изменений
После изменения proto в любом сервисе — нужно пересобрать Gateway:

bash
cd services/gateway
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go
cd /home/denismatveev/event_horizon
docker build -t eastwesser/gateway:latest -f Dockerfile.gateway.bin .
docker push eastwesser/gateway:latest
make deploy
🎯 Планы на завтра (16.07.2026)
Нагрузочное тестирование с K6

Оптимизация индексов в БД всех сервисов

Анализ узких мест

Автор: Денис Матвеев (Eastwesser)
Дата: 15.07.2026
Версия: v1.0.5

🔥 Event Horizon — играй, соревнуйся, побеждай! 🔥