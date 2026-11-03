📄 AUTHORS_AUDIT.md

🎯 Цель аудита

Проверить сквозной сценарий: 

Регистрация автора ➔ Создание карточек (Inventory) ➔ Синхронизация с магазином (Shop) ➔ Покупка (Billing) ➔ Отображение истории и метрик.

🛠️ Шаг 0: Подготовка окружения

    □ Перезагрузить виртуалку.
    □ Поднять весь кластер: docker compose -f docker-compose.cluster.yml up -d
    □ Проверить, что все контейнеры живы: docker ps
    □ Запустить фронтенд: cd frontend && npm run dev

🧑‍🎨 Шаг 1: Засев данных (Авторы и Карточки)

    □ Зайти под каждым из 5 авторов (*_2026).
    □ Создать по 1-2 карточки ККИ через UI (проверить загрузку картинок!).
    □ Указать корректные данные: Название, Описание, card_text, market_rub (например, 25), cost (золото/серебро).

Проверка SQL (Inventory):

docker exec -it event-horizon-postgres-inventory psql -U eventhorizon -d eventhorizon_inventory -c "SELECT name, author_id, price, attributes->>'market_rub' as market_rub, attributes->>'card_text' as card_text FROM inventory_items ORDER BY created_at DESC LIMIT 10;"

    □ Убедиться, что price = market_rub * 1000.
    □ Убедиться, что card_text не пустой.
    □ Убедиться, что author_id привязан к нужному пользователю.

🛒 Шаг 2: Синхронизация с магазином (NATS / Shop)

    □ Открыть страницу Магазина (/shop).
    □ Найти только что созданные карточки.
    □ Проверить, что цена в билетиках отображается корректно.
    □ Проверить работу фильтров (по типу, по стихии, по цене).

Проверка NATS / Outbox:

docker exec -it event-horizon-postgres-inventory psql -U eventhorizon -d eventhorizon_inventory -c "SELECT event_type, payload, created_at FROM outbox ORDER BY created_at DESC LIMIT 5;"

    □ Убедиться, что события inventory.item.created / updated генерируются.
    □ Посмотреть логи NATS-контейнера, чтобы убедиться, что сообщения доставлены.
    □ Посмотреть логи Shop, чтобы убедиться, что он их прочитал.

🛍️ Шаг 3: Покупка (Billing)

    □ Зайти под обычным пользователем.
    □ Пополнить баланс (или дать билетики через SQL, если платежка не подключена).
    □ Купить 3-4 карточки у разных авторов.

Проверка SQL (Billing / Shop):

# Проверить, что сток уменьшился
docker exec -it event-horizon-postgres-inventory psql -U eventhorizon -d eventhorizon_inventory -c "SELECT name, stock FROM inventory_items WHERE name IN ('Сова колдуна', 'Гронда', 'Крысиная нора', 'Латник Дахора', 'Ур-вайл', 'Предатель');"

# Проверить, что покупка записалась
# (В зависимости от того, как устроен ваш сервис биллинга, таблица может называться purchases или transactions)

    □ Убедиться, что stock уменьшился на 1.
    □ Убедиться, что баланс пользователя списался.
    □ Убедиться, что автору начислились 85% от суммы (проверить баланс автора).

📜 Шаг 4: История и События

    □ Зайти в раздел «История» (тот самый третий скриншот).
    □ Проверить, появились ли записи о покупках/регистрации.

Проверка SQL (History):


docker exec -it event-horizon-postgres-history psql -U eventhorizon -d eventhorizon_history -c "SELECT * FROM history_events ORDER BY created_at DESC LIMIT 5;"

    □ Если пусто — смотреть логи сервиса History и Gateway. Ищем ошибки парсинга или потери событий.
    □ Если записи есть, но фронт не показывает — смотреть HistoryPage.tsx и запросы в Network.

📈 Шаг 5: Метрики и Админка

    □ Зайти под админом.
    □ Открыть страницу метрик (/admin или /analytics).
    □ Проверить, что графики отображают покупки, активных пользователей и топ товаров.
    - ALSO, check Authors profiles, relogining and seeing info from them

Проверка SQL (Analytics):


# В зависимости от того, куда вы пишете метрики (ClickHouse / Prometheus / Postgres)

🚨 Решение проблем (Troubleshooting)

Если карточка не появилась в магазине: Проверить outbox. Если событие там есть, но его не прочитали — искать ошибку в консьюмере Shop.

Если сток не уменьшился: Проверить логи Inventory. Скорее всего, ReserveItem вернул FailedPrecondition (не хватает стока).

Если история пуста: Проверить, что Gateway отправляет события в NATS на топик истории, и что History-сервис на него подписан.

----

- idor (только ты меняешь свой товар) - check backend if it satisfies

- если купили все товары (у автора) Что происходит у авттора, у игрока, и тд - give answer

- gather metrics and make screenshots

- get NATS metrics

- what is going on with HISTORY service?



