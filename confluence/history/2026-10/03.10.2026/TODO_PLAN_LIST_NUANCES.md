Что Cursor сделал
Создал tickable debt list:

text
confluence/architecture/FINAL_DETAILS.md/TODO_FINAL_PRE_PROD_DETAILS.md
Содержит:

5 волн (Wave 1–5) с - [ ] чекбоксами.

C1–C4 под Author registration.

Progress table внизу (Done / Total).

Ссылки из README «Планы» + docs table.

Как отмечать: [ ] → [x] в этом файле.

Плюс: README обновлён — блок «Планы на следующие спринты» встроен.

Это правильно. Теперь у тебя один файл — маршрутная карта. Не надо держать в голове — открыл, посмотрел, отметил.

Мой разбор файла
Что хорошо
Структура — чистая:

Axes — complexity / risk / dependency.

Wave 1–5 — по приоритету.

Warm-up order — рекомендованный порядок.

Progress table — Done / Total.

Notes — правила (не смешивать Author reg, etc.).

Детализация — правильная:

Каждая задача — XS–XL, risk, ~время.

C1–C4 — под-волны, не монолит.

Dependency — где явно («Before 1000+», «Before C4»).

Ссылки — есть:

Из README.

Из docs table.

Ссылка на STILL_TECH_DEBT.md.

Что стоит уточнить
1. Wave 5 #1 — «Bottleneck fix» — dependency.

Depends on thin DTO story.

Правильно, что указана. Но — /shop/items thin DTO — это Wave 2 #1, другая задача. Возможно, надо явно:

text
Bottleneck (Wave 5) depends on:
  - Wave 2 #1 (/shop/items thin DTO) — DONE
  - Catalog pressure — 500+ cards
2. Wave 4 — «OpenAPI/Swagger для всех сервисов».

M · low · Docs across mesh

Что тут «для всех сервисов»? Ты уже имеешь gateway OpenAPI (docs/openapi.yaml). Для gRPC-сервисов — отдельный реестр? Или через gateway?

Стоит уточнить у Cursor'а — что именно.

3. Wave 2 — «Полиш игр — 5 штук».

~1 week if parallel.

«Parallel» — с чем? С другими задачами? Или 5 игр параллельно?

Логичнее: последовательно — одна игра = один небольшой PR. Не параллельно.

4. Wave 3 — «Author registration».

Before C4: decide monetization — what does an author get? (% of sales?)

Ключевой вопрос. Пока не решён — C4 не начать. C1–C3 — можно без монетизации.

5. Progress table.

*1 Quick wins: 0/6*
2 Content + debt: 0/6 (+ 0/5 games)
3 Author reg: 0/4 phases
4 Infra: 3/8 (CI/Ansible/k3s done)
5 Long term: 1/3 (NATS done)

Хорошо. Но — где задачи, которые уже сделаны в v1.0.9? Из этого списка — ничего, потому что Wave 1–5 — будущее. Ок.

Мой ответ Cursor'у
text
TODO_FINAL_PRE_PROD_DETAILS.md is clean. Three small
clarifications:

1. Wave 5 #1 Bottleneck — spell out the dependency:
     depends on Wave 2 #1 (/shop/items thin DTO) AND real
     catalog pressure (say, 500+ cards). Until then, defer.

2. Wave 4 — «OpenAPI/Swagger для всех сервисов» — what
   exactly? We already ship gateway OpenAPI (docs/
   openapi.yaml). Options:
     a) Keep gateway-only, no change.
     b) Publish per-service .proto as docs (no Swagger UI).
     c) Full per-service OpenAPI (grpc-gateway style).
   Pick one and note it.

3. Wave 2 games polish — not parallel across games. Do one
   game per PR, sequentially. Update the file.

Everything else stays.

Also: should there be a short note at the top — «Update the
Progress table at the bottom when ticking boxes»? Currently
buried in Notes.

DO NOT TOUCH code. Only this file + README link.
Что я держу в карте
Закрыто (v1.0.9 на GitHub, bd5b5ee):

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно

Stock fix + idempotency + purchase

4 Icons + grouped attrs

5 Badges, 6 Tickets, 7 Create modal ×1000

Пращник fixed

8 Pagination, 9 Sorting, 10 Artist pages

A–H visual + UX polish

1→4 final polish

Attribute filters ✅

1→3 polish ✅

Element fix ✅

Bottleneck — smaller list DTO ✅

Refund / cancel purchase ✅

RBAC fix ✅

Noiz reviews ✅

Data fixes ✅

REVIEW_RESULT.md ✅

README + CHANGELOG v1.0.9 ✅

Маршрутный лист создан:

🟧 confluence/architecture/FINAL_DETAILS.md/TODO_FINAL_PRE_PROD_DETAILS.md

🟧 5 волн, C1–C4, progress table.

🟧 Ссылка из README + docs table.

Следующая волна (по плану):

🟧 Wave 1 — Quick wins (XS–S): PUT 403, Emoji → SVG, Retry + jitter, Alerts Telegram, CB + Bulkhead, Rate limiter.

Долгосрочно:

🟧 Wave 2 — Content + debt (M)

🟧 Wave 3 — Author registration (XL, C1–C4)

🟧 Wave 4 — Infra (M-L)

🟧 Wave 5 — Bottleneck + observability (L)

Что делать сейчас
1. Отправь Cursor'у 3 уточнения (промпт выше) — он поправит файл.

2. Потом — определись с Wave 1.

PUT 403 — 30 минут, можно сейчас.

Emoji → SVG — 1-2 дня, визуально заметно.

Retry + jitter — 1 день, стабилизирует.

Alerts Telegram — 1 день, ops.

CB + Bulkhead — 1 день, resilience.

Rate Limiter — 1 день, защита.

Wave 1 = «разогрев». Каждая задача независима. Можешь выбрать любую.

3. Мой совет: PUT 403 — первым. 30 минут, закрывает подозрение. Потом — Emoji → SVG, потому что визуально и для kids важно.

