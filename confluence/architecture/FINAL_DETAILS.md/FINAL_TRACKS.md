# FINAL_TRACKS — living map (v1.1.0+)

Один трек за раз. Не распыляться.

| Doc | Role |
|-----|------|
| [`GAME_INSIGHTS.md`](../../history/2026-10/06.10.2026/GAME_INSIGHTS.md) | **Games only** — parked; tick there when on Track B |
| [`PLATFORM_INSIGHTS.md`](../../history/2026-10/06.10.2026/PLATFORM_INSIGHTS.md) | **Shop / chrome / infra** — active non-game map |
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
| Hotfix → Boost UX → per-game effects → Levels | 🟧 parked until Emma returns to games |

---

## Track D — Platform (shop / chrome) — ACTIVE

Чеклист: **PLATFORM_INSIGHTS.md**

| Step | Status |
|------|--------|
| **P1 Shop examples & art** | ✅ |
| **P2 Shop spinner + load** | ✅ |
| P3 Chrome polish | 🟧 **NEXT** |
| Berserk cards | 🔒 never without OK |

---

## Track C — Infra — LATER

- [ ] k3s NATS + Postgres StatefulSets (Helm data plane)

---

## Deferred (Emma OK only)

C4 payouts · Wave 2 #5b · multi-VU EXPLAIN · bottleneck 500+ · cursor pull-in · backfill-noiz · 3D/OSS games

---

## Что делать сейчас

```text
NEXT = Track D / PLATFORM_INSIGHTS P3
  — chrome / nav polish (Emma list)

P2 done (centered spinner + FE catalog waterfall fix).
Games: GAME_INSIGHTS parked until Emma says back to Track B.
```

---

## Карта

```
Wave 1–4   ████████████ DONE
Track A    ████████████ DONE
Track B    ██░░░░░░░░░░ PARKED (GAME_INSIGHTS)
Track D    ████░░░░░░░░ ACTIVE → P3 chrome  ← you are here
Track C    ░░░░░░░░░░░░ later
Deferred   ░░░░░░░░░░░░
```
