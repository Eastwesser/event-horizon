You're right — no pilot needed for the FORMAT. All cards are
ready in the seed tree, and *_cards_info.md already defines
the format. Go straight to seed.

But before the full run, we do a 5-card SMOKE to validate the
SCRIPT, not the format:

  1. Дирр               — plain
  2. Тхакрай (фойл)     — foil flag
  3. Мормолика (х2)     — two artists, two files
  4. Склеп Керсам       — artifact (no hp/attack_dice/move)
  5. Беллигемин         — companion

Seed script requirements:

  - Parse *_cards_info.md directly. Do NOT introduce a new
    meta.md format — the info files are the source of truth.
  - Source tree:
      services/inventory/internal/seed/berserk_cards/
        {year_set}/{year}/{Nth_year}/[{element}_Nth]/
  - For each card block in the info file:
      * parse: name, card_no, file_name, set, rarity, element,
        cost, cost_tier, type_main, type_sub, class, hp, move,
        attack_dice, attack_type, icons, unique, flying,
        companion, symbiont, parasite, artist, card_text,
        flavor_text
      * locate the JPG by file_name in the same folder
      * POST to /api/uploads → get URL
      * POST to /api/inventory/items with:
          - type: "карточка"
          - name
          - price: computed from market_rub × 2000 (or stock
            price — decide and confirm)
          - stock: 1 (default) or from info file
          - images: [uploaded_url]
          - attributes: { ...all parsed fields... }
  - Foil / noir → separate items. Foil is a flag in
    attributes; two files (regular + foil) → two items.
  - Duplicate artists (Мормолика) → two items, different
    artist_id, different image.
  - Artifacts / lands → type_main = artifact | land; leave
    hp / move / attack_dice / attack_type null.
  - Idempotent on (set_number, card_no, foil, noir, artist_id).
    Re-run must not duplicate.

  - --dry-run mode first: print what WOULD be created, no writes.
  - --limit N to run on 5 cards first (smoke).
  - Then full run.

  - Artist resolution: keep artist_id normalized
    (julia_alekseeva). If an artist_id doesn't exist in the
    artists table, either create it or report it as unresolved.

  - Do NOT modify images on disk. Only read + upload.

  - Do NOT delete existing items. Only upsert by the idempotency
    key.

Process:
  1. Show the plan for the seed script (parser + upload + create).
  2. Wait for my OK.
  3. Implement.
  4. Run with --dry-run --limit 5. Show output.
  5. I verify.
  6. Then --limit 5 real. Verify 5 items in shop.
  7. Then full run.

Do NOT skip steps 1–2. No code until I approve.