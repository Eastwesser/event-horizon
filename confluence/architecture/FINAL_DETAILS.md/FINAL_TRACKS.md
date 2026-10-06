# FINAL_TRACKS — living map (v1.1.0+)

Один трек за раз. Не распыляться.

| Doc | Role |
|-----|------|
| [`GAME_INSIGHTS.md`](../../history/2026-10/06.10.2026/GAME_INSIGHTS.md) | **Games** — Track B closed; deferred art parked |
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
| **8 games** · Boost · hotfix · UX · levels · polish | GAME_INSIGHTS |
| **Track D P1–P3** | shop art · spinner/load · chrome glossary |
| **Track C** | k3s data plane |

---

## Track A — Product — ✅ DONE

1. ✅ Refund · 2. ✅ Notifications · 3. ✅ JWT role refresh

---

## Track B — Games — ✅ DONE

Полный чеклист: **GAME_INSIGHTS.md**.

| Step | Status |
|------|--------|
| Shop cleanup (catalog) | ✅ |
| Boost Phase 1 allowlist | ✅ |
| Hotfix + Boost UX + per-game effects | ✅ |
| Levels + Flappy/Companion + Gears polish | ✅ |
| GameShell order + LB tabs + space skin | ✅ |
| Deferred art / 3D / tickets gift | 🟧 parked |

---

## Track D — Platform (shop / chrome) — ✅ P1–P3

Чеклист: **PLATFORM_INSIGHTS.md**

| Step | Status |
|------|--------|
| **P1 Shop examples & art** | ✅ |
| **P2 Shop spinner + load** | ✅ |
| **P3 Chrome polish** | ✅ (glossary + SVG; button list → Emma) |
| Berserk cards | 🔒 never without OK |

---

## Track C — Infra — ✅ DONE

- [x] k3s NATS + Postgres StatefulSets (Helm `dataPlane.enabled`; `make deploy-k3s-dataplane`)

---

## Deferred

Open: platform button list (Emma). Art/3D, Companion tickets economy.
Locked/skipped: C4=D, EXPLAIN/bottleneck/cursor/authors content, live backfill.

---

## Что делать сейчас

```text
Tracks A–D gameplay/platform DONE.

NEXT = push when ready · Emma chrome list · or deferred art when wanted.
```

---

## Карта

```
Wave 1–4   ████████████ DONE
Track A    ████████████ DONE
Track B    ████████████ DONE (deferred art parked)
Track D    ████████████ P1–P3 DONE
Track C    ████████████ DONE
Deferred   ████████░░░░ Emma + art/3D
```
