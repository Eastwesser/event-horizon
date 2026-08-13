# Frontend remaining — DO NOT DROP (13.08.2026)

## Status overview

| # | Task | Status |
|---|------|--------|
| F1 | API clients for Payment / Authors / Analytics | ✅ |
| F2 | **UI wired to those APIs** | ✅ pages + shop merch gate |
| F3 | **Ханойская башня** (`/game/hanoi`) | ✅ `Games/Hanoi/HanoiTower.tsx` |
| F4 | Subscription / Payment UX (`/subscription`) | ✅ `Payment/Subscription.tsx` |
| F5 | Authors UI (`/authors`) | ✅ `Authors/AuthorsPage.tsx` |
| F6 | Analytics dashboard (`/analytics`, admin) | ✅ `Analytics/AnalyticsDashboard.tsx` |
| F7 | Shop: `canPurchaseMerch` before merch buy | ✅ Shop + PurchaseModal |

## Clarification: two different “towers”

| Name | Path | What it is |
|------|------|------------|
| **Башенки (Towers)** | `/game/towers` | Stack falling blocks |
| **Ханойская башня (Hanoi)** | `/game/hanoi` | Classic 3 pegs, rings 3–8 |

Spec: `FRONTEND_KHANOY_TOWERS.md`

## Done criteria met

- Nav: Подписка, Авторы, Аналитика (admin-only), Hanoi on Home games list
- `useUserRole` caches role from `/api/auth/whoami`
- `npm run build` OK

## Soft backend debt (cross-ref)

| Task | Status |
|------|--------|
| Circuit breaker on all gRPC clients | ✅ |
| Inventory proto `version` field | ✅ |
| Critical package coverage | ✅ jwt 92.9%, converters ≥70% |
| Full-service coverage ≥70% | 🟡 deferred |
