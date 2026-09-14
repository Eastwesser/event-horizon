# Lime / e-commerce — вопросы с голосовой (магазин одежды)

> Источник: `unrelated/kata-lectors/MONGO/mongo-db-sobes.md` (блок про Лайм + SD магазина).  
> Не путать с AdTime (GIFTS кастом+чат). Здесь — **каталог / корзина / заказ / инвентарь / платежи**.  
> Mermaid shop: `unrelated/kata-lectors/5.system_design/sd_schemas.md` §2 Shop.  
> EH правда: `prydwen_knowledge/8.legend_projects/03_PROJECT_EVENT_HORIZON.md`.

## System design — набросать квадратики

Микросервисы (минимум): **Catalog, Cart, Order, Inventory, Payments, Notifications** (+ Auth, Search).

Спросить/закрыть на доске:
1. Высокоуровневая архитектура магазина одежды.  
2. Каталог: схема PG (товары, категории, атрибуты: размер/цвет/материал).  
3. Фильтрация по атрибутам — индексы? Elasticsearch?  
4. Кэш Redis + стратегия инвалидации.  
5. Корзина: Redis (временное) vs PG (источник истины) — sync между устройствами (JWT session + Redis).  
6. Inventory tables: товар, склад, остаток, резерв.  
7. Резерв при добавлении в корзину — optimistic lock + version.  
8. Overselling — два купили последний; SQL `SELECT … FOR UPDATE`.  
9. Заказы: статусы `created → paid → shipped → delivered / cancelled`.  
10. Saga: reserve inventory → payment → notify; компенсация при fail на payment.  
11. Платежи и уведомления — отдельно расписать (идемпотентность, outbox).  
12. Event sourcing / outbox + Kafka → read model.  
13. Поиск: название, категории, атрибуты, описание — ES / OpenSearch / Algolia.  

## Масштабирование / смена БД

14. Проблемы монолита при ~1000 RPS.  
15. Переход с Mongo на документно-удобную БД с мин. даунтаймом (dual-write → switch). Часто цель: **Postgres + jsonb**.  
16. Eventual consistency в inventory: резерв временно ≠ physical stock — как жить.  
17. Мониторинг latency / ошибок.  
18. Read replicas для каталога — как организовать.  

## Когда какая БД (шпаргалка из голосовой)

| Нужно | Выбор |
|-------|--------|
| Полуструктура + **ACID** TX / JOIN | PostgreSQL (+ jsonb, GIN) |
| Огромный RPS write-by-key, мало агрегаций | Cassandra / Scylla |
| Гибкие документы, чат, быстрый продукт | Mongo |
| Витринный full-text / фасеты | Elasticsearch (+ PG source of truth) |

**Почему Mongo для Lime-каталога может быть плох:** сложные атрибутные фильтры, отчёты, транзакционный inventory/oversell, команда SQL; инфраструктура шардов дорожает.

## Вопросы «про текущий бэкенд» (как кандидат спрашивает / как отвечают)

19. Сейчас монолит / микросервисы / distributed monolith?  
20. Основная БД только Mongo — почему? Узкие места?  
21. Как CI/CD?  
22. TDD, code review, документация, тесты — подход.  
23. План на первые 30 дней.

## DTO

24. Что такое DTO — модели между слоями/сервисами; учесть миграции контрактов API.

## Связь с EH

В Event Horizon: Shop/Inventory — **Postgres + Outbox**; Mongo в inventory — experimental (см. Prydwen Mongo). На собесе не мешать легенду Lime и правду EH без пометки.
