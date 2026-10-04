# Postmortem — VOID Home disk “floating square”

**Date:** 2026-10-04  
**Surface:** Home «Выбери игру» accretion disk (VOID v2→v3)  
**Status:** Root cause identified (`hide=core`); fix = replace core `box-shadow` with circular radial glow  
**Commits (local at investigation):** `6722de1` (v2), `c38bdfd` / `942a4cb` / `ebf6a92` (v3 + debug)

---

## Summary

On hover, a faint **axis-aligned square** appeared around the VOID disk and appeared to **drift toward the upper-left**. Layer toggles and visual QA proved the square was **not** parent `bg-void`, not `rotateX`, and not the outer glow alone. The source was **`.eh-disk-core`’s large gold `box-shadow`**: a circular element still produces a **rectangular compositing / blur AABB**; with a huge spread (up to ~200px) that rectangle’s soft edges read as a square, and hover scale/shadow changes made it look like it moved.

---

## Timeline (compressed)

| Phase | Hypothesis | Result |
|---|---|---|
| Reduced motion | OS/browser `prefers-reduced-motion: reduce` killed all motion | True for “no animation”; **not** the square. Added DEV `?motion=force`. |
| Glow blur clip | `.eh-disk-glow` `filter: blur` clipped to a rect | Removed blur; radial + `clip-path`. Square remained. |
| `isolation: isolate` | 420×420 stacking context edge | Removed. Square remained. |
| `rotateX(52deg)` on `.eh-disk-pull` | 3D plane foreshortening → NW drift | **Caused the drift.** Removed. **Square still visible when static/hover.** |
| Debug outlines `?debug-void=1` | Match outline color to square | No VOID layer outline sat on the square; ancestors also transparent. |
| Ancestor hunt | `backdrop-filter` / `bg-nebula` on parent | Computed: none. `hide=bg-nebula` / `hide=hero` → square stays. |
| Layer elimination `?hide=` | Binary search | `hide=glow|pull|rings|particles|hero|bg-nebula` → square **stays**. **`hide=core` → square GONE.** `hide=disk` → everything gone (consistent). |

---

## Root cause

`.eh-disk-core` used stacked **`box-shadow`** for the event-horizon corona, e.g. on pull:

```text
0 0 0 2px gold-hot
0 0 52px 16px gold
0 0 140px 42px gold
0 0 200px 64px cyan
inset …
```

Even with `border-radius: 50%`, large shadow blur/spread is painted in a **rectangular bounding box**. Soft shadow near that box’s corners reads as a square frame; hover (`scale` + stronger shadows) makes the frame appear to shift.

**Confirmed by:** Emma’s visual matrix — only `?motion=force&hide=core` removed the square while other disk layers remained.

---

## False leads (keep for next time)

1. **`prefers-reduced-motion`** — explains missing animation, not the square.  
2. **Parent `bg-void` tint mismatch** — attractive, but ancestry was transparent until full-page void; `hide=bg-nebula` did not help.  
3. **Outer `.eh-disk-glow`** — looked guilty; `hide=glow` alone did not remove the square.  
4. **`rotateX`** — real bug for NW *motion*, orthogonal to the remaining static/hover square.  
5. **Pixel A/B on CI/headless** — once showed `hide=glow+pull` cleaning corners (rings nest under pull); that **disagreed with human QA**. Trust `hide=` visual matrix on the target VM over headless deltas when they conflict.

---

## Fix (intended)

1. **`.eh-disk-core`:** drop outer `box-shadow` corona; paint black hole + gold/cyan falloff with **`radial-gradient(circle …)`** on a large enough circular layer (`clip-path: circle` optional).  
2. Keep **inset** darkening via gradient stops (not inset shadow) if needed.  
3. **Rings:** keep thicker band (~3.25px padding) so equator reads after flatten.  
4. Retain DEV tools for regressions: `?motion=force`, `?debug-void=1`, `?hide=core|glow|…`.

---

## Verification checklist

- [ ] `/?motion=force` — hover: no square, no NW drift  
- [ ] `/?motion=force&hide=core` — (control) still no square; with fix, full disk should match “no square” look while core visible  
- [ ] Equator flatten + particle spiral still readable  
- [ ] `prefers-reduced-motion` without `motion=force` still static  

---

## Follow-up (same day) — core looked “above” rings

**Symptom:** After the square fix, the black hole looked shifted **up** vs ring/particle center.  
**Not a layout bug:** `.eh-disk-core` / rings / particles bounding-box centers matched (`top/left: 50%` + `translate(-50%,-50%)`).  
**Cause:** Radial fill used `circle closest-side at 50% 42%` — paint origin above geometric center.  
**Fix:** `at 50% 50%` on idle / pull / hot / reduced-motion core gradients.

## Lessons

- **`box-shadow` ≠ “circular glow.”** Prefer radial gradients (or a dedicated larger circle layer) when the glow must extend far past the element.  
- **Layer elimination (`?hide=`) beats guessing** when outlines don’t match the artifact.  
- **Separate “moves” from “exists.”** Motion bugs (`rotateX`) and paint bugs (`box-shadow` AABB) can stack.  
- **Human VM QA is source of truth** for subtle compositing; headless pixel samples can mislead.  
- **Gradient center ≠ box center.** Check `radial-gradient(… at x% y%)` when the hole looks off-axis but DevTools boxes align.

---

## References

- `VOID_INSPECTION_1.md` — early `hide=` matrix (parents/hero/glow)  
- `VOID_INSPECTION_2.md` — `hide=core` confirmation  
- `VOID_HELL_*.md` / `VOID_SOLUTION_*.md` — prior hypotheses  
- Rollback point: `ebf6a92`  
