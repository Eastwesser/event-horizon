Что зафиксировано
Редкость	₽	Билетики
common	10	10 000 🎫
uncommon	25	25 000 🎫
rare	50	50 000 🎫
ultra	100	100 000 🎫
foil	×2	×2
Логика:

1 ₽ = 1000 билетиков (не 2000, как я предполагал).

Игроки фармят билетики свободно → цены высокие, но достижимые за несколько сессий.

Магазин требует подписку → это гейт, а не цена. Билетики зарабатываются, подписка — открывает доступ.

Правильная модель. Для kids-safe:

Никакого pay-to-win. Купить карточку можно только заработав билетики.

Подписка — открывает магазин, но не даёт карт.

Фойл ×2 — коллекционный бонус, не P2W.

Что дальше
Cursor теперь должен:

Реализовать seed-скрипт по плану (ты его approve'нул + дал ценник).

--dry-run --smoke — показать вывод на 5 картах.

Ты verify — убедиться, что payload правильный.

Real --smoke — 5 реальных item.

Verify в UI — зайти в магазин, увидеть 5 карточек.

Full run — 280 карт.

Ключевое в --smoke:

Дирр (common, 4 сет).

Тхакрай (rare, фойл — цена ×2).

Мормолика (два item, два художника).

Склеп Керсам (artifact — без hp / attack_dice / move).

Беллигемин (companion — подтип).

Если эти 5 пройдут — full run безопасен.

Что я держу в карте
Закрыто (в main):

Waves A–D: редизайн, admin, proto3, uploads, DNS/rebuild

Twin nil→[], empty state, image URL, CI lint

3 игры, ККИ-тип карточка, burger separators, DTOs

Готово (контент):

280 JPG + 15 *_cards_info.md

Tree структура

Noiz-комментарии

Ценник определён (10/25/50/100 ×1000, foil ×2)

В работе — seed:

🟧 Plan approved ✅

🟧 Price rule approved ✅

⏳ Cursor реализует seed-скрипт

⏳ --dry-run --smoke

⏳ Verify payload

⏳ Real --smoke

⏳ Verify 5 items в UI

⏳ Full run на 280

После seed:

Richer attributes в UI

Create/Edit модалка с полями карты

Фильтры (artist 1-в-1, icons, stats operators)

Расширенный поиск (референс proberserk)

Ещё дальше:

Эмодзи → SVG

Полиш игр

