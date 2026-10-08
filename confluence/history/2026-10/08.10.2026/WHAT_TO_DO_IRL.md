# WHAT TO DO IRL — close v1.1.0

Updated after tooling + seeds + CORE k6 (08.10 night).  
**Do not touch without OK:** refund, notifications, auth rewrite, Berserk cards, C4, hero, disk, games.

---

## Done tonight (Cursor + runtime)

- [x] `make test-k6` passes env (`EH_K6_*` / `.env.seed.admin`) → `browse.js` @ `:8079`
- [x] `make test-k6-purchase` added
- [x] Legacy blast → `deployments/k6/legacy/loadtest.js`
- [x] `loadtest_balancer.js` → `:8079` + LEGACY header
- [x] `make seed-shop` / `seed-shop-inventory` / `seed-themes` / `seed-history` / `seed-v110`
- [x] Seeds applied: inventory cards **281**, shop themes **2**, skins available, history demo, profile goose at `20261008120000`
- [x] CORE browse: **checks 100%**, p95≈705 ms &lt;800, errors 0% → `LOAD_RESULTS/browse-*.txt`
- [x] Interview patterns / WS / SQL / FE note → `INTERVIEW_PATTERNS.md`
- [x] Security + LOAD_RESULTS docs updated for bcrypt/rate-limit finding

---

## Still you IRL (cannot automate)

### Miro (~30–60 min)
Board: https://miro.com/app/board/uXjVJLLg9us=/

- [ ] Align ports + NATS + deploy vs deploy-heavy
- [ ] Draw happy path (`V1_1_0_FINAL_STATUS` §F)
- [ ] Export PNG → `08.10.2026/` + refresh `FINAL_SYSTEM_DESIGN_MIRO_SCHEME.md`
- [ ] Skim Mermaid / `EH_SCHEMAS.md` vs Miro

### Boosty (https://boosty.to/eastwesser)
Paste from `BOOSTY_FIX.md` / pack §E:

- [ ] Refresh Базовый if needed
- [ ] Update Расширенный (×2)
- [ ] Publish **v1.1.0** post

### Optional verify in browser
- [ ] Shop → Themes / Skins non-empty
- [ ] History has demo events for admin
- [ ] Play + save one ranked game → profile/LB not stuck at 0

### Tag
- [ ] When Miro + Boosty done: git tag `v1.1.0` (rollback point)

---

## Parked (not blocking tag)

Avatar upload · 108 authors · Tamagotchi tickets gift · notif deep-link · full CSRF pass · MCP/Tetiva · C4 · 1.1.1…1.1.8 posts · art wishes

---

## Commands cheat sheet

```bash
make seed-v110
make test-k6
make test-k6-purchase
# results under confluence/history/2026-10/08.10.2026/LOAD_RESULTS/
```
