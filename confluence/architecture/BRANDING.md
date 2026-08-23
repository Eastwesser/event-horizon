# Branding — Event Horizon logos

Assets collected 22.08.2026 for a future frontend redesign.

**Source folder (raw):**  
`confluence/history/2026-08/22.08.2026/news/event_horizon_logos/`

| File | Suggested use |
|------|----------------|
| `event_horizon_site_logo.md.png` | Site / marketing header, Open Graph |
| `event_horizon_minimal_logo.png` | Favicon, app icon, compact nav |
| `event_horizon_small_logo.png` | Inline UI (shop card, toast, loading) |

## Placement recommendation

When the FE redesign starts:

1. Copy chosen assets into `frontend/public/brand/` (or `frontend/src/assets/brand/`).
2. Keep this folder as the **design archive**; do not delete history copies.
3. Prefer SVG later for crisp scaling; PNG is fine for v1 splash/landing.

## Product UI notes (align with FE rules)

- Brand name should be a **hero-level** signal on landing — not only nav text.
- First viewport: brand + one headline + one short line + one CTA + one dominant visual.
- Avoid purple-on-white / cream-serif / broadsheet clichés unless Emma asks for that look.
- Logos should work on dark and light atmospheres (test contrast).

## Status

- Logos on disk: ✅
- Wired into React app: ❌ (not yet)
- Favicon / PWA icons: ❌

Next step when redesigning: pick primary mark (`minimal` or `site`), set CSS brand variables, then replace placeholder titles in `Home.tsx` / nav.
