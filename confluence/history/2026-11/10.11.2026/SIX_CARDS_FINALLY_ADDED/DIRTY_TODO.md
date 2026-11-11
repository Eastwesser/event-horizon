A. Засев (backend / data)

A1. Крысиная нора привязана к author_id=1502a3fa… (Event Horizon), а должна — к kruglov_2026 (fddf3f5e…). Нужен корректный author_id при засеве.

A2. Сова колдуна была market_rub: -1 — validateItem теперь отклоняет, но старые данные остались. Нужен «санитарный» SQL или миграция.

A3. Дубли Гронда (soft-deleted) — чистка deleted_at IS NOT NULL.

A4. authors count = 11, а в админке «По автору» — 108. Расхождение — или в админке джойн не тот, или в authors есть active=false. Проверить SELECT active, count(*) FROM authors GROUP BY active;.

A5. «Дважды один автор» (скрин 8) — дубль в authors (например, Event Horizon + kruglov_2026 на один user_id).

B. Фронтенд (UX карточки)

B1. Скрин 2 — «Введите URL!» при попытке сохранить с загруженной картинкой. Значит, image_url не подставляется в форму после POST /api/v1/uploads.

B2. Скрин 3 — нет интуитивной кнопки «Сохранить картинку». Нужен явный CTA.

B3. Скрин 1 — «Фронтенд добавления товара всратый». Общий UX формы.

B4. Скрин 5 / 11 — «Как оформить текст карточки как на скрине?» — нужен структурированный рендер attributes.card_text (сейчас, вероятно, plain text или [object Object]).

B5. Скрин 6 — у авторов нет профиля / портфолио. Отдельная фича.

B6. Скрин 12 — «Не получилось сохранить картинку даже админу». Тот же баг, что B1, но для admin.

B7. Скрин 13 — «После изменения товара админом его как будто стёрло — потом появляется». Гонка UpdateItem → Cache Invalidation → shop-inventory-sync. Нужно проверить порядок событий и TTL кэша.

B8. Скрин 14 — «Крысиная нора по фильтрам не находится». Фильтр по атрибутам не матчит element=woods / class=Постройка. Проверить toPostgresFilter и card_text-парсинг.

B9. Скрин 9 — «Пишет, что карт нет, а Круглов создал одну». Возможно, тот же author_id mismatch (A1).

B10. Скрин 7 — «Ник сохраняется в лидерборде, а в профиле — нет». event.user.nickname.updated → profile-nickname-updated consumer. В логе NATS он есть (profile-nickname-updated, last delivery 13m14s). Значит, событие летит, но Profile не обновляет. Проверить UpdateNickname в Profile.

B11. Скрин 10 — «Крысиная нора в магазине есть» — это OK, но в связке с B8/B9 даёт «карта есть, но не находится» → фильтры.

C. Покупка / деньги

C1. После покупки карточки — проверить, что billing.transactions получает -price с reference_id=shop:user:item:cycle, а revenue_splits — 85/15. По твоему логу: списание 25000 прошло, но revenue_splits в payment я в дампе не вижу — надо проверить.

C2. «Билетики за Gears не начислились» — по history.events видно score.updated с tickets_earned: 60 для Gears. Значит, событие есть. Проверить billing.transactions для этого user_id и reference_id (в логе — ...-tickets-1794000027). Если нет — RewardConsumer/Billing не обработал.

D. Наблюдаемость

D1. billing-durable и nats-hub-listener имеют Unprocessed: 5 — что-то не вычитывается. Проверить, какие subjects.

D2. nats stream info показал Number of Subjects: 2 при списке 15 — это норм (только с сообщениями), но balance.updated в списке subjects есть — значит, после «встряски» ок.