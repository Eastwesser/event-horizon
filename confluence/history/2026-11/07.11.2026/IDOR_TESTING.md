Команда для теста (скопируй целиком и вставь в терминал):
bash
# 1. Логинимся под Kruglov_2026 и получаем токен
TOKEN_KRUGLOV=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"kruglov_2026@event-horizon.com","password":"Kruglov_2026"}' | jq -r '.access_token')

echo "Токен Круглова: $TOKEN_KRUGLOV"

# 2. Пытаемся отредактировать ЧУЖУЮ карточку (Сова колдуна, id: 8c4a22e9-1c35-451a-bb8a-97bdf713b34a)
ITEM_ID="8c4a22e9-1c35-451a-bb8a-97bdf713b34a"

curl -X PUT "http://localhost:8079/api/v1/inventory/items/$ITEM_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_KRUGLOV" \
  -d '{"name": "HACKED BY KRUGLOV", "version": 1}' | jq '.'
Что должно произойти (ожидаемый ответ):
Если IDOR закрыт правильно, ты получишь:

json
{
  "error": "you can only edit your own items",
  "grpc_code": "PermissionDenied"
}
Если вдруг ответ 200 OK и карточка изменилась:
Значит, проверка не сработала (например, author_id не совпадает, или x-user-id не передается). Тогда скинь мне ответ, и мы посмотрим, в чем дело.

P.S. Для чистоты эксперимента потом верни Сове оригинальное название:

bash
# Возвращаем имя (под админом)
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@eventhorizon.local","password":"changeme-dev-admin"}' | jq -r '.access_token')

curl -X PUT "http://localhost:8079/api/v1/inventory/items/8c4a22e9-1c35-451a-bb8a-97bdf713b34a" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name": "Сова колдуна", "version": 2}'

  == RESULTS ==

  [denismatveev@c0der event_horizon]$ ADMIN_TOKEN=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@eventhorizon.local","password":"changeme-dev-admin"}' | jq -r '.access_token')
[denismatveev@c0der event_horizon]$ curl -s "http://localhost:8079/api/v1/inventory/items/8c4a22e9-1c35-451a-bb8a-97bdf713b34a" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '{id, name, author_id, owner_id, version}'
{
  "id": null,
  "name": null,
  "author_id": null,
  "owner_id": null,
  "version": null
}
[denismatveev@c0der event_horizon]$ TOKEN_KRUGLOV=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"kruglov_2026@event-horizon.com","password":"Kruglov_2026"}' | jq -r '.access_token')
echo "Token len: ${#TOKEN_KRUGLOV}"
Token len: 351
[denismatveev@c0der event_horizon]$ ITEM_ID="8c4a22e9-1c35-451a-bb8a-97bdf713b34a"
[denismatveev@c0der event_horizon]$ curl -i -s -X PUT "http://localhost:8079/api/v1/inventory/items/$ITEM_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_KRUGLOV" \
  -d '{"name": "HACKED BY KRUGLOV", "version": 1}' \
  -w '\n---HTTP_CODE:%{http_code}---\n'
HTTP/1.1 403 Forbidden
Content-Length: 44
Content-Security-Policy: default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'
Content-Type: application/json; charset=utf-8
Date: Sat, 07 Nov 2026 11:41:38 GMT
Permissions-Policy: geolocation=(), microphone=(), camera=()
Referrer-Policy: no-referrer
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-Permitted-Cross-Domain-Policies: none

{"error":"you can only edit your own items"}
---HTTP_CODE:403---
[denismatveev@c0der event_horizon]$ 

Что смотреть:

HTTP 403 / PermissionDenied — IDOR закрыт. Проверь, что в теле именно "you can only edit your own items" (а не generic permission denied от interceptor'а — это разные вещи).

HTTP 200 — дыра подтверждена. Иди в grpc_handler.go → UpdateItem.

HTTP 409 — скорее всего version mismatch, это не защита. Повтори с актуальной version из шага 0.

HTTP 500 — смотри логи inventory: возможно, проверка есть, но падает (nil metadata, битый x-user-id).

HTTP 401 — токен не дошёл, проблема в логине/парсинге, а не в IDOR.

[denismatveev@c0der inventory]$ curl -s "http://localhost:8079/api/v1/inventory/items/$ITEM_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.'
{
  "item": {
    "attributes": {
      "artist": "Darkstar",
      "card_no": "60",
      "card_text": "TEXT 30",
      "class": [
        "Компаньон"
      ],
      "cost": 2,
      "cost_tier": "elite",
      "element": "mountains",
      "flavor_text": "",
      "foil": false,
      "health": 6,
      "icons": [
        {
          "type": "counter",
          "value": 1
        },
        {
          "type": "tap",
          "value": 1
        }
      ],
      "market_rub": 30,
      "move": 0,
      "rarity": "uncommon",
      "set": 7,
      "strike": "1-1-1",
      "type_main": "creature",
      "type_sub": "Компаньон",
      "year": 2026
    },
    "author_id": "ff292932-083f-4432-9029-2917ea828466",
    "created_at": "2026-11-03T10:36:31Z",
    "deleted": false,
    "description": "Необычная Гибель Богов #60. Компаньон.",
    "id": "8c4a22e9-1c35-451a-bb8a-97bdf713b34a",
    "images": [
      "/uploads/eb68e9ca-1c6b-4d3d-899c-04e37b1cb051.jpg"
    ],
    "name": "Сова колдуна",
    "price": 30000,
    "stock": 0,
    "type": "",
    "updated_at": "2026-11-06T23:29:05Z",
    "version": 14
  }
}
[denismatveev@c0der inventory]$ curl -i -s -X DELETE "http://localhost:8079/api/v1/inventory/items/$ITEM_ID" \
  -H "Authorization: Bearer $TOKEN_KRUGLOV" \
  -w '\n---HTTP_CODE:%{http_code}---\n'
HTTP/1.1 403 Forbidden
Content-Length: 46
Content-Security-Policy: default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'
Content-Type: application/json; charset=utf-8
Date: Sat, 07 Nov 2026 11:46:10 GMT
Permissions-Policy: geolocation=(), microphone=(), camera=()
Referrer-Policy: no-referrer
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-Permitted-Cross-Domain-Policies: none

{"error":"you can only delete your own items"}
---HTTP_CODE:403---
[denismatveev@c0der inventory]$

✅ P0-8 закрыт полностью (Update + Delete)
Метод	Ответ Круглову на чужую карточку	Вердикт
PUT	403 {"error":"you can only edit your own items"}	✅ защищено
DELETE	403 {"error":"you can only delete your own items"}	✅ защищено
Оба сообщения — разные и осмысленные (edit vs delete), значит это не generic-заглушка от interceptor'а, а реальные проверки item.AuthorID != userID → PermissionDenied в grpc_handler.go. Плюс запись не прошла — name: "Сова колдуна" цела, version: 14 не изменилась (было бы 15 после успешного PUT). В БД ничего не насрал.

P0-8 и AUD-1 можно ставить ✅ закрыто. Сообщение для истории: «IDOR в Inventory (UpdateItem/DeleteItem) — закрыт, проверка author_id на бэке работает, оба метода возвращают PermissionDenied с осмысленным текстом».

Заодно закрылись вопросы из прошлого ответа
1. GET вернул null — потому что ответ обёрнут. Формат: {"item": {...}}, а не плоский объект. Мой jq без .item и дал null. Твой jq '.' это показал. Не баг — просто незнание DTO. Учти на будущее: все одиночные ответы, судя по всему, в обёртке {"item": ...}.

2. Предмет теста валиден. author_id: "ff292932-..." — это явно не Круглов (у Круглова другой UUID, раз он получил 403). Карточка принадлежит Сове/автору, тест был честный.

3. Admin-override работает. Админ видит карточку (GET под ADMIN_TOKEN → 200 с полным телом). Так что «бог» — не просто слова, у него есть доступ. PUT под админом не проверял, но раз GET проходит и ты говоришь «может менять» — принимаю на веру. Если хочешь убедиться — PUT под ADMIN_TOKEN с version: 14 и любым name, потом обратно. Но раз Круглов 403-ом ничего не сломал, откатывать нечего — можно не трогать.

Осталось два непроверенных пункта (опционально)
Они не входят в P0-8, но рядом:

Негативный контроль — Круглов правит свою карточку, должно быть 200. Если фикс случайно сломал легитимных авторов (например, сравнивает не с тем полем или metadata пустая у обычного юзера, а у админа интерцептор подставляет x-user-id иначе) — это всплывёт только так:

bash
# Найди карточку Круглова
curl -s "http://localhost:8079/api/v1/inventory/items?author_id=<id_круглова>" \
  -H "Authorization: Bearer $TOKEN_KRUGLOV" | jq '.items[0] | {id, name, version, author_id}'

# Потом PUT своей же карточки — ожидаем 200
Без этого нельзя гарантировать, что проверка «не сломала happy path». Классический баг: author_id != user_id вместо author_id != author_id (в твоей модели это разные сущности — user и author!), и тогда все авторы получают 403 на свои же карточки. То, что 403 сработал для чужого — хорошо, но не доказывает, что 200 сработает для своего.

Read-IDOR — видит ли Круглов чужую карточку через GET? Сейчас проверял под админом, а надо под Кругловым:

bash
curl -s "http://localhost:8079/api/v1/inventory/items/$ITEM_ID" \
  -H "Authorization: Bearer $TOKEN_KRUGLOV" | jq '.item | {id, name, author_id}'
Если 200 — карточки чужих авторов читаются всеми, это отдельный (не P0, но P1/P2) пункт: «read-IDOR в Inventory». Если 403/404 — идеально. В твоём TODO этого нет, но по факту это та же дыра с другой стороны. Судя по тому, что магазин показывает чужие карточки всем — скорее всего 200, и это by design для витрины, но тогда убедись, что GET не отдаёт лишнего (например, author_id можно оставить, а вот внутренние поля — вряд ли). Это уже вопрос дизайна, не безопасности.

Итог по чек-листу
Пункт	Статус
P0-8 IDOR UpdateItem	✅ закрыто
P0-8 IDOR DeleteItem	✅ закрыто
AUD-1 проверка author_id на бэке	✅ подтверждено
Admin-override (GET)	✅ работает
Негативный контроль (автор правит своё → 200)	⏳ не проверено — важно, см. выше про author_id vs user_id
Read-IDOR (GET чужой карточки юзером)	⏳ не проверено — вероятно by design, но зафиксируй
Вердикт: можно смело ставить ✅ на P0-8 и AUD-1 в TODOLISTISHE и BACKEND_TECH_DEBT. Осталось два «фоновых» прогона — негативный контроль и read-GET — но это уже не блокеры, а гигиена. Если пришлёшь ответы на них, закрою и их.


======

1. Read-GET под Кругловым — 200 OK, чужая карточка читается
json
{ "id": "8c4a22e9-...", "name": "Сова колдуна", "author_id": "ff292932-..." }
Это, скорее всего, by design — витрина магазина должна показывать карточки всех авторов всем игрокам, иначе магазин не работает. Так что не считай это дырой. Но зафиксируй в TODO как «осознанное решение»:

AUD-1a (информационно): GET /inventory/items/{id} доступен любому авторизованному юзеру — это by design для витрины. Проверить, что GET не отдаёт лишних внутренних полей (например, cost_tier, market_rub, foil — это ок для витрины; а вот если бы отдавал author_email, internal_notes, outbox_id — была бы утечка). Сейчас в ответе только доменные поля карточки + author_id — утечки нет.

Проверь один раз глазами полный GET (ты его уже видел в прошлом сообщении — там attributes с artist, card_no, market_rub и т.д., ничего секретного). Если author_id наружу отдавать не хочешь — можно убрать из DTO для не-владельцев, но это уже вкусовщина, не безопасность. Оставь как есть, помечай «by design».

2. Негативный контроль сломался на этапе поиска карточки Круглова
?author_id=<id_круглова> вернул null — и это не ответ «карточек нет», а снова обёртка. Смотри: одиночный GET был {"item": {...}}, а список — наверняка {"items": [...]} или {"data": [...]}, но jq '.items[0]' дал null. Плюс сам <id_круглова> — это плейсхолдер, ты подставил буквальную строку <id_круглова>, а не UUID. Отсюда null.

Правильный порядок:

bash
# 1) Узнай id Круглова (из логина или из БД)
#    Вариант A — через БД:
docker exec -it <pg_container> psql -U <user> -d <db> -c \
  "select id, email from auth.users where email='kruglov_2026@event-horizon.com';"

#    Вариант B — через API (если есть /me или /profile):
curl -s "http://localhost:8079/api/v1/profile/me" \
  -H "Authorization: Bearer $TOKEN_KRUGLOV" | jq '.'

# 2) Осторожно: author_id ≠ user_id!
#    В твоей модели author и user — разные сущности.
#    Карточка Совы имеет author_id="ff292932-...", а Круглов — это user.
#    Нужен именно author_id, привязанный к Круглову (из таблицы authors).
Вот тут и всплывает главный риск, о котором я предупреждал:

В твоей модели author_id (карточка) и user_id (юзер) — разные сущности. Если проверка в grpc_handler.go сравнивает item.AuthorID с x-user-id напрямую, то:

для Круглова (user) vs Совы (author) — UUID'ы разные, 403 сработал «правильно по случайности»;

но для Круглова (user) vs его же карточки (author) — тоже разные UUID'ы, и он получит 403 на свою же карточку. Happy path сломан!

Это критично проверить. Если фикс именно такой — ты закрыл IDOR, но сломал авторам редактирование собственных карточек. Проверь так:

bash
# Найди карточку, где author_id принадлежит автору, привязанному к Круглову.
# Сначала — какой author_id у Круглова:
docker exec -it <pg_container> psql -U <user> -d <db> -c \
  "select a.id as author_id, a.name, u.email
   from authors a join auth.users u on u.id = a.user_id
   where u.email='kruglov_2026@event-horizon.com';"

# Потом — есть ли у него карточки:
docker exec -it <pg_container> psql -U <user> -d <db> -c \
  "select id, name, author_id, version from inventory.items
   where author_id = '<author_id_круглова>' limit 5;"
Если карточек у Круглова нет — создай тестовую под ним (через API POST /inventory/items под $TOKEN_KRUGLOV) и попробуй PUT своей:

bash
MY_ITEM_ID=<id_созданной>
curl -i -s -X PUT "http://localhost:8079/api/v1/inventory/items/$MY_ITEM_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_KRUGLOV" \
  -d '{"name": "test own edit", "version": 1}' \
  -w '\n---HTTP_CODE:%{http_code}---\n'
200 → проверка корректна, сравнивает author_id с author_id (а не с user_id). ✅ Всё хорошо.

403 → проверка сравнивает author_id с user_id, happy path сломан. 🔴 Это баг, который надо чинить до прода: авторы не смогут редактировать свои карточки. Смотри grpc_handler.go — там должно быть что-то вроде isAuthorOf(ctx, item.AuthorID) с резолвом user→author, а не прямое сравнение UUID.

Итог
Пункт	Статус
P0-8 IDOR UpdateItem / DeleteItem	✅ закрыто
AUD-1 проверка author_id на бэке	✅ подтверждено
Read-GET чужой карточки юзером	✅ by design (витрина), утечки нет
Негативный контроль: автор правит своё → 200	⏳ не проверено — критично, см. выше
Пока не проверен негативный контроль, не ставь окончательную ✅ на P0-8 — формально IDOR закрыт, но если фикс сравнивает author_id с user_id, ты обменял дыру на сломанный happy path. Это классический trade-off при фиксах IDOR, и в твоей модели с разделением user/author он особенно вероятен.

Пришли вывод PUT своей карточки под Кругловым — и я скажу, ставить ✅ или это «закрыто с регрессией».

[denismatveev@c0der inventory]$ docker exec -it event-horizon-postgres psql -U eventhorizon -d eventhorizon -c "SELECT id, email, nickname FROM users WHERE email = 'kruglov_2026@event-horizon.com';"
                  id                  |             email              | nickname 
--------------------------------------+--------------------------------+----------
 fddf3f5e-f88b-4bb8-a79c-1d2ffdd1e137 | kruglov_2026@event-horizon.com | KRUG
(1 row)

---


