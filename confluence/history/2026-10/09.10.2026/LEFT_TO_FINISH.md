# What’s left — 10.10.2026

## Done this wave (engineering)

| Item | Notes |
|------|--------|
| `/api/v1` gateway live | Postmortem: [`POSTMORTEM_API_V1_GATEWAY.md`](./POSTMORTEM_API_V1_GATEWAY.md) |
| Pancaker score → LB/profile zeros | Game hexagon validator trusted empty board (=0); now trusts client score like memory. **Rebuild game done.** Replay one Pancaker run to refresh your admin row. |
| Billing/PG first-boot race | Compose `depends_on` + longer `start_period` |
| Boost at 0 lamps | FE disables checkbox when lamps &lt; 10 |
| Hexagon no pile-on cell | FE only drops on empty |
| Profile edit disclosure | Avatar/ID behind «Изменить» |
| Hanoi stage multipliers | ×1…×2 by ring count |
| Gears look / flower skins | Metal = gear teeth; flower skins = petals |
| Companion gift balance refresh | `invalidateBalanceCache` after claim |
| Interview EH answers | [`OCTOBER_INTERVIEW_EH_ANSWERS.md`](./OCTOBER_INTERVIEW_EH_ANSWERS.md) |
| Dream DevOps parking | [`DREAM_DEVOPS_IDEAS.md`](./DREAM_DEVOPS_IDEAS.md) — **not** v1.1 debt |
| Server sizing | [`SUITABLE_SERVER.md`](./SUITABLE_SERVER.md) — 8c/32G sweet spot |
| LB levels 2–3 | UI commented out; all boards fetch `level=1` |

---

## True “0 leftovers” for *you* (IRL / product — agent can’t finish alone)

| Item | Why agent can’t close it |
|------|---------------------------|
| **Miro sticker polish** | Physical/board edits — paste sheet is ready |
| **`git tag v1.1.0`** | Explicit release marker — say the word + which commit |
| **Unified game chrome** | Product layout choice (L/R controls) |
| **Tamagotchi art vs screenshots** | Needs Denis visual reference applied |
| **C4 / 108 authors / 3D** | Locked — do not open |

---

## Explicitly NOT leftovers

Loki, Tempo, node-exporter, multi-DB scrape, SLO burn alerts, full k8s prod → **`DREAM_DEVOPS_IDEAS.md`**.

---

## After power cut / smoke

```bash
make deploy
curl -sS -o /dev/null -w 'v1=%{http_code}\n' \
  'http://127.0.0.1:8079/api/v1/leaderboard?game_id=hexagon&limit=1'
# FE: play Pancaker → save → LB should show real score (not 0)
```
