# WHAT TO DO IRL

## Уже сделано

- [x] Код v1.1.0 + `/api/v1` + seeds scripts + CORE k6  
- [x] Boosty done (`BOOSTY_DONE/`)  
- [x] Miro PNGs + `MIRO_STICKERS_PASTE.md` / `MIRO_REVIEW.md`  
- [x] Leftovers wave: avatar, LB seed, card-artists authors, Tamagotchi tickets, record-beaten deep-link  

## Тебе осталось

1. **Miro stickers** — вставь routes/ports из `miro/MIRO_STICKERS_PASTE.md` (MCP стрелка Cursor→MCP, Outbox только где есть)  
2. После rebuild gateway/lb/notification:
   ```bash
   make seed-lb-demo
   make seed-card-artists   # нужен inventory с карточками
   ```
3. Опционально: ranked-игра + `git tag v1.1.0`  

## Parked (не делаем сейчас)

См. `PARKED_WISHES.md` — C4, 3D/Dodo/Sims/flower, crop UI, 108 login-авторов.
