# FINAL_TRACKS — living map (v1.1.0+)

Один трек за раз. Не распыляться.

| Doc | Role |
|-----|------|
| [`GAME_INSIGHTS.md`](../../history/2026-10/06.10.2026/GAME_INSIGHTS.md) | **Games only** — parked; tick there when on Track B |
| [`PLATFORM_INSIGHTS.md`](../../history/2026-10/06.10.2026/PLATFORM_INSIGHTS.md) | **Shop / chrome / infra** — P1–P3 done |
| [`TODO_FINAL_PRE_PROD_DETAILS.md`](./TODO_FINAL_PRE_PROD_DETAILS.md) | Wave 1–5 checklist |

---

## Закрыто (main)

| Chunk | Notes |
|-------|--------|
| v1.1.0 · Wave 4/4.5 · PR #2 · k6 correctness | |
| Wave 3 C1–C3 | |
| **Track A 3/3** | refund · notifications · JWT refresh |
| **Shop cleanup v2** | 281 Berserk untouched · 5 skins · 4 examples |
| **8 games playable** · Boost Phase 1 allowlist | games polish → GAME_INSIGHTS |
| **Track D P1–P3** | shop art · spinner/load · chrome glossary |

---

## Track A — Product — ✅ DONE

1. ✅ Refund · 2. ✅ Notifications · 3. ✅ JWT role refresh

---

## Track B — Games — PARKED

Полный чеклист: **GAME_INSIGHTS.md** (не править структуру без нужды).

| Step | Status |
|------|--------|
| Shop cleanup (catalog) | ✅ (platform-adjacent; done) |
| Boost Phase 1 allowlist | ✅ |
| Hotfix → Boost UX → per-game effects → Levels | 🟧 hotfix ✅ · UX/effects/levels next |

---

## Track D — Platform (shop / chrome) — ✅ P1–P3

Чеклист: **PLATFORM_INSIGHTS.md**

| Step | Status |
|------|--------|
| **P1 Shop examples & art** | ✅ |
| **P2 Shop spinner + load** | ✅ |
| **P3 Chrome polish** | ✅ (glossary + SVG; button list → P5/Emma) |
| Berserk cards | 🔒 never without OK |

---

## Track C — Infra — ✅ DONE

- [x] k3s NATS + Postgres StatefulSets (Helm `dataPlane.enabled`; `make deploy-k3s-dataplane`)

---

## Deferred (P5 slice done — see PLATFORM_INSIGHTS)

Open: platform button list (Emma). Locked/skipped: C4=D, #5b shipped, EXPLAIN/bottleneck/cursor/authors content, backfill dry-run OK.

---

## Что делать сейчас

```text
P4 + P5 actionable DONE.

NEXT = GAME_INSIGHTS / Track B games (hotfix → boost UX → effects).
```

---

## Карта

```
Wave 1–4   ████████████ DONE
Track A    ████████████ DONE
Track B    ██░░░░░░░░░░ NEXT (GAME_INSIGHTS)  ← you are here
Track D    ████████████ P1–P3 DONE
Track C    ████████████ DONE
Deferred   ████████░░░░ P5 slice done (Emma button list open)
```
