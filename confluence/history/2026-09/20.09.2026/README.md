CONTEXT
The Event Horizon redesign has a real visual problem: the site lost its
indigo identity and now reads as a "black-and-orange YouTube". That is
the wrong direction. Event Horizon is a GAMING PLATFORM FOR CHILDREN —
a safe, welcoming, slightly majestic place. Not dark-adult, not
corporate-SaaS, not "generic streaming site".

Also: alignment is off across pages. Cards, headers, grids and the new
hero don't sit on a consistent grid.

SCOPE OF THIS TASK
Two things only:
  A) Fix the color system so indigo returns as a primary identity color,
     and gold/orange becomes an ACCENT, not the dominant hue.
  B) Fix alignment / spacing / typography rhythm across migrated pages.

DO NOT start new features. DO NOT touch game mechanics. DO NOT add new
pages.

------------------------------------------------------------
PART A — COLOR SYSTEM
------------------------------------------------------------

Current problem (your own audit noted it):
  - The void/background bug was fixed, but the palette now leans too
    heavily into gold/orange (e.g. --color-horizon-gold used everywhere).
  - Indigo — which was part of the original identity — is missing or
    buried.

Target feeling:
  - Deep, calm, slightly majestic space. Safe. Welcoming to a child.
  - Indigo / deep blue-violet = PRIMARY identity.
  - Gold = ACCENT only (highlight, flagship, small moments of "wonder").
  - Cyan / soft violet = secondary support.
  - No harsh contrast, no "warning orange" as a dominant.

Required actions:

1. Open src/styles/theme.css. Show me the CURRENT full token list
   (colors, spacing, radii, typography) — as-is, no changes yet.

2. Propose a REVISED token palette where:
   - Primary background: deep indigo / near-black violet (not pure black,
     not neutral dark grey).
   - Surface layers: slightly lighter indigo shades (2–3 levels).
   - Primary accent: gold — used sparingly (flagship, key CTA, small
     highlights).
   - Secondary accent: soft cyan or violet — for links, badges, hover.
   - Text: warm off-white on indigo, not pure #fff.
   - Semantic colors (success / warning / danger) kept subtle.

3. Show me the proposed palette as a table:
   token name | current value | proposed value | where it's used.

4. STOP and wait for my approval before writing any code.

5. After approval: apply tokens, sweep the codebase for hardcoded colors
   (e.g. #ffd700, rgba gold, etc.) and replace them with tokens.

------------------------------------------------------------
PART B — ALIGNMENT & RHYTHM
------------------------------------------------------------

Goal: every page sits on ONE consistent grid. No more per-page
ad-hoc spacing.

Required actions:

1. Audit the shared primitives and layout:
   - PageHeader, Card, Button, Badge, Spinner, StatCard
   - Any layout wrapper / container / grid utilities
   Show me for each: current padding, margin, gap, max-width, alignment.

2. Identify inconsistencies, e.g.:
   - PageHeader padding differs between pages.
   - Cards use different internal padding.
   - Grids use different gap values.
   - Vertical rhythm (spacing between sections) is not on a scale.

3. Propose:
   - A single spacing scale (e.g. 4/8/12/16/24/32/48/64).
   - A single container max-width + horizontal padding.
   - A single vertical rhythm (section spacing, card padding, gap).
   - Typography scale (h1/h2/h3/body/caption — sizes + line-heights).

4. Show me the proposal as a short spec. Do NOT implement yet.

5. After approval: apply the scale to primitives, then sweep pages.

------------------------------------------------------------
CONSTRAINTS — IMPORTANT
------------------------------------------------------------

- This is a KIDS' SAFE SPACE. Tone check on every change:
  - Calm, warm, welcoming. No aggressive contrast, no "warning red/orange"
    dominance, no harsh neon.
  - "Majestic" is OK, "ominous" is NOT.
- Do not remove the accretion disk hero — refine it, don't redesign it.
- Do not touch game mechanic CSS (HexGrid, Tray, and the 4 games' internal
  canvases).
- Do not add new pages, new features, or new games.
- Everything must stay keyboard-accessible and pass prefers-reduced-motion.

------------------------------------------------------------
OUTPUT FORMAT
------------------------------------------------------------

Step 1: Report only. Show current tokens, current primitive spacing,
        and inconsistencies found. No code.
Step 2: Propose revised palette + spacing/typography scale. No code.
Step 3: Wait for my "go".
Step 4: Implement, then verify with tsc --noEmit + npm run build.
        Report exactly which files changed.

DO NOT do Steps 3–4 until I say "go".

