# MongoDB + AdTime DesChat — вопросы с собеса / голосовой (2026-09-08)

> Источник: `unrelated/kata-lectors/MONGO/mongo-db-sobes.md` (транскрипт + схема чатов).  
> Легенда: AdTime / GIFTS.ru — кастом мерча + **чат с дизайнерами** (WS), Mongo под сообщения.  
> Схемы SD (Mermaid): уже есть → `unrelated/kata-lectors/5.system_design/sd_schemas.md` §1 AdTime, §5 Messenger.  
> Теория Mongo (EH): `agents/prydwen_knowledge/2.data_bases/05_DB_MONGODB.md`.

## Схема коллекций (шпаргалка)

**chats**
```js
{
  _id: ObjectId,
  participants: [{ userId, lastReadAt }],
  lastMessage: { text, userId, createdAt },
  createdAt: Date,
  updatedAt: Date
}
```

**messages**
```js
{
  _id: ObjectId,
  chatId: ObjectId,
  userId: ObjectId,
  text: string,
  attachments: [...],
  createdAt: Date
}
```

### Индексы

| Назначение | Ключ |
|------------|------|
| Пагинация сообщений | `{ chatId: 1, createdAt: -1 }` |
| Фильтр по автору в чате | `{ chatId: 1, userId: 1 }` |
| Список чатов пользователя | `{ "participants.userId": 1, updatedAt: -1 }` (или `participants: 1, updatedAt: -1` как в конспекте) |
| TTL архив | TTL на `messages.createdAt` (в конспекте 90 дней; в SD-схемах иногда 30d — зафиксируй одну цифру в легенде) |

### Пагинация (агрегация)

```js
db.messages.aggregate([
  { $match: { chatId: ObjectId("...") } },
  { $sort: { createdAt: -1 } },
  { $skip: 0 },
  { $limit: 20 },
  { $lookup: {
      from: "users",
      localField: "userId",
      foreignField: "_id",
      as: "user"
  }}
])
```

На проде чаще **cursor** (`createdAt + _id`), не большой `$skip`.

---

## Вопросы интервьюера → что отвечать (каркас)

### A. Индексы и поиск

1. **Какие индексы у мессенджера?** → таблица выше + зачем left-prefix.  
2. **Как полнотекстовый поиск?** → Mongo text index / Atlas Search; для витрины товаров чаще Elasticsearch/OpenSearch проекция; для чата — редко full-text на всём архиве.  
3. **Список чатов пользователя** → индекс по participant + `updatedAt`; денормализованный `lastMessage` в `chats`.  
4. **explain** — COLLSCAN vs IXSCAN; составной индекс под `chatId + createdAt`.

### B. Транзакции в чате (отправка сообщения)

Поток из голосовой (multi-doc session, с 4.0+; сейчас ориентир 7.x/8.x):

1. `startSession` / `WithTransaction`  
2. Insert в `messages`  
3. Update `chats.lastMessage` + `updatedAt`  
4. Increment unread counters у participants  
5. Commit  

**Вопросы:**
5. **Как атомарно обновить `lastMessage` при новом сообщении?** → одна TX (replica set) или update chat + insert message в TX; показать псевдокод `session.WithTransaction` (Go mongo-driver).  
6. **Где в AdTime/DesChat нужны транзакции?** → сообщение + lastMessage + unread; не на каждый read receipt если eventual ок.  
7. **Проблемы TX в шардированном кластере?** → multi-shard TX дороже/ограничения; shard key = `chat_id` чтобы сообщение и метаданные чата реже cross-shard; retryable writes; latency.  
8. **Оптимистичная блокировка participants** (два юзера не добавятся гонкой) → `version` / `expected version` в update filter; или unique на пару.

### C. Рост данных / эксплуатация

9. **Шардирование** — ключ `chatId` (равномерность vs hot celebrity chat).  
10. **TTL + archive collections** — горячие N дней в hot, холод в archive / S3.  
11. **Медленные запросы** — explain, составные индексы, проекции, не тащить огромный `$lookup` без нужды.  
12. **Конкурентные обновления lastMessage** — last-write-wins по `createdAt` или TX.

### D. Миграции с Mongo

13. **Когда уходить с Mongo на PostgreSQL (+ jsonb)?** → нужны JOIN/репорты, строгие multi-entity TX, команда SQL-first; jsonb+GIN закрывает полуструктуру.  
14. **Как переписать схему чата на PG + jsonb metadata?** → `chats`, `messages` таблицы; metadata jsonb; индексы btree + gin.  
15. **Zero-downtime dual-write → switch?** → dual-write оба стора → backfill → read shadow → cutover → stop mongo writes.  
16. **Какие индексы на jsonb?** → GIN / expression indexes на частые пути.  
17. **Cassandra/Scylla vs Mongo** — Discord path: write-heavy timeline по partition key; когда выбрать C*/Scylla (линейный scale по ключу, без Mongo-style flexible ad-hoc) vs Mongo (документы, гибкость, проще secondary indexes в умеренных объёмах).  
18. **Пагинация в Cassandra без offset** — clustering key + paging state / `token` / `created_at` seek.  
19. **Почему в C* нет JOIN** — денормализация + multiple tables per query.

### E. Сторителлинг легенды (speak aloud)

20. **Как строил DesChat на Mongo?** — WS fanout (Redis PubSub), коллекции, индексы, TTL, проблемы медленных запросов, optimistic locking.  
21. **Почему Mongo ок для чата AdTime, но для Lime-каталога может быть плохо?** — каталог + сложные фильтры/атрибуты/транзакционный inventory → PG + ES; чат — append messages + денорм lastMessage.  
22. **Какая версия Mongo была / актуальна?** — в конспекте образ `7.0.5`; на собесе: «в проекте 7.x, рынок 8.x — multi-doc TX с 4.0». Не врать цифры KPI.

---

## Псевдокод Go (TX отправки)

```go
err := client.UseSession(ctx, func(sc mongo.SessionContext) error {
  return mongo.WithSession(sc, func(sc mongo.SessionContext) error {
    _, err := messages.InsertOne(sc, msg)
    if err != nil { return err }
    _, err = chats.UpdateOne(sc,
      bson.M{"_id": chatID},
      bson.M{"$set": bson.M{
        "lastMessage": bson.M{"text": msg.Text, "userId": msg.UserID, "createdAt": msg.CreatedAt},
        "updatedAt": msg.CreatedAt,
      },
      "$inc": /* unread for others */},
    )
    return err
  })
})
```

(Уточни API driver под свою версию; идея — одна session/TX.)

---

## Связанные файлы

| Куда | Файл |
|------|------|
| Prydwen кратко | `…/2.data_bases/05_DB_MONGODB.md` |
| Lime e-com вопросы | `../1.architecture/lime_ecommerce_sobes.md` |
| Mermaid AdTime/Messenger | `unrelated/kata-lectors/5.system_design/sd_schemas.md` |
| Raw | `unrelated/kata-lectors/MONGO/mongo-db-sobes.md` |
