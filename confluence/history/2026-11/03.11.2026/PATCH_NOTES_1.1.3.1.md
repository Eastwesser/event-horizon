📄 PATCH_NOTES_1.1.3.1.md
🎯 Patch 1.1.3.1 (Frontend Authors & Inventory Fixes)
Дата: 03.11.2026
Цель: Исправление критических багов в модальных окнах редактирования карточек (InventoryEditModal), синхронизация цены и текста, а также UI-полировка навигации.

🟩 Домен: Inventory (Редактирование карточек)
Файлы: frontend/src/components/Inventory/InventoryEditModal.tsx, frontend/src/services/inventoryApi.ts

Ширина модалки: Исправлен конфликт классов Tailwind (max-w-lg перебивал max-w-3xl). Теперь используется cn() с корректным мержем.

Скролл: Заголовок модалки (и кнопка закрытия) теперь зафиксирован. Скроллится только тело формы (flex-1 overflow-y-auto), что избавляет от «уезжающей» шапки.

Тип карточки (isCard): Исправлен баг, при котором поля ККИ (стихии, иконки) пропадали, если item.type не совпадал со строкой 'карточка'. Теперь проверка идет по наличию маркеров в attributes (market_rub, element).

Сохранение текста: Починен баг с потерей текста карты (card_text). Теперь он корректно сохраняется в attributes при любом типе карточки.

Пересчет цены: Починен баг с обнулением цены (25 → 10). Теперь цена в билетиках (suggestTickets) корректно пересчитывается при изменении поля «Рынок ≈ ₽» и отправляется на бэкенд.

Версионность: В UpdateItemRequest добавлено поле version. Это включает оптимистичную блокировку на бэкенде (защита от перезаписи устаревшими данными).

Двойной запрос: Удален дублирующийся вызов await updateItem, из-за которого улетало два запроса и возникал Version Conflict.

🟦 Домен: UI / Навигация
Файлы: frontend/src/components/Layout/AppNavbar.tsx, frontend/src/components/ui/Modal.tsx

Бургер-меню: Пункт «Автор» переименован в «Кабинет автора» для устранения путаницы с общей страницей «Авторы».

Модалка (Modal.tsx): Добавлен проп align?: 'center' | 'bottom'. Внутренний контейнер получил w-full и max-h-[90vh], что позволяет корректно переопределять ширину через className из родительских компонентов.

⚠️ Техдолг (Зафиксировано, но не исправлено на бэке)
IDOR в Inventory: На фронте кнопки редактирования чужих товаров скрыты, но на бэкенде (grpc_handler.go) отсутствует проверка author_id. Любой авторизованный пользователь может изменить чужой товар через gRPC/API.

🔍 Как проверить Outbox (6 сервисов)
Так как у тебя микросервисная архитектура с паттерном Outbox, события сначала пишутся в локальную таблицу БД, а потом воркер (или триггер) забирает их и отправляет в NATS.

Вот команды для проверки каждого сервиса (предполагая, что таблица называется outbox, а база соответствует сервису):

1. Inventory (Товары)

bash
docker exec -it event-horizon-postgres-inventory psql -U eventhorizon -d eventhorizon_inventory -c "SELECT event_type, payload, created_at FROM outbox ORDER BY created_at DESC LIMIT 5;"
Смотреть: события inventory.item.created или inventory.item.updated.

2. Game (Рекорды)

bash
docker exec -it event-horizon-postgres-game psql -U eventhorizon -d eventhorizon_game -c "SELECT event_type, payload, created_at FROM outbox ORDER BY created_at DESC LIMIT 5;"
Смотреть: события score.updated.

3. Billing (Деньги)

bash
docker exec -it event-horizon-postgres-billing psql -U eventhorizon -d eventhorizon_billing -c "SELECT event_type, payload, created_at FROM outbox ORDER BY created_at DESC LIMIT 5;"
Смотреть: события транзакций, списаний, пополнений.

4. Shop (Покупки)

bash
docker exec -it event-horizon-postgres-shop psql -U eventhorizon -d eventhorizon_shop -c "SELECT event_type, payload, created_at FROM outbox ORDER BY created_at DESC LIMIT 5;"
Смотреть: purchase.paid или purchase.created.

5. Authors (Авторы)

bash
docker exec -it event-horizon-postgres-authors psql -U eventhorizon -d eventhorizon_authors -c "SELECT event_type, payload, created_at FROM outbox ORDER BY created_at DESC LIMIT 5;"
Смотреть: author.upserted (создание/обновление профиля автора).

6. Payment (Платежи)

bash
docker exec -it event-horizon-postgres-payment psql -U eventhorizon -d eventhorizon_payment -c "SELECT event_type, payload, created_at FROM outbox ORDER BY created_at DESC LIMIT 5;"
Смотреть: события подписок или мерча.

📊 Твои графики (анализ перед ребутом)
Я посмотрел на твой скриншот Grafana.

Gateway RPS — есть всплески, система живет.

Gateway 5xx error — вот тут красный флаг. Есть скачки до 12.5%. Это либо падение gRPC-сервисов, либо таймауты (Jaeger покажет точнее). Скорее всего, это те самые ошибки Billing (500/503), которые мы видели в консоли.

NATS connections — стабильно, но in_msgs/s в нуле. Это значит, что либо события не пишутся, либо NATS Hub не читает их (проверим на аудите).

Tickets spent in the shop — ОГРОМНЫЙ НОЛЬ. Это подтверждает, что покупки через UI не проходят (или не фиксируются в метриках). Наша задача в аудите — сломать этот ноль.

History Service — scrape target health by job показывает, что History вроде живой, но на скрине истории пусто. Значит, либо Gateway не кидает события в History, либо запросы пользователя не передают user_id.

