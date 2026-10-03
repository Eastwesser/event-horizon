Что в плане правильно
A. CardImage — aspect-ratio: 5/7 + cover + rounded. Согласен. Именно это решает визуальные проблемы.

B. Grid alignment — h-full + flex-col + mt-auto для стока и кнопки. Правильно.

C. Detail page:

Сет · Год · Художник в одну строку — правильно.

Художник → /authors/:id — линк.

Card text collapsed — toggle, по умолчанию закрыто.

Скрыть artist_id, market_rub, idempotency_key — правильно, это внутренние поля.

D. Sort / filters / search:

+ element A–Z.

Toggles: в наличии / фойл / нуар / летающие — правильно, как фильтры (не сортировки).

Debounced search ?q= — правильно.

E. Modal preview — 240px, 5/7, cover, rounded. Согласен.

F. Authors:

Поиск по имени.

Плюрализация 1 карта / 2–4 карты / 5+ карт — правильно.

Multi-author — линк на first artist_id + показать полное display name. Разумно.

Top-N + «Показать всех» — правильно, чтобы 108 авторов не растягивали страницу.

G. Detail prev/next:

← → между items в текущем filter+sort list.

sessionStorage или URL list context — правильно, чтобы не терять контекст.

Light swipe на touch.

Browser back не сломан.

H. Perf:

loading="lazy" + decoding="async" — стандарт.

Кэш каталога, page change только slice — правильно.

Report bottleneck после замера.

Единственный нюанс — по A
Ты написал:

Detail hero: cover + 5/7 (or contain if you prefer full art — I'll use cover per your choice)

Уточни Cursor'у: cover для grid + modal, но detail hero — contain?

Причина: на detail пользователь хочет видеть карту целиком — это главный экран карты. cover обрежет края.

Мой выбор:

Grid (сетка) — cover (карты должны заполнять превью, равной высоты).

Modal preview — cover.

Detail hero — contain (пользователь смотрит карту целиком).

Letterbox на detail — ок, потому что фон тёмный и карта в центре — красиво.

В промпте: добавь уточнение.

Если Cursor настаивает на cover везде — принять. Тогда деталь тоже cover — не критично, просто небольшая обрезка по краям.

Финальный ответ Cursor'у
text
OK. Two clarifications:

A. Detail hero — use fit="contain" there (user is looking at
the full card). Grid + modal keep cover. Letterbox on detail
is fine — background is dark.

F. Multi-author rows («А. Кретов. А. Масалимова») — link to
the FIRST artist_id, show the full display string as text.
Report in the UI note that this is temporary until multi-
author pages exist.

Everything else in A–H as written.

Start with A+B. Report files changed. I'll verify before C.
Что я держу в карте
Закрыто:

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно, idempotency verified

Stock fix + idempotency + purchase

4 Icons + grouped attrs ✅

5 Badges ✅

6 Tickets ✅

7 Create modal ×1000 ✅

Пращник fixed ✅

8 Pagination ✅

9 Sorting ✅ (базовая)

10 Artist pages ✅

Текущая волна (visual + UX polish) — план approved:

🟧 A. CardImage (5/7, cover, rounded)

🟧 B. Grid alignment (h-full, mt-auto)

🟧 C. Detail page (Сет·Год·Художник, collapse text)

🟧 D. Sort + filters + search (element, toggles, ?q=)

🟧 E. Modal preview (240px)

🟧 F. Authors (search, плюрализация, multi-author)

🟧 G. Detail prev/next + swipe

🟧 H. Perf (lazy, кэш)

Следующая волна:

🟧 Attribute filters (element / rarity / artist 1-в-1 / foil / stats / icons)

🟧 Расширенный поиск (proberserk-style)

🟧 Noiz-комментарии

🟧 Author registration

🟧 Refund / cancel purchase

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

