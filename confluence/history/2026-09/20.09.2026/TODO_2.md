GO — with these adjustments before you implement:

COLOR PALETTE
- Accept the full proposed palette as-is, with these three changes:

  1. WARNING vs GOLD — they are too close now (#d4a017 vs #e8b84a).
     Change --color-warning to something clearly distinct, e.g. #c9911f,
     AND ensure warning states always carry a non-color signal (icon or
     label), not just hue. Kids-safe + a11y.

  2. CONTRAST CHECK — before applying, verify --color-indigo (#5b6cff)
     against --color-nebula (#151b33) and --color-nebula-elevated
     (#1e2748) for WCAG AA (text) and AAA where possible. If it fails,
     use --color-indigo-soft (#8b95ff) as the text-safe variant and keep
     #5b6cff for fills/borders only. Tell me which you used and why.

  3. FOCUS RINGS — decide explicitly: either move focus rings to
     indigo-soft for consistency with the new identity, OR keep them
     photon-cyan so focus never collides with hover. State your choice
     and reasoning. Do not leave it implicit.

EVERYTHING ELSE
- Spacing scale, container widths, PageShell helper, typography scale:
  approved as proposed.
- Accretion disk: keep gold, only refine opacity/ember tone to match
  the softer gold. Do not restructure it.
- Game mechanic CSS (HexGrid, Tray, canvases, memory/hanoi board):
  untouched.

EXECUTION ORDER
1. Apply tokens in theme.css.
2. Sweep hardcoded colors across the codebase → replace with tokens.
3. Apply spacing/typography scale to primitives (PageHeader, Card,
   Button, Badge, StatCard, Modal, Spinner).
4. Add PageShell helper and migrate pages onto it.
5. Verify: tsc --noEmit, npm run build, ReadLints.
6. Report: files changed, contrast results, focus-ring decision.

Do not add new pages, features, or games.