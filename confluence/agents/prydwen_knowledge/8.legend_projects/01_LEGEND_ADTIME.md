# AdTime — легенда проекта (GIFTS.ru / кастом мерча + DesChat)

> **Честный дисклеймер:** это **interview narrative** вокруг реального контура GIFTS.ru (кастом одежды/мерча) + чат с дизайнерами. Не выдумывай личные KPI (RPS, «200 дизайнеров одновременно»). Говори то, что помнишь; остальное — «так мы проектировали / типичная схема».  
> Старый черновик «ad-tech биддинг» в корпусе был **неверным якорем** — не использовать.  
> Mermaid: `unrelated/kata-lectors/5.system_design/sd_schemas.md` §1.  
> Mongo Q&A: `confluence/interview/2.database/mongodb_adtime_chat.md`.

## Elevator pitch (30 сек)

AdTime — IT-обёртка над GIFTS.ru: пользователь **генерирует принт нейросетью**, drag-and-drop накладывает на мерч (футболка / толстовка / кепка), рендерит превью → оформляет заказ → **общается с дизайнерами в чате (WebSocket)**, чтобы довести макет.

## Сервисы (якоря)

| Сервис | Роль | Стор |
|--------|------|------|
| `auth` | сессии / JWT | PG / Redis |
| `generation` | AI jobs (очередь) | Redis quota + RabbitMQ jobs + S3 артефакты |
| `order` / `store` | заказ и каталог мерча | PostgreSQL |
| `deschat` | чат с дизайнерами | **Mongo** (messages/chats) + Redis PubSub fanout + WS |

Стек вслух: Go, PG, Mongo (chat), Redis, RabbitMQ (generation), Kafka (order events), S3, Nginx, WS.

## DesChat — speakable (то, что чаще всего спросят)

1. **Коллекции:** `chats` (participants, lastMessage, updatedAt) + `messages` (chatId, userId, text, attachments, createdAt).  
2. **Индексы:** `{chatId, createdAt:-1}` пагинация; `{chatId, userId}` фильтр; `{participants.userId, updatedAt:-1}` список чатов; **TTL** на `messages.createdAt` (в конспекте 90d).  
3. **Отправка сообщения (TX):** session → insert message → update `lastMessage` + unread → commit (multi-doc с 4.0; образ в конспекте 7.0.5, рынок 8.x).  
4. **Realtime:** WS + Redis PubSub (не тащить fanout через Mongo).  
5. **Рост:** shard key ≈ `chatId`; archive / TTL; explain → составные индексы; optimistic lock на `participants` / version.  
6. **Мост к EH:** «в Event Horizon те же идеи consistency — Postgres + Outbox; Mongo в AdTime — именно под append-чат».

Подробные вопросы интервьюера → файл `mongodb_adtime_chat.md`.

## Почему Mongo для чата, а не для Lime-каталога

- **Чат:** документ-сообщение, денорм `lastMessage`, гибкие attachments, write-append.  
- **Магазин одежды (Lime):** атрибуты/фильтры, inventory oversell, saga payment → лучше **PG + ES**, Mongo как единственный стор — риск.

## Роль (STAR-шаблон без фейковых цифр)

1. **Ситуация:** нужен чат клиент↔дизайнер рядом с AI-генерацией мерча.  
2. **Задача:** надёжная доставка сообщений, список чатов, пагинация истории, рост архива.  
3. **Действие:** схема Mongo + индексы + TX lastMessage + WS/Redis; slow queries → explain/indexes.  
4. **Результат:** говори честно (что помнишь); не выдумывай latency KPI.

## Связь с Event Horizon

EH = правда репо (Shop/Inventory/Billing на Postgres + Outbox). AdTime = narrative про Mongo-чат и AI-pipeline. Не смешивать порты/сервисы EH с DesChat без пометки «в том проекте».
