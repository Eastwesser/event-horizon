# TODO_5 — results (after hard-refresh)

Hard-refresh confirmed by Emma. Fixes below re-checked end-to-end against TODO_3 + TODO_4.

## Root causes closed

| # | Symptom | Root cause | Fix |
|---|---|---|---|
| 1 | Content glued to left edge (Home + every PageShell page) | Unlayered `* { padding: 0 }` in `index.css` beat Tailwind `px-*` utilities | Reset moved into `@layer base` — utilities win again |
| 2 | 401 cascade / «no Bearer» logs | Interceptor cleared session even when no Bearer was sent; TTL 15m had no refresh | Attach token first; clear only after Bearer+401; refresh-once via `/auth/refresh` before logout |
| 3 | History / balance request storms | Unstable effect deps + StrictMode remounts | History: cancel-on-unmount, depend only on `eventType`. Balance: 5s in-flight cache + 30s poll |
| 4 | Black screens Authors/Leaderboard | Null crash + no boundary | Null-guards + App-level `ErrorBoundary` |
| 5 | Memory flip / Hanoi drag «gone» | `prefers-reduced-motion` killed *all* transitions | Reduced-motion scoped to UI chrome (`.eh-disk*`, `.eh-stagger-in`, `.eh-glow`); flip CSS hardened |
| 6 | Red «Комбо» | Score chips defaulted to gold/ember look | Default indigo-soft; Combo → photon-cyan; scores → gold |
| 7 | Hanoi win modal off-center | `fixed` trapped by transformed ancestors | `Modal` portaled to `document.body` |
| 8 | «Мемония» naming | Copy drift | Renamed to **Меморина** (Home, games, leaderboard, profile) |
| 9 | Hero still Блинопёк; 4-game grid | Flagship leftover | Disk core = `logo-minimal.png`; all **5** games in equal-height grid |

## Checklist from prior review

| # | Check | Status |
|---|---|---|
| 1 | Hard-refresh Home | Done (Emma) |
| 2 | Left inset on Home (`px-6` / `sm:px-8`) | Fixed — CSS layer root cause |
| 3 | First card left border visible | Fixed with #2 |
| 4 | Grid fills / 5 equal cards | `xl:grid-cols-5` + `items-stretch` + `min-h-[220px]` |
| 5 | Equal card heights | Yes |
| 6 | TTL 15m without logout | Refresh-on-401 path in `services/api.ts` |
| 7 | `/authors`, `/leaderboard` black screens | ErrorBoundary + null-guards |
| 8 | `/inventory` CRUD | Unblocked by auth refresh; re-test with live session |
| 9 | Memory flip / Hanoi drag | CSS present; reduced-motion no longer blanketing |
| 10 | `/history` + `/billing/balance/all` loops | Deduped / cancelled |

## Verified in this pass

- `npx tsc --noEmit` → **0**
- `npm run build` → **0**
- Login card padding measured ≈ **40px**; form inset ≈ **41px**; register-link `margin-top` ≈ **64px**
- Tailwind `px-6 sm:px-8` computes **non-zero** padding (utilities beat base reset)
- Home a11y tree: all 5 games including **Меморина**; disk uses brand logo (not Блинопёк)

## Files touched (high-signal)

- `frontend/src/index.css` — `@layer base` reset
- `frontend/src/styles/theme.css` — narrowed reduced-motion
- `frontend/src/services/api.ts` + `lib/auth.ts` — refresh + safe 401
- `frontend/src/components/Home/Home.tsx` — 5-game grid + logo disk
- `frontend/src/components/History/HistoryPage.tsx` — loop fix
- `frontend/src/components/Billing/Balance.tsx` — fetch cache
- `frontend/src/components/ui/Modal.tsx` — portal
- `frontend/src/components/ui/GameShell.tsx` — kids-safe ScoreChip tones
- Games / Leaderboard / Profile — Memorina + Combo cyan + Hanoi warning (not error red)

## Still for live session (needs real tokens)

1. Sit 16+ minutes → confirm silent refresh, no forced logout  
2. Inventory create/update/delete once  
3. Visual flip one Memory card + drag one Hanoi ring  

Auth login/register layout (TODO_3 Phase 1) already matches the brief.
