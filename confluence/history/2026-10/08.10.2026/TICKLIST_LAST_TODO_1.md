# Ticklist — last TODO map (08.10.2026)

**Rule:** `- [ ]` → `- [x]` when closed. Do **not** mutate real Berserk CCG merch without Denis OK.

**Order (Denis):**  
1. Repo cleanup → 2. Site / chrome / shop → 3. Games polish → 4. Mobile adaptive → 5. Load / metrics → 6. Interview docs + Miro → 7. MCP / Tetiva (**very last**)

**Overlap:** Track B (06.10) closed many game/boost items — re-verify in UI; re-open if regression.

---

## Sources audited (07.10.2026) — full re-check 08.10

Every file below was read end-to-end for this ticklist. Tick when you personally re-skim.

- [x] `CLEANUP.md` — scripts/Dockerfile hygiene; mind CI/Makefile refs
- [x] `MOBILE_ADAPTIVE.md` — smartphones **after** FE cleanup
- [x] `MCP_WHEN.md` — MCP only if needed; **last**
- [x] `BOOSTY_FIX.md` — https://boosty.to/eastwesser · tier ×2 + v1.1.0 content
- [x] `FUTURE_TODO_1.md` — parked wishlist after Tracks A–D
- [x] `INSIGHTS/1.IDEAS/VOICEMESSAGE_INSIGHTS_1.md` — interview / patterns / SQL / security / cleanup / services / achievements / Flappy cursor idea / Gears flower skin
- [x] `INSIGHTS/2.SHOP/VOICEMESSAGE_INSIGHTS_2.md` — home / LB / profile / shop / admin / authors / subs / history (full voice dump)
- [x] `INSIGHTS/2.SHOP/FEEDBACK_VoiceM_2_1.md` — written v1.1.0 site report + F12 bugs home
- [x] `INSIGHTS/3.GAMES/VOICEMESSAGE_INSIGHTS_3.md` — per-game deep dive (Pancaker→Companion)
- [x] `INSIGHTS/3.GAMES/FEEDBACK_VoiceM_3_1.md` — written v1.1.0 **games** report + F12
- [x] `INSIGHTS/3.GAMES/LEADERBOARD_LOGS.md` — spam `GET /leaderboard undefined`
- [x] `BUGS_FEEDBACKS/old_console_log.md` — `/api/api/profile` 404, achievements crash, notifications 401, submit 401, billing 401, API spam, VOID mount
- [x] `LOAD_TESTS/HIGHLOAD_TESTS.md` — CORE highload suite home
- [x] `LOAD_TESTS/METRICS.md` — targets (10k DAU model, latency, Go/DB/Grafana watchlist, Selectel ~$430)
- [x] `SYSTEM_DESIGN_MIRO/FINAL_SYSTEM_DESIGN_MIRO_SCHEME.md` — v1.1.0 pinnacle + **pull Miro image**
- [x] `SYSTEM_DESIGN_MIRO/INTERVIEW_QUESTIONS_EH.md` — where/ports/how
- [x] `INSIGHTS/3.GAMES/eh-screenshots/` — **80 files**, 12 folders (01–12) — visual proof for voices (audited 08.10)

**External (Denis links, 08.10):**
- Miro board: https://miro.com/app/board/uXjVJLLg9us=/
- Cursor Miro marketplace: https://cursor.com/marketplace/miro
- Boosty: https://boosty.to/eastwesser
- Tetiva (optional API client): see §17 — **not required for EH product**

---

## Screenshot evidence (`eh-screenshots/`) — voice ↔ pixels

Path base: `07.10.2026/INSIGHTS/3.GAMES/eh-screenshots/`  
Use these when implementing / QA; filenames are Denis’s presentation order.

### 01.mainpage
| Shot | What it proves | Ticklist |
|------|----------------|----------|
| `EVENT HORIZON MAINPAGE1.png` | Hero title sits high vs horizon center | §5 lower title |
| `Event_Horizon_yellow_letters.png` | Navbar yellow brand looks clickable, no About | §5 brand → About |
| `EVENT HORIZON MAINPAGE2.png` | Games grid; Builder blurb = «падающих блоков»; Void + CTAs tight | §5 blurbs + spacing |
| `EVENT HORIZON MAINPAGE3.png` | Lower home / footer area | §5 polish |
| `ADMIN_BELL.png` | Bell notifications OK (admin) | §9 keep bell |
| `ADMIN_BURGER.png` | Burger: inventory / history / authors / subs / admin | chrome OK |

### 02.leaderboard
| Shot | What it proves | Ticklist |
|------|----------------|----------|
| `LEADERBOARD_1_pancaker.png` | #1 nick `9d0ebedb` with **0** pts; tab still **Компаньон** | §5 LB zeros + EN name |
| `LEADERBOARD_2_flappy.png` … `8_tamagotchi.png` | Per-game LB chrome; thin ranks / test nicks | §5 seed nicknamed players |

### 03.profile
| Shot | What it proves | Ticklist |
|------|----------------|----------|
| `PROFILE_1.png` / `PROFILE_2.png` | Profile layout; scores zero; currency pills | §5 profile zeros |
| `PROFILE_3_NEED_TO_FIX_AND_CHANGE_AVA.png` | Browser `confirm` for nick; grid still **Орбиты** / **Компаньон**; no avatar upload | §5 nick UI + rename + ava |

### 04.shop
| Shot | What it proves | Ticklist |
|------|----------------|----------|
| `SHOP_1.png` | Badge / brelok / fenechka / C3 at **100** tickets (should be 100_000 physical) | §6 price floor |
| `SHOP_2`–`3` | Cards grid OK | keep |
| `SHOP_4.png` | **Темы** empty: «Нет товаров…» 0/284 | §6 seed themes |
| `SHOP_5`–`8` | Skins empty / merch / more catalog | §6 skins seed |
| `FILTER_BUG.png` | **Фенечка** selected but full **card** filters (стихия/класс/редкость) | §6 card filters only on Cards |
| `FILTER WORKS.png` / `SHOP_SORT_FILTER.png` | Sort/filter when OK | keep sort |
| `INVENTORY_DUPLICATES.png` | Rainbow pipes ×N different dates | §6 dedupe |
| `INVENTORY_PURCHASE_CANCEL.png` | Copy: «**Карта** уйдёт из инвентаря» | §6 «предмет будет удалён» |
| `INVENTORY*.png` | Cancel flow works | keep mechanic |

### 05.admin_panel
| Shot | What it proves | Ticklist |
|------|----------------|----------|
| `ADMIN_PENDING_1_USERS.png` | Users + subscription column | §7 sub UX |
| `ADMIN_PENDING_2`–`4` | Pending / approved / rejected apps | keep tabs |
| `ADMIN_INVENTORY.png` | 284 goods · **4 types** · **2 authors** · Top-5 with **₽** | §7 types/authors/tickets icon |
| `ADMIN ANALYTICS.png` | DAU/MAU/retention raw | §7 explain D0–D7 |

### 06.catalog
| Shot | What it proves | Ticklist |
|------|----------------|----------|
| `ALL_AVAILABLE_MERCH.png` | Type chips miss **Мерч**; foil/noir/flying global; physical @100 | §6 catalog merch + price |
| `MAKE_NEW_MERCH_ITEM.png` | Create form (tickets / file ≤2MB) | §7 create verify |
| `ADMINS_CONSOLE_INVENTORY.png` | Admin catalog console | §6 |

### 07–11 other
| Shot | What it proves | Ticklist |
|------|----------------|----------|
| `07.history/FULL_EMPTY_HISTORY.png` | All tabs empty «Пока нет событий» | §9 seed / fix history |
| `08.authors/ADMIN_AUTHOR_PAGE_0.png` | Demo multi-author note; «N карты»; 108 artists | §8 rewrite + товары |
| `08…PAGE_1/2` | Authors list / community | §8 unify |
| `09…AUTHOR_PROFILE.png` | Profile not created / URL avatar | §8 upload + create |
| `09…AUTHORS_PURCHASES.png` | Sales list | §8 payouts (C4) |
| `09…ADMIN_AUTHOR_LK.png` | «Мои карты» / Снять | §8 Мои товары / Удалить |
| `10.subscription/ADMIN_SUBSCTIPTION.png` | **Текущий/Будущий**; «Продление и отмена появятся позже» | §4 Базовый/Расширенный + renew/cancel |
| `11.analytics/*.png` | Analytics dashboards | §7 / §9 polish |

### 12.games (current + `referenceN/`)
| Shot | What it proves | Ticklist |
|------|----------------|----------|
| `games_8_eh.png` | All 8 tiles overview | isolation / home order |
| `1.pancaker/` vs `reference1/` | Flat hex vs intended stack look | §11 tray art deferred; space skin |
| `2.flappy/` vs `reference2/` | Current bird UI; boost optional; emoji currencies | §11 cosmic pipes; center prompt; icons |
| `3.builder/` vs `reference3/` | Arcade slab vs Dodo boxes | §11 Dodo deferred |
| `4.hanoi/` vs `reference4/` | Flat rings vs glossy/3D ref | §11 gloss; Sims deferred |
| `5.memonia/` + `reference5/` | Memory UX shots | §11 mapping bug re-verify |
| `6.2048/` + `reference6/` | Grid play | §11 swipe |
| `7.Gears/` + `reference7/` | Gears + goal **10** (Track B progress); boost copy | §11 flower skin wish; anti-abuse |
| `8.tamagotchi/` + `reference8/` | Soft care + daily +1000; still **Компаньон**; tickets later | §11 EN name; tickets gift |

**Note:** Game shots may predate Track B (06.10). Prefer live UI for “done”; screenshots still define **product intent** vs reference folders.

---

## Progress

| Block | Done / Total | Notes |
|-------|--------------|-------|
| 0 Meta / order | 0 / 5 | |
| 1 Repo cleanup | 7 / 8 | scripts moved; Dockerfiles stay; FE stores co-located |
| 2 Interview + Miro | 0 / 14 | after load preferred |
| 3 Security | 0 / 4 | |
| 4 Boosty / subs | 3 / 8 | Базовый/Расширенный + CTA live |
| 5 Home / chrome / profile / LB | 8 / 18 | About, blurbs, nick Modal, Gears/Tamagotchi |
| 6 Shop / inventory / catalog | 12 / 24 | merch chip, dedupe, themes/skins SQL, floors |
| 7 Admin panel | 2 / 12 | tickets icon Top-N; merch chip |
| 8 Authors | 1 / 12 | Мои товары |
| 9 History / analytics / notif | 5 / 8 | dual NATS + seed SQL + empty copy |
| 10 Games global | 10 / 14 | boost/GO/check/LB spam; Balance on all games |
| 11 Games per-title | 18 / 48 | cosmic skins, Memonia map, 2048 drag, Track B done |
| 12 Achievements | 0 / 4 | |
| 13 Mobile | 3 / 3 | safe-area + touch targets + shell inset |
| 14 Load / metrics | 0 / 16 | |
| 15 Bugs / console | 4 / 12 | profile path; LB undefined; API spam gated |
| 16 Parked wishlist | 0 / 8 | |
| 17 MCP / Tetiva | 0 / 5 | **very last** |

---

## 0. Meta / order

Sources: `CLEANUP.md`, `MOBILE_ADAPTIVE.md`, `MCP_WHEN.md`, `FUTURE_TODO_1.md`, ideas voice

- [ ] Confirm order: cleanup → FE/shop → games → mobile → load → docs/Miro → MCP/Tetiva last
- [ ] Pre-game rollback point (flash / tagged commit) — keep recoverable
- [ ] Re-verify Track B closed vs this list (boost UX, Towers levels, Companion gift, GO buttons…)
- [ ] Rename docs “Emma” → Denis where still present
- [ ] Fill `FEEDBACK_VoiceM_2_1.md` + `FEEDBACK_VoiceM_3_1.md` when v1.1.0 polish wave ends

---

## 1. Repo cleanup (**first**)

Sources: `CLEANUP.md`, ideas voice (professional look / folders)

- [x] Inventory root: which `.sh` / Dockerfiles **must** stay at root (CI / Makefile / Taskfile)
- [x] Move movable scripts → `scripts/` (update every Makefile / CI / compose ref **before** move)
- [x] Move movable Dockerfiles → folder **only if** all build refs updated — **kept at root** (Makefile `-f Dockerfile.*.bin .`)
- [x] FE: one folder per game (assets / icons / logic co-located) — stores co-located under `Games/*`
- [x] FE: shop / inventory / chrome in clear folders (stores under `Shop/`, `Inventory/`)
- [x] Backend: leave working Clean Architecture unless broken consistency found
- [x] Professional look: FE `tsc --noEmit` green after moves
- [ ] Smoke after moves: `make build-all && make docker-build-all && make deploy` (Denis runtime)

---

## 2. Interview / docs / Miro

Sources: ideas voice, `SYSTEM_DESIGN_MIRO/*`, Miro board

### Write-ups (v1.1.0 snapshot)
- [ ] **What have we done** (commit/state as of 07–08.10)
- [ ] **Which methods / patterns** we use + code links (LB, rate limiter, circuit breaker, outbox, …)
- [ ] **What could be better** (tech + design + product)
- [ ] Anti-patterns / risks note (or “none critical”) — after highload preferred
- [ ] Per-service plain-RU explainers: auth, game, billing, LB, profile, shop, inventory, authors, payment, history, analytics, notification, fulfillment, gateway, nats-hub, balancer
- [ ] Where WebSockets live (LB Redis→WS? notifications?) — document truth
- [ ] SQL practice list for interviews (admin joins: users / goods / purchases)
- [ ] `INTERVIEW_QUESTIONS_EH.md` — fill ports / “where is what”
- [ ] FE architecture note: backend-first then FE — is FE folder layout OK?

### Miro (Denis board)
- [ ] Open board https://miro.com/app/board/uXjVJLLg9us=/
- [ ] Align Miro with current v1.1.0 topology (ports, NATS subjects, deploy profiles)
- [ ] Export / screenshot scheme → refresh `FINAL_SYSTEM_DESIGN_MIRO_SCHEME.md` + store image under `08.10.2026/` or architecture/
- [ ] Optional: Cursor Miro marketplace plugin https://cursor.com/marketplace/miro — connect if useful for sync
- [ ] Cross-check Mermaid / `EH_SCHEMAS.md` vs Miro (no drift)

---

## 3. Security

Sources: ideas voice

- [ ] CSRF / XSS / SQL injection pass (gateway + FE)
- [ ] AuthZ on admin / author / inventory routes
- [ ] DDoS / rate-limit posture (limiter exists — verify surfaces)
- [ ] Confluence note: critical findings or “baseline OK”

---

## 4. Boosty / subscriptions

Sources: `BOOSTY_FIX.md`, shop voice (subscriptions)

Site: https://boosty.to/eastwesser

- [ ] Keep / refresh base tier copy (already drafted in `BOOSTY_FIX.md`)
- [ ] Update Boosty **tier 2 (×2)** benefits / copy
- [x] Align in-app names: **Базовый** / **Расширенный** (not «текущий/будущий»)
- [ ] Product decision: 3 tiers? (200 / 500 / 1000 ₽)
- [x] Make “future/расширенный” plan CTA live (Activate still hits checkout)
- [x] Add **продление** Boosty link + clearer footer (full cancel API later)
- [ ] v1.1.0 changelog content for Boosty / public share
- [ ] Optional follow-ups: 1.1.1…1.1.8 per-game posts

---

## 5. Home / chrome / profile / leaderboard

Sources: `VOICEMESSAGE_INSIGHTS_2.md`

### Home / hero
- [x] Lower «Event Horizon» toward vertical center (≈2× gap under navbar rule) — `01…/MAINPAGE1.png`
- [x] Navbar brand click → About (page **or** expand under hero) + Back — `01…/Event_Horizon_yellow_letters.png`
- [x] About copy: authors marketplace + games + tickets economy (professional, not “pet project”)
- [x] More vertical space between «Все игры» / «Лидерборд» and game grid (+ Void spacing ≈1.5–2×) — `01…/MAINPAGE2.png`
- [x] Home blurbs: Builder = floating/sliding blocks; Gears = connect to largest; Companion = Tamagotchi + soft care
- [ ] Trailing period cleanup on hero support line if still present
- [x] Keep Void mascot; don’t redesign black hole casually

### Chrome
- [ ] Platform button polish list (**Denis decides** which buttons) — from `FUTURE_TODO_1`
- [ ] Shop load spinner: true center H+V (not near header only)

### Profile
- [ ] Avatar upload (format hint; crop/size rules) — not URL-only — `03…/PROFILE_3_NEED_TO_FIX_AND_CHANGE_AVA.png`
- [ ] Fix all-zero game scores for admin after real saves (profile + LB consistency) — profile grid all 0
- [x] Nickname change **without** browser `confirm()` / host modal — in-app Modal
- [ ] Nickname change updates LB by **user id** (no split Admin/Nimda rows)
- [x] Rename profile tiles **Орбиты→Gears**, **Компаньон→Tamagotchi**

### Leaderboard chrome
- [x] Companion label → English (Tamagotchi) — `gameIcons` LB tabs
- [ ] Investigate Pancaker / others showing **0·0** tops — shot: #1 `9d0ebedb` @ 0
- [ ] Seed ~10 real nicknamed players × games for demo LB (not anon junk)
- [ ] LB game order matches Home
- [ ] Optional later: global combined score board — **skip** unless product asks (prefer per-game)

---

## 6. Shop / inventory / catalog

Sources: shop voice, `FEEDBACK_VoiceM_2_1.md`

### Catalog UX
- [x] Skins tab: show game skins (not empty) — Memonia animals etc.; restore rainbow-pipe era skins as cosmic where needed — `SHOP_5`+ *(SQL `seed-shop-themes-skins.sql` — apply on shop DB)*
- [x] Themes tab: seed ≥1 theme; **distinct icon** vs Skins — `SHOP_4.png` empty state *(star icon; SQL seed)*
- [x] Merch tab: seed ≥1 example if empty (catch-all for oversized / non-pocket goods) — cleanup-shop-content-v2
- [x] Mental model: **in-game cosmetic** (skins/themes) vs **physical** (cards / brelok / picture / fenechka / merch)
- [x] Card filters (element / class / rarity / foil-noir-flying) **only** on Cards — gated via `productType`
- [ ] Skins filters by game; themes niche (light/dark/cozy…); merch+brelok+picture+fenechka → **price** (+ shared basics)
- [x] Physical merch price floor **100_000** tickets — seed SQL updated (re-run cleanup script on DBs)
- [x] Deduplicate inventory list (rainbow pipes ×N) — `INVENTORY_DUPLICATES.png` *(group by item_id ×qty)*
- [x] Cancel copy: «предмет будет удалён из инвентаря» (not «карта»)
- [ ] Keep cosmic brelok art; regenerate badge / fenechka if too logo-like
- [x] Painting «Туманность Horizon» — delete/hide — platform-p1 + cleanup v2
- [ ] C3 smoke cards — OK to edit/delete (not real merch)

### Inventory / catalog admin
- [x] Catalog types include **merch**
- [x] Create-product: remove foil/noir/flying from non-card context; price label = **tickets** not ₽
- [ ] Types count / “4 types” → refresh (card, brelok, picture, fenechka, merch, skins, themes…)
- [x] Seed ≥1 picture, ≥1 theme; skins restored to filters *(themes/skins SQL; picture hidden)*
- [ ] Author seeding: ~108 card artists as authors (not only Event Horizon + Admin)

### Economy copy
- [ ] Clarify lamps = boosts; tickets = shop (tooltips in games too)
- [ ] Written site bugs/F12 report → `FEEDBACK_VoiceM_2_1.md`

---

## 7. Admin panel

Sources: shop voice (admin)

- [ ] Users: subscription column UX (active/inactive meaning clear)
- [ ] Seed author with **active** subscription for publish tests
- [ ] Inventory stats: authors count after seed — shot: **2 authors** / **4 types** (`ADMIN_INVENTORY.png`)
- [x] Top-N by price: tickets icon not ₽ — shot shows **100,000 ₽**
- [ ] Consider Top-100 by popularity / business metrics (not only Top-5 price)
- [x] Catalog type chips include **Мерч** — `06…/ALL_AVAILABLE_MERCH.png` missing chip
- [x] Analytics DAU/MAU/retention: explain «D0…D7»; fix empty retention if broken *(copy added; empty = no cohort data)*
- [ ] Applications tabs (pending / approved / rejected) — keep; polish labels
- [ ] Revenue view for admin later (subs ₽ → author payouts) — product lock with C4
- [ ] UX layout review: “are elements correctly placed?” — pass after polish
- [ ] Catalog create: image URL **or** file ≤2MB — verify works end-to-end
- [ ] Scroll / pagination for long author×goods lists after 108-author seed
- [ ] Written admin notes in site feedback file

---

## 8. Authors

Sources: shop voice (authors)

- [ ] Remove / rewrite demo line «временно до multi-author pages»
- [ ] Authors = all artists (not only «художники карт»); unify cards vs **goods** count
- [x] Author dashboard: «Мои карты» → **«Мои товары»**
- [x] Soft-delete label: «Удалить» (soft) instead of vague «Снять»
- [ ] Avatar / portfolio: upload from disk, not URL-only
- [ ] Sales → month-end payout % model (ties to C4 — product lock)
- [ ] Community authors list consistency with card artists + JWT smoke authors
- [ ] Author profile create flow if «not created yet»
- [ ] Everything non-(card|picture|brelok|fenechka) → **merch** catch-all (document for authors)
- [ ] Copy: authors pay subscription to list; players redeem with tickets; % to author
- [ ] Happy-path economics note (200 ₽ base sub, etc.) — for Boosty + About
- [ ] UX check of author cabinet layout

---

## 9. History / analytics / notifications

Sources: shop + ideas + console

- [x] History page: not empty on all tabs — seed registration / records / purchases / payments / authors — `07…/FULL_EMPTY_HISTORY.png` *(dual NATS publish + `seed-history-demo.sql`; apply after rebuild)*
- [x] Fix broken history if API returns nothing for admin *(subscribe `event.user.registered` → store as `user.registered`)*
- [ ] Notification: «your record beaten» deep-link to LB detail (player + scores)
- [x] Keep existing bell badge behaviour (don’t redraw)
- [x] Soft-handle notifications when logged out (no hard 401 spam)
- [ ] Analytics admin graphs: polish / explain MAU window / DAU series
- [x] F12: `/api/api/profile` 404 — fix double `/api` (also §15)
- [ ] Identity: no Telegram spam for players — in-app bell is enough

---

## 10. Games — global

Sources: `VOICEMESSAGE_INSIGHTS_3.md`, `FEEDBACK_VoiceM_3_1.md`, ideas

- [x] Currency icons in games (lamp/ticket SVG, not emoji) — BoostCheckbox + Icon lamp/ticket
- [x] Boost help **before** checkbox: what it does + not in LB + no shop profit from boosted runs — `gameBoostCopy`
- [x] Boost checkbox not eye-sore (collapsed / inside help) — `<details>` BoostCheckbox
- [x] GO buttons equal width / one row; «На главную» → `/#games` (not top of home)
- [x] Remove browser `confirm()` / debug submit overlays (`user id from localStorage`, `Sending to backend`…) — gone from Games; logs DEV-gated
- [x] Save toast: green **drawn** check icon (not emoji)
- [x] Isolate each game folder (FE) — same as §1
- [ ] Written report: what done in games for v1.1.0 → `FEEDBACK_VoiceM_3_1.md`
- [ ] 3 achievements per game (novice → amateur → pro → maestro → hero) — SVG, no emoji
- [ ] Optional OSS/3D engines later — don’t multi-language zoo without need
- [ ] Flappy LB eye-check after saves (manual QA)
- [x] `LEADERBOARD_LOGS.md`: stop spam `GET /leaderboard undefined` — fix client query params
- [x] Pancaker GO: correct Russian plural for “блинов/блина” — `pluralBliny`
- [ ] Pre-start UX: Start centered; boost optional below — polish if still awkward

---

## 11. Games — per title

### Pancaker / Hexagon
- [x] Explain boost = mono tray type; verify boosted run **not** ranked (was buggy: ranked while boost on)
- [ ] Level growth clarity **or** selectable difficulty / board size idea
- [x] Space skin: darker card under dark emoji (less pink+dark clash); more dark cosmic variety
- [ ] Cosmic pancakes button placement UX
- [ ] Top-10 not stuck at 0·0 for admin/anon after real saves
- [ ] Optional later: tray art instead of emoji; 2.5D stacks deferred

### Flappy
- [x] Rainbow pipes → cosmic (if any rainbow left) — kids-safe, no rainbow controversy
- [x] Center «Нажмите пробел» under score
- [x] Weaken click pulse / screen expand (hurts play)
- [x] Boost = world/pipes slower, bird normal (not cursor-follow — that kills the genre)
- [x] Gap jitter + score × level (re-verify)
- [ ] Optional: boosted sit-on-pipe grace — TBD
- [x] Rejected idea park: bird follows cursor on boost (ideas voice) — **do not ship**

### Builder / Towers
- [x] Disable meaningless «Идёт…» mid-run
- [x] Soft-fail / difficulty 1–10 vs floor (re-verify)
- [x] Boost = slower block (re-verify)
- [x] Screen shake on click — tone down if same as Flappy pulse
- [ ] Dodo Pizza boxes look — deferred wish

### Hanoi
- [x] Disk select: dark bg, white digits, gold hover (not grey-on-white)
- [x] Timer starts on «Старт», not on enter
- [x] Pre-game disk pick → Start (default 5; range 3–8)
- [x] Boost = auto-solve; link copy so players understand
- [x] Stop auto-solve button
- [ ] Ring gloss / volume polish
- [x] No skins for Hanoi (would hurt readability)
- [ ] 3D Sims camera — deferred

### Memonia
- [x] Fruit/animal mapping bug (peach+coconut → fox) — **re-verify fixed** *(shared FRUIT/ANIMAL lists)*
- [x] Home blurb without «фруктов»
- [x] Boost = one pair hint
- [ ] Wish: real card art later

### 2048
- [x] Swipe / drag on mobile (and mouse drag)
- [x] Boost = undo **or** cell swap (document; unranked)
- [x] Title optionally plain «2048»

### Gears (was Orbits)
- [x] Name: **Gears** (EN) everywhere
- [x] Gear visuals (not plain balls) — re-verify
- [x] Anti-abuse: lose line when field filled to spawn line; no merge-in-flight abuse
- [x] Goal copy not hardcoded «до 8» — “to the largest”
- [x] Boost = copy next nominal (not slower drop) — re-verify
- [ ] Optional flower skin (sunflower / rose / pansy / …) — wish from ideas voice
- [ ] Immediate merge on contact (no multi-touch delay) — from ideas voice

### Companion
- [x] EN name (Tamagotchi vs Companion) — decide + apply home/LB/game *(+ catalog)*
- [x] Soft care copy (no death FOMO) — “play with your pet”
- [x] Daily gift +1000 care points (re-verify); tickets gift when billing OK *(points done; tickets later)*
- [x] Remove weird boost-on-save unranked (or replace with gift)
- [ ] Room + cute pet art — wish
- [x] Not a spam-score game — daily care loop

---

## 12. Achievements

Sources: ideas voice

- [ ] Ladder: новичок → любитель → профессионал → маэстро → герой/король
- [ ] 3+ badges per game wired to scores / LB thresholds
- [ ] Draw SVG icons (no emoji)
- [ ] Profile surface for unlocks (depends on `/api/profile` fix)

---

## 13. Mobile adaptive (**after FE cleanup**)

Sources: `MOBILE_ADAPTIVE.md`

- [x] FE cleanup / game isolation done first (§1)
- [x] Responsive pass: home, shop, games, profile, LB on phone widths *(safe-area, shell px-4, brand/logout collapse, ScoreChip/controls)*
- [x] 2048 swipe + game touch targets usable on small screens

---

## 14. Load tests / metrics

Sources: `LOAD_TESTS/HIGHLOAD_TESTS.md`, `METRICS.md`

### Run
- [ ] Run CORE highload suite (k6 / existing scripts) on current stack
- [ ] Store results under `08.10.2026/LOAD_RESULTS/` (create when run)

### Capture vs targets (10k DAU model)
- [ ] RPS (avg ~17–50, peak ~35–100 depending on session×API model)
- [ ] Latency p50&lt;50 / p90&lt;150 / p95&lt;200 / p99&lt;300 ms
- [ ] Error rate / HTTP 5xx
- [ ] Concurrent HTTP (~5k) / WS (~2.5k) posture note
- [ ] CPU / RAM / I/O / net vs &lt;70% peak guidance
- [ ] Go: goroutines, GC, heap (&lt;512MB guidance), mutex waits
- [ ] DB: connections &lt;100, slow queries &lt;5%, deadlocks 0, repl lag &lt;1s
- [ ] NATS JetStream consumer lag
- [ ] Business: DAU/MAU/retention + popularity by `game_id` / authors

### Ops / docs
- [ ] Grafana 03:00 watchlist: Auth/Gateway p99, 5xx, free RAM, JetStream lag (&gt;1000 = bad)
- [ ] Post-load technical risk write-up (backend / FE / product crises)
- [ ] Optional: Selectel cost model refresh (~$430 / 5 servers) if infra changed
- [ ] Compare read:write ~2–3:1; avg payload &lt;10kb; raw events retention 30d

---

## 15. Bugs / console (F12)

Sources: `old_console_log.md`, `LEADERBOARD_LOGS.md`, feedback stubs

- [x] Fix `GET /api/api/profile` → 404 (double `/api` prefix) — now `api.get('/profile')`
- [x] Fix achievements uncaught promise when profile 404 (path fixed)
- [ ] Soft-handle `/notifications` 401 when logged out / expired
- [ ] Soft-handle `/game/submit` 401 — session refresh / clear message
- [ ] Soft-handle `/billing/balance/all` 401 — invalid/expired token UX
- [x] Strip or gate verbose `📡 API Request` console spam in prod (`import.meta.env.DEV`)
- [x] Leaderboard: `undefined` in console was GET body log spam — gated with DEV
- [ ] Investigate `[VOID] mount` double-log noise (dev only OK?)
- [x] React DevTools download hint — ignore
- [ ] Written site bugs report → `FEEDBACK_VoiceM_2_1.md`
- [ ] Written games bugs report → `FEEDBACK_VoiceM_3_1.md`
- [ ] Re-check console after fixes on Profile + each game submit + LB

---

## 16. Parked wishlist (`FUTURE_TODO_1.md`)

Ship path Tracks A–D = done on main. Still open:

- [ ] Platform chrome buttons — needs Denis list
- [ ] C4 payouts — product lock (deferred on purpose)
- [ ] Companion tickets + EN name — billing / naming
- [ ] 3D / Dodo / Sims / tray art — wish, not this sprint
- [ ] Flappy LB eye-check — manual QA
- [ ] EXPLAIN / bottleneck / live backfill — ops later
- [ ] Mass LB seeding 10×8 nicknamed players — QA
- [ ] Gears flower skin / Companion room art — wish

---

## 17. MCP / Tetiva (**very last**)

Sources: `MCP_WHEN.md` + Denis Tetiva note

### MCP
- [ ] Decide: do we need an MCP server for EH? (**very last**)
- [ ] If yes: inventory existing MCP in repo / Cursor; implement only after everything above
- [ ] If no: document “skipped — not needed for v1.1.0”

### Tetiva (https://tglink.io/c014cdc9e62dc7?erid=2W5zFGyuMvq)
Go+Wails+Vue API client (HTTP/gRPC/GraphQL/WS, local SQLite, Postman/cURL import, built-in MCP for Cursor). Free MIT.

- [ ] Decision: **optional developer tooling only** — not an EH product feature
- [ ] Recommendation: **skip for now** unless Denis wants it for local API QA / saved collections via MCP; does not replace OpenAPI/Swagger already on gateway
- [ ] If adopt later: install locally, import EH collections, wire MCP — still after polish/load/docs

---

## Next action

```text
START = §1 Repo cleanup (scripts/Docker refs, FE game folders).
THEN  = §5–9 site/shop/admin/authors + §4 Boosty copy.
THEN  = §10–12 games verify + leftovers + §15 console bugs.
THEN  = §13 mobile → §14 load → §2 interview+Miro → §17 MCP/Tetiva last.
```

### Source index (quick)

| Path | Role |
|------|------|
| `07.10.2026/CLEANUP.md` | Cleanup mandate |
| `07.10.2026/MOBILE_ADAPTIVE.md` | Mobile after FE cleanup |
| `07.10.2026/MCP_WHEN.md` | MCP last |
| `07.10.2026/BOOSTY_FIX.md` | Boosty tiers / changelog |
| `07.10.2026/FUTURE_TODO_1.md` | Parked after Tracks A–D |
| `07.10.2026/INSIGHTS/1.IDEAS/…` | Interview, security, achievements, services, Gears skin, Flappy anti-pattern |
| `07.10.2026/INSIGHTS/2.SHOP/…` | Site + shop + admin + authors + subs |
| `07.10.2026/INSIGHTS/3.GAMES/…` | Per-game UX + LB logs + games feedback stub |
| `07.10.2026/INSIGHTS/3.GAMES/eh-screenshots/` | **80 shots** · 01–12 presentation order · voice↔pixels (see evidence table above) |
| `07.10.2026/BUGS_FEEDBACKS/…` | Console / API bugs |
| `07.10.2026/LOAD_TESTS/…` | Highload home + metric targets |
| `07.10.2026/SYSTEM_DESIGN_MIRO/…` | Schema + interview ports; pull Miro image |
| Miro board | https://miro.com/app/board/uXjVJLLg9us=/ |
| Cursor Miro | https://cursor.com/marketplace/miro |
| Boosty | https://boosty.to/eastwesser |
| Tetiva | optional API client — §17 skip unless QA needs it |
