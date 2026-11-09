# TODO List — 09.11.2026 (v1.1.3.6)

Агрегированный список задач после сессии 08.11–09.11.
**P0-10 и P0-11 закрыты.** Баги `PROTOJSON` и `UpdateItem` зафиксированы в постмортемах.

## 🔥 Сделано в этой сессии (зафиксировать)
- ✅ **P0-10:** `UpdateItem` мержит поля. `attributes`, `images`, `price`, `stock`, `author_id` не обнуляются при частичном PUT.
- ✅ **P0-11:** `AdminCancelPurchase` обходит `fulfilled_at`. Админ может отменить товар, который уже «собран» (проверено вживую: 200 OK, `REFUNDED` в БД).
- ✅ **P0-9 (частично):** `AdminCancelPurchase` в Gateway и Shop работает, идемпотентность подтверждена.
- ✅ **Protojson:** `google.protobuf.Struct` не бывает `nil` — исправлен мерж `attributes` (`len(req.Attributes.AsMap()) > 0`).
- ✅ **Тесты:** `make test-all` зелёный (обновлён мок `RefundPurchase`).
- ✅ **Makefile:** починен `test-smoke` (сломанный цикл `for`).

---

## 🟥 P0 — Осталось (Блокеры)
| # | Сервис | Задача | Статус |
|---|--------|--------|--------|
| **P0-1** | Fulfillment | Transactional Outbox (событие теряется между `paid` и `fulfilled`) | [ ] |
| **P0-2** | Fulfillment | Заменить `time.After(10s)` на WMS API | [ ] |
| **P0-3** | Gateway | Разбить `runGateway` (God Function) на модули | [ ] |
| **P0-4** | Shop | `resumeSagas` не берёт `started > 60s` (деньги списаны, нет компенсации) | [ ] |
| **P0-5** | Billing | `RewardConsumer` использует `time.Now()` → двойное начисление. Использовать `event.ReferenceID` | [ ] |
| **P0-6** | Authors | `RevertApplication` hard delete → FK падает. Soft delete + check FKs | [ ] |
| **P0-7** | Balancer | Single point of failure. 2 Balancer + VIP/DNS (Nginx уже есть) | [ ] |
| **P0-8** | Inventory | IDOR: `UpdateItem`/`DeleteItem` — `checkAuthorAccess` есть, но проверить в `DeleteItem` | [ ] |
| **P0-9** | Shop | `paymentClient` может быть `nil`. Метрика `merch_gate_disabled` | [ ] |

---

## 🟧 P1 — Надёжность / Безопасность
(23 задачи — краткий список)
- **P1-1..2 (Gateway):** Два Hub (дублирование), Async gRPC без CB в `admin/users`.
- **P1-3..5 (Inventory):** Два драйвера (PG+Mongo) — оставить PG. `toPostgresFilter` без `fmt.Sprintf`. Outbox type assertion → явный интерфейс.
- **P1-6..7 (MCP):** SQL Sandbox (string matching) → read-only роль. Утечка `err.Error()`.
- **P1-8..10 (Balancer):** `/ready` не проверяет бэкенды, Static registry, нет health checks.
- **P1-11..12 (Notification):** Poison message (Ack+DLQ), Fan-Out без rate limit.
- **P1-13..15 (Analytics/Leaderboard):** `fmt.Sprintf` db name, `RestoreFromPostgres` (retry), нет rate limit на `record_beaten`.
- **P1-16..17 (Payment/History):** Нет сверки `amount` webhook, `RetentionWorker` в процессе.
- **P1-18..19 (Auth):** Нет ротации `JWT_SECRET`, reuse detection для refresh.
- **P1-20..23 (Shop/Общее):** `kafkaProd` не используется, `cancelPurchase` может задвоить, single-flight на кэш, WAF.

---

## 🟨 P2 — Гигиена / UX (Бэкенд)
- **P2-1..3:** `log.Printf` → `slog` (Gateway, Balancer), пустые пакеты (Balancer).
- **P2-4..6:** `newUUID` vs `uuid.NewString`, `redis-game` anti-cheat (INCR, SETNX, GET), `GetBalance` `-1` при miss.
- **P2-7..9:** `boosty.to` в Config, `subscriptionDays` в config, UX gap (JWT старый после approve).
- **P2-10..12:** 5 субъектов в 1 consumer (Analytics, History), Retry Fixed Delay (Telegram → exponential).
- **P2-13..16:** `UpdateProfile` наружу (без `total_score`), Gin routes (тест), `make seed-v110` заметка, ник в профиле (фронт-кэш).

---

## 🟨 P2 — Фронтенд: Карточки, Мобилки, Auth UX
- **FE-CARD-1..4:** Стоимость (cost), Ход (move), Бейдж «Карточка», Класс (class).
- **FE-MOB-1..4:** Горизонтальный скролл, Отступы, Центрирование модалок, Свайпы (2048/Pancaker).
- **FE-AUTH-1..3:** Автозаполнение, Глазик (пароль), Смена пароля (сценарий auth).

---

## 🟧 P1 — Новые сценарии (Модерация, История, Кейсы)
- **MOD-1..5:** Уведомления автору при удалении/редактировании, причина (soft-delete), кнопка «Дать/Отозвать подписку».
- **HIST-1..5:** Починить историю на фронте, 10 последних записей на пользователя, статистика, служебная инфа админа.
- **CASE-1..3:** 30 пицц (алерт), Бабуля натыкала (идемпотентность), Два игрока купили последнюю карту (Saga compensating).

---

## 🟨 P2 — Админка: Статистика инвентаря
- **ADMIN-INV-1:** Переименовать («Уникальных позиций», «Общее кол-во»).
- **ADMIN-INV-2:** Починить «По типу» (4 карточки с пустым типом).
- **ADMIN-INV-3:** Починить «По автору» (108 vs 11, дедупликация).

---

## 🟦 P3 — Roadmap (не трогаем)
i18n (RU/EN/DE/ES), Пагинация каталога, Уровни LB 2–3, Порог покрытия CI ≥70%, Retry с джиттером, 108 авторов (Locked), 3D (Locked), Helm (Locked).

---

## 🟩 P4 — Парковка (не трогаем)
Loki, Tempo, node-exporter, SLO burn, Consul, Alertmanager → Telegram, WAF/пентест, C4 payouts, Miro-стикеры.

---

## 🐛 BUGS FOUND (зафиксировано)
- **PROTOJSON_BUG:** `google.protobuf.Struct` не бывает `nil` (исправлено в `UpdateItem`). См. `confluence/postmortem/PROTOJSON_BUG.md`.
- **BUG-2026-11-08-2:** `price=0` неотличим от «не передали».
- **BUG-2026-11-08-3:** `stock=0` неотличим от «не передали».
- **BUG-2026-11-08-6:** `Images` в `UpdateItem` — проверить, приходит ли `[]` вместо `nil` (затирание).