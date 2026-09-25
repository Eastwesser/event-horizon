# P2d — Duplicate requests

## Finding

Observed doubles (`/auth/whoami` ×2, occasionally inventory/subscription) are consistent with **React StrictMode** double-mounting effects in **dev** (`main.tsx` wraps `<StrictMode>`).

## Mitigations already in place

- `useUserRole` — 60s module cache + in-flight dedupe for `/auth/whoami`
- `Balance` — 5s in-flight cache + invalidate on purchase
- History / other pages — cancel-on-unmount patterns from earlier TODOs

## Verdict

**Not a production bug.** In prod (no StrictMode double-invoke) whoami should hit once per cold mount / cache miss. Remaining 2× in local logs after cache = StrictMode remount before cache fills, or two components mounting before the first response lands (in-flight promise should collapse those).

No further code change required for P2d unless prod network tab still shows duplicate whoami without StrictMode.
